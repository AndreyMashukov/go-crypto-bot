<template>
<div>
  <v-tooltip location="top" v-if="!!stackItem.signal">
    <template v-slot:activator="{ props }">
      <v-icon icon="mdi-gesture-tap" v-bind="props" :color="'primary'" class="cursor-pointer"/>
    </template>
    <div>
      {{ stackItem.signal.percent }}% <b>Signal is received for {{ stackItem.symbol }}</b><br/>
      <small>AI has just generated new trading signal</small>
      <div><b>Buy price:</b> {{ stackItem.signal.buyPrice.toFixed((stackItem.price.toString().split('.')[1] || []).length) }} <small>USDT</small></div>
      <small v-for="(profitOption, index) in stackItem.signal.profitOptions" :key="index" class="d-block">
        <b>{{ profitOption.optionValue }}{{ profitOption.optionUnit }}</b> sell price is {{ profitOption.sellPrice.toFixed((stackItem.price.toString().split('.')[1] || []).length) }} USDT = {{ profitOption.optionPercent }}%
      </small>
      <div v-if="stackItem.signal.extraChargeOptions.length > 0"><b>Extra charge</b></div>
      <small v-for="(extraChargeOption, index) in stackItem.signal.extraChargeOptions" :key="index" class="d-block">
        If price fall down by {{ extraChargeOption.percent }}% then charge {{ extraChargeOption.budgetPercentage }}% of budget
      </small>
    </div>
  </v-tooltip>
  <div v-if="!!stackItem.binanceOrder || !!stackItem.buyPrice" class="d-inline">
    <span v-if="!!stackItem.binanceOrder">{{ stackItem.binanceOrder.price }}</span>
    <span v-else>{{ stackItem.buyPrice }}</span>
    <small>&nbsp;USDT</small>
    <v-icon
      v-if="!!stackItem.rating" 
      icon="mdi-circle-small"
      :color="stackItem.buyPrice > Number(stackItem.rating.avgBuyPrice).toFixed((stackItem.buyPrice.toString().split('.')[1] || {length: 2}).length) ? 'red' : 'primary'"
      size="x-large"
    ></v-icon>
  </div>
  <span v-else>n/a</span>
  <v-tooltip location="top">
    <template v-slot:activator="{ props }">
      <v-icon v-bind="props" icon="mdi-information-slab-circle-outline" class="cursor-pointer float-right"/>
    </template>
    <div>
      <div v-if="!!stackItem.signal">
        <b>{{stackItem.symbol }} signal price is: {{ stackItem.buyPrice }} <small>USDT</small></b><br/>
        <small>Can be overwritten by next values if they less:</small><br/>
        Predict: {{ stackItem.predictedPrice }} <small>USDT</small><br/>
        Close price: {{ stackItem.price }} <small>USDT</small><br/>
        Low: {{ stackItem.lowPrice }} <small>USDT</small><br/>
        <div v-if="stackItem.interpolation.btcInterpolationUsdt > 0">BTC interpolation: {{ stackItem.interpolation.btcInterpolationUsdt }} <small>USDT</small></div>
        <div v-if="stackItem.interpolation.ethInterpolationUsdt > 0">ETH interpolation: {{ stackItem.interpolation.ethInterpolationUsdt }} <small>USDT</small></div>
        <div v-if="!!stackItem.binanceOrder">Binance order: {{ stackItem.binanceOrder.price }} <small>USDT</small></div>
      </div>
      <div v-else>
        <b>{{stackItem.symbol }} calculated price is: {{ stackItem.buyPrice }} <small>USDT</small></b><br/>
        <small>Can be overwritten by next values if they less:</small><br/>
        Predict: {{ stackItem.predictedPrice }} <small>USDT</small><br/>
        Close price: {{ stackItem.price }} <small>USDT</small><br/>
        Low: {{ stackItem.lowPrice }} <small>USDT</small><br/>
        <div v-if="stackItem.interpolation.btcInterpolationUsdt > 0">BTC interpolation: {{ stackItem.interpolation.btcInterpolationUsdt }} <small>USDT</small></div>
        <div v-if="stackItem.interpolation.ethInterpolationUsdt > 0">ETH interpolation: {{ stackItem.interpolation.ethInterpolationUsdt }} <small>USDT</small></div>
        <div v-if="!!stackItem.binanceOrder">Binance order: {{ stackItem.binanceOrder.price }} <small>USDT</small></div>
      </div>
    </div>
  </v-tooltip>
</div>
</template>
<script lang="ts">
export default {
  components: {},
  name: "StackBuyPrice",
  props: {
    stackItem: {
      type: Object,
      default: () => {},
    },
  }
}
</script>
<style scoped>
.float-right {
  float: right;
}
.cursor-pointer {
  cursor: pointer;
}
</style>