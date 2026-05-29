<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useLimitsStore, type TradeLimit } from '~/stores/limits'

const { t } = useI18n()
const limits = useLimitsStore()

const dialog = ref(false)
const editing = ref<TradeLimit>({ symbol: '', isEnabled: true, usdtLimit: 100, minProfitPercent: 0.5 })
const isNew = ref(true)

function openCreate() {
  editing.value = { symbol: '', isEnabled: true, usdtLimit: 100, minProfitPercent: 0.5 }
  isNew.value = true
  dialog.value = true
}

function openEdit(row: TradeLimit) {
  editing.value = { ...row }
  isNew.value = false
  dialog.value = true
}

async function save() {
  if (isNew.value) {
    await limits.create(editing.value)
  } else {
    await limits.update(editing.value)
  }
  dialog.value = false
}

onMounted(() => limits.load())
</script>

<template>
  <v-container fluid>
    <div class="d-flex align-center justify-space-between mb-4">
      <h1 class="text-h5">{{ t('config.title') }}</h1>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">
        {{ t('config.add_limit') }}
      </v-btn>
    </div>

    <v-card>
      <v-data-table
        :items="limits.items"
        :headers="[
          { title: t('config.symbol'),             key: 'symbol' },
          { title: t('config.usdt_limit'),         key: 'usdtLimit' },
          { title: t('config.min_profit_percent'), key: 'minProfitPercent' },
          { title: t('config.is_enabled'),         key: 'isEnabled' },
          { title: '',                             key: 'actions', sortable: false },
        ]"
        :loading="limits.loading"
        density="compact"
        :items-per-page="50"
      >
        <template #item.isEnabled="{ item }">
          <v-switch
            :model-value="item.isEnabled"
            density="compact"
            color="primary"
            hide-details
            @update:model-value="limits.toggle(item)"
          />
        </template>
        <template #item.actions="{ item }">
          <v-btn size="small" variant="text" icon="mdi-pencil" @click="openEdit(item)" />
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="dialog" max-width="500">
      <v-card>
        <v-card-title>{{ isNew ? t('config.add_limit') : editing.symbol }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="editing.symbol" :label="t('config.symbol')" :disabled="!isNew" />
          <v-text-field v-model.number="editing.usdtLimit" :label="t('config.usdt_limit')" type="number" />
          <v-text-field v-model.number="editing.minProfitPercent" :label="t('config.min_profit_percent')" type="number" step="0.1" />
          <v-switch v-model="editing.isEnabled" :label="t('config.is_enabled')" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">{{ t('config.cancel') }}</v-btn>
          <v-btn color="primary" variant="flat" @click="save">{{ t('config.save') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>
