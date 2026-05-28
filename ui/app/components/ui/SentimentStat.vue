<template>
  <div class="news-sentiment">
    <div v-if="!!entity.sentiment">
      <v-icon v-if="entity.sentiment.label === 'BULLISH'" color="primary" icon="mdi-rocket-launch" size="small"/>
      <v-icon v-else-if="entity.sentiment.label === 'NEUTRAL'" color="primary" icon="mdi-scale-balance" size="small"/>
      <v-icon v-else color="rgb(189,47,38)" icon="mdi-trending-down" size="small"/>
      <b
        :style="entity.sentiment.label === 'BULLISH' ? 'color: primary;' : (entity.sentiment.label === 'NEUTRAL' ? 'color: primary;' : 'color: rgb(189,47,38);')"
      >
        {{entity.sentiment.label}}
        <small>{{entity.sentiment.score.toFixed(2)}}</small>
      </b>
      <v-progress-linear
          class="mt-0"
          height="2px"
          max="100"
          :model-value="entity.sentiment.score*100"
          :color="entity.sentiment.label === 'BULLISH' ? 'primary' : (entity.sentiment.label === 'NEUTRAL' ? 'primary' : 'rgb(189,47,38)')"
      />
    </div>
    <span v-else>
      <span>NEWS:</span> <small>N/A</small>
    </span>
  </div>
</template>
<script lang="ts">
export default {
  name: "SentimentStat",
  components: {},
  props: {
    entity: {
      type: Object,
      default: () => '',
    },
  },
}
</script>
<style scoped>
.news-sentiment {
  font-size: 12px;
  height: 20px;
}
.news-sentiment span {
  font-weight: 500;
}
</style>