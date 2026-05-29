<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  useLimitsStore,
  emptyTradeLimit,
  FRAME_INTERVALS,
  HISTORY_INTERVALS,
  PROFIT_UNITS,
  type TradeLimit,
} from '~/stores/limits'
import { useSettingsStore, TRADE_STACK_SORTINGS } from '~/stores/settings'

const limits = useLimitsStore()
const settings = useSettingsStore()

const tab = ref<'pairs' | 'bot'>('pairs')

// ─── pairs CRUD ────────────────────────────────────────────────────────
const dialog = ref(false)
const editing = ref<TradeLimit>(emptyTradeLimit())
const isNew = ref(true)

function openCreate() {
  editing.value = emptyTradeLimit()
  isNew.value = true
  dialog.value = true
}

function openEdit(row: TradeLimit) {
  editing.value = JSON.parse(JSON.stringify(row)) as TradeLimit
  isNew.value = false
  dialog.value = true
}

function addProfitOption() {
  editing.value.profitOptions.push({
    index: editing.value.profitOptions.length,
    isTriggerOption: false,
    optionValue: 1,
    optionUnit: 'h',
    optionPercent: 0.5,
  })
}

function removeProfitOption(idx: number) {
  editing.value.profitOptions.splice(idx, 1)
}

function addExtraChargeOption() {
  editing.value.extraChargeOptions.push({
    index: editing.value.extraChargeOptions.length,
    percent: -3,
    amountUsdt: 50,
  })
}

function removeExtraChargeOption(idx: number) {
  editing.value.extraChargeOptions.splice(idx, 1)
}

async function save() {
  if (isNew.value) {
    await limits.create(editing.value)
  } else {
    await limits.update(editing.value)
  }
  dialog.value = false
}

// ─── bot settings ─────────────────────────────────────────────────────
const holdScore = ref<number>(75)
const stackSorting = ref<'percent' | 'diff'>('percent')
const savedToast = ref(false)

async function saveBot() {
  await settings.save({ holdScore: holdScore.value, tradeStackSorting: stackSorting.value })
  savedToast.value = true
}

onMounted(async () => {
  await Promise.all([limits.load(), settings.load()])
  holdScore.value = settings.settings.holdScore
  stackSorting.value = settings.settings.tradeStackSorting
})
</script>

<template>
  <v-container fluid>
    <h1 class="text-h5 mb-2">{{ $t('Configuration') }}</h1>

    <v-tabs v-model="tab" color="primary" data-test="config-tabs">
      <v-tab value="pairs" data-test="tab-pairs">{{ $t('Trade pairs') }}</v-tab>
      <v-tab value="bot" data-test="tab-bot">{{ $t('Bot settings') }}</v-tab>
    </v-tabs>

    <v-window v-model="tab" class="mt-4">
      <!-- ──────────────── pairs tab ──────────────── -->
      <v-window-item value="pairs">
        <div class="d-flex justify-end mb-2">
          <v-btn
            color="primary"
            prepend-icon="mdi-plus"
            data-test="add-pair-btn"
            @click="openCreate"
          >
            {{ $t('Add pair') }}
          </v-btn>
        </div>

        <v-card>
          <v-data-table
            :items="limits.items"
            :headers="[
              { title: $t('Symbol'),         key: 'symbol' },
              { title: $t('USDT limit'),     key: 'USDTLimit' },
              { title: $t('Frame interval'), key: 'frameInterval' },
              { title: $t('Profit options'), key: 'profitOptions', sortable: false },
              { title: $t('Enabled'),        key: 'isEnabled' },
              { title: '',                   key: 'actions', sortable: false },
            ]"
            :loading="limits.loading"
            density="compact"
            :items-per-page="50"
            data-test="limits-table"
          >
            <template #item.profitOptions="{ item }">
              <v-chip
                v-for="opt in item.profitOptions"
                :key="opt.index"
                size="x-small"
                class="mr-1"
              >
                {{ opt.optionValue }}{{ opt.optionUnit }} → {{ opt.optionPercent }}%
              </v-chip>
              <span v-if="item.profitOptions.length === 0" class="text-disabled">—</span>
            </template>

            <template #item.isEnabled="{ item }">
              <v-switch
                :model-value="item.isEnabled"
                density="compact"
                color="primary"
                hide-details
                :data-test="`toggle-${item.symbol}`"
                @update:model-value="limits.toggle(item)"
              />
            </template>

            <template #item.actions="{ item }">
              <v-btn
                size="small"
                variant="text"
                icon="mdi-pencil"
                :data-test="`edit-${item.symbol}`"
                @click="openEdit(item)"
              />
            </template>
          </v-data-table>
        </v-card>
      </v-window-item>

      <!-- ──────────────── bot settings tab ──────────────── -->
      <v-window-item value="bot">
        <v-card class="pa-4">
          <h2 class="text-h6 mb-4">{{ $t('Global bot settings') }}</h2>
          <v-row>
            <v-col cols="12" md="4">
              <v-text-field
                v-model.number="holdScore"
                :label="$t('Hold score threshold')"
                type="number"
                step="1"
                data-test="hold-score-input"
              />
            </v-col>
            <v-col cols="12" md="4">
              <v-select
                v-model="stackSorting"
                :items="TRADE_STACK_SORTINGS"
                :label="$t('Trade stack sorting')"
                data-test="stack-sorting-select"
              />
            </v-col>
          </v-row>
          <v-btn
            color="primary"
            :loading="settings.saving"
            data-test="bot-save-btn"
            @click="saveBot"
          >
            {{ $t('Save') }}
          </v-btn>
        </v-card>
      </v-window-item>
    </v-window>

    <!-- ──────────────── trade-pair dialog ──────────────── -->
    <v-dialog v-model="dialog" max-width="780" persistent>
      <v-card data-test="pair-dialog">
        <v-card-title>
          <span>{{ isNew ? $t('Add pair') : editing.symbol }}</span>
        </v-card-title>

        <v-card-text>
          <v-row>
            <v-col cols="12" md="6">
              <v-text-field
                v-model="editing.symbol"
                :label="$t('Symbol')"
                :disabled="!isNew"
                data-test="form-symbol"
              />
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field
                v-model.number="editing.USDTLimit"
                :label="$t('USDT limit')"
                type="number"
                data-test="form-usdt-limit"
              />
            </v-col>
            <v-col cols="12" md="6">
              <v-select
                v-model="editing.frameInterval"
                :items="FRAME_INTERVALS"
                :label="$t('Frame interval')"
                data-test="form-frame-interval"
              />
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field
                v-model.number="editing.framePeriod"
                :label="$t('Frame period')"
                type="number"
                data-test="form-frame-period"
              />
            </v-col>
            <v-col cols="12" md="6">
              <v-select
                v-model="editing.buyPriceHistoryCheckInterval"
                :items="HISTORY_INTERVALS"
                :label="$t('History interval')"
                data-test="form-history-interval"
              />
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field
                v-model.number="editing.buyPriceHistoryCheckPeriod"
                :label="$t('History period')"
                type="number"
                data-test="form-history-period"
              />
            </v-col>
          </v-row>

          <!-- profit options sub-form -->
          <div class="d-flex align-center justify-space-between mt-4 mb-2">
            <h3 class="text-subtitle-1">{{ $t('Profit options') }}</h3>
            <v-btn
              size="small"
              prepend-icon="mdi-plus"
              data-test="add-profit-option"
              @click="addProfitOption"
            >
              {{ $t('Add') }}
            </v-btn>
          </div>
          <v-row
            v-for="(opt, idx) in editing.profitOptions"
            :key="idx"
            class="profit-option-row"
            :data-test="`profit-option-${idx}`"
            align="center"
          >
            <v-col cols="3">
              <v-text-field
                v-model.number="opt.optionValue"
                :label="$t('Value')"
                type="number"
                density="compact"
                :data-test="`profit-value-${idx}`"
              />
            </v-col>
            <v-col cols="3">
              <v-select
                v-model="opt.optionUnit"
                :items="PROFIT_UNITS"
                :label="$t('Unit')"
                density="compact"
                :data-test="`profit-unit-${idx}`"
              />
            </v-col>
            <v-col cols="3">
              <v-text-field
                v-model.number="opt.optionPercent"
                :label="$t('Percent')"
                type="number"
                step="0.1"
                density="compact"
                :data-test="`profit-percent-${idx}`"
              />
            </v-col>
            <v-col cols="2">
              <v-switch
                v-model="opt.isTriggerOption"
                :label="$t('Trigger')"
                density="compact"
                hide-details
                :data-test="`profit-trigger-${idx}`"
              />
            </v-col>
            <v-col cols="1">
              <v-btn
                icon="mdi-delete"
                size="small"
                variant="text"
                :data-test="`remove-profit-${idx}`"
                @click="removeProfitOption(idx)"
              />
            </v-col>
          </v-row>

          <!-- extra-charge options sub-form -->
          <div class="d-flex align-center justify-space-between mt-4 mb-2">
            <h3 class="text-subtitle-1">{{ $t('Extra charges') }}</h3>
            <v-btn
              size="small"
              prepend-icon="mdi-plus"
              data-test="add-extra-charge"
              @click="addExtraChargeOption"
            >
              {{ $t('Add') }}
            </v-btn>
          </div>
          <v-row
            v-for="(opt, idx) in editing.extraChargeOptions"
            :key="idx"
            class="extra-charge-row"
            :data-test="`extra-charge-${idx}`"
            align="center"
          >
            <v-col cols="4">
              <v-text-field
                v-model.number="opt.percent"
                :label="$t('Fall percent')"
                type="number"
                step="0.1"
                density="compact"
                :data-test="`extra-percent-${idx}`"
              />
            </v-col>
            <v-col cols="4">
              <v-text-field
                v-model.number="opt.amountUsdt"
                :label="$t('Amount USDT')"
                type="number"
                density="compact"
                :data-test="`extra-amount-${idx}`"
              />
            </v-col>
            <v-col cols="2">
              <v-btn
                icon="mdi-delete"
                size="small"
                variant="text"
                :data-test="`remove-extra-${idx}`"
                @click="removeExtraChargeOption(idx)"
              />
            </v-col>
          </v-row>

          <v-switch
            v-model="editing.isEnabled"
            :label="$t('Enabled')"
            color="primary"
            class="mt-4"
            data-test="form-enabled"
          />
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" data-test="form-cancel" @click="dialog = false">
            {{ $t('Cancel') }}
          </v-btn>
          <v-btn color="primary" variant="flat" data-test="form-save" @click="save">
            {{ $t('Save') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-snackbar v-model="savedToast" :timeout="2000" color="success">
      {{ $t('Settings saved') }}
    </v-snackbar>
  </v-container>
</template>
