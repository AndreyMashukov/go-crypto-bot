<template>
  <div class="financial-background">
    <div class="market-control-block">
      <div class="market-control">
        <v-btn color="success" size="small" class="btn-ord" @click="makeOrder('buy')">{{$t('financial.buy_btn')}}</v-btn>
        <v-btn color="red" size="small" class="btn-ord" @click="makeOrder('sell')">{{$t('financial.sell_btn')}}</v-btn>
      </div>
    </div>

    <div class="canvas-block">
      <span class="profit">
        <v-chip v-if="profitPercent === 0" variant="flat" color="primary" size="x-small">0.00%</v-chip>
        <v-chip v-if="profitPercent > 0" variant="flat" color="success" size="x-small">+{{ profitPercent }}%</v-chip>
        <v-chip v-if="profitPercent < 0" variant="flat" color="red" size="x-small">-{{ profitPercent }}%</v-chip>
      </span>
      <h2>{{ symbol }}</h2>
      <span class="current-price"><small>{{$t('financial.current_price')}} </small>{{ currentPrice }}<small>USDT</small></span>
      <canvas :id="canvasId"></canvas>
    </div>
    <v-dialog
        v-model="orderDialog.flag"
        width="auto"
        persistent
    >
      <v-card>
        <v-card-title>{{ orderDialog.operation }} <small>{{ symbol }}</small></v-card-title>
        <v-card-subtitle v-if="orderDialog.opened > 0">{{ orderDialog.operation === 'SELL' ? 'OPEN' : 'PRICE' }}: {{ orderDialog.opened }} <small v-if="orderDialog.operation === 'SELL'">{{$t('financial.operation_card_subtitle')}} ~ {{  ((quantity * orderDialog.price) - (quantity * orderDialog.opened)).toFixed(2) }}$</small></v-card-subtitle>
        <v-card-text>
          <div class="price-input-block">
            <v-btn
              class="price-btn"
              variant="flat"
              color="primary"
              @click="orderDialog.price = Number((orderDialog.price - priceStep).toFixed(priceLength))"
            >-</v-btn>
            <v-text-field
              class="price-input-value"
              :label="$t('financial.price_label').replace('[orderDialog.operation]', orderDialog.operation)"
              type="number"
              v-model="orderDialog.price"
            ></v-text-field>
            <v-btn
              class="price-btn"
              variant="flat"
              color="primary"
              @click="orderDialog.price = Number((orderDialog.price + priceStep).toFixed(priceLength))"
            >+</v-btn>
          </div>
          <div class="price-options">
            <v-btn size="x-small" v-for="(option, index) in orderDialog.options" :key="index" color="primary" @click="orderDialog.price = option.value">
              {{ option.percent }}
            </v-btn>
          </div>
        </v-card-text>
        <v-card-actions class="order-actions">
          <v-btn variant="flat" color="default" class="action" @click="closeOrderDialog">{{$t('financial.cancel_btn')}}</v-btn>
          <v-btn variant="flat" :color="orderDialog.operation === 'SELL' ? 'red' : 'success'" class="action" @click="doOrder">{{ orderDialog.operation }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts">
export default {
  components: {},
  name: "Financial",
  props: {
    symbol: {
      type: String,
      default: () => '',
    },
    candles: {
      type: Array,
      default: () => [],
    },
    profitPercent: {
      type: Number,
      default: () => 0,
    },
    currentPrice: {
      type: Number,
      default: () => 0,
    },
    ordersSell: {
      type: Array,
      default: () => [],
    },
    quantity: {
      type: Number,
      default: () => 0,
    },
    orderSellPending: {
      type: Number,
      default: () => 0,
    },
    orderBuyPending: {
      type: Number,
      default: () => 0,
    },
    ordersBuy: {
      type: Array,
      default: () => [],
    },
    icebergBuy: {
      type: Array,
      default: () => [],
    },
    icebergSell: {
      type: Array,
      default: () => [],
    },
    predicted: {
      type: Array,
      default: () => [],
    },
    btcIndex: {
      type: Array,
      default: () => [],
    },
    ethIndex: {
      type: Array,
      default: () => [],
    },
    orderOpened: {
      type: Number,
      default: () => 0,
    },
    updateSubject: {
      type: Object,
      default: () => null,
    }
  },
  unmounted() {
    if (this.updateSubscription) {
      this.updateSubscription.unsubscribe();
    }
    if (this.chartRef) {
      this.chartRef.destroy();
    }
  },

  setup(props: any, {emit}: any) {
    function onOrder(data: any) {
      emit('onOrder', data);
    }

    return {
      onOrder,
    }
  },
  mounted() {
    var ctx = document.getElementById(this.canvasId).getContext('2d');

    const updateDatasets = () => {
      var datasets = [
        {
          type: 'candlestick',
          label: this.symbol,
          data: this.candles,
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                }
              }]
            },
          },
          order: 10,
        },
      ];

      if (this.ordersBuy.length > 0) {
        datasets.push({
          type: 'scatter',
          label: this.$t('financial.mounted_label_buy'),
          data: this.ordersBuy,
          elements: {
            point:{
              radius: 3,
            },
          },
          order: 0,
          borderWidth: 0.5,
          borderColor: '#6aff00',
          pointBackgroundColor: '#6aff00',
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                },
              }],
            },
          },
        });
      }

      if (this.ordersSell.length > 0) {
        datasets.push({
          type: 'scatter',
          label: this.$t('financial.mounted_label_sell'),
          data: this.ordersSell,
          elements: {
            point:{
              radius: 3,
            },
          },
          order: 0,
          borderWidth: 0.5,
          borderColor: '#ff6363',
          pointBackgroundColor: '#ff6363',
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                },
              }],
            },
          },
        });
      }

      if (this.orderOpened > 0) {
        datasets.push(      {
          type: 'line',
          label: this.$t('financial.mounted_label_current'),
          data: this.candles.map((point) => {
            return {
              x: point.x,
              y: this.orderOpened,
            };
          }),
          elements: {
            point:{
              radius: 0,
            }
          },
          order: 0,
          borderWidth: 1.5,
          borderColor: '#8dec4d',
          pointBackgroundColor: '#99f596',
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                },
              }]
            },
          },
        });
      }

      if (this.predicted.length > 0) {
        datasets.push(      {
          type: 'scatter',
          label: this.$t('financial.mounted_label_predict'),
          data: this.predicted,
          elements: {
            point:{
              radius: 2,
            }
          },
          order: 0,
          borderWidth: 0.5,
          borderColor: 'rgb(0,0,0)',
          pointBackgroundColor: 'rgb(253,241,78)',
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                },
              }]
            },
          },
        });
      }

      if (this.btcIndex.length > 0) {
        datasets.push(      {
          type: 'scatter',
          label: this.$t('financial.mounted_label_btc'),
          data: this.btcIndex,
          elements: {
            point:{
              radius: 2,
            }
          },
          order: 0,
          borderWidth: 0.5,
          borderColor: 'rgb(0,0,0)',
          pointBackgroundColor: 'rgb(128,62,243)',
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                },
              }]
            },
          },
        });
      }

      if (this.ethIndex.length > 0) {
        datasets.push(      {
          type: 'scatter',
          label: this.$t('financial.mounted_label_eth'),
          data: this.ethIndex,
          elements: {
            point:{
              radius: 2,
            }
          },
          order: 0,
          borderWidth: 0.5,
          borderColor: 'rgb(0,0,0)',
          pointBackgroundColor: 'rgb(255,130,8)',
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                },
              }]
            },
          },
        });
      }

      if (this.orderSellPending > 0) {
        datasets.push(      {
          type: 'line',
          label: this.$t('financial.mounted_label_limit_sell'),
          data: this.candles.map((point) => {
            return {
              x: point.x,
              y: this.orderSellPending,
            };
          }),
          elements: {
            point:{
              radius: 0,
            }
          },
          order: 0,
          borderWidth: 1.5,
          borderColor: '#980000',
          pointBackgroundColor: '#980000',
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                },
              }]
            },
          },
        });
      }

      if (this.orderBuyPending > 0) {
        datasets.push(      {
          type: 'line',
          label: this.$t('financial.mounted_label_limit_buy'),
          data: this.candles.map((point) => {
            return {
              x: point.x,
              y: this.orderBuyPending,
            };
          }),
          elements: {
            point:{
              radius: 0,
            }
          },
          order: 0,
          borderWidth: 1.5,
          borderColor: '#0b5700',
          pointBackgroundColor: '#0b5700',
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                },
              }]
            },
          },
        });
      }

      if (this.icebergBuy.length > 0) {
        datasets.push(      {
          type: 'line',
          label: 'Iceberg Buy',
          data: this.icebergBuy.map((point) => {
            return {
              x: point.x,
              y: point.y,
            };
          }),
          elements: {
            point:{
              radius: 1,
            }
          },
          order: 0,
          borderWidth: 1.5,
          borderColor: 'rgba(87,180,32,0.59)',
          pointBackgroundColor: 'rgba(87,180,32,0.59)',
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                },
              }]
            },
          },
        });
      }

      if (this.icebergSell.length > 0) {
        datasets.push(      {
          type: 'line',
          label: 'Iceberg Sell',
          data: this.icebergSell.map((point) => {
            return {
              x: point.x,
              y: point.y,
            };
          }),
          elements: {
            point:{
              radius: 1,
            }
          },
          order: 0,
          borderWidth: 1.5,
          borderColor: 'rgba(203,44,44,0.51)',
          pointBackgroundColor: 'rgba(203,44,44,0.51)',
          options: {
            responsive         : true,
            maintainAspectRatio: false,
            scales: {
              yAxes: [{
                ticks: {
                  beginAtZero: false,
                },
              }]
            },
          },
        });
      }

      return datasets;
    };

    const chartRef = new Chart(ctx, {
      type: 'candlestick',
      data: {
        datasets: updateDatasets(),
      },
      options: {
        animation: {
          duration: 0
        },
        plugins: {
          tooltip: {
            callbacks: {
              label: (ctx) => {
                if (ctx.formattedValue.includes('High') && ctx.formattedValue.includes('Low')) {
                  return ctx.formattedValue;
                }
                return (`${ctx.dataset.label}: ${ctx.parsed.y.toFixed(10)}`)
              }
            }
          }
        },
      }
    });

    if (this.updateSubject) {
      this.updateSubscription = this.updateSubject.subscribe(() => {
        const newDatasets = updateDatasets();
        for (let i = 0; i < chartRef.data.datasets.length; i++) {
          chartRef.data.datasets[i].data = newDatasets[i].data;
        }
        chartRef.update();
      });
    }

    this.chartRef = chartRef;
  },
  computed: {
    canvasId: {
      get() {
        return `canvas-${this.symbol}`;
      },
    },
  },
  data() {
    let priceLength = 8;
    if (this.currentPrice) {
      const priceSplit = this.currentPrice.toString().split('.');
      if (priceSplit.length > 1) {
        priceLength = priceSplit[1].length
      }
    }

    const priceStep = Number(Math.pow(10, -1 *priceLength).toFixed(priceLength))
    let openedOrderPrice = 0.00;

    if (this.orderOpened > 0) {
      openedOrderPrice = Number(this.orderOpened.toFixed(priceLength));
    }

    return {
      openedOrderPrice,
      priceLength,
      priceStep,
      updateSubscription: null,
      orderDialog: {
        opened: 0.00,
        flag: false,
        operation: '',
        price: 0.00,
        options: [],
      },
      chartRef: null,
    };
  },
  methods: {
    closeOrderDialog() {
      this.orderDialog.flag = false;
      this.orderDialog.operation = '';
      this.orderDialog.price = 0.00;
      this.orderDialog.opened = 0.00;
      this.orderDialog.options = [];
    },
    makeOrder(operation) {
      this.orderDialog.flag = true;
      this.orderDialog.operation = operation.toUpperCase();
      this.orderDialog.price = this.currentPrice;

      const options = [
        {
          percent: '0.5%',
          multiplier: 0.005,
        },
        {
          percent: '1%',
          multiplier: 0.015,
        },
        {
          percent: '2%',
          multiplier: 0.02,
        },
        {
          percent: '2.5%',
          multiplier: 0.025,
        },
        {
          percent: '4%',
          multiplier: 0.04,
        },
        {
          percent: '5%',
          multiplier: 0.05,
        },
      ];

      if ('sell' === operation && this.openedOrderPrice) {
        this.orderDialog.price = Number((this.openedOrderPrice + (this.openedOrderPrice * 0.005)).toFixed(this.priceLength));
        this.orderDialog.opened = this.openedOrderPrice;
        this.orderDialog.options = [];
        options.forEach((option) => {
          this.orderDialog.options.push({
            ...option,
            percent: `+${option.percent}`,
            value: Number((this.openedOrderPrice + (this.openedOrderPrice * option.multiplier)).toFixed(this.priceLength)),
          });
        });
        this.orderDialog.options.push({
          percent: 'Now',
          value: Number(this.currentPrice),
        });
      }

      if ('buy' === operation && this.currentPrice) {
        this.orderDialog.price = Number((this.currentPrice - (this.currentPrice * 0.005)).toFixed(this.priceLength));
        this.orderDialog.opened = Number(this.currentPrice.toFixed(this.priceLength));
        this.orderDialog.options = [];

        options.forEach((option) => {
          this.orderDialog.options.push({
            ...option,
            percent: `-${option.percent}`,
            value: Number((this.currentPrice - (this.currentPrice * option.multiplier)).toFixed(this.priceLength)),
          });
        });
        this.orderDialog.options.push({
          percent: 'Now',
          value: Number(this.currentPrice),
        });
      }
    },
    doOrder() {
      this.onOrder({
        operation: this.orderDialog.operation.toUpperCase(),
        price: this.orderDialog.price,
        symbol: this.symbol,
      });
      this.closeOrderDialog();
    },
  }
}
</script>

<style scoped>
.canvas-block {
  width: 100% !important;
  position: relative;
}
.canvas-block h2 {
  position: absolute;
  top: 6px;
  left: 65px;
  font-size: 16px;
  color: #000000;
}
.canvas-block .profit {
  position: absolute;
  top: 4px;
  left: 10px;
}
.canvas-block .current-price {
  position: absolute;
  top: 4px;
  right: 15px;
  font-weight: bold;
  color: #000;
}
.canvas-block .current-price small {
  font-size: 10px;
}
canvas {
  background-color: #efefef !important;
  padding: 10px;
}

@media only screen and (max-width: 700px) {
  canvas {
    padding-top: 35px;
  }
}
.market-control-block {
  text-align: right;
  background-color: #efefef;
  padding: 6px;
  max-width: 400px;
}
.market-control {
  flex-direction: row;
  display: flex;
  column-gap: 4px;
}
.order-actions {
  flex-direction: row;
  display: flex;
  column-gap: 4px;
}
.order-actions .action {
  flex: 1;
}
.market-control .btn-ord {
  flex: 1;
}
.financial-background {
  background-color: #efefef;
  text-align: right;
  text-align: -webkit-right;
}
.price-input-block {
  display: flex;
  flex-flow: row nowrap;
}
.price-input-block > * {
  flex: 1;
  align-items: stretch;
  min-height: 56px;
  border-radius: 0;
}
.price-input-value {
  flex-grow: 7;
}
.price-btn {
  font-size: 26px;
  font-weight: bold;
  touch-action: manipulation;
}
.price-options {
  display: flex;
  flex-flow: row nowrap;
  gap: 2px;
  margin-bottom: 10px;
}
.price-options > * {
  flex: 1;
  row-gap: 1px;
  touch-action: manipulation;
}
</style>