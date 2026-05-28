<template>
  <div>
    <v-row v-if="tradeFilter.mode === 'multi'" class="filter-item">
      <v-col cols="1" class="p-0 numeric-column">
        <v-icon :icon="`mdi-numeric-${optionNumber}-box`" color="primary"/>
      </v-col>
      <v-col cols="10" class="p-0">
        <v-row>
          <v-col cols="6" class="p-0 py-0">
            <v-select
                v-model="tradeFilter.type"
                :label="$t('trade_condition.col_label_type')"
                variant="underlined"
                :rules="[rules.required]"
                :items="conditions"
                type="text"
            />
          </v-col>
          <v-col v-if="canBeMulti" cols="6" class="p-0 py-0">
            <v-select
                v-model="tradeFilter.mode"
                :label="$t('trade_condition.col_label_mode')"
                variant="underlined"
                :rules="[rules.required]"
                :items="conditionMode"
                type="text"
                @update:model-value="onChangeMode"
            />
          </v-col>
        </v-row>
        <TradeCondition
            v-for="(childFilter, childFilterIndex) in tradeFilter.children"
            :can-be-multi="false"
            :option-number="childFilterIndex+1"
            :trade-filter="childFilter"
            class="child-filter-container"
            :symbols="symbols"
            @on-click-remove="tradeFilter.children.splice(childFilterIndex, 1)"
        />
        <div class="parameters">
          <v-btn
              v-for="(parameter, paramIndex) in ['price', 'daily_percent', 'position_time_minutes', 'extra_orders_today', 'has_signal', 'sentiment_label', 'sentiment_score'].filter((x) => !tradeFilter.children.map((y) => y.parameter).includes(x))"
              :key="`${paramIndex}-${parameter}`"
              size="x-small"
              class="mb-2 mt-2 mr-1"
              color="primary"
              @click="clickAddChildren(parameter)"
          >{{ `+ ${parameter}` }}</v-btn>
        </div>
      </v-col>
      <v-col cols="1" class="p-0">
        <v-icon icon="mdi-close" color="red" style="margin-left: -10px" @click="clickRemove"/>
      </v-col>
    </v-row>
    <v-row v-if="tradeFilter.mode === 'single'" :class="canBeMulti? 'filter-item' : 'filter-item-child'">
      <v-col v-if="canBeMulti" cols="1" class="p-0 numeric-column">
        <v-icon :icon="`mdi-numeric-${optionNumber}-box`" color="primary"/>
      </v-col>
      <v-col :cols="canBeMulti ? 10 : 11" class="p-0 py-0">
        <v-row>
          <v-col :cols="canBeMulti ? 3 : 6" class="p-0 py-0">
            <v-select
                v-model="tradeFilter.type"
                :label="$t('trade_condition.col_label_type')"
                variant="underlined"
                :rules="[rules.required]"
                :items="conditions"
                type="text"
            />
          </v-col>
          <v-col :cols="canBeMulti ? 5 : 6" class="p-0 py-0">
            <v-select
                v-model="tradeFilter.symbol"
                :label="$t('trade_condition.col_label_symbol')"
                variant="underlined"
                :rules="[rules.required]"
                :items="symbols"
                type="text"
            />
          </v-col>
          <v-col v-if="canBeMulti" cols="4" class="p-0 py-0">
            <v-select
                v-model="tradeFilter.mode"
                :label="$t('trade_condition.col_label_mode')"
                variant="underlined"
                :rules="[rules.required]"
                :items="conditionMode"
                type="text"
                @update:model-value="onChangeMode"
            />
          </v-col>
        </v-row>
        <v-row>
          <v-col cols="5" class="p-0 py-0">
            <v-select
                v-model="tradeFilter.parameter"
                :label="$t('trade_condition.col_label_parameter')"
                variant="underlined"
                :rules="[rules.required]"
                :items="['price', 'daily_percent', 'position_time_minutes', 'extra_orders_today', 'has_signal', 'sentiment_label', 'sentiment_score']"
                type="text"
            />
          </v-col>
          <v-col cols="3" class="p-0 py-0">
            <v-select
                v-model="tradeFilter.condition"
                :label="$t('trade_condition.col_label_condition')"
                variant="underlined"
                :rules="[rules.required]"
                :items="conditionMap[tradeFilter.parameter] || []"
                type="text"
            />
          </v-col>
          <v-col cols="4" class="p-0 py-0">
            <v-text-field
              v-if="inputTypeMap[tradeFilter.parameter] === 'text'"
              v-model="tradeFilter.value"
              :label="$t('trade_condition.col_label_value')"
              type="text"
              variant="underlined"
              :rules="[rules.required].concat(inputRules[tradeFilter.parameter])"
              :persistent-hint="!!inputHint[tradeFilter.parameter]"
              :hint="inputHint[tradeFilter.parameter]"
            />
            <v-checkbox
              v-if="inputTypeMap[tradeFilter.parameter] === 'checkbox'"
              v-model="tradeFilter.value"
              type="text"
              variant="underlined"
            />
            <v-select
              v-if="inputTypeMap[tradeFilter.parameter] === 'select'"
              v-model="tradeFilter.value"
              type="text"
              :items="selectOptions[tradeFilter.parameter]"
              variant="underlined"
            />
          </v-col>
        </v-row>
      </v-col>
      <v-col cols="1" class="p-0 numeric-column" style="padding: 0;">
        <v-icon icon="mdi-close" color="red" :style="canBeMulti ? 'margin-left: -5px; padding:0;' : ''" @click="clickRemove"/>
      </v-col>
    </v-row>
  </div>
</template>
<script lang="ts">
import {TradeLimit} from '~/model/trade-limit';

export default {
  name: "TradeCondition",
  props: {
    optionNumber: {
      type: Number,
      default: () => 0,
    },
    tradeFilter: {
      type: Object,
      default: () => {},
    },
    canBeMulti: {
      type: Boolean,
      default: () => false,
    },
    symbols: {
      type: Array,
      default: () => [],
    },
  },
  setup(props: any, {emit}: any) {
    function onClickRemove(symbol: any) {
      emit('onClickRemove', symbol);
    }
    function onAddChildren(symbol: any, parameter: any) {
      emit('onAddChildren', {symbol, parameter});
    }
    return {
      onAddChildren,
      onClickRemove,
    }
  },
  data() {
    const tradeFilter: any = {...this.tradeFilter};

    const conditionEqNeq = [
      {
        title: '=',
        value: 'eq',
      },
      {
        title: '!=',
        value: 'neq',
      },
    ];
    const conditions = [
      {
        title: '<',
        value: 'lt',
      },
      {
        title: '<=',
        value: 'lte',
      },
      {
        title: '>',
        value: 'gt',
      },
      {
        title: '>=',
        value: 'gte',
      },
    ];

    const rulePositive: any = (value: any) => value > 0.00 || this.$t('trade_condition.data_rules_positive');
    const ruleMax1: any     = (value: any) => value <= 1.00 || this.$t('trade_condition.data_rules_max1');

    return {
      rules: {
        required: (value: any) => !!value || this.$t('trade_condition.data_rules_required'),
      },
      conditionMap: {
        'price': conditions.concat(conditionEqNeq),
        'daily_percent': conditions.concat(conditionEqNeq),
        'position_time_minutes': conditions.concat(conditionEqNeq),
        'extra_orders_today': conditions.concat(conditionEqNeq),
        'has_signal': conditionEqNeq,
        sentiment_label: conditionEqNeq,
        sentiment_score: conditions.concat(conditionEqNeq),
      },
      inputRules: {
        'price': [],
        'daily_percent': [],
        'position_time_minutes': [],
        'extra_orders_today': [],
        'has_signal': 'checkbox',
        sentiment_label: [],
        sentiment_score: [rulePositive, ruleMax1],
      },
      inputTypeMap: {
        'price': 'text',
        'daily_percent': 'text',
        'position_time_minutes': 'text',
        'extra_orders_today': 'text',
        'has_signal': 'checkbox',
        sentiment_label: 'select',
        sentiment_score: 'text',
      },
      selectOptions: {
        sentiment_label: ['BEARISH', 'BULLISH'],
      },
      inputHint: {
        sentiment_score: '0.00 -> 1.00',
      },
      cached: {
        tradeFilter,
      }
    };
  },
  computed: {
    conditionMode: {
      get() {
        return [
          {
            title: this.$t('conditions.mode.single'),
            value: 'single',
          },
          {
            title: this.$t('conditions.mode.multi'),
            value: 'multi',
          },
        ];
      },
    },
    conditions: {
      get() {
        return [
          {
            title: this.$t('conditions.or'),
            value: 'or',
          },
          {
            title: this.$t('conditions.and'),
            value: 'and',
          },
        ];
      },
    },
  },
  methods: {
    clickRemove() {
      this.onClickRemove(this.tradeFilter);
    },
    clickAddChildren(parameter: string) {
      this.onAddChildren(this.tradeFilter.symbol, parameter);
    },
    onChangeMode(mode: any) {
      if (mode === 'multi' && this.tradeFilter.children.length === 0) {
        this.tradeFilter.children.push({
          symbol: this.tradeFilter.symbol,
          parameter: this.tradeFilter.parameter,
          condition: this.tradeFilter.condition,
          value: this.tradeFilter.value,
          type: this.tradeFilter.type,
          mode: 'single',
        });
      }

      if (mode === 'single' && this.tradeFilter.children.length > 0) {
        this.tradeFilter.symbol = this.tradeFilter.children[0].symbol;
        this.tradeFilter.parameter = this.tradeFilter.children[0].parameter;
        this.tradeFilter.condition = this.tradeFilter.children[0].condition;
        this.tradeFilter.value = this.tradeFilter.children[0].value;
        this.tradeFilter.type = this.tradeFilter.children[0].type;
        this.tradeFilter.mode = 'single';
      }
    },
  }
}
</script>
<style scoped>
.numeric-column {
  align-items: center;
  display: flex;
}
.filter-item {
  border: 1px solid #dedddd;
  padding-top: 15px;
  padding-bottom: 10px;
  border-radius: 6px;
  margin-bottom: 15px;
}
.filter-item-child {
  border-top: 1px solid #dedddd;
  padding-top: 20px;
  padding-bottom: 10px;
  margin-left: 2px;
}
.child-filter-container {
  padding-top: 15px;
  padding-bottom: 10px;
}
.parameters * {
  font-size: 9px !important;
  padding: 1px 3px !important;
}
</style>