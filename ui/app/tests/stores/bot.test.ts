import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useBotStore } from '~/stores/bot'

// Stub Nuxt's useNuxtApp so the store's $api injection resolves to a
// vi.fn() under test. Same shape Nuxt exposes at runtime.
vi.stubGlobal('useNuxtApp', () => ({
  $api: vi.fn(async (path: string) => {
    if (path === '/bot/health/check') {
      return {
        bot: { id: 1, botUuid: 'u', exchange: 'binance', isMasterBot: false },
        binanceStatus: 'ok', dbStatus: 'ok', redisStatus: 'ok', mlStatus: 'ready',
        updates: {}, orderBook: {}, dateTimeNow: '',
      }
    }
    if (path === '/bot/account') {
      return { balances: [{ asset: 'USDT', free: 10000, locked: 0 }] }
    }
    if (path === '/bot/order/position/list') return []
    if (path === '/bot/order/trade/list') return []
    throw new Error(`unexpected path ${path}`)
  }),
}))

describe('useBotStore', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it('refresh fills balance, positions, trades, health', async () => {
    const store = useBotStore()
    expect(store.usdtBalance).toBeNull()

    await store.refresh()

    expect(store.usdtBalance).toBe(10000)
    expect(store.positions).toEqual([])
    expect(store.recentTrades).toEqual([])
    expect(store.health?.binanceStatus).toBe('ok')
  })
})
