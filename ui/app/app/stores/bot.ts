// useBotStore — health snapshot + balance + positions + recent trades.
// All getters fan out through $api so the same Pinia state powers the
// dashboard, the orders page and any live indicators in the chrome.

import { defineStore } from 'pinia'

export interface HealthSnapshot {
  bot: { id: number; botUuid: string; exchange: string; isMasterBot: boolean }
  binanceStatus: string
  dbStatus: string
  redisStatus: string
  mlStatus: string
  updates: Record<string, [string, string, string]>
  orderBook: Record<string, string>
  dateTimeNow: string
}

export interface Position {
  symbol: string
  qty: number
  avgEntry: number
  unrealizedPnL: number
}

export interface RecentTrade {
  symbol: string
  side: string
  price: number
  qty: number
  createdAt: string
}

export interface AssetBalance {
  asset: string
  free: number
  locked: number
}

export const useBotStore = defineStore('bot', () => {
  const { $api } = useNuxtApp()

  const health = ref<HealthSnapshot | null>(null)
  const balances = ref<AssetBalance[]>([])
  const positions = ref<Position[]>([])
  const recentTrades = ref<RecentTrade[]>([])
  const loading = ref(false)

  const usdtBalance = computed(() => balances.value.find(b => b.asset === 'USDT')?.free ?? null)

  async function refresh() {
    loading.value = true
    try {
      const [h, account, pos, trades] = await Promise.all([
        $api<HealthSnapshot>('/bot/health/check'),
        $api<{ balances: AssetBalance[] }>('/bot/account'),
        $api<Position[]>('/bot/order/position/list'),
        $api<RecentTrade[]>('/bot/order/trade/list'),
      ])
      health.value = h
      balances.value = account.balances ?? []
      positions.value = pos ?? []
      recentTrades.value = trades ?? []
    } finally {
      loading.value = false
    }
  }

  return { health, balances, positions, recentTrades, loading, usdtBalance, refresh }
})
