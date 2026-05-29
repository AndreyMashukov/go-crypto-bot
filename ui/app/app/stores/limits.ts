import { defineStore } from 'pinia'

export interface TradeLimit {
  symbol: string
  isEnabled: boolean
  usdtLimit: number
  minProfitPercent: number
}

export const useLimitsStore = defineStore('limits', () => {
  const { $api } = useNuxtApp()

  const items = ref<TradeLimit[]>([])
  const loading = ref(false)

  async function load() {
    loading.value = true
    try {
      items.value = await $api<TradeLimit[]>('/bot/trade/limit/list')
    } finally {
      loading.value = false
    }
  }

  async function create(limit: TradeLimit) {
    await $api('/bot/trade/limit/create', { method: 'POST', body: limit })
    await load()
  }

  async function update(limit: TradeLimit) {
    await $api('/bot/trade/limit/update', { method: 'POST', body: limit })
    await load()
  }

  async function toggle(limit: TradeLimit) {
    await $api(`/bot/trade/limit/switch/${limit.symbol}`, {
      method: 'PUT',
      body: { isEnabled: !limit.isEnabled },
    })
    await load()
  }

  return { items, loading, load, create, update, toggle }
})
