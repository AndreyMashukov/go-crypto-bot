<template>
  <v-container fluid>
    <div>
      <v-alert
          class="mt-2 mb-4"
          border="bottom"
          border-color="info"
      >
        <template v-slot:text>
          {{$t(`dashboard.documentation.text`)}} <a class="documentation-link" :href="$t(`dashboard.documentation.link.url`)" target="_blank">{{$t(`dashboard.documentation.link.text`)}}</a>
        </template>
      </v-alert>
    </div>

    <v-row v-if="!!positions" class="mb-1">
      <v-col cols="12" v-if="stack && stack.restartRequired">
        <v-alert
            type="warning"
            :title="$t('dashboard.restart_required.title')"
            :text="$t('dashboard.restart_required.text')"
            class="mb-2"
        />
        <v-btn
            :loading="dataLoading.deploy"
            variant="flat"
            color="success"
            @click="deployBot"
            class="mb-4 mt-2"
            append-icon="mdi-restart"
        >
          <b>{{$t('bot.restart_btn')}}</b>
        </v-btn>
      </v-col>
      <v-col cols="12" class="profits">
        <div>
          <div class="profit-period-selector">
            <v-select
              v-model="profitPeriod"
              :items="periodOptions"
              @update:modelValue="refreshProfit"
              hide-details
              variant="filled"
            />
          </div>
          <div class="profit-scroll-horizontal">
            <ProfitItem
              :title="portfolioReport.title"
              :profit="portfolioReport.profit"
              :trades="portfolioReport.trades"
              :trade-volume="portfolioReport.tradeVolume"
              :key="`${portfolioReport.title}-portfolio`"
              :background-color="portfolioReport.backgroundColor"
              :font-color="portfolioReport.color"
              :avg-percent="portfolioReport.avgPercent"
            />
            <ProfitItem
              v-for="(reportItem, index) in profits"
              :title="reportItem.title"
              :profit="Number(reportItem.profit)"
              :trades="Number(reportItem.trades)"
              :trade-volume="Number(reportItem.tradeVolume)"
              :key="`${reportItem.title}-${index}`"
              :avg-percent="Number(reportItem.avgPercent)"
            />
          </div>
        </div>
      </v-col>
      <v-col cols="12" class="available-symbol-list">
        <div v-for="(symbol, index) in availableSymbolList" :key="index" class="symbol-list-item">
          <span class="symbol-title">{{ symbol.symbol }}</span>
          <v-btn size="x-small" color="primary" @click="addQuickSymbol(symbol)" :loading="quickSymbolLoading[symbol.symbol]">+ ADD</v-btn>
        </div>
      </v-col>
      <v-col cols="12" class="trade-stack">
        <Stack
          v-if="!!stack && stack.stack && stack.stack.length > 0"
          :available-symbols="availableSymbols"
          :stack-items="stack.stack"
          :sorting="stack.sorting"
          :has-active-signal-subscription="stack.hasActiveSignalSubscription"
          :loading-map="stackLoadingMap"
          :exchange="currentBot.provider"
          @onSymbol="onSymbolClick"
          @onSymbolSwitch="onSymbolSwitch"
          @onSignalSwitch="onSignalSwitch"
          @onSortingSwitch="onSortingSwitch"
          @onFilterClick="onFilterClick"
          @onBudgetChange="onBudgetChange"
          @onSignalConfig="onSignalConfig"
        />
        <v-progress-linear v-else :indeterminate="true" color="success" height="4px"/>
      </v-col>
      <v-col cols="12" class="positions">
        <Position
            @click="tab = position.symbol"
            class="position-item"
            v-for="(position, index) in positionsSorted"
            :key="`pos-${index}-${position.symbol}`"
            :position="position"
            :show-bottom-btn="true"
            :show-used-budget="true"
            :show-profit-percent="true"
            :bottom-btn-text="$t('dashboard.positions_bottom_btn_text')"
            :show-buy-switch="true"
            :show-cancel-manual="true"
            :show-condition-btn="true"
            :show-avg-price="true"
            :show-price-change-speed="true"
            :show-pivots="true"
            :show-sentiment="true"
            @onCancelManual="onCancelManual(position.symbol)"
            @onSwitchEnabled="onSymbolSwitch({symbol: position.symbol})"
            @onBottomBtn="settingsModal(position)"
            @onFilterClick="onFilterPosClick"
        />
      </v-col>
      <v-col cols="12" class="trade-stack my-2">
        <v-card max-width="350px" class="ml-0 bg-success">
          <v-card-title class="mb-0 pb-0">{{$t('dashboard.target_profit')}}</v-card-title>
          <v-card-subtitle class="mt-0 pb-2"><b>+{{ getTargetProfit }}</b> <small>USDT</small></v-card-subtitle>
        </v-card>
      </v-col>
    </v-row>
    <div v-if="!!chartData" class="charts">
      <v-tabs
          v-model="tab"
          bg-color="success"
      >
        <v-tab v-for="(chart, index) in chartData" :value="chart.symbol" :key="`tab-${index}`">{{ chart.symbol }}</v-tab>
      </v-tabs>
      <div v-for="(chart, index) in chartData" :key="`chart-${index}`">
        <div v-if="chart.symbol === tab">
          <TradingView
            :symbol="chart.symbol"
          />
          <Financial
              :symbol="chart.symbol"
              :quantity="chart.quantity"
              :candles="chart.candles"
              :orders-sell="chart.orderSell"
              :orders-buy="chart.orderBuy"
              :predicted="chart.predicted"
              :btc-index="chart.btcIndex"
              :eth-index="chart.ethIndex"
              :order-opened="chart.orderBuyOpened"
              :order-buy-pending="chart.orderBuyPending"
              :order-sell-pending="chart.orderSellPending"
              :iceberg-buy="chart.icebergBuy"
              :iceberg-sell="chart.icebergSell"
              :current-price="chart.lastPrice"
              :profit-percent="chart.profit"
              :update-subject="subjectMap[chart.symbol]"
              @onOrder="sendManualOrder"
          />
          <h4 class="text-center" style="background-color: #efefef;">{{$t('dashboard.text_price_change_speed')}}</h4>
          <PriceSpeedChart
              :symbol="chart.symbol"
              :avg-price-points="chart.avgPriceChangeSpeed"
              :min-price-points="chart.minPriceChangeSpeed"
              :max-price-points="chart.maxPriceChangeSpeed"
              :update-subject="subjectMap[chart.symbol]"
          />
          <h4 class="text-center" style="background-color: #efefef;">{{$t('dashboard.text_trade_quantity')}}</h4>
          <TradeVolumeChart
              :symbol="chart.symbol"
              :sell-volume-points="chart.sellTradeVolume"
              :buy-volume-points="chart.buyTradeVolume"
              :iceberg-buy-qty="chart.icebergBuyQty"
              :iceberg-sell-qty="chart.icebergSellQty"
              :cummulative-volume-points="chart.cummulativeTradeQty"
              :update-subject="subjectMap[chart.symbol]"
          />
          <div v-if="chart.marketCapPrice.filter((x) => x.y > 0).length > 10">
            <h4 class="text-center" style="background-color: #efefef;">{{$t('dashboard.text_market_price')}}</h4>
            <MarketCapPriceChart
                :symbol="chart.symbol"
                :market-cap-price-points="chart.marketCapPrice"
                :update-subject="subjectMap[chart.symbol]"
            />
          </div>
          <div v-if="chart.marketCapValue.filter((x) => x.y > 0).length > 10">
            <h4 class="text-center" style="background-color: #efefef;">{{$t('dashboard.text_market_capital')}}</h4>
            <MarketCapChart
                :symbol="chart.symbol"
                :market-cap-points="chart.marketCapValue"
                :update-subject="subjectMap[chart.symbol]"
            />
          </div>
        </div>
      </div>
    </div>
    <v-row class="last-trades mt-2" v-if="!!lastOrders">
      <v-col cols="12">
        <h2>{{$t('dashboard.completed_order_list')}}</h2>
        <TradeTable :last-orders="lastOrders"/>
      </v-col>
    </v-row>
    <v-row>
      <v-col cols="12" class="swap-action-list">
        <Swap
            v-for="(swap, index) in swaps" :key="`swap-${index}`"
            :swap="swap"
        />
      </v-col>
    </v-row>
    <v-dialog
        persistent
        v-model="filtersDialog.flag"
        max-width="650px"
        min-width="380px"
        z-index="9999"
    >
      <div class="position-relative" style="width: 100%; z-index: 999;">
        <v-btn icon variant="text" style="top: 0;right:0;position: absolute" @click="closeFilters">
          <v-icon icon="mdi-close"/>
        </v-btn>
      </div>
      <v-card v-if="filtersDialog.flag">
        <v-form @submit.prevent="saveFilters" ref="form5">
          <v-card-title class="mb-3">{{ filtersDialog.operation.toUpperCase() }} <small>{{ filtersDialog.symbol }}</small> {{$t('dashboard.save_filters_only_if')}}</v-card-title>
          <v-card-text>
            <TradeCondition
                v-for="(tradeFilter, tradeFilterIndex) in filtersDialog.filters"
                :key="`trade-filter-${tradeFilterIndex}`"
                class="mb-0 position-relative"
                :trade-filter="tradeFilter"
                :option-number="tradeFilterIndex+1"
                :can-be-multi="true"
                :symbols="symbolList"
                @onAddChildren="tradeFilter.children.push({symbol: filtersDialog.symbol, parameter: ($event.parameter || 'price'), condition: null, value: null, type: 'and', mode: 'single' })"
                @onClickRemove="filtersDialog.filters.splice(tradeFilterIndex, 1)"
            />

            <v-btn size="x-small" class="mb-2 mt-2" color="primary" @click="filtersDialog.filters.push({symbol: filtersDialog.symbol, parameter: 'price', condition: null, value: '0.00', type: 'or', mode: 'single', children: [] })">{{$t('dashboard.save_filters_condition')}}</v-btn>
            <v-divider class="mb-2"/>
          </v-card-text>

          <v-card-actions class="justify-end">
            <v-btn variant="flat" color="default" class="action" @click="closeFilters">{{$t('dashboard.common_cancel_btn')}}</v-btn>
            <v-btn variant="flat" color="success" type="submit" class="action w-33">{{$t('dashboard.common_apply_btn')}}</v-btn>
          </v-card-actions>
        </v-form>
      </v-card>
    </v-dialog>
    <v-dialog
        persistent
        v-model="settingsDialog.flag"
        max-width="500px"
        min-width="380px"
        z-index="9999"
    >
      <div class="position-relative" style="width: 100%; z-index: 999;">
        <v-btn icon variant="text" style="top: 0;right:0;position: absolute" @click="closeSettings">
          <v-icon icon="mdi-close"/>
        </v-btn>
      </div>
      <v-card v-if="settingsDialog.flag">
        <v-card-title class="pb-0">{{$t('dashboard.setting_dialog_flag')}} <small>{{ settingsDialog.order.symbol }}</small></v-card-title>
        <v-tabs
            v-model="settingsDialog.tab"
            bg-color="default"
        >
          <v-tab v-for="(tab, index) in settingsDialogTabs" :value="index" :key="`tab-${index}`">{{ tab.name }}</v-tab>
        </v-tabs>
        <v-window v-model="settingsDialog.tab" class="mt-4" style="overflow-y: scroll !important;">
          <v-window-item value="0">
            <v-form @submit.prevent="saveProfitSettings" ref="form1" class="mt-2">
              <v-card-text>
                <v-row
                  v-for="(profitOption, profitOptionIndex) in settingsDialog.profitOptions"
                  class="mb-0 position-relative"
                  :key="`profit-${profitOptionIndex}`"
                  :style="{'background-color': !!settingsDialog.positionTime && (getNumericTime(profitOption) > settingsDialog.positionTime || profitOptionIndex === (settingsDialog.profitOptions.length - 1)) ? 'rgba(153,245,150,0.77)' : 'rgba(244,67,54,0.67)'}"
                >
                  <v-col cols="1" class="p-0">
                    <v-icon :icon="`mdi-numeric-${profitOptionIndex+1}-box`" color="primary"/>
                  </v-col>
                  <v-col cols="3" class="p-0 py-0">
                    <v-text-field
                        :label="$t('dashboard.settings_dialog_label_1')"
                        type="number"
                        v-model="profitOption.optionValue"
                        variant="underlined"
                        :rules="[rules.required, rules.positive]"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="3" class="p-0 py-0">
                    <v-select
                        :label="$t('dashboard.settings_dialog_label_2')"
                        type="text"
                        v-model="profitOption.optionUnit"
                        :rules="[rules.required]"
                        :items="profitPeriodLabels"
                        variant="underlined"
                    ></v-select>
                  </v-col>
                  <v-col cols="4" class="p-0 py-0">
                    <v-text-field
                        :label="$t('dashboard.settings_dialog_label_3')"
                        type="number"
                        v-model="profitOption.optionPercent"
                        variant="underlined"
                        :rules="[rules.required, rules.positive, rules.min05]"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="1" class="p-0">
                    <v-icon icon="mdi-close" color="red" style="margin-left: -10px" @click="settingsDialog.profitOptions.splice(profitOptionIndex, 1)"/>
                  </v-col>
                </v-row>

                <v-btn size="x-small" class="mb-2" color="primary" @click="settingsDialog.profitOptions.push({optionUnit: 'h', optionValue: 1, optionPercent: 0.5, isTriggerOption: false})">{{$t('dashboard.settings_dialog_option_btn')}}</v-btn>
                <v-divider class="mb-2"/>
              </v-card-text>

              <v-card-actions class="justify-end">
                <v-btn variant="flat" color="default" class="action" @click="closeSettings">{{$t('dashboard.common_cancel_btn')}}</v-btn>
                <v-btn variant="flat" color="success" type="submit" class="action w-33">{{$t('dashboard.common_apply_btn')}}</v-btn>
              </v-card-actions>
            </v-form>
          </v-window-item>
          <v-window-item value="1">
            <v-form @submit.prevent="saveChargeSettings" ref="form" class="mt-2">
              <v-card-text>
                <v-row
                  v-for="(chargeOption, optionIndex) in settingsDialog.chargeOptions"
                  class="mb-0"
                  :key="`charge-${optionIndex}`"

                  :style="{'background-color': !!settingsDialog.positionTime && getExtraBudgetSum(optionIndex) > Math.round(settingsDialog.order.usedExtraBudget) ? 'rgba(153,245,150,0.77)' : 'rgba(244,67,54,0.67)'}"
                >
                  <v-col cols="1" class="p-0">
                    <v-icon :icon="`mdi-numeric-${optionIndex+1}-box`" color="primary"/>
                  </v-col>
                  <v-col cols="5" class="p-0 py-0">
                    <v-text-field
                        :label="$t('dashboard.settings_dialog_label_4')"
                        type="number"
                        v-model="chargeOption.percent"
                        variant="underlined"
                        :rules="[rules.required, rules.negative]"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="5" class="p-0 py-0">
                    <v-text-field
                        :label="$t('dashboard.settings_dialog_label_5')"
                        type="number"
                        v-model="chargeOption.amountUsdt"
                        variant="underlined"
                        :rules="[rules.required, rules.min15]"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="1" class="p-0">
                    <v-icon icon="mdi-close" color="red" style="margin-left: -10px" @click="settingsDialog.chargeOptions.splice(optionIndex, 1)"/>
                  </v-col>
                </v-row>

                <v-btn size="x-small" class="mb-2" color="primary" @click="settingsDialog.chargeOptions.push({percent: 0.00, amountUsdt: 0.00})">{{$t('dashboard.settings_dialog_option_btn')}}</v-btn>
                <v-divider class="mb-2"/>
                <div v-if="settingsDialog.order.usedExtraBudget > 0">
                  <b>{{$t('dashboard.used_extra_budget')}}</b> {{ Math.round(settingsDialog.order.usedExtraBudget) }}
                  <small>USDT</small>
                </div>
              </v-card-text>

              <v-card-actions class="justify-end">
                <v-btn variant="flat" color="default" class="action" @click="closeSettings">{{$t('dashboard.common_cancel_btn')}}</v-btn>
                <v-btn variant="flat" color="success" type="submit" class="action w-33">{{$t('dashboard.common_apply_btn')}}</v-btn>
              </v-card-actions>
            </v-form>
          </v-window-item>
        </v-window>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script lang="ts">
import {Financial, Stack} from '#components';
import {Subject} from 'rxjs';
import Position from '~/components/Position.vue';
import {Alert} from '~/model/alert';
import {AlertEvent} from '~/model/alert-event';
import ProfitItem from '~/components/ProfitItem.vue';
import PriceSpeedChart from '~/components/PriceSpeedChart.vue';
import MarketCapChart from '~/components/MarketCapChart.vue';
import MarketCapPriceChart from '~/components/MarketCapPriceChart.vue';
import TradeCondition from '~/components/TradeCondition.vue';
import Swap from '~/components/ui/Swap.vue';

const uniqArray = (array) => {
  const map = {};
  for (let i = 0; i < array.length; ++i) {
    map[array[i].x] = array[i];
  }

  return Object.keys(map).map((i) => {
    return map[i];
  })
};

const parseChart = (charts: Array<any>) => {
  const chartMap: any = {};
  for (let i = 0; i < charts.length; i++) {
    let lastPrice = 0.00;
    let profit = 0.00;
    let openPrice = 0.00;

    const chart = charts[i];
    if (Object.keys(chart).length === 0) {
      continue
    }
    Object.keys(chart).map((key: string) => {
      const symbol: any = Object.keys(chart)[0].split('-').pop();

      if (!chartMap[symbol]) {
        chartMap[symbol] = {
          candles: [],
          predicted: [],
          btcIndex: [],
          ethIndex: [],
          orderSell: [],
          orderBuy: [],
          orderSellPending: 0,
          orderBuyPending: 0,
          orderBuyOpened: 0,
          profit: 0,
          openPrice: 0,
          lastPrice: 0,
          avgPriceChangeSpeed: [],
          minPriceChangeSpeed: [],
          maxPriceChangeSpeed: [],
          sellTradeVolume: [],
          buyTradeVolume: [],
          cummulativeTradeQty: [],
          marketCapValue: [],
          marketCapPrice: [],
          icebergBuy: [],
          icebergSell: [],
          icebergBuyQty: [],
          icebergSellQty: [],
        }
      }

      if (key.includes('predict-')) {
        chartMap[symbol].predicted = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y > 0 ? point.y : null}
        });
        return;
      }

      if (key.includes('capitalization-value-')) {
        chartMap[symbol].marketCapValue = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y > 0 ? point.y : null}
        });
        return;
      }

      if (key.includes('iceberg-price-buy-value-')) {
        chartMap[symbol].icebergBuy = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y > 0 ? point.y : null}
        });
        return;
      }

      if (key.includes('iceberg-qty-buy-value-')) {
        chartMap[symbol].icebergBuyQty = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y > 0 ? point.y : null}
        });
        return;
      }

      if (key.includes('iceberg-qty-sell-value-')) {
        chartMap[symbol].icebergSellQty = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y > 0 ? point.y : null}
        });
        return;
      }

      if (key.includes('iceberg-price-sell-value-')) {
        chartMap[symbol].icebergSell = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y > 0 ? point.y : null}
        });
        return;
      }

      if (key.includes('cummulative-trade-qty-')) {
        chartMap[symbol].cummulativeTradeQty = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y}
        });
        return;
      }

      if (key.includes('capitalization-price-')) {
        chartMap[symbol].marketCapPrice = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y > 0 ? point.y : null}
        });
        return;
      }

      if (key.includes('avg-change-speed-')) {
        chartMap[symbol].avgPriceChangeSpeed = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y}
        });
        return;
      }

      if (key.includes('trade-volume-buy-')) {
        chartMap[symbol].buyTradeVolume = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y}
        });
        return;
      }

      if (key.includes('trade-volume-sell-')) {
        chartMap[symbol].sellTradeVolume = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y}
        });
        return;
      }

      if (key.includes('min-change-speed-')) {
        chartMap[symbol].minPriceChangeSpeed = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y}
        });
        return;
      }

      if (key.includes('max-change-speed-')) {
        chartMap[symbol].maxPriceChangeSpeed = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y}
        });
        return;
      }

      if (key.includes('interpolation-btc-')) {
        chartMap[symbol].btcIndex = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y > 0 ? point.y : null}
        });
        return;
      }

      if (key.includes('interpolation-eth-')) {
        chartMap[symbol].ethIndex = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {x: date.valueOf(), y: point.y > 0 ? point.y : null}
        });
        return;
      }

      if (key.includes('kline-')) {
        chartMap[symbol].candles = chart[key].map((point: any) => {
          if (point['h']) {
            const date = new Date(point.x);
            date.setSeconds(0);
            date.setMilliseconds(0);

            return {
              x: date.valueOf(),
              o: point.o,
              c: point.c,
              l: point.l,
              h: point.h,
            };
          }

          return {...point, y: point.y > 0 ? point.y : null}
        });
        const lastValue: any = [...new Set(chart[key])].pop()
        lastPrice = lastValue.c;
        return;
      }

      if (key.includes('order-buy-opened')) {
        chartMap[symbol].orderBuyOpened = Math.max(...[...new Set(chart[key])].map((point: any) => {
          return Number(point.y);
        }));
        openPrice = Math.max(...[...new Set(chart[key])].map((x: any) => x.y))
        return;
      }

      if (key.includes('order-buy-pending')) {
        chartMap[symbol].orderBuyPending = Math.max(...[...new Set(chart[key])].map((point: any) => {
          return Number(point.y);
        }));
        return;
      }

      if (key.includes('order-sell-pending')) {
        chartMap[symbol].orderSellPending = Math.max(...[...new Set(chart[key])].map((point: any) => {
          return Number(point.y);
        }));
        return;
      }

      if (key.includes('order-sell')) {
        chartMap[symbol].orderSell = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {
            x: date.valueOf(),
            y: point.y > 0 ? point.y : null
          };
        });
        return;
      }

      if (key.includes('order-buy')) {
        chartMap[symbol].orderBuy = chart[key].map((point: any) => {
          const date = new Date(point.x);
          date.setSeconds(0);
          date.setMilliseconds(0);

          return {
            x: date.valueOf(),
            y: point.y > 0 ? point.y : null
          };
        });
        return;
      }
    });

    if (openPrice > 0) {
      const diff = lastPrice - openPrice;
      profit = Math.round(diff * 100 / openPrice * 100)/100
    }

    const symbol: any = Object.keys(chart)[0].split('-').pop();
    chartMap[symbol]['profit'] = profit;
    chartMap[symbol]['openPrice'] = openPrice;
    chartMap[symbol]['lastPrice'] = lastPrice;
  }

  return chartMap;
};

export default defineNuxtComponent({
  name: "[id].vue",
  components: {
    Swap,
    TradeCondition,
    MarketCapPriceChart, MarketCapChart, PriceSpeedChart, ProfitItem, Position, Financial, Stack},
  async asyncData(ctx: any) {
    const authToken = ctx.$services.authService.getToken();
    if (!authToken) {
      await ctx.$services.routerService.navigate('/');
      return;
    }

    const route = useRoute();
    const {t} = useI18n();

    const [chartResponse, lastOrders, profits, currentBot, symbolList] = await Promise.all([
      ctx.$services.httpClient.secureGetServer(`/v1/dashboard/${route.params.id}/chart`),
      ctx.$services.httpClient.secureGetServer(`/v1/dashboard/${route.params.id}/trades`),
      ctx.$services.httpClient.secureGetServer(`/v1/dashboard/${route.params.id}/profit?period=month`),
      ctx.$services.httpClient.secureGetServer(`/v1/cryptobot/${route.params.id}`),
      ctx.$services.httpClient.secureGetServer(`/v1/cryptobot/${route.params.id}/symbol/list`),
    ]);

    if (!chartResponse) {
      // todo: 404 error...
      await ctx.$services.routerService.navigate('/');
      return;
    }

    useHead({
      title: t('dashboard.header').replace('[route.params.id]', route.params.id),
    })

    let chartData = [];
    if (chartResponse) {
      chartData = parseChart(chartResponse)
    }

    return {
      availableSymbols: symbolList.map((x) => x.symbol),
      symbolList: symbolList.map((x) => x.symbol),
      stack: [],
      profits: (profits || []),
      chart: chartData,
      lastOrders: lastOrders,
      positions: [],
      tab: chartData.length > 0 ? chartData[0].symbol : '',
      botId: route.params.id,
      currentBot,
      swaps: [],
    };
  },
  mounted() {
    setTimeout(() => {
      this.refreshPositions();
      this.refreshStack();
      this.refreshAvailableSymbols();
      this.refreshSwaps();
    })

    const route = useRoute();

    this.updateSubscription = setInterval(() => {
      if (!route.params.id) {
        return;
      }

      this.refreshChart();
      this.refreshTrades();
      this.refreshProfit();
      this.refreshStack();
      this.refreshPositions();
      this.refreshSwaps();
    }, 10000);
  },
  unmounted() {
    if (this.updateSubscription) {
      clearInterval(this.updateSubscription);
    }
  },
  data(): any {
    let subjectMap = {}

    Object.keys(this.chart).forEach((prop: string) => {
      const symbol = prop.split('-').pop();
      subjectMap[symbol] = new Subject();
    });

    return {
      stackLoadingMap: {},
      periodOptions: [
        {title: this.$t('dashboard.data_period_title_1'), value: 'month'},
        {title: this.$t('dashboard.data_period_title_2'), value: 'week'},
        {title: this.$t('dashboard.data_period_title_3'), value: 'day'},
      ],
      profitPeriod: 'month',
      subjectMap,
      settingsDialogTabs: [
        {
          name: 'Profit',
          value: 0,
        },
        {
          name: 'Extra Charge',
          value: 1,
        }
      ],
      updateSubscription: null,
      settingsDialog: {
        flag: false,
        order: null,
        chargeOptions: [],
        profitOptions: [],
        tab: 1,
        positionTime: 0,
      },
      filtersDialog: {
        operation: '',
        flag: false,
        symbol: '',
        filters: [],
      },
      rules: {
        required: (value: any) => !!value || this.$t('dashboard.data_rules_required'),
        negative: (value: any) => value < 0.00 || this.$t('dashboard.data_rules_negative'),
        min15: (value: any) => value >= 15 || this.$t('dashboard.data_rules_min15'),
        positive: (value: any) => value > 0.00 || this.$t('dashboard.data_rules_positive'),
        min05: (value: any) => value >= 0.5 || this.$t('dashboard.data_rules_min05'),
      },
      availableSymbolList: [],
      quickSymbolLoading: {},
      dataLoading: {
        deploy: false,
        position: false,
        stack: false,
        report: false,
        swaps: false,
        chart: false,
        trades: false,
      },
    };
  },
  computed: {
    getTargetProfit: {
      get() {
        const profits = this.positionsSorted.map((x) => {
          return (x.sellPrice * x.order.executedQuantity) - (x.order.price * x.order.executedQuantity);
        });

        let sum = 0.00;

        profits.forEach((p) => {
          sum += Number(p);
        })

        return sum.toFixed(2);
      },
    },
    profitPeriodLabels: {
      get() {
        return [
          {
            title: this.$t('periods.minute'),
            value: 'i',
          },
          {
            title: this.$t('periods.hour'),
            value: 'h',
          },
          {
            title: this.$t('periods.day'),
            value: 'd',
          },
          {
            title: this.$t('periods.month'),
            value: 'm',
          },
        ];
      },
    },
    portfolioReport: {
      get() {
        let profit = 0.00
        let tradeVolume = 0.00
        let trades = 0;
        let backgroundColor = '#1DE9B6FF';
        let color = '#000000';

        this.positionsSorted.forEach((position) => {
          profit += position.profit;
          tradeVolume += (position.order.price * position.order.executedQuantity);
          trades++;
        });

        if (profit < 0) {
          backgroundColor = 'rgb(189,47,38)';
          color = '#FFFFFF'
        }

        const avgPercent = Number((profit * 100 / tradeVolume).toFixed(2));
        const {t} = useI18n();

        return {
          title: t('dashboard.computed_title'),
          profit,
          tradeVolume: Number(Number(tradeVolume).toFixed(2)),
          trades,
          backgroundColor,
          color,
          avgPercent,
        };
      },
    },
    positionsSorted: {
      get() {
        return this.positions.sort((a, b) => a.percent < b.percent ? 1 : -1);
      },
    },
    chartData: {
      get(): Array<any> {
        return Object.keys(this.chart).map((objectKey: string) => {
          const object = this.chart[objectKey];
          object['symbol'] = objectKey;
          object['quantity'] = 0.00;

          if (this.positions) {
            this.positions.forEach((position: any) => {
              if (objectKey === position.symbol) {
                object['quantity'] = position.order.executedQuantity;
              }
            });
          }

          return object;
        });
      },
    },
  },
  methods: {
    ...parseChart,
    deployBot() {
      this.dataLoading.deploy = true;
      const route = useRoute();
      this.$services.httpClient.securePutClient(`/v1/cryptobot/${route.params.id}/deploy`, {}).then(() => {
        this.refreshStack();
      }).finally(() => this.dataLoading.deploy = false);
    },
    addQuickSymbol(symbol: any) {
      this.quickSymbolLoading[symbol.symbol] = true;
      const route = useRoute();
      this.$services.httpClient.securePostClient(`/v1/dashboard/${route.params.id}/quick/symbol`, {
        exchangeSymbol: symbol.id,
        restartBot: 0,
      }).then(() => {
        let message = `Symbol ${symbol.symbol} has been added.`;
        let alertType = Alert.TYPE_SUCCESS;

        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(message, alertType),
            4000
          )
        );
      }).finally(() => {
        this.refreshAvailableSymbols();
        this.refreshStack();
        this.quickSymbolLoading[symbol.symbol] = false;
      });
    },
    refreshChart() {
      if (this.dataLoading.chart) {
        return;
      }

      this.dataLoading.chart = true;
      const route = useRoute();
      this.$services.httpClient.secureGetClient(`/v1/dashboard/${route.params.id}/chart`).then((chart: any) => {
        this.chart = parseChart(chart);
        Object.keys(this.chart).forEach((symbol) => {
          this.subjectMap[symbol].next();
        });
      }).finally(() => this.dataLoading.chart = false);
    },
    refreshTrades() {
      if (this.dataLoading.trades) {
        return;
      }

      this.dataLoading.trades = true;
      const route = useRoute();
      this.$services.httpClient.secureGetClient(`/v1/dashboard/${route.params.id}/trades`).then((lastOrders: any) => {
        this.lastOrders = lastOrders;
      }).finally(() => this.dataLoading.trades = false);
    },
    refreshProfit() {
      if (this.dataLoading.report) {
        return;
      }

      this.dataLoading.report = true;
      const route = useRoute();
      if (route.params.id) {
        this.$services.httpClient.secureGetClient(`/v1/dashboard/${route.params.id}/profit?period=${this.profitPeriod}`).then((profits) => {
          this.profits = profits
        }).finally(() => this.dataLoading.report = false);
      }
    },
    refreshAvailableSymbols() {
      const route = useRoute();
      if (route.params.id) {
        this.$services.httpClient.secureGetClient(`/v1/dashboard/${route.params.id}/symbol/available`).then((availableSymbolList: any) => {
          this.availableSymbolList = availableSymbolList
        });
      }
    },
    refreshStack() {
      if (this.dataLoading.stack) {
        return;
      }

      this.dataLoading.stack = true;
      const route = useRoute();
      if (route.params.id) {
        this.$services.httpClient.secureGetClient(`/v1/dashboard/${route.params.id}/stack/v2`).then((stack: any) => {
          this.stack = stack;
        }).finally(() => this.dataLoading.stack = false);
      }
    },
    refreshPositions() {
      if (this.dataLoading.position) {
        return;
      }

      this.dataLoading.position = true;
      const route = useRoute();
      if (route.params.id) {
        this.$services.httpClient.secureGetServer(`/v1/dashboard/${route.params.id}/positions`).then((positions: any) => {
          this.positions = positions;
        }).finally(() => this.dataLoading.position = false);
      }
    },
    refreshSwaps() {
      if (this.dataLoading.swaps) {
        return;
      }

      this.dataLoading.swaps = true;
      const route = useRoute();
      if (route.params.id) {
        this.$services.httpClient.secureGetServer(`/v1/dashboard/${route.params.id}/swap/list`).then((swaps: any) => {
          this.swaps = swaps;
        }).finally(() => this.dataLoading.swaps = false);
      }
    },
    onSymbolClick(symbol) {
      this.tab = symbol;
    },
    onFilterPosClick({position, operation}) {
      this.onFilterClick({stackItem: position, operation});
    },
    onSignalConfig({signalConfig, stackItem}) {
      const route = useRoute();
      this.$services.httpClient.securePatchClient(`/v1/dashboard/${route.params.id}/${stackItem.symbol}/update`, {
        signalConfig,
      }).then(() => {
        let message = `Signal config for '${stackItem.symbol}' is updated`;
        let alertType = Alert.TYPE_SUCCESS;

        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(message, alertType),
            4000
          )
        );
      }).finally(() => {
        this.refreshStack();
      });
    },
    onBudgetChange(stackItem: any) {
      const route = useRoute();
      if (!this.stackLoadingMap[stackItem.symbol]) {
        this.stackLoadingMap[stackItem.symbol] = {};
      }
      this.stackLoadingMap[stackItem.symbol].budget = true;
      this.$services.httpClient.securePatchClient(`/v1/dashboard/${route.params.id}/${stackItem.symbol}/update`, {
        usdtLimit: Math.max(Number(stackItem.budgetUsdt), 15),
      }).then((data) => {
        let message = `Budget is changed for ${data.symbol} to: ${data.usdtLimit} USDT`;
        let alertType = Alert.TYPE_SUCCESS;

        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(message, alertType),
            4000
          )
        );
      }).finally(() => {
        this.refreshStack();
        this.stackLoadingMap[stackItem.symbol].budget = false;
      });
    },
    onFilterClick({stackItem, operation}) {
      const options = {
        'buy': 'tradeFiltersBuy',
        'sell': 'tradeFiltersSell',
        'avg': 'tradeFiltersExtraCharge',
      }

      this.filtersDialog = {
        operation: operation,
        flag: true,
        symbol: stackItem.symbol,
        filters: (stackItem[options[operation]] || []).map((x) => {
          if (x.parameter === 'has_signal') {
            x.value = (x.value.toString() === "true" || x.value.toString() === "1");
          }

          return {
            ...x,
            children: (x.children || []).map((y) => {
              if (y.parameter === 'has_signal') {
                y.value = (y.value.toString() === "true" || y.value.toString() === "1");
              }

              return {...y, mode: 'single'};
            }),
            mode: (x.children || []).length > 0 ? 'multi' : 'single',
          }
        }),
      };
    },
    closeFilters() {
      this.filtersDialog = {
        operation: '',
        flag: false,
        symbol: '',
        filters: [],
      };
    },
    saveFilters() {
      this.$refs.form5.validate().then(({valid}: any) => {
        if (!valid) {
          return;
        }

        this.$services.httpClient.securePutClient(`/v1/cryptobot/${this.botId}/${this.filtersDialog.operation}/conditions`, {
          symbol: this.filtersDialog.symbol,
          conditions: this.filtersDialog.filters.map((item) => {
            if (item.mode === 'multi') {
              return {
                type: item.type,
                children: (item.children || []).map((child) => {
                  return {
                    symbol: child.symbol,
                    parameter: child.parameter,
                    condition: child.condition,
                    value: child.value.toString(),
                    type: child.type,
                  }
                }),
              }
            }

            return {
              symbol: item.symbol,
              parameter: item.parameter,
              condition: item.condition,
              value: item.value.toString(),
              type: item.type,
            }
          }),
        }).then(() => {
          let message = `${this.$t('dashboard.methods_save_filters')}`;
          let alertType = Alert.TYPE_SUCCESS;

          this.$services.eventManager.alert(
            new AlertEvent(
              new Alert(message, alertType),
              4000
            )
          );
        }).finally(() => {
          this.refreshPositions();
          this.refreshStack();
        });
      });
    },
    onSortingSwitch(sorting: string) {
      this.$services.httpClient.securePutClient(`/v1/dashboard/${this.botId}/stack/${sorting}/sort`).then(() => {
        let message = `${this.$t('dashboard.methods_sorting')}`;
        let alertType = Alert.TYPE_SUCCESS;

        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(message, alertType),
            4000
          )
        );
      }).finally(() => this.refreshStack());
    },
    onSymbolSwitch(stackItem: any) {
      this.$services.httpClient.securePutClient(`/v1/dashboard/${this.botId}/stack/${stackItem.symbol}/switch`).then((data: any) => {
        let message = `${this.$t('dashboard.methods_symbol_disabled').replace('[stackItem.symbol]', stackItem.symbol)}`;
        if (data.isEnabled) {
          message = `${this.$t('dashboard.methods_symbol_enabled').replace('[stackItem.symbol]', stackItem.symbol)}`;
        }

        let alertType = Alert.TYPE_SUCCESS;
        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(message, alertType),
            4000
          )
        );
      }).finally(() => this.refreshStack());
    },
    onSignalSwitch(stackItem: any) {
      this.$services.httpClient.securePutClient(`/v1/dashboard/${this.botId}/stack/${stackItem.symbol}/signal-switch`).then((data: any) => {
        let message = `${this.$t('dashboard.methods_signal_disabled').replace('[stackItem.symbol]', stackItem.symbol)}`;
        if (data.signalTrading) {
          message = `${this.$t('dashboard.methods_signal_enabled').replace('[stackItem.symbol]', stackItem.symbol)}`;
        }

        let alertType = Alert.TYPE_SUCCESS;
        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(message, alertType),
            4000
          )
        );
      }).finally(() => this.refreshStack());
    },
    onCancelManual(symbol: string) {
      this.$services.httpClient.secureDeleteClient(`/v1/cryptobot/${this.botId}/order/${symbol}`).then(() => {
        let message = `${this.$t('dashboard.methods_cancel_manual').replace('[symbol]', symbol)}`;

        let alertType = Alert.TYPE_SUCCESS;
        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(message, alertType),
            4000
          )
        );
      });
    },
    sendManualOrder(order) {
      this.$services.httpClient.securePostClient(`/v1/cryptobot/${this.botId}/order`, order).then(() => {
        let message = `${this.$t('dashboard.methods_send_manual')}`;
        let alertType = Alert.TYPE_SUCCESS;
        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(message, alertType),
            4000
          )
        );
      }).finally(() => this.refreshPositions());
    },
    settingsModal(position: any) {
      this.settingsDialog.flag = true;
      this.settingsDialog.order = position.order;
      this.settingsDialog.tab = 0;
      this.settingsDialog.positionTime = position.positionTime;
      this.settingsDialog.chargeOptions = (position.order.extraChargeOptions || []);
      this.settingsDialog.profitOptions = (position.order.profitOptions || []).map((option) => {
        return {...option, isTriggerOption: Boolean(option.isTriggerOption)};
      });
      if (this.settingsDialog.chargeOptions.length === 0) {
        this.settingsDialog.chargeOptions = [
          {
            percent: 0.00,
            amountUsdt: 0.00,
          }
        ];
      }
      if (this.settingsDialog.profitOptions.length === 0) {
        this.settingsDialog.profitOptions = [
          {
            optionValue: 1,
            optionUnit: 'h',
            optionPercent: Number(2.15),
            isTriggerOption: true,
          }
        ];
      }
    },
    closeSettings() {
      this.settingsDialog.flag = false;
      this.settingsDialog.order = null;
      this.settingsDialog.chargeOptions = [];
      this.settingsDialog.profitOptions = [];
      this.settingsDialog.tab = 0;
      this.settingsDialog.positionTime = 0;
    },
    getNumericTime(option: any) {
      let value = 0;

      if ('i' === option.optionUnit) {
        value = option.optionValue * 60;
      }

      if ('h' === option.optionUnit) {
        value = option.optionValue * 3600;
      }

      if ('d' === option.optionUnit) {
        value = option.optionValue * 3600 * 24;
      }

      if ('m' === option.optionUnit) {
        value = option.optionValue * 3600 * 24 * 30;
      }

      return value;
    },
    getExtraBudgetSum(search) {
      let budget = 0;

      this.settingsDialog.chargeOptions.forEach((option, index) => {
        if (index <= search) {
          budget += Number(option.amountUsdt);
        }
      })

      return budget;
    },
    saveProfitSettings() {
      this.$refs.form1.validate().then(({valid}: any) => {
        this.settingsDialog.profitOptions.sort((a, b) => this.getNumericTime(a) > this.getNumericTime(b) ? 1 : -1)

        if (!valid) {
          return;
        }

        let index = 0;
        const hasTrigger = this.settingsDialog.profitOptions.filter((x) => x.isTriggerOption).length > 0;
        this.$services.httpClient.securePutClient(`/v1/cryptobot/${this.botId}/multi/profit`, {
          orderId: this.settingsDialog.order.id,
          profitOptions: this.settingsDialog.profitOptions.map((option) => {
            const indexVal = index;
            index++;

            let isTriggerOption = option.isTriggerOption;
            if (!hasTrigger) {
              isTriggerOption = indexVal === 0;
            }

            return {
              index: indexVal,
              optionValue: Number(option.optionValue),
              optionUnit: option.optionUnit,
              optionPercent: Number(option.optionPercent),
              isTriggerOption: isTriggerOption,
            };
          }),
        }).then(() => {
          let message = `${this.$t('dashboard.methods_save_profit')}`;
          let alertType = Alert.TYPE_SUCCESS;
          this.$services.eventManager.alert(
            new AlertEvent(
              new Alert(message, alertType),
              4000
            )
          );
        }).finally(() => this.refreshPositions());
      });
    },
    saveChargeSettings() {
      this.$refs.form.validate().then(({valid}: any) => {
        this.settingsDialog.chargeOptions.sort((a, b) => Number(a.percent) < Number(b.percent) ? 1 : -1)

        if (!valid) {
          return;
        }

        let index = 0;

        this.$services.httpClient.securePutClient(`/v1/cryptobot/${this.botId}/multi/charge`, {
          orderId: this.settingsDialog.order.id,
          extraChargeOptions: this.settingsDialog.chargeOptions.map((value) => {
            const indexVal = index;
            index++;

            return {
              amountUsdt: Number(value.amountUsdt),
              percent: Number(value.percent),
              index: indexVal,
            };
          }),
        }).then(() => {
          let message = `${this.$t('dashboard.methods_save_charge')}`;
          let alertType = Alert.TYPE_SUCCESS;
          this.$services.eventManager.alert(
            new AlertEvent(
              new Alert(message, alertType),
              4000
            )
          );
        }).finally(() => this.refreshPositions());
      });
    },
  },
})
</script>

<style scoped>
.financial-block {
  margin-top: 10px;
}
.last-trades h2 {
  margin-left: 10px;
}
.divider {
  margin-top: 20px;
  margin-bottom: 20px;
}
.positions {
  display: inline-table;
  position: relative;
  padding: 0;
}
.trade-stack {
  padding: 0;
  margin-bottom: 4px;
  height: auto;
}
.charts {
  margin: 0 -16px;
}
.profit-scroll-horizontal {
  overflow-x: scroll;
  width: 100%;
  display: flex;
  flex-direction: row;
}
.profit-scroll-horizontal::-webkit-scrollbar {
  display: none;
}
.profit-params-block small {
  font-size: 10px;
}
.profit-period-selector {
  margin-bottom: 10px;
}
.available-symbol-list {
  overflow-x: scroll;
  width: 100%;
  display: flex;
  flex-direction: row;
  gap: 4px
}
.available-symbol-list .symbol-list-item {
  background: rgba(29,233,182,0.58);
  padding: 10px;
  border-radius: 4px;
  vertical-align: middle;
}
.symbol-list-item .symbol-title {
  font-weight: bold;
  margin-right: 5px;
}
.swap-action-list {
  padding-top: 25px;
  padding-bottom: 25px;
}
</style>