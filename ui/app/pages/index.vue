<template>
  <v-container fluid>
    <div class="main-image-block">
      <h1>{{$t('index.cloud_crypto_bot')}}</h1>
      <ul>
        <li>{{$t('index.easy_to_use')}}</li>
        <li>{{$t('index.uptime_24_7')}}</li>
        <li>{{$t('index.modern_technologies')}}</li>
        <li>{{$t('index.technical_support')}}</li>
      </ul>
      <div class="start-trade">
        <v-btn variant="flat" size="large" color="secondary" @click="loginNow">
          {{$t('index.trade_now')}}
        </v-btn>
        <v-btn variant="flat" :href="$t('index.documentation_link')" target="_blank" color="info">
          {{$t('index.documentation')}}
        </v-btn>
      </div>
    </div>
    <v-row class="advantages">
      <v-col cols="12" lg="3" md="4" sm="6" xs="6">
        <v-card
            class="mx-auto"
            max-width="344"
            :title="$t('index.safety_title')"
            :subtitle="$t('index.safety_subtitle')"
            prepend-icon="mdi-github"
            append-icon="mdi-check"
        >
          <v-card-text>{{$t('index.safety_text')}}<a style="text-decoration: underline;color:#1f659b;font-weight: bold;" href="https://github.com/AndreyMashukov/go-crypto-bot" target="_blank">{{$t('index.safety_text_link')}}</a></v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" lg="3" md="4" sm="6" xs="6">
        <v-card
            class="mx-auto"
            max-width="344"
            :title="$t('index.cloud_title')"
            :subtitle="$t('index.cloud_subtitle')"
            prepend-icon="mdi-cloud"
            append-icon="mdi-check"
        >
          <v-card-text>{{$t('index.cloud_text')}}</v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" lg="3" md="4" sm="6" xs="6">
        <v-card
            class="mx-auto"
            max-width="344"
            :title="$t('index.dev_team_title')"
            :subtitle="$t('index.dev_team_subtitle')"
            prepend-icon="mdi-star"
            append-icon="mdi-check"
        >
          <v-card-text>{{$t('index.dev_team_text')}}</v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" lg="3" md="4" sm="6" xs="6">
        <v-card
            class="mx-auto"
            max-width="344"
            :title="$t('index.support_title')"
            :subtitle="$t('index.support_subtitle')"
            prepend-icon="mdi-face-agent"
            append-icon="mdi-check"
        >
          <v-card-text>{{$t('index.support_text')}}</v-card-text>
        </v-card>
      </v-col>
    </v-row>
    <h2 class="bg-teal-darken-4 text-center" style="margin: 12px -16px 0 -16px !important;padding-top: 10px;">
      {{$t('index.exchanges_title')}}
    </h2>
    <v-row class="bg-teal-darken-4" style="margin: 0 -16px 0 -16px !important;">
      <v-col cols="12" lg="6" md="6" sm="12" xs="12" class="text-center">
        <a href="https://binance.com" target="_blank">
          <Binance/>
        </a>
      </v-col>
      <v-col cols="12" lg="6" md="6" sm="12" xs="12" class="text-center">
        <a href="https://bybit.com" target="_blank">
          <ByBit/>
        </a>
      </v-col>
    </v-row>
    <p class="bg-teal-darken-4" style="margin: 0 -16px 0 -16px !important;padding: 16px">
      {{$t('index.exchanges_description')}}
    </p>
    <v-row class="opened-position-carousel">
      <v-col cols="12" class="opened-position-carousel-container">
        <div class="scroll-block">
          <div class="position-scroll">
            <v-badge
              v-for="(position, index) in positions"
              :key="`pos-${index}`"
              :content="position.position.order.exchange"
              :color="'primary'"
              :offset-y="'29'"
              :offset-x="position.position.order.exchange === 'binance' ? '52' : '35'"
            >
              <Position
                class="position-item"
                :position="position.position"
                :trader="position.trader"
                :orientation="'horizontal'"
              />
            </v-badge>
          </div>
        </div>
      </v-col>
    </v-row>
    <div class="swap-options">
      <div v-for="(swap, index) in swaps" :key="`swap-${swap.title}-${swap.percent}-${index}`" class="swap-option">
        <span><b>{{$t('index.triangular')}}</b></span><br>
        <span>{{ swap.title }}</span><br>
        <small>{{$t('index.percent_now')}} = {{ swap.percent }}% {{$t('index.percent_max')}} = {{ swap.maxPercent }}%</small>
        <div class="swap-steps">
          <v-chip style="height: 16px;" size="xs" :color="swap.swapOne.operation === 'BUY' ? 'primary' : 'red'" class="pr-2 mr-1 mt-1"><div class="step-number">1</div><span>{{ swap.swapOne.operation }}</span>&nbsp;<b>{{ swap.swapOne.symbol }}</b><small>&nbsp;{{ swap.swapOne.price }}&nbsp;</small></v-chip>
          <v-chip style="height: 16px;" size="xs" :color="swap.swapTwo.operation === 'BUY' ? 'primary' : 'red'" class="pr-2 mr-1 mt-1"><div class="step-number">2</div><span>{{ swap.swapTwo.operation }}</span>&nbsp;<b>{{ swap.swapTwo.symbol }}</b><small>&nbsp;{{ swap.swapTwo.price }}&nbsp;</small></v-chip>
          <v-chip style="height: 16px;" size="xs" :color="swap.swapThree.operation === 'BUY' ? 'primary' : 'red'" class="pr-2 mt-1"><div class="step-number">3</div><span>{{ swap.swapThree.operation }}</span>&nbsp;<b>{{ swap.swapThree.symbol }}</b><small>&nbsp;{{ swap.swapThree.price }}&nbsp;</small></v-chip>
        </div>
      </div>
    </div>
    <v-row class="last-trades">
      <h2>{{$t('index.discover_trades')}}</h2>
      <v-col cols="12">
        <TradeTable :last-orders="lastOrders"/>
      </v-col>
    </v-row>
    <v-row class="pricing bg-teal-darken-4">
      <h2>{{$t('index.two_options')}}</h2>
      <v-col cols="12" lg="6" md="6" sm="12" xs="12">
        <v-card
          class="mx-auto"
          max-width="344"
          :title="$t('index.flexible_title')"
          :subtitle="$t('index.flexible_subtitle')"
          prepend-icon="mdi-percent"
          append-icon="mdi-check"
        >
          <v-card-text>{{$t('index.flexible_text')}}</v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" lg="6" md="6" sm="12" xs="12">
        <v-card
          class="mx-auto"
          max-width="344"
          :title="$t('index.fixed_title')"
          :subtitle="$t('index.fixed_subtitle')"
          prepend-icon="mdi-account-cash"
          append-icon="mdi-check"
        >
          <v-card-text>{{$t('index.fixed_text')}}</v-card-text>
        </v-card>
      </v-col>
    </v-row>
    <v-row class="commission-block">
      <h2>{{$t('commission_table.header')}}</h2>
      <v-col cols="12">
        <v-table v-if="commission && commission.length > 0" fixed-header>
          <thead>
            <tr>
              <th class="text-left">{{$t('commission_table.level')}}</th>
              <th class="text-left">{{$t('commission_table.volume')}}</th>
              <th class="text-left">{{$t('commission_table.commission_percent')}}</th>
              <th class="text-left">{{$t('commission_table.min_commission_usd')}}</th>
            </tr>
            </thead>
          <tbody>
            <tr
                v-for="item in commission"
                :key="item.level"
            >
              <td>{{ item.level }}</td>
              <td><span v-html="item.title"/></td>
              <td>{{ item.commissionPercent.toFixed(2) }}%</td>
              <td>{{ item.minCommissionUsd.toFixed(2) }}$</td>
            </tr>
          </tbody>
        </v-table>
      </v-col>
    </v-row>
  </v-container>
</template>
<script lang="ts">
import {TradeTable} from '#components'
import ByBit from '~/components/logos/ByBit.vue';
import Binance from '~/components/logos/Binance.vue';

export default defineNuxtComponent({
  components: {Binance, ByBit, TradeTable},
  async asyncData() {
    const config = useRuntimeConfig();
    const {t} = useI18n();

    const [{data: lastOrders}, {data: positions}, {data: swaps}, {data: commission}] = await Promise.all([
      useFetch('/public/cryptobot/trades', {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
      }),
      useFetch('/public/cryptobot/positions', {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
      }),
      useFetch('/public/cryptobot/swaps', {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
      }),
      useFetch('/public/commission/info', {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
      }),
    ]);

    useHead({
      title: t('index.header'),
    })

    let positionsArray = [];

    if (positions.value && positions.value.length > 0) {
      positionsArray = (positions.value).concat(positions.value).concat(positions.value)
    }

    const swapsValue = swaps.value === null ? [] : swaps.value;
    const commissionValue = commission.value === null ? [] : commission.value;

    return {
      lastOrders: lastOrders.value,
      positions: positionsArray,
      swaps: swapsValue,
      commission: commissionValue,
    };
  },
  data() {
    return {
      updateSubscription: null,
    };
  },
  computed: {},
  mounted() {
    const config = useRuntimeConfig();
    this.updateSubscription = setInterval(() => {
      useFetch('/public/cryptobot/positions', {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
      }).then(({data: positions}) => {
        this.positions = (positions.value).concat(positions.value).concat(positions.value);
      });
      useFetch('/public/cryptobot/swaps', {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
      }).then(({data: swaps}) => {
        this.swaps = swaps.value === null ? [] : swaps.value;
      });
    }, 10000);
  },
  unmounted() {
    if (this.updateSubscription) {
      clearInterval(this.updateSubscription);
    }
  },
  methods: {
    loginNow() {
      if (this.$services.authService.getToken()) {
        setTimeout(async () => {
          await this.$services.routerService.navigate('/account');
        });
      } else {
        this.$services.authService.openAuthDialog.next();
      }
    },
  },
})
</script>
<style scoped>
.main-image-block {
  background-image: url("/phone.png");
  background-position: right -30px top 20px;
  background-size: 500px;
  background-color: rgba(0, 0, 0, 0.80);
  margin: -16px !important;
  margin-top: -22px !important;
  position: relative;
  box-sizing: border-box;
  width: 100vw;
  height: 520px;
}

.main-image-block h1 {
  position: absolute;
  top: 30px;
  left: 50px;
  font-size: 36px;
  border-radius: 12px;
  padding: 0 10px;
  margin-right: 50px;
  color: #FFFFFF;
  text-shadow: 2px 2px rgba(0, 0, 0, 0.99);
  font-weight: bold;
}
.main-image-block ul {
  position: absolute;
  top: 180px;
  left: 80px;
  font-size: 20px;
  color: #FFFFFF;
  text-shadow: 2px 2px #000;
  font-weight: bold;
}
.main-image-block .start-trade {
  top: 350px;
  left: 60px;
  position: absolute;
}
.main-image-block .start-trade button {
  display: block;
}
.main-image-block .start-trade button:first-child {
  margin-bottom: 8px;
}
.last-trades h2 {
  text-align: left;
  width: 100%;
  margin-left: 10px;
  margin-top: 10px;
}
.pricing {
  margin: 0 -16px;
  padding: 10px;
  min-height: 350px;
}
.commission-block {
  margin: 0 -16px;
  padding: 10px;
}
.commission-block h2 {
  text-align: left;
  width: 100%;
  margin-left: 10px;
  margin-top: 10px;
}
.plans h2 {
  text-align: left;
  width: 100%;
  margin-left: 10px;
  margin-top: 10px;
}
.pricing h2 {
  text-align: left;
  width: 100%;
  margin-left: 10px;
  margin-top: 10px;
}
.order-trade div {
  font-size: 14px;
}
.order-trade small {
  font-size: 10px;
}
.plan-prop {
  margin: 5px;
}
.plan-top {
  min-height: 80px;
}
@media only screen and (max-width: 700px) {
  .main-image-block {
    background-position: right -80px top -10px;
  }
  .main-image-block h1 {
    left: 20px;
  }
  .main-image-block ul {
    left: 45px;
    top: 250px;
  }
  .main-image-block .start-trade {
    top: 410px;
    left: calc(50% - 130px);
  }
}
.scroll-block {
  position: relative;
  width: 100%;
  overflow: hidden;
  z-index: 1;
  margin: 0;
  padding: 0;
}
.position-scroll {
  overflow: hidden;
  height: 100%;
  white-space: nowrap;
  animation: scrollText 100s infinite linear;
  margin: 0;
  font-size: 0;
  display: flex;
  justify-content: space-between;
  width: fit-content;
}
@keyframes scrollText {
  from {
    transform: translateX(0%);
  }
  to {
    transform: translateX(-50%);
  }
}
.opened-position-carousel-container {
  padding-top: 2px !important;
}
.opened-position-carousel {
  margin: 0 -16px 0 -16px !important;
  padding-top: 0 !important;
  position: relative;
}
.advantages {
  margin: 15px -16px -16px -16px !important;
  background-color: rgba(0, 0, 0, 0.80);
  padding-bottom: 16px;
}
.opened-position-carousel > div {
  padding: 10px 0;
}
.swap-option {
  display: inline-block;
  margin: 4px;
  padding: 8px;
  background: rgb(230, 238, 156);
  font-size: 9px;
  border-radius: 5px;
  max-width: 280px;
  text-align: center;
}
.swap-option .step-number {
  display: inline-block;
  border-radius: 4px;
  background-color: #8d9ceb;
  margin-right: 4px;
  padding: 5px;
  font-weight: bold;
  color: #000000;
}
.swap-option .swap-steps {
  margin-top: 5px;
  font-size: 9px;
  white-space: break-spaces;
}
.swap-options {
  padding: 0;
  text-align: left;
  margin: 0 0 6px;
  overflow: auto;
  white-space: nowrap;
}
</style>
