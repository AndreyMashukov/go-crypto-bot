<template>
  <div class="d-inline">
    <v-btn
      @click="show = true"
      variant="flat"
      elevation="0"
      size="x-small"
      append-icon="mdi-chart-timeline-variant-shimmer"
      color="primary"
      width="100%"
    >
      PIVOT
    </v-btn>
    <v-dialog
        persistent
        v-model="show"
        max-width="500px"
        min-width="380px"
        z-index="9999"
    >
      <v-card>
        <v-card-title><small>PIVOT</small> {{symbol}}</v-card-title>
        <v-card-text class="pivot-card">
          <v-table v-if="pivots.length > 0" class="bg-primary pivot-table">
            <thead>
            <tr>
              <th class="text-left">
                <v-icon icon="mdi-calendar" size="small" color="white"/>
              </th>
              <th class="text-left">S1</th>
              <th class="text-left">S2</th>
              <th class="text-left">S3</th>
              <th class="text-left">R1</th>
              <th class="text-left">R2</th>
              <th class="text-left">R3</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="(pivot, index) in pivotsSorted" :key="index">
              <td class="text-left">{{pivot.periodDays}}<small>D</small></td>
              <td class="text-left">{{pivot.s1.toFixed(precision)}}</td>
              <td class="text-left">{{pivot.s2.toFixed(precision)}}</td>
              <td class="text-left">{{pivot.s3.toFixed(precision)}}</td>
              <td class="text-left">{{pivot.r1.toFixed(precision)}}</td>
              <td class="text-left">{{pivot.r2.toFixed(precision)}}</td>
              <td class="text-left">{{pivot.r3.toFixed(precision)}}</td>
            </tr>
            </tbody>
          </v-table>
          <div v-else>
            Pivot points is not calculated
          </div>
        </v-card-text>

        <v-card-actions class="justify-end">
          <v-btn variant="flat" color="default" class="action" @click="show = false">CLOSE</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>
<script lang="ts">
export default {
  components: {},
  name: "PivotPointsPopup",
  props: {
    symbol: {
      type: String,
      default: () => '',
    },
    pivots: {
      type: Array,
      default: () => [],
    },
    precision: {
      type: Number,
      default: () => 8,
    }
  },
  data() {
    return {
      show: false,
    };
  },
  computed: {
    pivotsSorted(): any {
      return this.pivots.sort((a, b) => a.periodDays > b.periodDays ? 1 : -1)
    }
  }
}
</script>
<style scoped>
.pivot-table thead th {
  color: #FFFFFF !important;
  font-weight: bold;
}
.pivot-table {
  border-radius: 6px;
  font-size: 10px !important;
}
.pivot-table tr {
  line-height: 1px;
}
.pivot-table td,
.pivot-table th {
  padding: 2px !important;
  height: 35px !important;
}
.pivot-table td:nth-child(1),
.pivot-table th:nth-child(1) {
  padding-left: 12px !important;
}
.pivot-card {
  padding: 6px !important;
}
.pivot-table tbody tr:nth-of-type(odd) {
  background-color: rgba(255, 255, 255, .2);
}
</style>