<template>
  <v-container fluid>
    <v-card
      class="mx-auto profile-top-image subscription-card-block"
      rounded="0"
    >
      <Ticker
:options="{
    colorTheme: 'dark',
    autosize: true,
    symbols: [
        {
          proName: 'BINANCE:BTCUSDT',
          title: 'Binance BTCUSDT',
        },
        {
          proName: 'BYBIT:BTCUSDT',
          title: 'ByBit BTCUSDT',
        },
        {
          proName: 'BINANCE:ETHUSDT',
          title: 'Binance ETHUSDT',
        },
        {
          proName: 'BYBIT:ETHUSDT',
          title: 'ByBit ETHUSDT',
        },
      ],
      isTransparent: false,
      showSymbolLogo: true,
      locale: 'en',
  }"/>
      <v-list-item
        class="text-white"
        :title="`${userVm.nickname}`"
      />
      <v-card
        width="280"
        class="subscription-card"
      >
        <template #title>
          <div v-if="!!userVm">
            <div>
              <v-tooltip location="top">
                <template #activator="{ props }">
                  <small>{{$t('account.budget')}} {{ userVm.budget.toFixed(2) }}$</small>&nbsp;
                  <v-icon icon="mdi-information" v-bind="props" :color="'primary'" size="x-small" class="cursor-pointer"/>
                </template>
                <div>
                  <small v-if="!userVm.hasActiveBasicSubscription">{{$t('account.budget_hint.no_basic_subscription')}}</small>
                  <small v-else>{{$t('account.budget_hint.has_basic_subscription')}}</small>
                </div>
              </v-tooltip>
              <div>
                <v-btn key="budget-recharge" size="x-small" color="primary" @click="rechargeBudget">{{$t('account.recharge_budget')}}</v-btn>
              </div>
            </div>
            <div v-if="userVm.isPartner" class="partner-budget-block">
              <div>
                <b>{{$t('account.partner_budget')}}:</b> {{ userVm.partnerBudget.toFixed(2) }} <small>USDT</small>
              </div>
              <div>
                <a href="mailto:support@example.com">{{$t('account.withdraw_request')}}</a>
              </div>
            </div>
          </div>
        </template>
      </v-card>
    </v-card>

    <div>
      <v-alert
          class="mt-2 mb-4"
          border="bottom"
          border-color="info"
      >
        <template #text>
          {{$t(`account.documentation.text`)}} <a class="documentation-link" :href="$t(`account.documentation.link.url`)" target="_blank">{{$t(`account.documentation.link.text`)}}</a>
        </template>
      </v-alert>
    </div>

    <v-row class="mt-1">
      <v-col cols="12" lg="8" md="8" sm="12" xs="12">
        <v-row>
          <v-col v-for="(botInfo, index) in botList" v-if="botList.length > 0" :key="index" cols="12" lg="4" md="6" sm="12" xs="12">
            <v-card
                color="teal-lighten-5"
            >
              <v-btn icon="mdi-cog" variant="plain" style="position: absolute;right: 0;" elevation="0" @click="openBot(botInfo.cryptobot.id)"/>
              <v-card-item>
                <div>
                  <div class="text-overline mb-1">
                    {{ $t('account.crypto_bot').replace('[provider]', botInfo.cryptobot.provider) }}
                  </div>
                  <div class="text-h6 mb-1">
                    {{ $t('account.bot_instance').replace('[id]', botInfo.cryptobot.id) }}
                  </div>
                  <small>IP: {{ botInfo.cryptobot.ipAddress || $t('account.is_not_running') }}</small>
                  <div class="text-caption">
                    {{$t('account.bot_status')}}
                    <v-chip v-if="botInfo.cryptobot.status === 'new'" variant="flat" color="primary" size="x-small">
                      {{ botInfo.cryptobot.status }}
                    </v-chip>
                    <v-chip v-if="botInfo.cryptobot.status === 'stopped'" variant="flat" color="warning" size="x-small">
                      {{ botInfo.cryptobot.status }}
                    </v-chip>
                    <v-chip v-if="botInfo.cryptobot.status === 'running'" variant="flat" color="success" size="x-small">
                      {{ botInfo.cryptobot.status }}
                    </v-chip>
                    <v-chip v-if="botInfo.cryptobot.status === 'deploy'" variant="flat" color="accent" size="x-small">
                      {{ botInfo.cryptobot.status }}
                    </v-chip>
                    <v-chip v-if="botInfo.cryptobot.status === 'error'" variant="flat" color="red" size="x-small">
                      {{ botInfo.cryptobot.status }}
                    </v-chip>
                  </div>
                  <div>
                    <v-chip v-if="!botInfo.cryptobot.dedicated" prepend-icon="mdi-server" color="primary" size="x-small">
                      {{$t('account.server.shared')}}
                    </v-chip>
                    <v-chip v-else prepend-icon="mdi-server" color="primary"  size="x-small">
                      {{$t('account.server.dedicated').replace('[ip]', botInfo.cryptobot.dedicated.ip)}}
                    </v-chip>
                  </div>
                  <div>
                    <v-chip v-if="!botInfo.commission.percent" prepend-icon="mdi-label-percent" color="grey-darken-4" size="x-small">
                      {{$t('account.commission.no_commission')}}
                    </v-chip>
                    <v-chip v-else prepend-icon="mdi-label-percent" color="grey-darken-4"  size="x-small">
                      {{$t('account.commission.has_commission').replace('[percent]', botInfo.commission.percent).replace('[minValueUsd]', botInfo.commission.minValueUsd.toFixed(2))}}
                    </v-chip>
                  </div>
                  <div>
                    <v-chip prepend-icon="mdi-finance" color="grey-darken-4" size="x-small">
                      {{$t('account.monthly_trade_volume').replace('[tradeVolume]', botInfo.commission.botMonthlyVolume.toFixed(2))}}
                    </v-chip>
                  </div>
                  <div v-if="!!botInfo.cryptobot.errorMessage" class="error-message">
                    <small>{{ botInfo.cryptobot.errorMessage }}</small>
                  </div>
                </div>
              </v-card-item>

              <v-card-actions>
                <v-btn
                    color="primary"
                    variant="flat"
                    :loading="dashboardLoading[botInfo.cryptobot.provider]"
                    :disabled="botInfo.cryptobot.status !== 'running' || (dashboardLoading['binance'] || dashboardLoading['bybit'])"
                    @click="openDashboard(botInfo.cryptobot)"
                >
                  {{$t('account.dashboard_btn')}}
                </v-btn>
              </v-card-actions>
            </v-card>
          </v-col>
          <v-col v-for="(provider, index) in availableProviderList" v-if="availableProviderList.length > 0" :key="index" cols="12" lg="4" md="6" sm="12" xs="12">
            <v-card
                color="teal-lighten-5"
            >
              <v-card-item>
                <div>
                  <div class="text-overline mb-1">
                    {{ $t('account.crypto_bot').replace('[provider]', provider) }}
                  </div>
                  <div class="text-h6 mb-1">
                    {{ $t('account.new_instance') }}
                  </div>
                  <div class="text-caption">{{$t('account.new_instance_setup')}}</div>
                </div>
              </v-card-item>

              <v-card-actions>
                <v-btn color="primary" variant="flat" @click="botSetupDialog = provider">
                  {{$t('account.setup_btn')}}
                </v-btn>
              </v-card-actions>
            </v-card>
          </v-col>
        </v-row>
      </v-col>
      <v-col cols="12" lg="4" md="4" sm="12" xs="12">
        <v-card
            color="teal-lighten-5"
        >
          <v-card-title>{{$t(`account.paid_services.title`)}}</v-card-title>
          <v-list lines="three">
            <v-list-item
                v-for="paidService in paidServices"
                :key="paidService.code"
            >
              <template #title>
                <div class="paid-service-title">{{$t(`account.paid_services.${paidService.code}.title`)}}</div>
              </template>
              <template #subtitle>
                <div>
                  <span class="service-active-until">{{!!paidService.expiresAt ? $t(`account.paid_services.until`).replace('[date]', TimeHelper.getFormatted(paidService.expiresAt)) : $t(`account.paid_services.payment_required`)}}</span>
                </div>
                <div>
                  <small>{{$t('account.paid_services.price_subtitle').replace('[days]', paidService.days).replace('[price]', paidService.price)}}</small>
                </div>
                <div v-if="paidService.code === 'api_subscription'" class="api-docs">
                  <div v-html="$t('account.paid_services.api_doc_url')"/>
                </div>
              </template>
              <template #prepend>
                <v-avatar color="primary">
                  <v-icon color="white">{{$t(`account.paid_services.${paidService.code}.icon`)}}</v-icon>
                </v-avatar>
              </template>

              <template #append>
                <v-btn
                  color="primary"
                  variant="text"
                  append-icon="mdi-cart"
                  size="small"
                  :loading="paidServiceStatus.loading"
                  @click="purchaseService(paidService)"
                >
                  <span v-if="!paidService.isActive">{{$t(`account.paid_services.buy`)}}</span>
                  <span v-else>{{$t(`account.paid_services.extend`).replace('[days]', paidService.days)}}</span>
                </v-btn>
              </template>
            </v-list-item>
          </v-list>
        </v-card>
      </v-col>
    </v-row>
    <v-dialog v-model="botSetupDialog" width="600">
      <template #default="{ isActive }">
        <v-form ref="form" @submit.prevent="setupBot(botSetupDialog)">
          <v-card>
            <v-card-title class="mt-2">{{$t('account.bot_setup_dialog')}}</v-card-title>
            <v-row class="px-4 pt-4">
              <v-col cols="12" lg="12" md="12" sm="12" xs="12" class="mt-0 pt-0">
                <v-text-field
                    v-model="apiKey"
                    :label="$t('account.bot_setup_key_label').replace('[provider]', botSetupDialog)"
                    variant="solo-filled"
                    :rules="[rules.required, rules.apiKeyPattern]"
                />
              </v-col>
              <v-col cols="12" lg="12" md="12" sm="12" xs="12" class="mt-0 pt-0">
                <v-text-field
                    v-model="apiSecret"
                    :label="$t('account.bot_setup_secret_label').replace('[provider]', botSetupDialog)"
                    variant="solo-filled"
                    :rules="[rules.required, rules.apiKeyPattern]"
                />
              </v-col>
              <v-col cols="12" lg="12" md="12" sm="12" xs="12" class="mt-0 pt-0">
                <v-alert
                    type="warning"
                    :title="$t('account.bot_setup_dialog_config_title')"
                    :text="$t('account.bot_setup_dialog_config_text')"
                />
              </v-col>
            </v-row>

            <v-card-actions class="mt-2">
              <v-spacer/>

              <v-btn
                  variant="flat"
                  color="primary"
                  :text="$t('account.bot_setup_dialog_setup_btn')"
                  type="submit"
              />
              <v-btn
                  variant="text"
                  color="primary"
                  :text="$t('account.bot_setup_dialog_close_btn')"
                  @click="botSetupDialog = false"
              />
            </v-card-actions>
          </v-card>
        </v-form>
      </template>
    </v-dialog>
    <v-dialog v-model="isBudgetRecharge" width="500" persistent>
      <template #default="{ isActive }">
        <v-card class="payment-card">
          <v-card-title class="ml-2 mt-2">
            {{$t('account.buy_sub_recharge')}}
          </v-card-title>
          <div v-if="!awaitingPayment.loading">
            <div class="pay-btn-block">
              <div>
                <v-btn
                    :loading="paymentInProcess"
                    variant="flat"
                    color="success"
                    class="pay-btn-budget"
                    @click="payNow(100)"
                >
                  100 USDT
                </v-btn>
                <v-btn
                    :loading="paymentInProcess"
                    variant="flat"
                    color="success"
                    class="pay-btn-budget"
                    @click="payNow(250)"
                >
                  250 USDT
                </v-btn>
                <v-btn
                    :loading="paymentInProcess"
                    variant="flat"
                    color="success"
                    class="pay-btn-budget"
                    @click="payNow(500)"
                >
                  500 USDT
                </v-btn>
                <v-btn
                    :loading="paymentInProcess"
                    variant="flat"
                    color="success"
                    class="pay-btn-budget"
                    @click="payNow(1000)"
                >
                  1000 USDT
                </v-btn>
              </div>
            </div>
          </div>
          <div v-else>
            <v-progress-linear v-if="awaitingPayment.loading" color="success" indeterminate=""/>
            <div v-if="!awaitingPayment.status">
              <p class="px-6 py-4">{{$t('account.awaiting_payment')}}</p>
              <div class="pay-btn-block">
                <v-btn
                    variant="flat"
                    color="success"
                    class="pay-btn"
                    @click="payNow(rechargeBudgetModal.amount, 'budget')"
                >
                  {{$t('account.pay_now_btn')}} {{ rechargeBudgetModal.amount }} USDT
                </v-btn>
              </div>
            </div>
            <div v-else>
              <div v-if="awaitingPayment.status === 'paid'" class="payment-alert text-center">
                <v-icon
                    class="mb-6"
                    color="success"
                    icon="mdi-check-circle-outline"
                    size="128"
                />

                <div class="text-h4 font-weight-bold">{{$t('account.payment_is_paid')}}</div>
              </div>
              <div v-if="awaitingPayment.status === 'expired'" class="payment-alert text-center">
                <v-icon
                    class="mb-6"
                    color="error"
                    icon="mdi-close-circle-outline"
                    size="128"
                />

                <div class="text-h4 font-weight-bold">{{$t('account.payment_is_expired')}}</div>
              </div>
            </div>
          </div>
          <HowToBuyBitcoin/>
          <v-card-actions>
            <v-spacer/>

            <v-btn
              variant="text"
              color="primary"
              :text="$t('account.close_payment_btn')"
              @click="closePayment()"
            />
          </v-card-actions>
        </v-card>
      </template>
    </v-dialog>

    <v-row class="mb-2">
      <v-col cols="12" lg="12" md="12" sm="12" xs="12">
        <v-alert
            v-if="botList.length === 0"
            style="margin-top: 10px;"
            type="success"
            :title="$t('account.setup_bot_title')"
            :text="$t('account.setup_bot_text')"
        />
        <v-alert
            v-else
            style="margin-top: 10px;"
            type="info"
            color="teal-lighten-4"
            :title="$t('account.contact_support_title')"
            :text="$t('account.contact_support_text')"
        >
          <NuxtLink to="https://t.me/autotrade_cloud" class="text-decoration-none support-btn" target="_blank">
            <v-btn variant="flat" color="white">
              {{$t('account.contact_us_btn')}}
            </v-btn>
          </NuxtLink>
        </v-alert>
      </v-col>
    </v-row>
  </v-container>
</template>
<script lang="ts">
import HowToBuyBitcoin from '~/components/HowToBuyBitcoin.vue';
import {AlertEvent} from '~/model/alert-event';
import {Alert} from '~/model/alert';
import {Confirmation} from '~/model/confirmation';
import {TimeHelper} from '~/services/time-helper';

// todo: how to buy
// 1) https://academy.binance.com/en/articles/binance-beginner-s-guide
// 2) https://www.coinbase.com/en/how-to-buy/bitcoin
// 3) https://metamask.io/buy-crypto/
export default defineNuxtComponent({
  components: {HowToBuyBitcoin},
  async asyncData(ctx: any) {
    const {t} = useI18n();

    const authToken = ctx.$services.authService.getToken();
    if (!authToken) {
      await ctx.$services.routerService.navigate('/');
      return;
    }
    const [user, paidServices, botList, availableBotMap] = await Promise.all([
      ctx.$services.httpClient.secureGetServer('/v1/user/me'),
      ctx.$services.httpClient.secureGetServer('/v1/service/list'),
      ctx.$services.httpClient.secureGetServer('/v1/cryptobot/list/extended'),
      ctx.$services.httpClient.secureGetServer('/v1/cryptobot/available'),
    ]);

    useHead({
      title: t('account.header'),
    });

    return {
      userVm: user,
      botList,
      availableBotMap,
      breadcrumbs: [
        {
          title: 'Home',
          disabled: false,
          href: '/',
        },
        {
          title: 'My Account',
          disabled: true,
          href: '',
        },
      ],
      botSetupDialog: false,
      paidServices,
    }
  },
  data() {
    return {
      dashboardLoading: {
        bybit: false,
        binance: false,
      },
      apiKey: '',
      apiSecret: '',
      paymentInProcess: false,
      awaitingPayment: {
        loading: false,
        timer: null,
        paymentLink: null,
        status: null,
      },
      rules: {
        required: (value: any) => !!value || this.$t('account.data_required'),
        apiKeyPattern: (value: any) => value.match(/^[a-z0-9]+$/ui) || this.$t('account.api_key_pattern'),
      },
      paidServiceStatus: {
        loading: false,
      },
      rechargeBudgetModal: {
        flag: false,
        amount: 0,
      },
    }
  },
  computed: {
    TimeHelper() {
      return TimeHelper
    },
    availableProviderList: {
      get() {
        const list = [];

        Object.keys(this.availableBotMap).forEach((provider) => {
          if (this.availableBotMap[provider]) {
            list.push(provider);
          }
        })

        return list;
      },
    },
    isBudgetRecharge: {
      get() {
        return this.rechargeBudgetModal.flag;
      },
    },
  },
  methods: {
    refreshUser() {
      this.$services.httpClient.secureGetServer('/v1/user/me').then((user) => {
        this.userVm = user;
      });
    },
    refreshPaidServices() {
      this.$services.httpClient.secureGetServer('/v1/service/list').then((paidServices) => {
        this.paidServices = paidServices;
      });
    },
    purchaseService(paidService: any) {
      let action = this.$t('account.paid_services.action_buy');
      if (paidService.isActive) {
        action = this.$t('account.paid_services.action_extend');
      }

      const text = this.$t('account.paid_services.purchase_message')
        .replace('[action]', action)
        .replace('[service]', this.$t(`account.paid_services.${paidService.code}.title`))
        .replace('[days]', paidService.days)
        .replace('[price]', paidService.price);

      this.$services.eventManager.confirmation(new Confirmation(
        this.$t('account.paid_services.confirm_title'),
        text,
        this.$t('account.paid_services.action_cancel'),
       action,
       () => {},
       () => {
         this.paidServiceStatus.loading = true;
         this.$services.httpClient.securePostClient('/v1/service/purchase', {
           code: paidService.code,
         }).then(() => {
           const successMessage = this.$t('account.paid_services.purchase_done')
             .replace('[service]', this.$t(`account.paid_services.${paidService.code}.title`));

           this.$services.eventManager.alert(
             new AlertEvent(
               new Alert(successMessage, Alert.TYPE_SUCCESS),
               4000
             )
           );
           this.refreshUser();
           this.refreshPaidServices();
         }).finally(() => {
           this.paidServiceStatus.loading = false;
         });
       },
      ));
    },
    rechargeBudget() {
      this.rechargeBudgetModal.flag = true;
    },
    closePayment() {
      if (this.awaitingPayment.timer) {
        clearInterval(this.awaitingPayment.timer);
      }

      this.awaitingPayment = {
        loading: false,
        timer: null,
        paymentLink: null,
        status: null,
      };
      this.rechargeBudgetModal.flag = false;
      this.rechargeBudgetModal.amount = 0;
    },
    payNow(amount: number) {
      const url = `/v1/subscription/budget`;
      const method = 'POST';
      this.rechargeBudgetModal.amount = amount;

      const config = useRuntimeConfig();
      const tab = window.open('');

      if (this.awaitingPayment.paymentLink != null) {
        tab.location = this.awaitingPayment.paymentLink;
        tab.focus();
        return;
      }

      this.paymentInProcess = true;

      useFetch(url, {
        method,
        baseURL: config.public.baseUrl,
        server: true,
        headers: {
          Authorization: `Bearer ${this.$services.authService.getToken()}`,
        },
        body: {amount},
      }).then(({data: payment}) => {
        if (payment.value.paymentLink) {
          tab.location = payment.value.paymentLink;
          tab.focus();
          this.paymentInProcess = false;
          // listener...
          this.awaitingPayment.loading = true;
          this.awaitingPayment.paymentLink = payment.value.paymentLink;

          this.awaitingPayment.timer = setInterval(() => {
            useFetch(`/v1/payment/${payment.value.id}`, {
              method: 'GET',
              baseURL: config.public.baseUrl,
              server: true,
              headers: {
                Authorization: `Bearer ${this.$services.authService.getToken()}`,
              },
            }).then(({data: paymentDetails}) => {
              if (['paid', 'expired'].includes(paymentDetails.value.status)) {
                this.awaitingPayment.status = paymentDetails.value.status;

                clearInterval(this.awaitingPayment.timer);
                setTimeout(() => {
                  window.location.reload();
                }, 2000);
              }
            });
          }, 3500);
        }
      });
    },
    async openDashboard(bot: any) {
      this.dashboardLoading[bot.provider] = true;
      await this.$services.routerService.navigate('/dashboard/'+bot.id);
    },
    async openBot(id: number) {
      await this.$services.routerService.navigate('/bot/'+id);
    },
    setupBot(provider: string) {
      this.$refs.form.validate().then(({valid}: any) => {
        if (!valid) {
          return;
        }

        const config = useRuntimeConfig();
        useFetch('/v1/cryptobot', {
          method: 'POST',
          baseURL: config.public.baseUrl,
          server: true,
          body: {
            apiKey: this.apiKey,
            apiSecret: this.apiSecret,
            provider,
          },
          headers: {
            Authorization: `Bearer ${this.$services.authService.getToken()}`,
          },
        }).then(({data: data, error: error}: any) => {
          if (error.value) {
            this.$services.eventManager.alert(
              new AlertEvent(
                new Alert(error.value.data, Alert.TYPE_ERROR),
                4000
              )
            );
            return;
          }

          window.location.reload();
        }).finally();
      });
    }
  },
})
</script>
<style scoped>
.profile-top-image {
  margin: -20px !important;
  margin-bottom: 6px !important;
}
.paid-service-title {
  white-space: normal;
  font-size: 14px;
  font-weight: 500;
}
.service-active-until {
  font-size: 12px;
  color: #000000 !important;
  opacity: 0.99 !important;
}
.pay-btn-block {
  margin: 10px 25px;
  box-sizing: border-box;
}
.payment-alert {
  margin: 25px;
  box-sizing: border-box;
}
.payment-card {
  box-sizing: border-box;
}
.pay-btn-block .pay-btn {
  width: 100%;
  margin: auto;
}
.pay-btn-block .pay-btn-budget {
  width: 49%;
  margin: 2px;
}
.plan-prop {
  margin: 5px;
}
.plan-top {
  min-height: 80px;
}
.profile-top-image {
  position: relative;
}
.subscription-card {
  margin: 6px !important;
  background: rgba(255, 255, 255, 0.90);
}
.subscription-card-block {
  background: #2a2e39;
}
.v-container {
  margin-top: 0;
}
.support-btn {
  float: right;
  vertical-align: middle;
}
.support-btn .v-btn {
  margin-top: 4px;
}
.error-message {
  color: rgb(255, 0, 0);
  background-color: rgba(255, 0, 0, 0.05);
  padding: 10px;
  border-radius: 4px;
  margin: 0;
  margin-top: 10px;
}
.partner-budget-block {
  font-size: 12px;
  margin-top: 10px;
}
.partner-budget-block div {
  line-height: 16px;
}
.partner-budget-block a {
  color: #174d77;
  text-decoration: underline;
}
</style>