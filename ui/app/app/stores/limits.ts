// useLimitsStore — trade-pair configuration store.
//
// Mirrors model.TradeLimit on the bot side (server/src/model/trade_limit.go)
// so the form bound in pages/config.vue can build a payload the bot's
// /trade/limit/create + /trade/limit/update endpoints can deserialise
// straight into TradeLimit without any field-name translation.

import { defineStore } from 'pinia'

export interface ProfitOption {
  index: number
  isTriggerOption: boolean
  optionValue: number
  optionUnit: 'i' | 'h' | 'd'
  optionPercent: number
}

export interface ExtraChargeOption {
  index: number
  percent: number
  amountUsdt: number
}

export interface TradeLimit {
  symbol: string
  USDTLimit: number
  minPrice: number
  minQuantity: number
  minNotional: number
  isEnabled: boolean
  minPriceMinutesPeriod: number
  frameInterval: string
  framePeriod: number
  buyPriceHistoryCheckInterval: string
  buyPriceHistoryCheckPeriod: number
  profitOptions: ProfitOption[]
  extraChargeOptions: ExtraChargeOption[]
}

export const FRAME_INTERVALS = ['1m', '5m', '15m', '30m', '1h', '2h', '4h', '6h', '12h', '1d'] as const
export const HISTORY_INTERVALS = ['1h', '4h', '1d', '1w'] as const
export const PROFIT_UNITS = ['i', 'h', 'd'] as const

export function emptyTradeLimit(): TradeLimit {
  return {
    symbol: '',
    USDTLimit: 100,
    minPrice: 0.0001,
    minQuantity: 0.0001,
    minNotional: 10,
    isEnabled: true,
    minPriceMinutesPeriod: 200,
    frameInterval: '2h',
    framePeriod: 20,
    buyPriceHistoryCheckInterval: '1d',
    buyPriceHistoryCheckPeriod: 14,
    profitOptions: [],
    extraChargeOptions: [],
  }
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
