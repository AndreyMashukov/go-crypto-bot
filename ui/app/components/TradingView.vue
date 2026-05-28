<template>
  <Chart
:options="{
    width: '100%',
    height: '600',
    symbol: `${exchange}:${symbol}`,
    interval: interval,
    timezone: 'Etc/UTC',
    theme: 'dark',
    style: '1',
    backgroundColor: '#0C0C0D',
    locale: 'en',
    enable_publishing: false,
    allow_symbol_change: false,
    calendar: false,
  }"/>
</template>
<script lang="ts">
import {Chart} from '#components';
export default {
  name      : "TradingView",
  components: {
    Chart,
  },
  props: {
    exchange: {
      type: String,
      default: () => 'BINANCE',
    },
    symbol: {
      type: String,
      default: () => 'BTCUSDT',
    },
    interval: {
      type: String,
      default: () => 'D',
    }
  }
}
</script>
<style scoped>

</style>