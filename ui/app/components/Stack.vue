<template>
  <div class="trade-stack-block">
    <table class="stack-items">
      <thead>
        <tr class="stack-item header-row">
          <th colspan="22" class="stack-column">{{$t('stack.column').replace('[provider]', exchange)}}</th>
        </tr>
        <tr class="stack-item header-row">
          <th class="stack-column first-col" style="border-right: none;"/>
          <th class="stack-column" style="border-right: none;"/>
          <th class="stack-column" style="border-right: none;"/>
          <th class="stack-column" style="border-right: none;"/>
          <th class="stack-column"/>
          <th class="stack-column">{{$t('stack.column_symbol')}}</th>
          <th class="stack-column">{{$t('stack.column_rating')}}</th>
          <th class="stack-column">{{$t('stack.column_news')}}</th>
          <th class="stack-column">
            {{$t('stack.column_signal')}}
            <v-tooltip location="top">
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" icon="mdi-information-slab-circle-outline" class="cursor-pointer"/>
              </template>
              <span>{{$t('stack.signal_trading_hint')}}</span>
            </v-tooltip>
          </th>
          <th class="stack-column">{{$t('stack.column_price')}}</th>
          <th class="stack-column">{{$t('stack.column_predicted_price')}}</th>
          <th class="stack-column">{{$t('stack.column_change_speed')}}</th>
          <th class="stack-column">{{$t('stack.column_buy_price')}}</th>
          <th class="stack-column fixed-width-column">
            <div class="switch-row-container">
              {{$t('stack.column_points')}}
              <v-switch
                  v-model="sortingPoints"
                  color="success"
                  size="10"
                  hide-details
                  @change="setSorting('diff')"
              />
            </div>
          </th>
          <th class="stack-column fixed-width-column">
            <div class="switch-row-container">
              {{$t('stack.column_percent')}}
              <v-switch
                  v-model="sortingPercent"
                  color="success"
                  size="10"
                  hide-details
                  @change="setSorting('percent')"
              />
            </div>
          </th>
          <th class="stack-column">{{$t('stack.column_budget')}}</th>
          <th class="stack-column">{{$t('stack.column_balance')}}</th>
          <th class="stack-column">{{$t('stack.column_type')}}</th>
          <th class="stack-column">STS</th>
          <th class="stack-column">BKS</th>
          <th class="stack-column">MDS</th>
          <th class="stack-column">OBS</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(stackItem, index) in getStackItems" :key="`stack-${index}-${stackItem.symbol}`" class="stack-item" :style="{'background-color': 'rgba(29,233,182,0.58)', 'color': '#000000'}">
          <td class="stack-column first-col">
            <div class="switch-row-container">
              <span class="row-number">{{ index + 1 }}</span>
              <v-switch
                  v-model="stackItem.isEnabled"
                  color="success"
                  size="10"
                  hide-details
                  @change="switchSymbol(stackItem)"
                  :disabled="!availableSymbols.includes(stackItem.symbol) && !stackItem.isEnabled"
              />
            </div>
          </td>
          <td class="stack-column filter-btn">
            <v-btn size="x-small" color="primary" @click="pressFilterBtn(stackItem, 'buy')">
              {{$t('stack.filter_btn_buy_if')}}
            </v-btn>
          </td>
          <td class="stack-column filter-btn">
            <v-btn size="x-small" color="primary" @click="pressFilterBtn(stackItem, 'avg')">
              {{$t('stack.filter_btn_avg_if')}}
            </v-btn>
          </td>
          <td class="stack-column filter-btn">
            <v-btn size="x-small" color="primary" @click="pressFilterBtn(stackItem, 'sell')">
              {{$t('stack.filter_btn_sell_if')}}
            </v-btn>
          </td>
          <td class="stack-column filter-btn">
            <PivotPointsPopup
              :pivots="stackItem.pivots"
              :precision="(stackItem.price.toString().split('.')[1] || {length: 2}).length"
              :symbol="stackItem.symbol"
            />
          </td>
          <td class="stack-column text-left">
            <StackSymbol @click="clickSymbol(stackItem.symbol)" :stack-item="stackItem" :exchange="exchange"/>
          </td>
          <td class="stack-column rating-td">
            <v-icon v-if="!!stackItem.rating" color="primary" icon="mdi-star" size="small"/>
            <v-icon v-else color="primary" icon="mdi-star-off" size="small"/>
            <span v-if="!!stackItem.rating" class="rated">
              {{$t('stack.rating_top_prefix')}}-{{stackItem.rating.rating}}
              <v-tooltip location="top">
                <template v-slot:activator="{ props }">
                  <v-icon v-bind="props" icon="mdi-information" color="primary" class="cursor-pointer" size="small"/>
                </template>
                <div>
                  <p>{{$t('stack.avg_sell_price')}}: {{ Number(stackItem.rating.avgSellPrice).toFixed((stackItem.price.toString().split('.')[1] || {length: 2}).length) }} <small>USDT</small></p>
                  <p>{{$t('stack.avg_buy_price')}}: {{ Number(stackItem.rating.avgBuyPrice).toFixed((stackItem.price.toString().split('.')[1] || {length: 2}).length) }} <small>USDT</small></p>
                </div>
              </v-tooltip>
            </span>
            <span v-else><small>{{$t('stack.no_rating')}}</small></span>
          </td>
          <td class="stack-column rating-td">
            <SentimentStat :entity="stackItem"/>
          </td>
          <td class="stack-column">
            <div class="switch-row-container cursor-pointer">
              <v-switch
                  v-model="stackItem.signalTrading"
                  color="primary"
                  size="10"
                  hide-details
                  :disabled="!hasActiveSignalSubscription"
                  @change="switchSignal(stackItem)"
              />
              <div>
                <v-btn
                  icon="mdi-cog"
                  :disabled="!stackItem.signalTrading"
                  variant="text"
                  @click.prevent="signalConfigure(stackItem)"
                  rounded="0"
                  color="primary"
                  elevation="0"
                  class="signal-configure-btn"
                >
                </v-btn>
              </div>
            </div>
          </td>
          <td class="stack-column" :style="{'background-color': stackItem.isPriceValid ? '#1DE9B6FF' : 'rgb(189,47,38)', 'color': stackItem.isPriceValid ? '#000000' : '#FFFFFF'}">
            <v-icon v-if="stackItem.isPriceValid" color="primary" icon="mdi-shield-check"/>
            <v-icon v-else color="white" icon="mdi-shield-alert"/>
            {{ stackItem.price }}
            <small>USDT</small>
            <v-icon
              v-if="!!stackItem.rating"
              icon="mdi-circle-small"
              :color="stackItem.price > Number(stackItem.rating.avgBuyPrice).toFixed((stackItem.price.toString().split('.')[1] || {length: 2}).length) ? 'red' : 'primary'"
              size="x-large"
            ></v-icon>
          </td>
          <td class="stack-column">
            <div v-if="!!stackItem.predictedPrice">
              <span>{{ stackItem.predictedPrice }}</span>
              <small>&nbsp;USDT</small>
            </div>
            <span v-else>n/a</span>
          </td>
          <td class="stack-column" :style="{'background-color': stackItem.priceChangeSpeedAvg >= 0 ? '#1DE9B6FF' : 'rgb(189,47,38)', 'color': stackItem.priceChangeSpeedAvg >= 0 ? '#000000' : '#FFFFFF'}">
            <v-icon v-if="stackItem.priceChangeSpeedAvg == 0" icon="mdi-arrow-right-thick" color="primary"/>
            <v-icon v-if="stackItem.priceChangeSpeedAvg > 0" icon="mdi-arrow-top-right-thick" color="primary"/>
            <v-icon v-if="stackItem.priceChangeSpeedAvg < 0" icon="mdi-arrow-bottom-right-thick" color="white"/>
            {{ stackItem.priceChangeSpeedAvg > 0 ? `+${stackItem.priceChangeSpeedAvg}` : stackItem.priceChangeSpeedAvg.toFixed(2) }}
            <small>{{$t('stack.column_pts')}}</small>
          </td>
          <td class="stack-column">
            <StackBuyPrice :stack-item="stackItem"/>
          </td>
          <td class="stack-column" :style="{'background-color': stackItem.pricePointsDiff >= 0 ? '#1DE9B6FF' : 'rgb(189,47,38)', 'color': stackItem.pricePointsDiff >= 0 ? '#000000' : '#FFFFFF'}">
            <v-icon v-if="stackItem.pricePointsDiff > 0" icon="mdi-arrow-top-right-thick" color="primary"/>
            <v-icon v-if="stackItem.pricePointsDiff < 0" icon="mdi-arrow-bottom-right-thick" color="white"/>
            {{ stackItem.pricePointsDiff > 0 ? `+${stackItem.pricePointsDiff}` : stackItem.pricePointsDiff }}
            <small>PTS</small>
          </td>
          <td class="stack-column" :style="{'background-color': stackItem.percent >= 0 ? '#1DE9B6FF' : 'rgb(189,47,38)', 'color': stackItem.percent >= 0 ? '#000000' : '#FFFFFF'}">
            <v-icon v-if="stackItem.percent > 0" icon="mdi-arrow-top-right-thick" color="primary"/>
            <v-icon v-if="stackItem.percent < 0" icon="mdi-arrow-bottom-right-thick" color="white"/>
            {{ stackItem.percent > 0 ? `+${stackItem.percent}` : stackItem.percent }}%
          </td>
          <td class="stack-column">
            <v-text-field
              class="stack-budget-field"
              :label="null"
              v-model="stackItem.budgetUsdt"
              variant="solo-filled"
              hide-details
              hint="You can change budget here"
              append-inner-icon="mdi-cash"
              @input="changeBudget(stackItem)"
              :loading="loadingMapValue[stackItem.symbol].budget"
              :disabled="stackItem.isExtraCharge"
              pattern="[0-9]*"
              inputmode="numeric"
            />
          </td>
          <td class="stack-column" :style="{'background-color': stackItem.hasEnoughBalance ? '#1DE9B6FF' : 'rgb(189,47,38)', 'color': stackItem.hasEnoughBalance ? '#000000' : '#FFFFFF'}">
            {{ stackItem.balanceAfter.toFixed(2) }}
            <small>USDT</small>
          </td>
          <td class="stack-column">{{ stackItem.isExtraCharge ? 'Extra Charge' : 'Position Open' }}</td>
          <td class="stack-column" v-html="getStrategyText(stackItem, 'sma_trade_strategy')"/>
          <td class="stack-column" v-html="getStrategyText(stackItem, 'base_kline_strategy')"/>
          <td class="stack-column" v-html="getStrategyText(stackItem, 'market_depth_strategy')"/>
          <td class="stack-column" v-html="getStrategyText(stackItem, 'order_based_strategy')"/>
        </tr>
      </tbody>
    </table>
    <v-dialog
      persistent
      v-model="signalConfigDialog.active"
      max-width="500px"
      min-width="380px"
      z-index="9999"
    >
      <div class="position-relative" style="width: 100%; z-index: 999;">
        <v-btn icon variant="text" style="top: 0;right:0;position: absolute" @click="signalConfigDialog.active = false">
          <v-icon icon="mdi-close"/>
        </v-btn>
      </div>
      <v-card v-if="signalConfigDialog.active">
        <v-card-title class="mb-3">{{$t('stack.signal_config.heading')}} {{signalConfigDialog.stackItem.symbol}}</v-card-title>
        <v-card-text>
          <div class="mb-4">
            <div v-if="!!signalConfigDialog.stackItem.rating" class="signal-config-description">
              {{$t('stack.signal_config.description').replace('[symbol]', signalConfigDialog.stackItem.symbol).replace('[trades]', signalConfigDialog.stackItem.rating.trades)}}<br>
              <span><b>{{$t('stack.signal_config.average_buy')}}:</b> {{signalConfigDialog.stackItem.rating.avgBuyPrice}}</span> <small>USDT</small><br>
              <span><b>{{$t('stack.signal_config.average_sell')}}:</b> {{signalConfigDialog.stackItem.rating.avgSellPrice}}</span> <small>USDT</small>
            </div>
            <div v-else class="signal-config-description text-red-darken-2">
              {{$t('stack.signal_config.no_rating_text').replace('[symbol]', signalConfigDialog.stackItem.symbol)}}<br>
              <b>{{$t('stack.signal_config.no_rating_bold_alert')}}</b>
            </div>
          </div>
          <v-row>
            <v-col rows="12" lg="6" md="6" sm="6" xs="6">
              <v-select
                :label="$t('stack.signal_config.labels.signal_period_days')"
                v-model="signalConfigDialog.config.signalPeriodDays"
                variant="underlined"
                :items="[
                  {title: $t('stack.signal_config.labels.signal_period_days_options.1d'), value: 1},
                  {title: $t('stack.signal_config.labels.signal_period_days_options.7d'), value: 7},
                  {title: $t('stack.signal_config.labels.signal_period_days_options.14d'), value: 14},
                  {title: $t('stack.signal_config.labels.signal_period_days_options.30d'), value: 30},
                 ]"
                @update:model-value="onSignalConfig(signalConfigDialog.config, signalConfigDialog.stackItem)"
              ></v-select>
            </v-col>
          </v-row>
          <h4 class="mb-4">{{$t('stack.signal_config.correction_title')}}</h4>
          <v-row class="mt-2">
            <v-col cols="12" lg="6" md="6" sm="6" xs="6" class="p-0 py-0">
              <v-text-field
                :label="$t('stack.signal_config.labels.percent_filter')"
                type="number"
                min="0.5"
                v-model="signalConfigDialog.config.percentFilter"
                @update:model-value="delaySignalUpdate(signalConfigDialog.config, signalConfigDialog.stackItem)"
                variant="underlined"
              ></v-text-field>
            </v-col>
            <v-col cols="12" lg="6" md="6" sm="6" xs="6" class="p-0 py-0">
              <v-switch
                :label="$t('stack.signal_config.labels.avg_buy_correction')"
                v-model="signalConfigDialog.config.avgBuyCorrection"
                @update:model-value="onSignalConfig(signalConfigDialog.config, signalConfigDialog.stackItem)"
                color="primary"
                size="10"
                hide-details
                :disabled="!signalConfigDialog.stackItem.rating"
                class="stack-signal-config-switch"
              />
            </v-col>
          </v-row>
          <v-row class="mb-4">
            <v-col cols="12" lg="6" md="6" sm="6" xs="6" class="p-0 py-0">
              <v-select
                :label="$t('stack.signal_config.labels.sell_price_correction_mode')"
                v-model="signalConfigDialog.config.sellPriceCorrectionMode"
                variant="underlined"
                :items="[
                  {title: $t('stack.signal_config.labels.sell_price_correction_mode_options.max'), value: 'max'},
                  {title: $t('stack.signal_config.labels.sell_price_correction_mode_options.min'), value: 'min'},
                  {title: $t('stack.signal_config.labels.sell_price_correction_mode_options.equal'), value: 'equal'}
                ]"
                type="text"
                :disabled="!signalConfigDialog.stackItem.rating"
                @update:model-value="onSignalConfig(signalConfigDialog.config, signalConfigDialog.stackItem)"
              />
            </v-col>
            <v-col cols="12" lg="6" md="6" sm="6" xs="6" class="p-0 py-0">
              <v-switch
                :label="$t('stack.signal_config.labels.avg_sell_correction')"
                v-model="signalConfigDialog.config.avgSellCorrection"
                @update:model-value="onSignalConfig(signalConfigDialog.config, signalConfigDialog.stackItem)"
                color="primary"
                size="10"
                hide-details
                class="stack-signal-config-switch"
                :disabled="!signalConfigDialog.stackItem.rating"
              />
            </v-col>
          </v-row>
          <h4 class="mb-2">{{$t('stack.signal_config.filtration_title')}}</h4>
          <v-row class="mb-4 mt-2">
            <v-col cols="12" lg="4" md="4" sm="6" xs="6" class="p-0 py-0">
              <v-switch
                :label="$t('stack.signal_config.labels.rating_filter')"
                v-model="signalConfigDialog.config.ratingFilter"
                @update:model-value="onSignalConfig(signalConfigDialog.config, signalConfigDialog.stackItem)"
                color="primary"
                size="10"
                hide-details
                class="stack-signal-config-switch"
              />
            </v-col>
            <v-col cols="12" lg="4" md="4" sm="6" xs="6" class="p-0 py-0">
              <v-switch
                :label="$t('stack.signal_config.labels.avg_buy_filter')"
                v-model="signalConfigDialog.config.avgBuyFilter"
                @update:model-value="onSignalConfig(signalConfigDialog.config, signalConfigDialog.stackItem)"
                color="primary"
                size="10"
                hide-details
                class="stack-signal-config-switch"
                :disabled="!signalConfigDialog.stackItem.rating"
              />
            </v-col>
            <v-col cols="12" lg="4" md="4" sm="6" xs="6" class="p-0 py-0">
              <v-switch
                :label="$t('stack.signal_config.labels.avg_sell_filter')"
                v-model="signalConfigDialog.config.avgSellFilter"
                @update:model-value="onSignalConfig(signalConfigDialog.config, signalConfigDialog.stackItem)"
                color="primary"
                size="10"
                hide-details
                class="stack-signal-config-switch"
                :disabled="!signalConfigDialog.stackItem.rating"
              />
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
    </v-dialog>
  </div>
</template>
<script lang="ts">
import StackBuyPrice from '~/components/ui/StackBuyPrice.vue';
import StackSymbol from '~/components/ui/StackSymbol.vue';
import PivotPointsPopup from '~/components/ui/PivotPointsPopup.vue';
import SentimentStat from '~/components/ui/SentimentStat.vue';

export default {
  name: "Stack",
  components: {SentimentStat, StackSymbol, StackBuyPrice, PivotPointsPopup},
  props: {
    exchange: {
      type: String,
      default: () => '',
    },
    stackItems: {
      type: Array,
      default: () => null,
    },
    sorting: {
      type: String,
      default: () => 'percent',
    },
    hasActiveSignalSubscription: {
      type: Boolean,
      default: () => false,
    },
    budgetUpdateProgress: {
      type: Boolean,
      default: () => false,
    },
    loadingMap: {
      type: Object,
      default: () => {},
    },
    availableSymbols: {
      type: Array,
      default: () => [],
    },
  },
  data() {
    return {
      signalConfigDialog: {
        active: false,
        config: {
          signalPeriodDays: 1,
          percentFilter: 0.5,
          ratingFilter: false,
          avgBuyFilter: false,
          avgSellFilter: false,
          avgBuyCorrection: true,
          avgSellCorrection: false,
          sellPriceCorrectionMode: 'equal',
        },
        stackItem: null,
      },
      sortingPoints: this.sorting === 'diff',
      sortingPercent: this.sorting === 'percent',
      budgetTimer: null,
      signalUpdateTimer: null,
    }
  },
  setup(props: any, {emit}: any) {
    function onSymbol(symbol: any) {
      emit('onSymbol', symbol);
    }
    function onSymbolSwitch(stackItem: any) {
      emit('onSymbolSwitch', stackItem);
    }
    function onSignalSwitch(stackItem: any) {
      emit('onSignalSwitch', stackItem);
    }
    function onSignalConfig(signalConfig: any, stackItem: any) {
      emit('onSignalConfig', {signalConfig, stackItem});
    }
    function onSortingSwitch(stackItem: any) {
      emit('onSortingSwitch', stackItem);
    }
    function onFilterClick(stackItem: any) {
      emit('onFilterClick', stackItem);
    }
    function onBudgetChange(stackItem: any) {
      emit('onBudgetChange', stackItem);
    }

    return {
      onSymbol,
      onSymbolSwitch,
      onSignalSwitch,
      onSortingSwitch,
      onFilterClick,
      onBudgetChange,
      onSignalConfig,
    }
  },
  computed: {
    loadingMapValue() {
      let map: any = {};

      this.stackItems.forEach((item: any) => {
        let budget = (this.loadingMap[item.symbol] || {budget: false}).budget;
        map[item.symbol] = {
          budget,
        };
      });

      return map;
    },
    getStackItems(): any {
      return this.stackItems;
    },
  },
  methods: {
    signalConfigure(stackItem: any) {
      this.signalConfigDialog.config = {
        ...stackItem.signalConfig,
        percentFilter: stackItem.signalConfig.percentFilter.toFixed(2),
      };
      this.signalConfigDialog.stackItem = Object.assign({}, {...stackItem});
      this.signalConfigDialog.active = true;
    },
    getStrategyText(item, name) {
      let strategy = item.strategyDecisions.filter((x) => x.strategyName === name);

      if (strategy.length === 0) {
        return '<span>n/a</span>';
      }

      strategy = strategy[0];

      return `<span>${strategy.operation} = ${strategy.score}</span>`
    },
    clickSymbol(symbol) {
      this.onSymbol(symbol);
    },
    switchSymbol(stackItem) {
      this.onSymbolSwitch(stackItem);
    },
    switchSignal(stackItem) {
      this.onSignalSwitch(stackItem);
    },
    setSorting(sorting) {
      switch (sorting) {
        case 'percent':
          this.sortingPoints = false;
          this.sortingPercent = true;
          break;
        case 'diff': {
          this.sortingPoints = true;
          this.sortingPercent = false;
          break;
        }
      }

      this.onSortingSwitch(sorting);
    },
    pressFilterBtn(stackItem, operation) {
      this.onFilterClick({stackItem, operation});
    },
    changeBudget(stackItem: any) {
      if (this.budgetTimer) {
        clearTimeout(this.budgetTimer);
      }

      this.budgetTimer = setTimeout(() => {
        this.onBudgetChange(stackItem);
      }, 1000)
    },
    delaySignalUpdate(config, stackItem) {
      if (this.signalUpdateTimer) {
        clearTimeout(this.signalUpdateTimer);
      }

      this.signalUpdateTimer = setTimeout(() => {
        this.onSignalConfig(config, stackItem);
      }, 2000);
    },
  },
}
</script>
<style scoped>
.stack-items {
  border: 1px solid #888888;
  width: 100%;
  min-width: 2150px;
  border-collapse: collapse;
  border-spacing: 0;
}
.stack-item {
  border-bottom: 1px solid #888888;
}
.stack-column {
  padding: 2px;
  font-size: 12px;
  border-right: 1px solid #888888;
}
.cursor-pointer {
  cursor: pointer;
}
.stack-column:last-child {
  border-right: none;
}
.header-row {
  background-color: rgb(0, 77, 64);
  color: #FFFFFF;
  text-transform: uppercase;
  font-weight: bold;
}
.trade-stack-block {
  overflow-x: scroll;
  width: 100%;
  height: auto;
  overflow-y: hidden;
}
.trade-stack-block::-webkit-scrollbar {
  display: none;
}
.row-number {
}
.first-col {
  background-color: rgb(0, 77, 64);
  text-align: center;
  font-weight: bold;
  color: #FFFFFF;
  max-width: 72px;
  width: 72px;
  overflow: hidden;
}
.switch-row-container {
  height: 23px;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding-left: 5px;
  padding-right: 5px;
}
.fixed-width-column {
  max-width: 110px;
  width: 110px;
  overflow: hidden;
}
.filter-btn {
  width: 64px !important;
  margin: auto;
  text-align: center;
}
.rating-td {
  min-width: 84px;
  text-align: left;
}
.rating-td .rated {
  font-weight: bold;
}
.signal-configure-btn {
  width: 30px !important;
  height: 30px !important;
}
.signal-config-description {
  font-size: 12px;
}
</style>