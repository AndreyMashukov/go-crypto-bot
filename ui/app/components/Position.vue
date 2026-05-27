<template>
  <div :style="{'background-color': position.percent > 0 ? 'rgba(29,233,182,0.33)' : 'rgba(244,67,54,0.4)'}" :class="`position-item-${orientation}`">
    <div v-if="!!position.changeToday" class="daily-position-change" :style="{'background-color': position.changeToday.profit >= 0 ? '#1DE9B6FF' : 'rgb(189,47,38)', 'color': position.changeToday.profit >= 0 ? '#000000' : '#FFFFFF'}">
      <v-icon v-if="position.changeToday.profit == 0" icon="mdi-arrow-right-thick" color="primary"/>
      <v-icon v-if="position.changeToday.profit > 0" icon="mdi-arrow-top-right-thick" color="primary"/>
      <v-icon v-if="position.changeToday.profit < 0" icon="mdi-arrow-bottom-right-thick" color="white"/>
      {{ position.changeToday.profit > 0 ? `+${position.changeToday.profit.toFixed(2)}` : position.changeToday.profit.toFixed(2) }}<small class="usdt">USDT</small>
      <small>
        ({{ position.changeToday.percent > 0 ? `+${position.changeToday.percent.toFixed(2)}` : position.changeToday.percent.toFixed(2) }}%)
      </small>
    </div>
    <div v-else-if="position.priceChangeSpeedAvg != null && showPriceChangeSpeed" class="price-change-speed" :style="{'background-color': position.priceChangeSpeedAvg >= 0 ? '#1DE9B6FF' : 'rgb(189,47,38)', 'color': position.priceChangeSpeedAvg >= 0 ? '#000000' : '#FFFFFF'}">
      <v-icon v-if="position.priceChangeSpeedAvg == 0" icon="mdi-arrow-right-thick" color="primary"/>
      <v-icon v-if="position.priceChangeSpeedAvg > 0" icon="mdi-arrow-top-right-thick" color="primary"/>
      <v-icon v-if="position.priceChangeSpeedAvg < 0" icon="mdi-arrow-bottom-right-thick" color="white"/>
      {{ position.priceChangeSpeedAvg > 0 ? `+${position.priceChangeSpeedAvg.toFixed(2)}` : position.priceChangeSpeedAvg.toFixed(2) }}
      <small>{{$t('position.position_pts_sec')}}</small>
    </div>
    <div v-if="showSentiment">
      <SentimentStat :entity="position"/>
    </div>
    <div class="position-symbol">
      {{ position.symbol }}
      <v-icon
        v-if="!!position.rating"
        icon="mdi-circle-small"
        :color="sellPrice > Number(position.rating.avgSellPrice).toFixed(precision) ? 'red' : 'primary'"
        size="small"
        class="position-sell-price-dot"
      ></v-icon>
      <small v-if="trader.length > 0">{{ trader }}</small>
      <div class="position-rating" v-if="isVertical">
        <v-icon v-if="!!position.rating" color="primary" icon="mdi-star" size="small"/>
        <v-icon v-else color="primary" icon="mdi-star-off" size="small"/>
        <span v-if="!!position.rating" class="rated">{{$t('position.rating_top_prefix')}}-{{position.rating.rating}}</span>
        <span v-else>{{$t('position.no_rating')}}</span>
        <v-tooltip location="top" v-if="!!position.rating">
          <template v-slot:activator="{ props }">
            <v-icon v-bind="props" icon="mdi-information" color="primary" class="cursor-pointer" size="small"/>
          </template>
          <div>
            <p>{{$t('position.avg_sell_price')}}: {{ Number(position.rating.avgSellPrice).toFixed(precision) }} <small>USDT</small></p>
            <p>{{$t('position.avg_buy_price')}}: {{ Number(position.rating.avgBuyPrice).toFixed(precision) }} <small>USDT</small></p>
          </div>
        </v-tooltip>
      </div>
      <div class="swap-info">
        <small v-if="position.order.swap" class="swap-text">{{$t('position.position_swap')}}</small>
      </div>
    </div>
    <div :class="`position-inner-${orientation}`">
      <div :class="`position-percent-${orientation}`" :style="{'background-color': position.percent > 0 ? '#1DE9B6FF' : 'rgba(244,67,54,0.67)'}">
        <span v-if="position.percent > 0">+</span>{{ position.percent }}%<br>
        <span v-if="position.percent > 0">+</span><span>{{ position.profit }}</span><small>USDT</small>
      </div>
      <div :class="`position-prices-${orientation}`">
        <div class="price-now" :style="{'background-color': position.isPriceExpired ? 'rgba(244,67,54)' : '', 'color': position.isPriceExpired ? '#FFFFFF' : ''}">
          <v-icon v-if="!position.isPriceExpired" color="primary" icon="mdi-shield-check" size="small"/>
          <v-icon v-else color="white" icon="mdi-shield-alert" size="small"/>
          <b>{{$t('position.position_now')}}</b> {{ position.kLine.c }}
        </div><br>
        <div v-if="showAvgPrice">
          <v-icon color="primary" icon="mdi-finance" size="small"/>
          <b>{{$t('position.position_avg')}}</b> {{ avgPrice }}
          <v-icon
              v-if="!!position.rating"
              icon="mdi-circle-small"
              :color="avgPrice > Number(position.rating.avgBuyPrice).toFixed(precision) ? 'red' : 'primary'"
              size="x-large"
              class="position-sell-price-dot"
          ></v-icon>
        </div>
        <div>
          <v-icon v-if="!!position.binanceOrder" color="primary" icon="mdi-cart" size="small"/>
          <v-icon v-else-if="!position.canSell" color="primary" icon="mdi-hand-back-right" size="small"/>
          <v-icon v-else color="primary" icon="mdi-timer" size="small"/>
          <b>{{$t('position.position_sell')}}</b>&nbsp;
          <span>{{ sellPrice }}</span>
          <v-icon
            v-if="!!position.rating"
            icon="mdi-circle-small"
            :color="sellPrice > Number(position.rating.avgSellPrice).toFixed(precision) ? 'red' : 'primary'"
            size="x-large"
            class="position-sell-price-dot"
          ></v-icon>
        </div>
        <div>
          <v-icon color="primary" icon="mdi-hand-coin" size="small"/>
          <b>{{$t('position.position_qty')}}</b>&nbsp;
          <span v-if="!!position.binanceOrder">{{ position.binanceOrder.origQty }}</span>
          <span v-else>{{ executedQuantity }}</span>
        </div>
        <span class="predict"><b>{{$t('position.position_predict')}}</b> {{ position.predictedPrice > 0 ? position.predictedPrice : 'n/a' }}</span>
      </div>
    </div>
    <div v-if="showUsedBudget" class="used-budget">
      <v-icon icon="mdi-lock" color="primary" v-if="position.order.swap" size="x-small"/>
      <v-icon icon="mdi-check-circle" color="primary" v-else size="x-small"/>
      {{ (position.order.price * position.order.executedQuantity).toFixed(2) }}
      <small>$</small>
    </div>
    <div v-if="showProfitPercent" class="profit-percent">
      <div v-if="!!position.manualOrder">
        <v-icon icon="mdi-account" color="primary" size="x-small"/>
        {{ position.manualOrder.price }}
      </div>
      <div v-else>
        <v-icon icon="mdi-percent" color="primary" size="x-small"/>
        {{ position.closeStrategy.minProfitPercent.toFixed(2) }}
      </div>
      <div class="target-profit-block">
        +<b>{{getTargetProfit}}</b> <small>USDT</small>
      </div>
    </div>
    <div v-if="showBuySwitch" class="position-switch">
      <v-switch
          v-model="isEnabled"
          :color="position.percent > 0 ? '#1DE9B6FF' : 'rgba(244,67,54,0.67)'"
          size="10"
          hide-details
          @change="switchEnabled"
      />
    </div>
    <v-btn
      append-icon="mdi-cog"
      v-if="showBottomBtn"
      color="primary"
      variant="elevated"
      size="x-small"
      @click.prevent="pressBottomBtn"
      class="w-100"
    >
      {{ bottomBtnText }}
    </v-btn>
    <div v-if="showConditionBtn" class="condition-btn-container">
      <v-btn
        append-icon="mdi-hand-back-right"
        color="primary"
        variant="elevated"
        size="x-small"
        class="condition-btn"
        elevation="0"
        @click.prevent="pressFilterBtn(position, 'buy')"
      >
        {{$t('position.position_btn_buy_if')}}
      </v-btn>
      <v-btn
        append-icon="mdi-hand-back-right"
        color="primary"
        variant="elevated"
        size="x-small"
        class="condition-btn"
        elevation="0"
        @click.prevent="pressFilterBtn(position, 'avg')"
      >
        {{$t('position.position_btn_avg_if')}}
      </v-btn>
      <v-btn
        append-icon="mdi-hand-back-right"
        color="primary"
        variant="elevated"
        size="x-small"
        class="condition-btn"
        elevation="0"
        @click.prevent="pressFilterBtn(position, 'sell')"
      >
        {{$t('position.position_btn_sell_if')}}
      </v-btn>
    </div>
    <PivotPointsPopup
      :pivots="position.pivots"
      :precision="precision"
      v-if="showPivots"
      class="condition-btn"
      :symbol="position.symbol"
    />
    <div v-if="showCancelManual">
      <v-btn
        append-icon="mdi-close-octagon"
        color="red"
        variant="elevated"
        size="x-small"
        @click.prevent="cancelManual"
        class="w-100"
        :disabled="!position.manualOrder"
      >
        {{$t('position.position_btn_manual')}}
      </v-btn>
    </div>
    <div v-if="!position.binanceOrder || Number(position.binanceOrder.executedQty) === 0">
      <v-progress-linear
          class="mt-2"
          height="5px"
          :max="position.targetProfit*1000"
          :model-value="position.profit*1000"
          color="success"
          v-if="position.profit > 0"
      ></v-progress-linear>
      <v-progress-linear
          class="mt-2"
          height="5px"
          max="100"
          value="0"
          color="red"
          v-else
      ></v-progress-linear>
    </div>
    <div v-else>
      <v-progress-linear
          class="mt-2"
          height="5px"
          :max="Number(position.binanceOrder.origQty)*1000"
          :model-value="Number(position.binanceOrder.executedQty)*1000"
          color="info"
      ></v-progress-linear>
    </div>
  </div>
</template>
<script lang="ts">
import PivotPointsPopup from '~/components/ui/PivotPointsPopup.vue';
import SentimentStat from '~/components/ui/SentimentStat.vue';

export default {
  name: "Position",
  components: {SentimentStat, PivotPointsPopup},
  props: {
    position: {
      type: Object,
      default: () => null,
    },
    trader: {
      type: String,
      default: () => "",
    },
    orientation: {
      type: String,
      default: () => 'vertical'
    },
    showBottomBtn: {
      type: Boolean,
      default: () => false,
    },
    showConditionBtn: {
      type: Boolean,
      default: () => false,
    },
    showAvgPrice: {
      type: Boolean,
      default: () => false,
    },
    showPriceChangeSpeed: {
      type: Boolean,
      default: () => false,
    },
    showPivots: {
      type: Boolean,
      default: () => false,
    },
    showProfitPercent: {
      type: Boolean,
      default: () => false,
    },
    bottomBtnText: {
      type: String,
      default: () => '',
    },
    showUsedBudget: {
      type: Boolean,
      default: () => false,
    },
    showSentiment: {
      type: Boolean,
      default: () => false,
    },
    showBuySwitch: {
      type: Boolean,
      default: () => false,
    },
    showCancelManual: {
      type: Boolean,
      default: () => false,
    },
  },
  data() {
    return {
      isEnabled: this.position.isEnabled,
    };
  },
  setup(props: any, {emit}: any) {
    function onBottomBtn(data: any) {
      emit('onBottomBtn', data);
    }
    function onSwitchEnabled(data: any) {
      emit('onSwitchEnabled', data);
    }
    function onCancelManual(data: any) {
      emit('onCancelManual', data);
    }
    function onFilterClick(position: any) {
      emit('onFilterClick', position);
    }

    return {
      onBottomBtn,
      onSwitchEnabled,
      onCancelManual,
      onFilterClick,
    }
  },
  computed: {
    getTargetProfit: {
      get() {
        return ((this.sellPrice * this.position.order.executedQuantity) - (this.position.order.price * this.position.order.executedQuantity)).toFixed(2);
      },
    },
    sellPrice() {
      if (!!this.position.binanceOrder) {
        return this.position.binanceOrder.price;
      }

      return this.position.sellPrice;
    },
    precision(): number {
      let split = this.position.sellPrice.toString().split('.');
      if (split.length > 1) {
        return split[1].length;
      }

      split = this.position.kLine.c.toString().split('.');

      if (split.length > 1) {
        return split[1].length;
      }

      return 10;
    },
    executedQuantity() {
      const split = this.position.order.quantity.toString().split('.') || [];

      if (split.length > 1) {
        return this.position.order.executedQuantity.toFixed(split[1].length);
      }

      return this.position.order.executedQuantity;
    },
    avgPrice() {
      if (!this.position.kLine) {
        return this.position.order.price;
      }

      return this.position.order.price.toFixed(this.precision);
    },
    isVertical() {
      return this.orientation === 'vertical';
    },
  },
  methods: {
    pressBottomBtn() {
      this.onBottomBtn();
    },
    switchEnabled() {
      this.onSwitchEnabled();
    },
    cancelManual() {
      this.onCancelManual();
    },
    pressFilterBtn(position, operation) {
      this.onFilterClick({position, operation});
    }
  },
}
</script>

<style scoped>
.target-profit-block {
  font-size: 10px;
  color: rgba(0, 0, 0, 0.65);
}
.position-inner-vertical {
  display: block;
}
.position-inner-horizontal {
  display: inline-block;
}
.position-inner-horizontal .position-percent-horizontal {
  margin-right: 10px;
  padding: 5px !important;
}
.position-inner-horizontal > div {
  display: inline-block;
}
.position-item-vertical {
  display: inline-block;
  margin: 1px;
  padding: 6px;
  width: calc((100% / 14) - 2px);
  box-sizing: border-box;
  cursor: pointer;
  position: relative;
}
@media only screen and (max-width: 1280px) {
  .position-item-vertical {
    width: calc((100% / 10) - 2px);
  }
}
@media only screen and (max-width: 1024px) {
  .position-item-vertical {
    width: calc((100% / 8) - 2px);
  }
}
@media only screen and (max-width: 900px) {
  .position-item-vertical {
    width: calc((100% / 7) - 2px);
  }
}
@media only screen and (max-width: 750px) {
  .position-item-vertical {
    width: calc((100% / 6) - 2px);
  }
}
@media only screen and (max-width: 650px) {
  .position-item-vertical {
    width: calc((100% / 5) - 2px);
  }
}
@media only screen and (max-width: 500px) {
  .position-item-vertical {
    width: calc((100% / 3) - 2px);
  }
}
@media only screen and (max-width: 360px) {
  .position-item-vertical {
    width: calc((100% / 3) - 2px);
  }
}
.position-item-horizontal {
  display: inline-block;
  margin: 1px;
  padding: 6px;
  width: 200px;
  box-sizing: border-box;
  cursor: pointer;
  position: relative;
}
.position-symbol {
  font-weight: bold;
  font-size: 14px;
}
.position-symbol small {
  display: block;
  font-weight: normal;
}
.position-percent-vertical {
  color: #000000;
  font-size: 16px;
  font-weight: bold;
  background-color: #fff;
  border-radius: 5px;
  text-align: center;
  margin-top: 3px;
  border: 2px solid rgba(239, 239, 239, 0.7);
}
.position-percent-horizontal {
  color: #000000;
  font-size: 14px;
  font-weight: bold;
  background-color: #fff;
  border-radius: 5px;
  text-align: center;
  margin-top: 3px;
  border: 2px solid rgba(239, 239, 239, 0.7);
  padding: 0 10px;
}
.position-percent span {
  font-size: 12px;
  color: #1F1F1FFF;
}
.position-percent small {
  font-size: 8px;
  color: #1F1F1FFF;
}
.position-prices-vertical {
  margin-top: 4px;
  font-size: 9px;
  overflow: hidden;
  white-space: nowrap;
}
.position-prices-horizontal {
  margin-top: 4px;
  font-size: 9px;
}
.position-prices-vertical .predict,
.position-prices-horizontal .predict {
  font-size: 9px;
}
.position-prices-horizontal .predict {
  position: absolute;
  top: 6px;
  right: 6px;
}
.swap-text {
  font-size: 6px;
  display: inline;
}
.swap-info {
  min-height: 9px;
}
.used-budget {
  margin-top: 6px;
  font-size: 14px;
  text-align: center;
  font-weight: bold;
}
.profit-percent {
  font-size: 14px;
  text-align: center;
  font-weight: bold;
}
.price-now {
  display: inline;
  padding: 1px 1px 1px 0;
  border-radius: 2px;
}
.position-switch {
  display: flex;
  flex-direction: row;
  align-items: center;
  height: 28px;
}
.condition-btn-container {
  display: flex;
  flex-direction: column;
  gap: 1px;
  margin-top: 4px;
}
.condition-btn {
}
.daily-position-change {
  font-size: 10px;
  margin-bottom: 2px;
}
.daily-position-change .usdt {
  font-size: 7px;
}
.price-change-speed {
  font-size: 10px;
  margin-bottom: 2px;
}
.price-change-speed small {
  font-size: 6px;
}
.position-rating {
  font-size: 10px;
  font-weight: normal;
  height: 12px;
}
.position-rating .rated {
  font-weight: bold;
}
.position-sell-price-dot {
  margin-left: 0 !important;
  left: -3px;
  top: -1px;
  height: 12px;
}
</style>