// useSettingsStore — bot-level config (holdScore + tradeStackSorting).
//
// Mirrors model.Bot (server/src/model/bot.go). PUT /bot/update is the
// bot's single mutate endpoint; the GET is folded into /health/check
// which returns the bot's current snapshot.

import { defineStore } from 'pinia'

export type TradeStackSorting = 'percent' | 'diff'

export interface BotSettings {
  botUuid: string
  exchange: string
  isMasterBot: boolean
  tradeStackSorting: TradeStackSorting
  holdScore: number
}

export const TRADE_STACK_SORTINGS: TradeStackSorting[] = ['percent', 'diff']

export const useSettingsStore = defineStore('settings', () => {
  const { $api } = useNuxtApp()

  const settings = ref<BotSettings>({
    botUuid: '',
    exchange: 'binance',
    isMasterBot: false,
    tradeStackSorting: 'percent',
    holdScore: 75,
  })
  const loading = ref(false)
  const saving = ref(false)

  async function load() {
    loading.value = true
    try {
      const snap = await $api<{ bot: BotSettings }>('/bot/health/check')
      settings.value = { ...settings.value, ...snap.bot }
    } finally {
      loading.value = false
    }
  }

  async function save(patch: Partial<BotSettings>) {
    saving.value = true
    try {
      await $api('/bot/bot/update', { method: 'PUT', body: { ...settings.value, ...patch } })
      settings.value = { ...settings.value, ...patch }
    } finally {
      saving.value = false
    }
  }

  return { settings, loading, saving, load, save }
})
