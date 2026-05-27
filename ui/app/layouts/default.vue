<template>
  <v-app>
    <v-app-bar :elevation="2" color="teal-darken-4">
      <template v-slot:title>
        <nuxt-link class="cursor-pointer" @click="() => {
          $services.routerService.navigate('/')
        }" style="text-decoration: none;color: #FFFFFF;">
          <v-card-title class="site-title">AUTOTRADE.cloud</v-card-title>
          <v-card-subtitle class="site-subtitle">{{$t('default.cloud_crypto_trade')}}</v-card-subtitle>
        </nuxt-link>
      </template>
      <template v-slot:append>
      <div class="language-selector-container">
        <LanguageSelector :value="locale" :locales="availableLocales"/>
      </div>
        <v-btn icon="mdi-account" @click="clickAccount"></v-btn>
        <v-btn v-if="isLoggedIn" icon="mdi-logout" @click="logout"></v-btn>
      </template>
    </v-app-bar>
    <div style="margin-top: 68px !important; margin-bottom: 22px !important; box-sizing: border-box;">
      <NuxtPage />
    </div>
    <v-bottom-navigation bg-color="teal-darken-4">
      <div class="footer-links">
        <a :href="$t('index.documentation_link')" target="_blank">{{$t('index.documentation')}}</a>
      </div>
    </v-bottom-navigation>
    <v-dialog width="500" v-model="accountDialog">
      <template v-slot:default="{ isActive }">
        <v-card>
          <v-tabs
              v-model="accountDialogTab"
              align-tabs="end"
          >
            <v-tab :value="1">{{$t('default.registration_form')}}</v-tab>
            <v-tab :value="2">{{$t('default.login_form')}}</v-tab>
          </v-tabs>
          <v-window v-model="accountDialogTab">
            <v-window-item
                :value="1"
                class="pt-2"
            >
              <RegistrationForm
                @onRegister="onRegister"
                @onGoNext="accountDialogTab = 2"
                :loading="registrationProcess"
                :error-message="registrationError"
              />
            </v-window-item>
            <v-window-item
                :value="2"
            >
              <v-container fluid>
                <LoginForm
                  @onLogin="onLogin"
                  @onGoBack="accountDialogTab = 1"
                  :loading="loginProcess"
                  :error-message="loginError"
                  :default-email="userEmail"
                />
              </v-container>
            </v-window-item>
          </v-window>

          <v-card-actions>
            <v-spacer></v-spacer>

            <v-btn
                variant="text"
                color="primary"
                :text="$t('default.close_form')"
                @click="closeAccountDialog"
            ></v-btn>
          </v-card-actions>
        </v-card>
      </template>
    </v-dialog>

    <div class="text-center ma-2">
      <v-snackbar
          v-model="snackbar.flag"
          :color="snackbar.color"
      >
        {{ snackbar.text }}

        <template v-slot:actions>
          <v-btn
              color="black"
              variant="text"
              @click="snackbar.flag = false"
          >
            {{$t('default.close_snackbar')}}
          </v-btn>
        </template>
      </v-snackbar>

      <v-dialog
          v-model="confirmationModal.flag"
          max-width="400"
          persistent
      >
        <v-card
            prepend-icon="mdi-information"
            :text="confirmationModal.text"
            :title="confirmationModal.title"
        >
          <template v-slot:actions>
            <v-spacer></v-spacer>

            <v-btn @click="() => {
              confirmationModal.close.callback();
              confirmationModal.flag = false;
            }">
              {{confirmationModal.close.text}}
            </v-btn>

            <v-btn color="primary" variant="flat" @click="() => {
              confirmationModal.ok.callback();
              confirmationModal.flag = false;
            }">
              {{confirmationModal.ok.text}}
            </v-btn>
          </template>
        </v-card>
      </v-dialog>
    </div>
  </v-app>
</template>
<script lang="ts">
import RegistrationForm from '~/components/RegistrationForm.vue';
import LoginForm from '~/components/LoginForm.vue';
import {AlertEvent} from '~/model/alert-event';
import {Alert} from '~/model/alert';

export default defineNuxtComponent({
  components: {LoginForm, RegistrationForm},
  async asyncData(ctx: any) {
    const config = useRuntimeConfig();
    const authToken = ctx.$services.authService.getToken();
    let initialData = {};
    if (authToken) {
      let [{data: cities}, {data: user}] = await Promise.all([
        useFetch('/location/list', {
          method: 'GET',
          baseURL: config.public.baseUrl,
          server: true,
        }),
        useFetch('/v1/user/me', {
          method: 'GET',
          baseURL: config.public.baseUrl,
          server: true,
          headers: {
            Authorization: `Bearer ${authToken}`,
          },
        }),
      ]);

      initialData = {cities, user: user.value};
    } else {
      let [{data: cities}] = await Promise.all([
        useFetch('/location/list', {
          method: 'GET',
          baseURL: config.public.baseUrl,
          server: true,
        }),
      ]);
      initialData = {cities};
    }

    return {
      ...initialData,
      accountDialog: false,
      accountDialogTab: 0,
      registrationProcess: false,
      registrationError: '',
      loginProcess: false,
      loginError: '',
      isLoggedIn: authToken != null,
    }
  },
  data() {
    return {
      openAuthSubscription: null,
      updateUserSubscription: null,
      eventSubscription: null,
      snackbar: {
        flag: false,
        text: '',
        color: '',
      },
      userEmail: this.$services.authService.getEmail(),
      locale: this.$i18n.locale,
      availableLocales: this.$i18n.locales,
      confirmationModal: {
        flag: false,
        title: '',
        text: '',
        close: {
          text: '',
          callback: () => {},
        },
        ok: {
          text: '',
          callback: () => {},
        },
      },
      confirmSubscription: null,
    }
  },
  mounted() {
    const route = useRoute();
    const router = useRouter();

    if (route.query.promocode) {
      this.$services.authService.setPromoCode(route.query.promocode.toString());
      let query = Object.assign({}, route.query);
      delete query.promocode;
      router.replace({query})
    }

    const config = useRuntimeConfig();

    if (this.user) {
      this.$services.authService.setUser(this.user);
    }

    this.updateUserSubscription = this.$services.authService.updateUser.subscribe(() => {
      const token = this.$services.authService.getToken();
      useFetch('/v1/user/me', {
        method: 'GET',
        baseURL: config.public.baseUrl,
        server: true,
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }).then(({data: user}) => {
        this.$services.authService.setUser(user.value);
        this.isLoggedIn = true;
        this.accountDialog = false;
      });
    });
    this.confirmSubscription = this.$services.eventManager.confirmationSubject.subscribe((confirmation) => {
      this.confirmationModal = {
        flag: true,
        title: confirmation.title,
        text: confirmation.text,
        close: {
          text: confirmation.closeBtnText,
          callback: confirmation.closeCallback
        },
        ok: {
          text: confirmation.okBtnText,
          callback: confirmation.okCallback,
        },
      };
    })
    this.openAuthSubscription = this.$services.authService.openAuthDialog.subscribe(() => {
      this.accountDialog = true;
    });
    this.eventSubscription = this.$services.eventManager.alertSubject.subscribe((alert: AlertEvent) => {
      this.snackbar.flag = true;
      this.snackbar.text = alert.alert.text;
      this.snackbar.color = alert.alert.getColor();

      setTimeout(() => {
        this.snackbar = {
          flag: false,
          text: '',
          color: '',
        };
      }, alert.timeout)
    })
  },
  unmounted() {
    if (this.openAuthSubscription) {
      this.openAuthSubscription.unsubscribe();
    }
    if (this.updateUserSubscription) {
      this.updateUserSubscription.unsubscribe();
    }
    if (this.eventSubscription) {
      this.eventSubscription.unsubscribe();
    }
    if (this.confirmSubscription) {
      this.confirmSubscription.unsubscribe();
    }
  },
  methods: {
    closeAccountDialog() {
      this.accountDialog = false;
      this.registrationError = '';
      this.registrationProcess = false;
    },
    processLogin(token: string) {
      this.$services.authService.setToken(token);
      setTimeout(async () => {
        this.$services.authService.updateUser.next();
        await this.$services.routerService.navigate('/account');
      });
    },
    onLogin(data: any) {
      this.loginError = '';
      this.loginProcess = true;
      const config = useRuntimeConfig();

      const email = data.email;

      useFetch('/oauth2/token', {
        method: 'POST',
        baseURL: config.public.baseUrl,
        server: true,
        body: {
          client_id: config.public.clientId,
          client_secret: config.public.clientSecret,
          grant_type: 'password',
          username: email,
          password: data.password
        }
      }).then(({data: token, error: error}: any) => {
        if (error.value) {
          this.loginError = error.value.data.message;
          this.$services.eventManager.alert(
            new AlertEvent(
              new Alert(error.value.data.message, Alert.TYPE_ERROR),
              6000
            )
          );
          setTimeout(() => {
            this.loginError = '';
          }, 10000);
          return;
        }

        this.$services.authService.setEmail(email);
        this.processLogin(token.value.access_token);
      }).finally(() => this.loginProcess = false);
    },
    onRegister(data: any) {
      this.registrationError = '';
      this.registrationProcess = true;
      const config = useRuntimeConfig();

      let email = data.email;

      useFetch('/public/register/code', {
        method: 'POST',
        baseURL: config.public.baseUrl,
        server: true,
        body: {
          ...data,
        }
      }).then(({data: data, error: error}: any) => {
        if (error.value) {
          this.registrationError = error.value.data;
          this.$services.eventManager.alert(
            new AlertEvent(
              new Alert(error.value.data, Alert.TYPE_ERROR),
              6000
            )
          );

          setTimeout(() => {
            this.registrationError = '';
          }, 10000);
          return;
        }

        this.$services.eventManager.alert(
          new AlertEvent(
            new Alert(`${this.$t('default.register_alert1')}${email}${this.$t('default.register_alert2')}`, Alert.TYPE_SUCCESS),
            6000
          )
        );

        this.$services.authService.setEmail(email);
        this.userEmail = email;
        this.accountDialogTab = 2;
      }).finally(() => this.registrationProcess = false);
    },
    clickAccount() {
      if (this.isLoggedIn) {
        setTimeout(async () => {
          await this.$services.routerService.navigate('/account');
        });
        return;
      }

      const email = this.$services.authService.getEmail();

      if (email) {
        this.userEmail = email;

        this.accountDialog = true;
        this.accountDialogTab = 2;
        return;
      }

      this.accountDialog = true;
      this.accountDialogTab = 1;
    },
    logout() {
      this.$services.authService.clearToken();
      this.$services.authService.setUser(null);
      this.isLoggedIn = false;
      setTimeout(async () => {
        await this.$services.routerService.navigate('/');
      });
    },
  },
})
</script>
<style>
* {
  font-family: 'Manrope', serif !important;
}
.footer-links {
  position: absolute;
  left: 15px;
  top: 16px;
}
.footer-links a {
  text-decoration: underline;
  color: #fff;
}
.site-title {
  font-size: 16px;
  margin-bottom: 0;
  padding-bottom: 0;
}
.site-subtitle {
  padding-top: 0;
  margin-top: -5px;
  text-transform: uppercase;
  font-size: 10px;
}
.language-selector-container {
  max-width: 80px;
  margin-right: 20px;
}
@media only screen and (max-width: 700px) {
  .language-selector-container {
    max-width: 40px;
    margin-right: 5px;
  }
}
</style>
