<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useBotStore } from '~/stores/bot'

const { t } = useI18n()
const bot = useBotStore()

let timer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  bot.refresh()
  timer = setInterval(() => bot.refresh(), 15000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <v-container fluid>
    <h1 class="text-h5 mb-4">{{ t('dashboard.title') }}</h1>

    <v-row>
      <v-col cols="12" md="4">
        <v-card>
          <v-card-title>{{ t('dashboard.balance') }}</v-card-title>
          <v-card-text class="text-h4">
            <span v-if="bot.usdtBalance === null">—</span>
            <span v-else>{{ bot.usdtBalance.toFixed(2) }} USDT</span>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="8">
        <v-card>
          <v-card-title>{{ t('dashboard.open_positions') }}</v-card-title>
          <v-data-table
            v-if="bot.positions.length > 0"
            :items="bot.positions"
            :headers="[
              { title: t('orders.symbol'), key: 'symbol' },
              { title: t('orders.qty'),    key: 'qty' },
              { title: t('orders.price'),  key: 'avgEntry' },
              { title: 'PnL',              key: 'unrealizedPnL' },
            ]"
            density="compact"
            hide-default-footer
          />
          <v-card-text v-else>{{ t('dashboard.no_data') }}</v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-4">
      <v-col cols="12">
        <v-card>
          <v-card-title>{{ t('dashboard.recent_trades') }}</v-card-title>
          <v-data-table
            v-if="bot.recentTrades.length > 0"
            :items="bot.recentTrades"
            :headers="[
              { title: t('orders.symbol'),     key: 'symbol' },
              { title: t('orders.side'),       key: 'side' },
              { title: t('orders.price'),      key: 'price' },
              { title: t('orders.qty'),        key: 'qty' },
              { title: t('orders.created_at'), key: 'createdAt' },
            ]"
            density="compact"
            :items-per-page="20"
          />
          <v-card-text v-else>{{ t('dashboard.no_data') }}</v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-overlay :model-value="bot.loading && bot.health === null" persistent>
      <v-progress-circular indeterminate color="primary" />
    </v-overlay>
  </v-container>
</template>
