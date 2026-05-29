import { defineStore } from 'pinia'

export interface OrderRow {
  symbol: string
  side: string
  price: number
  qty: number
  status: string
  createdAt: string
}

type Tab = 'active' | 'pending' | 'positions'

const PATH: Record<Tab, string> = {
  active: '/bot/order/list',
  pending: '/bot/order/pending/list',
  positions: '/bot/order/position/list',
}

export const useOrdersStore = defineStore('orders', () => {
  const { $api } = useNuxtApp()

  const items = ref<OrderRow[]>([])
  const tab = ref<Tab>('active')
  const loading = ref(false)

  async function load() {
    loading.value = true
    try {
      items.value = await $api<OrderRow[]>(PATH[tab.value])
    } finally {
      loading.value = false
    }
  }

  async function setTab(next: Tab) {
    tab.value = next
    await load()
  }

  return { items, tab, loading, load, setTab }
})
