<template>
  <v-card class="swap-main">
    <template v-slot:title>
      <div class="swap-id-block">SWAP ID {{swap.action.id}}</div>
    </template>
    <template v-slot:subtitle>
      <small>{{swap.action.startQuantity}}</small> -> <small>{{!!swap.action.endQuantity ? swap.action.endQuantity : '?'}}</small>&nbsp;
      <span v-if="swap.action.status === 'success' && !!swap.action.startQuantity && !!swap.action.endQuantity">
        {{ profitCoin }} <small>{{swap.action.asset}}</small>
      </span>
    </template>
    <v-row>
      <v-col cols="12" lg="4" md="4" sm="12" xs="12" class="mt-0 pt-0">
        <SwapStep
          :symbol="swap.action.swapOneSymbol"
          :status="swap.action.swapOneExternalStatus"
          :asset="swap.action.asset"
          :side="swap.action.swapOneSide"
          :price="swap.action.swapOnePrice"
          :quantity="swap.action.swapOneQuantity"
          :step="1"
          :swap-id="swap.action.id"
          :swap-status="swap.action.status"
          :balance="swap.balance[swap.action.asset]"
          :buyPrice="swap.action.priceOneBuy"
          :sellPrice="swap.action.priceOneSell"
        />
      </v-col>
      <v-col cols="12" lg="4" md="4" sm="12" xs="12" class="mt-0 pt-0">
        <SwapStep
          :symbol="swap.action.swapTwoSymbol"
          :status="swap.action.swapTwoExternalStatus"
          :asset="swap.action.swapOneSymbol.replace(swap.action.asset, '')"
          :side="swap.action.swapTwoSide"
          :price="swap.action.swapTwoPrice"
          :quantity="swap.action.swapTwoQuantity"
          :step="2"
          :swap-id="swap.action.id"
          :swap-status="swap.action.status"
          :balance="swap.balance[swap.action.swapOneSymbol.replace(swap.action.asset, '')]"
          :buyPrice="swap.action.priceTwoBuy"
          :sellPrice="swap.action.priceTwoSell"
        />
      </v-col>
      <v-col cols="12" lg="4" md="4" sm="12" xs="12" class="mt-0 pt-0">
        <SwapStep
          :symbol="swap.action.swapThreeSymbol"
          :status="swap.action.swapThreeExternalStatus"
          :asset="swap.action.swapTwoSymbol.replace(swap.action.swapOneSymbol.replace(swap.action.asset, ''), '')"
          :side="swap.action.swapThreeSide"
          :price="swap.action.swapThreePrice"
          :quantity="swap.action.swapThreeQuantity"
          :step="3"
          :swap-id="swap.action.id"
          :swap-status="swap.action.status"
          :balance="swap.balance[swap.action.swapTwoSymbol.replace(swap.action.swapOneSymbol.replace(swap.action.asset, ''), '')]"
          :buyPrice="swap.action.priceThreeBuy"
          :sellPrice="swap.action.priceThreeSell"
        />
      </v-col>
    </v-row>
  </v-card>
</template>
<script lang="ts">
import SwapStep from '~/components/ui/SwapStep.vue';

export default {
  components: {SwapStep},
  name: "Swap",
  props: {
    swap: {
      type: Object,
      default: () => {},
    },
  },
  computed: {
    profitCoin(): string {
      const profit = this.swap.action.endQuantity - this.swap.action.startQuantity;

      if (profit < 0) {
        return profit.toFixed(10).toString();
      }

      if (profit === 0) {
        return '0.00';
      }

      return `+${profit.toFixed(10)}`;
    },
  },
}
</script>
<style scoped>
.swap-main {
  margin: 5px;
  padding: 10px;
}
</style>
