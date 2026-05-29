<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { useOrdersStore } from '~/stores/orders'

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
    <h1 class="text-h5 mb-4">{{ $t('Orders') }}</h1>

    <v-tabs :model-value="orders.tab" color="primary" @update:model-value="orders.setTab($event as any)">
      <v-tab value="active">{{ $t('Active') }}</v-tab>
      <v-tab value="pending">{{ $t('Pending') }}</v-tab>
      <v-tab value="positions">{{ $t('Positions') }}</v-tab>
    </v-tabs>

    <v-card class="mt-2">
      <v-data-table
        v-if="orders.items.length > 0"
        :items="orders.items"
        :headers="[
          { title: $t('Symbol'),  key: 'symbol' },
          { title: $t('Side'),    key: 'side' },
          { title: $t('Quantity'), key: 'qty' },
          { title: $t('Price'),   key: 'price' },
          { title: $t('Status'),  key: 'status' },
          { title: $t('Created'), key: 'createdAt' },
        ]"
        :loading="orders.loading"
        density="compact"
        :items-per-page="50"
      />
      <v-card-text v-else>{{ $t('No orders') }}</v-card-text>
    </v-card>
  </v-container>
</template>
