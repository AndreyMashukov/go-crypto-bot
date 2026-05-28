<template>
  <v-card
      max-width="100%"
      :append-icon="`mdi-numeric-${step}-box`"
  >
    <template #title>
      {{symbol}}
      <small v-if="!!status">{{status}}</small>
      <small v-else>WAITING</small>
    </template>
    <template #subtitle>
      <div v-if="side && side === 'SELL' && quantity && price">
        <small>Swap price: 1 {{asset}} = {{price}} {{getQuote}}</small>
        <div>
          <small>Now price: <b>BUY:</b> {{buyPrice}} {{getQuote}} <b>SELL:</b> {{sellPrice}} {{getQuote}}</small>
        </div>
        <b>{{side}}:</b> {{quantity}}<small>{{asset}}</small> x {{price}}<small>{{getQuote}}</small> = {{(quantity*price).toFixed(precision)}}<small>{{getQuote}}</small>
      </div>
      <div v-else-if="side && side === 'BUY' && quantity && price">
        <small>Swap price: 1 {{getQuote}} = {{price}} {{asset}}</small>
        <div>
          <small>Now price: <b>BUY:</b> {{buyPrice}} {{asset}} <b>SELL:</b> {{sellPrice}} {{asset}}</small>
        </div>
        <b>{{side}}:</b> {{quantity}}<small>{{getQuote}}</small> x {{price}}<small>{{asset}}</small> = {{(quantity*price).toFixed(precision)}}<small>{{asset}}</small>
      </div>
      <div v-if="!!balance">
        <b>Free: </b>{{balance.free}}<small>{{asset}}</small>&nbsp;
        <b>Locked: </b>{{balance.locked}}<small>{{asset}}</small>
      </div>
    </template>
    <template #prepend>
      <v-icon v-if="swapStatus === 'success' && (!status || status === 'NEW')" size="30" icon="mdi-cancel" color="gray"/>
      <v-icon v-else-if="status === 'FILLED'" size="30" icon="mdi-check-circle" color="primary"/>
      <v-icon v-else-if="status === 'FILLED_FORCE'" size="30" icon="mdi-alert-circle-check" color="primary"/>
      <v-icon v-else-if="status === 'FILLED_RB'" size="30" icon="mdi-close-circle" color="warning"/>
      <v-icon v-else-if="status === 'PARTIALLY_FILLED'" size="30" icon="mdi-timelapse" color="info"/>
      <v-progress-circular v-else size="30" width="4" color="info" indeterminate/>
<!--      <v-icon icon="mdi-timer" color="info" v-else/>-->
    </template>
  </v-card>
</template>
<script lang="ts">
export default {
  name: "SwapStep",
  components: {},
  props: {
    symbol: {
      type: String,
      default: () => '',
    },
    swapStatus: {
      type: String,
      default: () => '',
    },
    status: {
      type: String,
      default: () => '',
    },
    asset: {
      type: String,
      default: () => '',
    },
    side: {
      type: String,
      default: () => '',
    },
    price: {
      type: Number,
      default: () => 0,
    },
    buyPrice: {
      type: Number,
      default: () => 0,
    },
    sellPrice: {
      type: Number,
      default: () => 0,
    },
    quantity: {
      type: Number,
      default: () => 0,
    },
    step: {
      type: Number,
      default: () => 0,
    },
    swapId: {
      type: Number,
      default: () => 0,
    },
    balance: {
      type: Object,
      default: () => {},
    },
  },
  computed: {
    getQuote(): string {
      return this.symbol.replace(this.asset, '');
    },
    precision(): number {
      if (this.price) {
        const split = this.price.toString().split('.');

        if (split[1]) {
          return split[1].length;
        }

        return 0;
      }

      return 10;
    }
  }
}
</script>
<style scoped>
.swap-id-block {
  font-size: 12px;
  line-height: 12px;
}
</style>
