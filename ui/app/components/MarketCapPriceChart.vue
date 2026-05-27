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
  name: "MarketCapPriceChart",
  props: {
    symbol: {
      type: String,
      default: () => '',
    },
    marketCapPricePoints: {
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

      if (this.marketCapPricePoints.length > 0) {
        datasets.push({
          label: this.$t('market_cap_price_chart.mounted_label'),
          data: this.marketCapPricePoints.map((point) => {
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
          borderColor: '#0152fd',
          pointBackgroundColor: 'rgba(142,170,245,0.4)',
          backgroundColor: 'rgba(142,170,245,0.4)',
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
        plugins: {
          tooltip: {
            callbacks: {
              label: (ctx) => (`${ctx.dataset.label}: ${ctx.raw.y}`)
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
        return `canvas-market-cap-price-${this.symbol}`;
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
  max-height: 300px;
}

@media only screen and (max-width: 700px) {
  canvas {
    padding-top: 35px;
  }
}
</style>