<script setup lang="ts">
// MetricChart is the only chart primitive after Phase G. Every chart on
// the dashboard is a thin call into this component pointing at a
// Prometheus PromQL expression. The FE never reads raw klines / market
// candles from the bot's own DB — the architecture-doc §8.3 rule.
//
// Renders the {series: [{label, points: [[unixTs, value], ...]}]} JSON
// envelope the ui/server PrometheusController returns. Multi-series
// queries (e.g. by symbol or by outcome) get one line per series with
// the series label as the legend.

import {ref, onMounted, onBeforeUnmount, computed, watch} from 'vue'
import {use} from 'echarts/core'
import {CanvasRenderer} from 'echarts/renderers'
import {LineChart} from 'echarts/charts'
import {TitleComponent, TooltipComponent, LegendComponent, GridComponent} from 'echarts/components'
import VChart from 'vue-echarts'

use([CanvasRenderer, LineChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent])

interface SeriesPoint {
  0: number
  1: string
}

interface Series {
  label: string
  points: SeriesPoint[]
}

interface PrometheusRangeResponse {
  series: Series[]
}

const props = withDefaults(defineProps<{
  query: string
  start?: string | number
  end?: string | number
  step?: string
  title?: string
  height?: string
  refreshSeconds?: number
}>(), {
  start: '-30 minutes',
  end: 'now',
  step: '15s',
  title: '',
  height: '320px',
  refreshSeconds: 15,
})

const loading = ref(true)
const error = ref<string | null>(null)
const series = ref<Series[]>([])

let timer: ReturnType<typeof setInterval> | null = null

async function fetchSeries() {
  error.value = null
  try {
    const params = new URLSearchParams({
      query: props.query,
      start: String(props.start),
      end: String(props.end),
      step: props.step,
    })
    const response = await $fetch<PrometheusRangeResponse>(`/api/prometheus/range?${params.toString()}`)
    if (!response || !Array.isArray(response.series)) {
      throw new Error('malformed prometheus range response: missing series array')
    }
    series.value = response.series
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'unknown error'
    series.value = []
  } finally {
    loading.value = false
  }
}

const hasData = computed(() => series.value.some(s => s.points.length > 0))

const chartOption = computed(() => {
  const allTimes = new Set<number>()
  for (const s of series.value) {
    for (const p of s.points) {
      allTimes.add(p[0])
    }
  }
  const sortedTimes = Array.from(allTimes).sort((a, b) => a - b)
  const labels = sortedTimes.map(t => new Date(t * 1000).toISOString().slice(11, 19))

  return {
    tooltip: {trigger: 'axis'},
    legend: {data: series.value.map(s => s.label), bottom: 0},
    grid: {left: '3%', right: '4%', bottom: '15%', containLabel: true},
    xAxis: {type: 'category', data: labels, axisLabel: {rotate: 30, fontSize: 10}},
    yAxis: {type: 'value'},
    series: series.value.map(s => ({
      name: s.label,
      type: 'line',
      smooth: true,
      showSymbol: false,
      data: sortedTimes.map(t => {
        const found = s.points.find(p => p[0] === t)
        return found ? Number(found[1]) : null
      }),
    })),
  }
})

watch(() => [props.query, props.start, props.end, props.step], () => {
  loading.value = true
  fetchSeries()
})

onMounted(() => {
  fetchSeries()
  if (props.refreshSeconds > 0) {
    timer = setInterval(fetchSeries, props.refreshSeconds * 1000)
  }
})

onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})
</script>

<template>
  <div class="metric-chart">
    <h4 v-if="title" class="metric-chart__title">{{ title }}</h4>
    <div v-if="loading" class="metric-chart__state">Loading…</div>
    <div v-else-if="error" class="metric-chart__state metric-chart__state--error">
      Prometheus query failed: {{ error }}
    </div>
    <div v-else-if="!hasData" class="metric-chart__state">No data for the selected range.</div>
    <VChart v-else :option="chartOption" :style="{height}" autoresize/>
  </div>
</template>

<style scoped>
.metric-chart {
  padding: 8px;
  border-radius: 6px;
  background-color: #fff;
}

.metric-chart__title {
  text-align: center;
  background-color: #efefef;
  padding: 4px;
  margin: 0 0 8px;
  font-size: 14px;
  font-weight: 600;
}

.metric-chart__state {
  text-align: center;
  padding: 24px;
  color: #777;
}

.metric-chart__state--error {
  color: #c00;
}
</style>
