<template>
  <div class="financial-background">
    <div class="canvas-block">
      <canvas :id="canvasId"/>
    </div>
  </div>
</template>

<script lang="ts">
export default {
  name: "PriceSpeedChart",
  components: {},
  props: {
    symbol: {
      type: String,
      default: () => '',
    },
    avgPricePoints: {
      type: Array,
      default: () => [],
    },
    maxPricePoints: {
      type: Array,
      default: () => [],
    },
    minPricePoints: {
      type: Array,
      default: () => [],
    },
    updateSubject: {
      type: Object,
      default: () => null,
    },
  },
  data() {
    return {
      updateSubscription: null,
      chartRef: null,
    };
  },
  computed: {
    canvasId: {
      get() {
        return `canvas-price-speed-${this.symbol}`;
      },
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
    const ctx = document.getElementById(this.canvasId).getContext('2d');

    const updateDatasets = () => {
      const datasets = [
        {
          label: this.$t('price_speed_chart.mounted_label_avg'),
          data: this.avgPricePoints.map((point) => {
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
          borderColor: '#0748a8',
          pointBackgroundColor: '#5999e3',
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
        }
      ];

      if (this.minPricePoints.length > 0) {
        datasets.push({
          label: this.$t('price_speed_chart.mounted_label_min_pcs'),
          data: this.minPricePoints.map((point) => {
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

      if (this.maxPricePoints.length > 0) {
        datasets.push({
          label: this.$t('price_speed_chart.mounted_label_max_pcs'),
          data: this.maxPricePoints.map((point) => {
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

    const datasets = updateDatasets();

    const chartRef = new Chart(ctx, {
      type: 'line',
      data: {
        labels: datasets[0].data.map((point) => {
          const date = new Date(point.x);

          return date.toLocaleTimeString();
        }),
        datasets: datasets,
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