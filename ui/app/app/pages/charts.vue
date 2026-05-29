<script setup lang="ts">
const { t } = useI18n()

const panels = [
  { key: 'tick_publish_failure', query: 'rate(tick_publish_failure_total[1m])' },
  { key: 'tick_drop',            query: 'rate(tick_drop_total[1m])' },
  { key: 'pubsub_reconnect',     query: 'rate(pubsub_reconnect_total[5m])' },
  { key: 'tick_consume_lag',     query: 'histogram_quantile(0.95, sum(rate(tick_consume_lag_seconds_bucket[1m])) by (le))' },
  { key: 'handle_duration',      query: 'histogram_quantile(0.95, sum(rate(handle_duration_seconds_bucket[1m])) by (le))' },
  { key: 'order_place_duration', query: 'histogram_quantile(0.95, sum(rate(order_place_duration_seconds_bucket[5m])) by (le, outcome))' },
  { key: 'pnl_realized',         query: 'sum(rate(pnl_realized_total[5m])) by (symbol)' },
  { key: 'position_open_seconds', query: 'histogram_quantile(0.95, sum(rate(position_open_seconds_bucket[5m])) by (le, symbol))' },
] as const
</script>

<template>
  <v-container fluid>
    <h1 class="text-h5 mb-4">{{ t('charts.title') }}</h1>

    <v-row>
      <v-col v-for="panel in panels" :key="panel.key" cols="12" md="6">
        <MetricChart
          :title="t(`charts.${panel.key}`)"
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
