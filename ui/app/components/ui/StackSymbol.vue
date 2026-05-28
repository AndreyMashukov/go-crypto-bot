<template>
<div>
  <v-tooltip v-if="!!stackItem.binanceOrder" location="top">
    <template #activator="{ props }">
      <v-icon icon="mdi-cart" v-bind="props" :color="'primary'" class="cursor-pointer"/>
    </template>
    <div>
      <b>{{$t('stack.binance_order_tooltip.title').replace('[provider]', exchange)}}</b><br>
      <div><b>{{$t('stack.binance_order_tooltip.symbol')}}:</b>{{ stackItem.binanceOrder.symbol }}</div>
      <div><b>{{$t('stack.binance_order_tooltip.order_id')}}:</b>{{ stackItem.binanceOrder.orderId }}</div>
      <div><b>{{$t('stack.binance_order_tooltip.price')}}:</b>{{ stackItem.binanceOrder.price }} <small>USDT</small></div>
      <div><b>{{$t('stack.binance_order_tooltip.qty')}}:</b>{{ stackItem.binanceOrder.origQty }} <small>{{ stackItem.binanceOrder.symbol.replace('USDT', '') }}</small></div>
      <div><b>{{$t('stack.binance_order_tooltip.status')}}:</b>{{ stackItem.binanceOrder.status }}</div>
      <div><b>{{$t('stack.binance_order_tooltip.side')}}:</b>{{ stackItem.binanceOrder.side }}</div>
    </div>
  </v-tooltip>
  <v-tooltip v-if="stackItem.isBuyLocked" location="top">
    <template #activator="{ props }">
      <v-icon icon="mdi-lock" v-bind="props" :color="'primary'" class="cursor-pointer"/>
    </template>
    <div>
      <b>Loss security system temporary locked trading for this symbol!</b><br>
    </div>
  </v-tooltip>
  <v-tooltip v-if="stackItem.isFiltered" location="top">
    <template #activator="{ props }">
      <v-icon icon="mdi-hand-back-right" v-bind="props" :color="'primary'" class="cursor-pointer"/>
    </template>
    <div>
      <b>Your conditions is not passed for making operations of type {{ stackItem.isExtraCharge ? 'Extra Charge' : 'BUY' }}!</b><br>
    </div>
  </v-tooltip>
  <v-tooltip v-if="!stackItem.isEnabled" location="top">
    <template #activator="{ props }">
      <v-icon icon="mdi-close-octagon" v-bind="props" :color="'primary'" class="cursor-pointer"/>
    </template>
    <div>
      <b>Symbol '{{ stackItem.symbol }}' is disabled for any kind of BUY operations!</b><br>
    </div>
  </v-tooltip>
  <span class="font-weight-bold stack-symbol">{{ stackItem.symbol }}</span>
  <small>[{{ stackItem.index+1 }}]</small>
</div>
</template>
<script lang="ts">
export default {
  name: "StackSymbol",
  components: {},
  props: {
    exchange: {
      type: String,
      default: () => '',
    },
    stackItem: {
      type: Object,
      default: () => {},
    },
  }
}
</script>
<style scoped>
.cursor-pointer {
  cursor: pointer;
}
.stack-symbol {
  cursor: pointer;
}
.stack-symbol:hover {
  color: rgb(3, 28, 166);
}
</style>