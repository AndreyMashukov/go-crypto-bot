<template>
  <div>
    <v-table height="500px" fixed-header v-if="lastOrders && lastOrders.length > 0">
      <thead>
      <tr>
        <th class="text-left">{{$t('trade_table.nickname')}}</th>
        <th class="text-left">{{$t('trade_table.symbol')}}</th>
        <th class="text-left">{{$t('trade_table.open_price')}}</th>
        <th class="text-left">{{$t('trade_table.close_price')}}</th>
        <th class="text-left">{{$t('trade_table.profit')}}</th>
        <!--            <th class="text-left">Profit percent</th>-->
      </tr>
      </thead>
      <tbody>
      <tr
          v-for="item in lastOrders"
          :key="item.id"
      >
        <td>{{ item.nickname }}</td>
        <td>{{ item.symbol }}</td>
        <td class="order-trade">
          <div>{{ item.buy }}<small>USDT</small> x {{ item.buyQuantity }}</div>
          <small>{{ item.open }}</small>
        </td>
        <td class="order-trade">
          <div>{{ item.sell }}<small>USDT</small> х {{ item.sellQuantity }}</div>
          <small>{{ item.close }}</small>
        </td>
        <td>
          <v-chip variant="flat" color="success" size="small" v-if="item.profit >= 0">
            +{{ item.profit }}$
          </v-chip>
          <v-chip variant="flat" color="red" size="small" v-else>
            -{{ item.profit }}$
          </v-chip>
        </td>
      </tr>
      </tbody>
    </v-table>
    <div class="mx-auto" v-else>
      <p>{{$t('trade_table.not_completed_trade')}}</p>
    </div>
  </div>
</template>

<script>
export default {
  name: "TradeTable",
  props: {
    lastOrders: {
      type: Array,
      default: () => [],
    }
  }
}
</script>

<style scoped>
.order-trade {
  font-size: 14px;
  min-width: 140px;
}
.order-trade small {
  font-size: 10px;
}
</style>