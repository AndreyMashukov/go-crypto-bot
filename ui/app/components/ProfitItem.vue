<template>
  <v-badge
      :content="trades"
      :color="badgeColor"
      class="profit-item"
      offset-y="8"
      offset-x="10"
      :style="{'background-color': backgroundColor, 'color': fontColor}"
  >
    <div>
      <h4 class="profit-title">{{ title }}</h4>
      <div class="profit-params-block">
        <div class="profit-value"><b>{{$t('profit_item.profit')}}</b> {{ Number(profit) > 0 ? '+' : (Number(profit) < 0 ? '-' : '') }}{{ Number(profit).toFixed(2) }} <small>USDT</small></div>
        <div class="volume-value"><b>{{$t('profit_item.volume')}}</b> {{ Number(tradeVolume).toFixed(2) }} <small>USDT</small></div>
        <div class="volume-value"><b>{{$t('profit_item.average')}}</b> {{ Number(avgPercent).toFixed(2) > 0 ? '+' : '' }}{{ Number(avgPercent).toFixed(2) }}<small>%</small></div>
      </div>
    </div>
  </v-badge>
</template>
<script lang="ts">
export default {
  name: "ProfitItem",
  props: {
    trades: {
      type: Number,
      default: () => 0,
    },
    title: {
      type: String,
      default: () => '',
    },
    profit: {
      type: Number,
      default: () => 0.00,
    },
    avgPercent: {
      type: Number,
      default: () => 0.00,
    },
    tradeVolume: {
      type: Number,
      default: () => 0.00,
    },
    badgeColor: {
      type: String,
      default: () => 'primary',
    },
    backgroundColor: {
      type: String,
      default: () => '#1DE9B6FF',
    },
    fontColor: {
      type: String,
      default: () => '#000000',
    },
  },
}
</script>
<style scoped>

.profit-item {
  display: inline-block;
  margin-right: 15px;
  padding: 5px 10px;
  border-radius: 5px;
  min-width: 175px;
}
.profit-params-block {
  font-size: 14px;
  line-height: 16px;
}
.profit-title {
  padding-bottom: 4px;
}
</style>