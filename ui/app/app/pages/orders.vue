<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { useOrdersStore } from '~/stores/orders'

const { t } = useI18n()
const orders = useOrdersStore()

let timer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  orders.load()
  timer = setInterval(() => orders.load(), 10000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

watch(() => orders.tab, () => orders.load())
</script>

<template>
  <v-container fluid>
    <h1 class="text-h5 mb-4">{{ t('orders.title') }}</h1>

    <v-tabs :model-value="orders.tab" color="primary" @update:model-value="orders.setTab($event as any)">
      <v-tab value="active">{{ t('orders.active') }}</v-tab>
      <v-tab value="pending">{{ t('orders.pending') }}</v-tab>
      <v-tab value="positions">{{ t('orders.positions') }}</v-tab>
    </v-tabs>

    <v-card class="mt-2">
      <v-data-table
        v-if="orders.items.length > 0"
        :items="orders.items"
        :headers="[
          { title: t('orders.symbol'),     key: 'symbol' },
          { title: t('orders.side'),       key: 'side' },
          { title: t('orders.qty'),        key: 'qty' },
          { title: t('orders.price'),      key: 'price' },
          { title: t('orders.status'),     key: 'status' },
          { title: t('orders.created_at'), key: 'createdAt' },
        ]"
        :loading="orders.loading"
        density="compact"
        :items-per-page="50"
      />
      <v-card-text v-else>{{ t('orders.no_orders') }}</v-card-text>
    </v-card>
  </v-container>
</template>
