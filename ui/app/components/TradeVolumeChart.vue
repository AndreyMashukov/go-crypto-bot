<template>
  <div class="financial-background">
    <div class="canvas-block">
      <canvas :id="canvasId"></canvas>
    </div>
  </div>
</template>

<script lang="ts">
export default {
  components: {},
  name: "TradeVolumeChart",
  props: {
    symbol: {
      type: String,
      default: () => '',
    },
    sellVolumePoints: {
      type: Array,
      default: () => [],
    },
    buyVolumePoints: {
      type: Array,
      default: () => [],
    },
    cummulativeVolumePoints: {
      type: Array,
      default: () => [],
    },
    icebergBuyQty: {
      type: Array,
      default: () => [],
    },
    icebergSellQty: {
      type: Array,
      default: () => [],
    },
    updateSubject: {
      type: Object,
      default: () => null,
    },
  },
  unmounted() {
    if (this.updateSubscription) {
      this.updateSubscription.unsubscribe();
    }
    if (this.chartRef) {
      this.chartRef.destroy();
    }
  },
  mounted() {
    var ctx = document.getElementById(this.canvasId).getContext('2d');

    const updateDatasets = () => {
      const datasets = [];

      if (this.sellVolumePoints.length > 0) {
        datasets.push({
          label: this.$t('trade_volume_chart.mounted_label_sqs'),
          data: this.sellVolumePoints.map((point) => {
            return {
              x: point.x,
              y: point.y*-1,
            };
          }),
          elements: {
            point:{
              radius: 1,
            }
          },
          fill: true,
          order: 0,
          borderWidth: 1.5,
          borderColor: '#e80c0c',
          pointBackgroundColor: 'rgba(217,112,112,0.58)',
          backgroundColor: 'rgba(217,112,112,0.58)',
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

      if (this.icebergSellQty.length > 0) {
        datasets.push({
          label: 'Iceberg Sell Qty',
          data: this.icebergSellQty.map((point) => {
            return {
              x: point.x,
              y: point.y*-1,
            };
          }),
          elements: {
            point:{
              radius: 0.5,
            }
          },
          order: 0,
          borderWidth: 1.5,
          borderColor: 'rgba(232,12,12,0.5)',
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

      if (this.icebergBuyQty.length > 0) {
        datasets.push({
          label: 'Iceberg Buy Qty',
          data: this.icebergBuyQty.map((point) => {
            return {
              x: point.x,
              y: point.y,
            };
          }),
          elements: {
            point:{
              radius: 0.5,
            }
          },
          order: 0,
          borderWidth: 1.5,
          borderColor: 'rgba(91,217,13,0.5)',
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

      if (this.cummulativeVolumePoints.length > 0) {
        datasets.push({
          label: this.$t('trade_volume_chart.mounted_label_cq'),
          data: this.cummulativeVolumePoints.map((point) => {
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
          fill: true,
          order: 0,
          borderWidth: 1.5,
          borderColor: '#8010fa',
          pointBackgroundColor: 'rgba(170,100,238,0.24)',
          backgroundColor: 'rgba(170,100,238,0.24)',
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

      if (this.buyVolumePoints.length > 0) {
        datasets.push({
          label: this.$t('trade_volume_chart.mounted_label_bqs'),
          data: this.buyVolumePoints.map((point) => {
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
          fill: true,
          order: 0,
          borderWidth: 1.5,
          borderColor: '#5bd90d',
          pointBackgroundColor: 'rgba(153,245,150,0.55)',
          backgroundColor: 'rgba(153,245,150,0.55)',
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
      type: 'line',
      data: {
        labels: updateDatasets()[0].data.map((point) => {
          const date = new Date(point.x);

          return date.toLocaleTimeString();
        }),
        datasets: updateDatasets(),
      },
      options: {
        animation: {
          duration: 0
        },
      }
    });

    if (this.updateSubject) {
      this.updateSubscription = this.updateSubject.subscribe(() => {
        const newDatasets = updateDatasets();
        for (let i = 0; i < chartRef.data.datasets.length; i++) {
          chartRef.data.datasets[i].data = newDatasets[i].data;
          chartRef.data.labels = newDatasets[i].data.map((point) => {
            const date = new Date(point.x);

            return date.toLocaleTimeString();
          });
        }
        chartRef.update();
      });
    }

    this.chartRef = chartRef;
  },
  computed: {
    canvasId: {
      get() {
        return `canvas-trade-volume-${this.symbol}`;
      },
    },
  },
  data() {
    return {
      updateSubscription: null,
      chartRef: null,
    };
  },
  methods: {}
}
</script>

<style scoped>
.financial-background {
  background-color: #efefef;
  text-align: right;
  text-align: -webkit-right;
}
.canvas-block {
  width: 100% !important;
  position: relative;
}
canvas {
  background-color: #efefef !important;
  padding: 10px;
  max-height: 400px;
}

@media only screen and (max-width: 700px) {
  canvas {
    padding-top: 35px;
  }
}
</style>