<script setup lang="ts">
const panels = [
  { title: 'Tick publish failures',     query: 'rate(tick_publish_failure_total[1m])' },
  { title: 'Dropped ticks',             query: 'rate(tick_drop_total[1m])' },
  { title: 'Pub/sub reconnects',        query: 'rate(pubsub_reconnect_total[5m])' },
  { title: 'Tick consume lag, s',       query: 'histogram_quantile(0.95, sum(rate(tick_consume_lag_seconds_bucket[1m])) by (le))' },
  { title: 'Handler duration, s',       query: 'histogram_quantile(0.95, sum(rate(handle_duration_seconds_bucket[1m])) by (le))' },
  { title: 'Order place duration, s',   query: 'histogram_quantile(0.95, sum(rate(order_place_duration_seconds_bucket[5m])) by (le, outcome))' },
  { title: 'Realized P&L',              query: 'sum(rate(pnl_realized_total[5m])) by (symbol)' },
  { title: 'Position open, s',          query: 'histogram_quantile(0.95, sum(rate(position_open_seconds_bucket[5m])) by (le, symbol))' },
] as const
</script>

<template>
  <v-container fluid>
    <h1 class="text-h5 mb-4">{{ $t('Prometheus metrics') }}</h1>

    <v-row>
      <v-col v-for="panel in panels" :key="panel.title" cols="12" md="6">
        <MetricChart
          :title="$t(panel.title)"
          :query="panel.query"
          start="-30 minutes"
          end="now"
          step="15s"
          :refresh-seconds="15"
        />
      </v-col>
    </v-row>
  </v-container>
</template>
