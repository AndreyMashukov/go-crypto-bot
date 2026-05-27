<template>
  <v-container fluid>
    <div v-if="!getDedicatedServer && getAvailableSlots === 0 && botDetails.status !== 'running'">
      <v-alert
          type="error"
          :title="$t('bot.no_available_slots.title')"
          :text="$t('bot.no_available_slots.text')"
          class="mb-2"
      />
    </div>
    <div v-else>
      <v-alert
          v-if="!!botDetails.errorMessage"
          type="error"
          :title="$t('bot.error_title')"
          :text="botDetails.errorMessage"
          class="mb-2"
      />
      <v-alert
          v-if="isServicePaid"
          class="mt-0"
          type="warning"
          :title="$t('bot.attention_title')"
          :text="$t('bot.attention_text')"
      ></v-alert>
      <v-alert
          v-else
          class="mt-0"
          border="bottom"
          border-color="error"
          :title="$t('bot.payment_req_title')"
          :text="$t('bot.payment_req_text')"
      >
        <template v-slot:append>
          <NuxtLink to="/account" class="text-decoration-none">
            <v-btn variant="flat" color="success">
              {{$t('bot.pay_now_btn')}}
            </v-btn>
          </NuxtLink>
        </template>
      </v-alert>

      <v-btn
          :loading="dataLoading.stop"
          variant="flat"
          color="red"
          @click="stopBot"
          class="mb-4 mt-2 mr-4"
          append-icon="mdi-stop"
          :disabled="botDetails.status !== 'running'"
      >
        <b>{{$t('bot.stop_btn')}}</b>
      </v-btn>
      <v-btn
          :loading="dataLoading.deploy"
          variant="flat"
          color="success"
          @click="deployBot"
          class="mb-4 mt-2"
          append-icon="mdi-play"
          v-if="botDetails.status !== 'running'"
          :disabled="!isServicePaid"
      >
        <b>{{$t('bot.start_btn')}}</b>
      </v-btn>
      <v-btn
          :loading="dataLoading.deploy"
          variant="flat"
          color="success"
          @click="deployBot"
          class="mb-4 mt-2"
          append-icon="mdi-restart"
          v-if="botDetails.status === 'running'"
          :disabled="!isServicePaid"
      >
        <b>{{$t('bot.restart_btn')}}</b>
      </v-btn>
    </div>
    <div>
      <v-alert
        class="mt-2 mb-4"
        border="bottom"
        border-color="info"
      >
        <template v-slot:text>
          {{$t(`bot.documentation.${botDetails.provider}.text`)}} <a class="documentation-link" :href="$t(`bot.documentation.${botDetails.provider}.link.url`)" target="_blank">{{$t(`bot.documentation.${botDetails.provider}.link.text`)}}</a>
        </template>
      </v-alert>
    </div>
    <div class="mb-2">
      <v-chip prepend-icon="mdi-server" color="primary" v-if="!getDedicatedServer">
        {{$t('bot.server.shared').replace('[slots]', getAvailableSlots)}}
      </v-chip>
      <v-chip prepend-icon="mdi-server" color="primary" v-else>
        {{$t('bot.server.dedicated').replace('[ip]', getDedicatedServer.ip)}}
      </v-chip>
    </div>
    <h1>{{$t('bot.inst')}} #{{ botDetails.id }}</h1>
    <v-form @submit.prevent="saveConfig" ref="form" class="mb-4">
      <v-btn :loading="dataLoading.update" variant="flat" color="primary" type="submit" class="mb-4 mt-2 mr-4">{{$t('bot.inst_save_config')}}</v-btn>
      <v-btn :loading="dataLoading.sync" variant="flat" color="warning" class="mb-4 mt-2" append-icon="mdi-sync" @click="syncConfig">{{$t('bot.inst_sync_config')}}</v-btn>
      <v-row class="mt-4">
        <v-col cols="12" lg="6" md="6" sm="6" xs="6" class="mt-0 pt-0">
          <v-text-field
              :label="$t('bot.key_label').replace('[provider]', botDetails.provider)"
              :append-icon="showKey ? 'mdi-eye' : 'mdi-eye-off'"
              :type="showKey ? 'text' : 'password'"
              @click:append="showKey = !showKey"
              v-model="apiKey"
              variant="solo-filled"
              :rules="[rules.required]"
              :hint="$t('bot.key_hint').replace('[get_ips]', getIps).replace('[provider]', botDetails.provider)"
              persistent-hint
              aria-autocomplete="none"
          />
        </v-col>
        <v-col cols="12" lg="6" md="6" sm="6" xs="6" class="mt-0 pt-0">
          <v-text-field
              :label="$t('bot.secret_label').replace('[provider]', botDetails.provider)"
              :append-icon="showSecret ? 'mdi-eye' : 'mdi-eye-off'"
              :type="showSecret ? 'text' : 'password'"
              @click:append="showSecret = !showSecret"
              v-model="apiSecret"
              variant="solo-filled"
              :rules="[rules.required]"
              :hint="$t('bot.secret_hint').replace('[get_ips]', getIps).replace('[provider]', botDetails.provider)"
              persistent-hint
              aria-autocomplete="none"
          />
        </v-col>
      </v-row>

      <v-btn variant="flat" color="primary" size="small" class="mt-4" @click="addSymbol">{{$t('bot.add_symbol')}}</v-btn>
      <v-row>
        <v-col cols="12" lg="6" md="12" sm="12" xs="12" v-for="(tradeConfig, index) in cryptoTradeConfigs" :key="`${tradeConfig.symbol}-${index}`" class="mt-0 pt-0">
          <div class="pt-4">
            <h3>{{ tradeConfig.symbol }}</h3>
            <v-btn size="small" color="red" class="mt-1" @click="deleteSymbol(tradeConfig, index)">{{$t('bot.delete_btn')}}</v-btn>
            <v-btn size="small" color="info" class="ml-1 mt-1" @click="applyForAll(index)">{{$t('bot.apply_btn')}}</v-btn>
            <v-switch
                v-model="tradeConfig.enabled"
                color="primary"
                :rules="[]"
                :label="$t('bot.add_symbol_label')"
                hide-details
            ></v-switch>
            <v-row class="mt-4">
              <v-col cols="6" lg="3" md="3" sm="3" xs="3" class="mt-0 pt-0">
                <v-select
                    :label="$t('bot.symbol_label')"
                    v-model="tradeConfig.symbol"
                    variant="solo-filled"
                    :rules="[rules.required]"
                    :items="symbols"
                    type="text"
                    :hint="$t('bot.symbol_hint').replace('[provider]', botDetails.provider)"
                    persistent-hint
                    class="can-change-by-user"
                />
              </v-col>
              <v-col cols="6" lg="3" md="3" sm="3" xs="3" class="mt-0 pt-0">
                <v-text-field
                    :label="$t('bot.budget_label')"
                    v-model="tradeConfig.usdtLimit"
                    variant="solo-filled"
                    :rules="[rules.required, rules.min15]"
                    type="number"
                    :hint="$t('bot.budget_hint')"
                    persistent-hint
                    class="can-change-by-user"
                />
              </v-col>
              <v-col cols="6" lg="3" md="3" sm="3" xs="3" class="mt-0 pt-0">
                <v-btn
                  append-icon="mdi-cog"
                  variant="flat"
                  color="success"
                  @click="configSettingsDialog(tradeConfig, index)"
                >
                  {{$t('bot.buy_sell_btn')}}
                </v-btn>
              </v-col>
            </v-row>
            <v-expansion-panels class="mt-4">
              <v-expansion-panel
                  :title="$t('bot.tech_setting')"
              >
                <v-expansion-panel-text class="pt-4">
                  <v-row>
                    <v-col cols="6" lg="3" md="3" sm="3" xs="3" class="mt-0 pt-0">
                      <v-text-field
                          :label="$t('bot.min_price_label')"
                          v-model="tradeConfig.minPriceMinutesPeriod"
                          variant="solo-filled"
                          :rules="[rules.required]"
                          type="number"
                          :hint="$t('bot.min_price_hint')"
                          persistent-hint
                          class="change-only-if-know"
                      />
                    </v-col>
                    <v-col cols="6" lg="3" md="3" sm="3" xs="3" class="mt-0 pt-0">
                      <v-select
                          :label="$t('bot.frame_interval_label')"
                          v-model="tradeConfig.frameInterval"
                          variant="solo-filled"
                          :rules="[rules.required]"
                          :items="['1m', '15m', '30m', '1h', '2h', '4h', '6h', '12h', '1d', '1w', '1M']"
                          type="text"
                          :hint="$t('bot.frame_interval_hint')"
                          persistent-hint
                          class="change-only-if-know"
                      />
                    </v-col>
                    <v-col cols="6" lg="3" md="3" sm="3" xs="3" class="mt-0 pt-0">
                      <v-text-field
                          :label="$t('bot.frame_period_label')"
                          v-model="tradeConfig.framePeriod"
                          variant="solo-filled"
                          :rules="[rules.required, rules.max200]"
                          type="number"
                          :hint="$t('bot.frame_period_hint')"
                          persistent-hint
                          class="change-only-if-know"
                      />
                    </v-col>
                    <v-col cols="6" lg="3" md="3" sm="3" xs="3" class="mt-0 pt-0">
                      <v-select
                          :label="$t('bot.check_interval_label')"
                          v-model="tradeConfig.buyPriceHistoryCheckInterval"
                          variant="solo-filled"
                          :rules="[rules.required]"
                          :items="['1s', '1m', '15m', '30m', '1h', '2h', '4h', '6h', '8h', '12h', '1d', '3d', '1w']"
                          type="text"
                          :hint="$t('bot.check_interval_hint')"
                          persistent-hint
                          class="change-only-if-know"
                      />
                    </v-col>
                    <v-col cols="6" lg="3" md="3" sm="3" xs="3" class="mt-0 pt-0">
                      <v-text-field
                          :label="$t('bot.check_period_label')"
                          v-model="tradeConfig.buyPriceHistoryCheckPeriod"
                          variant="solo-filled"
                          :rules="[rules.required, rules.max100]"
                          type="number"
                          :hint="$t('bot.check_period_hint')"
                          persistent-hint
                          class="change-only-if-know"
                      />
                    </v-col>
                  </v-row>
                </v-expansion-panel-text>
              </v-expansion-panel>
            </v-expansion-panels>
          </div>
        </v-col>
      </v-row>
      <v-btn :loading="dataLoading.update" v-if="cryptoTradeConfigs.length >= 2" variant="flat" color="primary" type="submit" class="mt-6">{{$t('bot.save_config_btn')}}</v-btn>
    </v-form>
    <v-dialog
        persistent
        v-model="settingsDialog.flag"
        max-width="500px"
        min-width="380px"
        z-index="9999"
        min-height="500px"
        :scrollable="true"
    >
      <div class="position-relative" style="width: 100%; z-index: 999;">
        <v-btn icon variant="text" style="top: 0;right:0;position: absolute" @click="closeConfigSettings">
          <v-icon icon="mdi-close"/>
        </v-btn>
      </div>
      <v-card v-if="settingsDialog.flag">
        <v-card-title class="pb-0">{{$t('bot.config')}} <small>{{ settingsDialog.config.symbol }}</small></v-card-title>
        <v-tabs
            v-model="settingsDialog.tab"
            bg-color="default"
        >
          <v-tab v-for="(tab, index) in settingsDialogTabs" :value="index" :key="`tab-${index}`">{{ tab.name }}</v-tab>
        </v-tabs>
        <v-window v-model="settingsDialog.tab" class="mt-4" style="overflow-y: scroll !important;">
          <v-window-item value="0">
            <v-form @submit.prevent="saveProfitOptions" ref="form3" class="mt-2">
              <v-card-text>
                <v-row
                  v-for="(profitOption, profitOptionIndex) in settingsDialog.profitOptions"
                  class="mb-0 position-relative"
                  :key="`profit-${profitOptionIndex}`"
                  :style="{'background-color': profitOption.isTriggerOption ? 'rgba(153,245,150,0.77)' : '#FFFFFF'}"
                >
                  <v-col cols="1" class="p-0">
                    <v-icon :icon="`mdi-numeric-${profitOptionIndex+1}-box`" color="primary"/>
                    <v-checkbox color="primary" v-model="profitOption.isTriggerOption" @change="triggerOptionChanged(profitOptionIndex)" class="trigger-option"/>
                  </v-col>
                  <v-col cols="3" class="p-0 py-0">
                    <v-text-field
                        :label="$t('bot.trigger_options_label_1')"
                        type="number"
                        v-model="profitOption.optionValue"
                        variant="underlined"
                        :rules="[rules.required, rules.positive]"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="3" class="p-0 py-0">
                    <v-select
                        :label="$t('bot.trigger_options_label_2')"
                        type="text"
                        v-model="profitOption.optionUnit"
                        :rules="[rules.required]"
                        :items="profitPeriodLabels"
                        variant="underlined"
                    ></v-select>
                  </v-col>
                  <v-col cols="4" class="p-0 py-0">
                    <v-text-field
                        :label="$t('bot.trigger_options_label_3')"
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

                <v-btn size="x-small" class="mb-2" color="primary" @click="settingsDialog.profitOptions.push({optionUnit: 'h', optionValue: 1, optionPercent: 0.5, isTriggerOption: false})">{{$t('bot.setting_dialog_btn')}}</v-btn>
                <v-divider class="mb-2"/>
              </v-card-text>

              <v-alert
                  type="info"
                  color="rgba(153,245,150,0.77)"
                  :text="$t('bot.setting_dialog_alert_text')"
                  class="mx-2"
              />

              <v-card-actions class="justify-end">
                <v-btn variant="flat" color="default" class="action" @click="closeConfigSettings">{{$t('bot.setting_cansel_btn')}}</v-btn>
                <v-btn variant="flat" color="success" type="submit" class="action w-33">{{$t('bot.setting_apply_btn')}}</v-btn>
              </v-card-actions>
            </v-form>
          </v-window-item>
          <v-window-item value="1">
            <v-form @submit.prevent="saveExtraCharge" ref="form2" class="mt-2">
              <v-card-text>
                <v-row v-for="(chargeOption, optionIndex) in settingsDialog.extraChargeOptions" class="mb-0" :key="`charge-${optionIndex}`">
                  <v-col cols="1" class="p-0">
                    <v-icon :icon="`mdi-numeric-${optionIndex+1}-box`" color="primary"/>
                  </v-col>
                  <v-col cols="5" class="p-0 py-0">
                    <v-text-field
                        :label="$t('bot.save_extra_charge_label_1')"
                        type="number"
                        v-model="chargeOption.percent"
                        variant="underlined"
                        :rules="[rules.required, rules.negative]"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="5" class="p-0 py-0">
                    <v-text-field
                        :label="$t('bot.save_extra_charge_label_2')"
                        type="number"
                        v-model="chargeOption.amountUsdt"
                        variant="underlined"
                        :rules="[rules.required, rules.min15]"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="1" class="p-0">
                    <v-icon icon="mdi-close" color="red" style="margin-left: -10px" @click="settingsDialog.extraChargeOptions.splice(optionIndex, 1)"/>
                  </v-col>
                </v-row>

                <v-btn size="x-small" class="mb-2" color="primary" @click="settingsDialog.extraChargeOptions.push({percent: 0.00, amountUsdt: 0.00})">{{$t('bot.setting_dialog_btn')}}</v-btn>
                <v-divider class="mb-2"/>
              </v-card-text>

              <v-card-actions class="justify-end">
                <v-btn variant="flat" color="default" class="action" @click="closeConfigSettings">{{$t('bot.setting_cansel_btn')}}</v-btn>
                <v-btn variant="flat" color="success" type="submit" class="action w-33">{{$t('bot.setting_apply_btn')}}</v-btn>
              </v-card-actions>
            </v-form>
          </v-window-item>
        </v-window>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script lang="ts">
import {AlertEvent} from '~/model/alert-event';
import {Alert} from '~/model/alert';
import {TradeLimit} from '~/model/trade-limit';

export default defineNuxtComponent({
  name: "[id].vue",
  async asyncData(ctx: any) {
    const {t} = useI18n();
    const authToken = ctx.$services.authService.getToken();
    if (!authToken) {
      await ctx.$services.routerService.navigate('/');
      return;
    }

    const route = useRoute();
    const config = useRuntimeConfig();

    const [{data: user}, {data: botDetails}, {data: serverList}, {data: symbolList}] = await Promise.all([
      useFetch('/v1/user/me', {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
      }),
      useFetch(`/v1/cryptobot/${route.params.id}`, {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
      }),
      useFetch(`/v1/cryptobot/${route.params.id}/server/list`, {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
      }),
      useFetch(`/v1/cryptobot/${route.params.id}/symbol/list`, {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
      }),
    ]);

    if (!botDetails.value) {
      // todo: 404 error...
      await ctx.$services.routerService.navigate('/');
      return;
    }

    useHead({
      title: t('bot.header').replace('[botDetails.value.id]', botDetails.value.id),
    })

    return {
      botDetails: botDetails.value,
      userVm: user.value,
      serverList: serverList.value,
      symbolList: (symbolList.value || []).map((x) => x.symbol),
    };
  },
  mounted() {
  },
  data(): any {
    return {
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
      settingsDialog: {
        flag: false,
        config: null,
        symbol: '',
        extraChargeOptions: [],
        profitOptions: [],
        index: 0,
        tab: 0,
      },
      showKey: false,
      showSecret: false,
      apiKey: this.botDetails.apiKey,
      apiSecret: this.botDetails.apiSecret,
      rules: {
        required: (value: any) => !!value || this.$t('bot.data_required'),
        max200  : (value: any) => value <= 200 || this.$t('bot.data_max200'),
        max100  : (value: any) => value <= 100 || this.$t('bot.data_max100'),
        min15   : (value: any) => value >= 15 || this.$t('bot.data_min15'),
        negative: (value: any) => value < 0.00 || this.$t('bot.data_negative'),
        positive: (value: any) => value > 0.00 || this.$t('bot.data_positive'),
        min05: (value: any) => value >= 0.5 || this.$t('bot.data_min05'),
      },
      newCryptoTradeConfigs: [],
      dataLoading: {
        deploy: false,
        update: false,
        sync: false,
        stop: false,
      },
      symbols: this.symbolList,
    }
  },
  computed: {
    getAvailableSlots() {
      let available = 0;
      this.serverList.servers.forEach((x) => {
        available += Number(x.available);
      });

      return Math.max(0, available);
    },
    getDedicatedServer() {
      if (!this.botDetails.hasDedicatedServer) {
        return null;
      }

      return this.botDetails.dedicated;
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
    getIps: {
      get(): string {
        return this.serverList.list.map((x) => x.ip).join(', ');
      }
    },
    isServicePaid: {
      get() {
        return this.userVm.hasActiveBasicSubscription || this.userVm.budget >= 1.00;
      }
    },
    cryptoTradeConfigs: {
      get() {
        const additional = this.newCryptoTradeConfigs;

        return this.botDetails.cryptoTradeConfigs.concat(additional);
      },
    },
  },
  methods: {
    triggerOptionChanged(profitOptionIndex) {
      this.settingsDialog.profitOptions.forEach((option, index) => {
        if (index !== profitOptionIndex) {
          option.isTriggerOption = false;
        }
      })
    },
    applyForAll(index) {
      const config = this.cryptoTradeConfigs[index];

      this.newCryptoTradeConfigs.forEach((element, elementIndex) => {
        this.newCryptoTradeConfigs[elementIndex] = {
          ...config,
          symbol: element.symbol,
          id: element.id,
        }
      });

      this.botDetails.cryptoTradeConfigs.forEach((element, elementIndex) => {
        this.botDetails.cryptoTradeConfigs[elementIndex] = {
          ...config,
          symbol: element.symbol,
          id: element.id,
        }
      });
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
    saveProfitOptions() {
      this.$refs.form3.validate().then(({valid}: any) => {
        this.settingsDialog.profitOptions.sort((a, b) => this.getNumericTime(a) > this.getNumericTime(b) ? 1 : -1)

        if (!valid) {
          return;
        }

        const hasTrigger = this.settingsDialog.profitOptions.filter((x) => x.isTriggerOption).length > 0;

        let index = 0;
        const indexedOptions = this.settingsDialog.profitOptions.map((option) => {
          const indexVal = index
          ++index;

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
        });

        let elementIndex = this.settingsDialog.index;
        if (!this.settingsDialog.config.id) {
          elementIndex = elementIndex - this.botDetails.profitOptions.length;
          this.newCryptoTradeConfigs[elementIndex].profitOptions = indexedOptions;
        }

        this.botDetails.cryptoTradeConfigs[elementIndex].profitOptions = indexedOptions;
        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(`${this.$t('bot.save_config_success')}`, Alert.TYPE_SUCCESS),
            4000
          )
        );
      });
    },
    saveExtraCharge() {
      this.$refs.form2.validate().then(({valid}: any) => {
        this.settingsDialog.extraChargeOptions.sort((a, b) => Number(a.percent) < Number(b.percent) ? 1 : -1)

        if (!valid) {
          return;
        }

        let index = 0;
        const indexedOptions = this.settingsDialog.extraChargeOptions.map((option) => {
          let indexVal = index
          ++index;

          return {
            amountUsdt: Number(option.amountUsdt),
            percent: Number(option.percent),
            index: indexVal,
          };
        });

        let elementIndex = this.settingsDialog.index;
        if (!this.settingsDialog.config.id) {
          elementIndex = elementIndex - this.botDetails.cryptoTradeConfigs.length;
          this.newCryptoTradeConfigs[elementIndex].extraChargeOptions = indexedOptions;
        }

        this.botDetails.cryptoTradeConfigs[elementIndex].extraChargeOptions = indexedOptions;
        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(`${this.$t('bot.save_config_success')}`, Alert.TYPE_SUCCESS),
            4000
          )
        );
      });
    },
    configSettingsDialog(config, index) {
      let options = [];
      (config.extraChargeOptions || []).forEach((option) => {
        options.push(option);
      });
      let profitOptions = [];
      (config.profitOptions || []).forEach((option) => {
        profitOptions.push({...option, isTriggerOption: Boolean(option.isTriggerOption)});
      });

      this.settingsDialog = {
        flag: true,
        config: config,
        symbol: config.symbol,
        index: index,
        extraChargeOptions: options,
        profitOptions: profitOptions,
        tab: 0,
      };
    },
    closeConfigSettings() {
      this.settingsDialog = {
        flag: false,
        config: null,
        symbol: '',
        index: 0,
        extraChargeOptions: [],
        profitOptions: [],
        tab: 0,
      };
    },
    deployBot() {
      this.dataLoading.deploy = true;
      const config = useRuntimeConfig();
      useFetch(`/v1/cryptobot/${this.botDetails.id}/deploy`, {
        method: 'PUT',
        baseURL: config.public.baseUrl,
        server: false,
        body: {},
        headers: {
          Authorization: `Bearer ${this.$services.authService.getToken()}`,
        },
      }).then(({data: data, error: error}: any) => {
        if (error.value) {
          let message = error.value.data

          if (error.value.data.message) {
            message = error.value.data.message
          }

          this.$services.eventManager.alert(
            new AlertEvent(
              new Alert(message, Alert.TYPE_ERROR),
              4000
            )
          );

          return;
        }

        window.location.reload();
      }).finally(() => this.dataLoading.deploy = false);
    },
    addSymbol() {
      this.newCryptoTradeConfigs.push({
        symbol: "BTCUSDT",
        usdtLimit: 15.00,
        enabled: false,
        minPriceMinutesPeriod: 200,
        frameInterval: '2h',
        framePeriod: 20,
        buyPriceHistoryCheckInterval: '1d',
        buyPriceHistoryCheckPeriod: 14,
        extraChargeOptions: [],
        profitOptions: [
          {
            index: 0,
            optionValue: 1,
            optionUnit: 'h',
            optionPercent: 2.50,
          }
        ],
      });

      this.$forceUpdate();
    },
    deleteSymbol(tradeConfig: any, index: number) {
      if (!tradeConfig.id) {
        index = index - this.botDetails.cryptoTradeConfigs.length;
        this.newCryptoTradeConfigs.splice(index, 1);
        return;
      }

      this.botDetails.cryptoTradeConfigs.splice(index, 1);
    },
    stopBot() {
      const config = useRuntimeConfig();
      this.dataLoading.stop = true;
      useFetch(`/v1/cryptobot/${this.botDetails.id}/stop`, {
        method: 'PUT',
        baseURL: config.public.baseUrl,
        server: false,
        headers: {
          Authorization: `Bearer ${this.$services.authService.getToken()}`,
        },
      }).then(({data: data, error: error}: any) => {
        window.location.reload();
      });
    },
    syncConfig() {
      const config = useRuntimeConfig();
      this.dataLoading.sync = true;
      useFetch(`/v1/cryptobot/${this.botDetails.id}/config/sync`, {
        method: 'PUT',
        baseURL: config.public.baseUrl,
        server: false,
        headers: {
          Authorization: `Bearer ${this.$services.authService.getToken()}`,
        },
      }).then(({data: data, error: error}: any) => {
        window.location.reload();
      });
    },
    saveConfig() {
      this.$refs.form.validate().then(({valid}: any) => {
        if (!valid) {
          this.$services.eventManager.alert(
            new AlertEvent(
              new Alert(`${this.$t('bot.save_config_alert')}`, Alert.TYPE_ERROR),
              8000
            )
          );

          return;
        }

        try {
          this.cryptoTradeConfigs.forEach((config) => {
            if (config.profitOptions.length === 0) {
              throw new Error(`${this.$t('bot.save_config_new_error').replace('[config.symbol]', config.symbol)}`);
            }
          });
        } catch (e) {
          this.$services.eventManager.alert(
              new AlertEvent(
                  new Alert(e, Alert.TYPE_ERROR),
                  8000
              )
          );

          return;
        }

        this.dataLoading.update = true
        const config = useRuntimeConfig();
        useFetch(`/v1/cryptobot/${this.botDetails.id}`, {
          method: 'PATCH',
          baseURL: config.public.baseUrl,
          server: false,
          body: {
            apiKey: this.apiKey,
            apiSecret: this.apiSecret,
            cryptoTradeConfigs: this.cryptoTradeConfigs.map((config: any) => {
              return {
                symbol: config.symbol,
                usdtLimit: config.usdtLimit,
                enabled: config.enabled,
                minPriceMinutesPeriod: config.minPriceMinutesPeriod,
                frameInterval: config.frameInterval,
                framePeriod: config.framePeriod,
                buyPriceHistoryCheckInterval: config.buyPriceHistoryCheckInterval,
                buyPriceHistoryCheckPeriod: config.buyPriceHistoryCheckPeriod,
                extraChargeOptions: config.extraChargeOptions,
                profitOptions: config.profitOptions,
              };
            })
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
            this.dataLoading.update = false
            return;
          }

          window.location.reload();
          this.dataLoading.update = false
        });
      });
    }
  },
})
</script>

<style>
.can-change-by-user .v-field {
  background-color: #1DE9B6FF !important;
}
.change-only-if-know .v-field {
  background-color: rgba(244,67,54,0.67) !important;
}
.trigger-option {
  margin: 0;
  padding: 0;
  position: absolute;
  top: 20px;
  left: 2.5px;
}
</style>