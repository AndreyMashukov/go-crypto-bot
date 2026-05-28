<template>
  <v-container fluid>
    <v-form ref="form" @submit.prevent="signUp">
      <v-row>
        <v-col cols="12" lg="6" md="6" sm="6" xs="6" class="mt-0 pt-0">
          <v-text-field
              v-model="nickname"
              :label="$t('registration_form.nickname_label')"
              variant="solo-filled"
              :rules="[rules.required, rules.nickname]"
          />
        </v-col>
        <v-col cols="12" lg="6" md="6" sm="6" xs="6" class="mt-0 pt-0">
          <v-text-field
              v-model="email"
              :label="$t('registration_form.email_label')"
              variant="solo-filled"
              :rules="[rules.required, rules.email]"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12" lg="6" md="6" sm="6" xs="6" class="mt-0 pt-0">
          <v-text-field
              v-model="promocode"
              :label="$t('registration_form.promocode')"
              variant="solo-filled"
              :error="promocodeStatus.error"
              :error-messages="!!promocodeStatus.error ? [promocodeStatus.error] : []"
              :persistent-hint="true"
              :hint="promocodeStatus.valid ? $t('registration_form.promocode_is_valid') : ''"
              :bg-color="promocodeStatus.valid ? 'success' : ''"
              :loading="promocodeStatus.loading"
              @update:model-value="checkPromoCode"
          />
        </v-col>
      </v-row>
      <div class="text-center">
        {{$t('registration_form.existing_account')}} <a class="text-blue text-decoration-underline" @click="onGoNext">{{$t('registration_form.login')}}</a>
      </div>
      <v-btn
          :disabled="loading"
          :loading="loading"
          type="submit"
          class="mt-4"
          width="100%"
          variant="flat"
          color="primary"
          :text="$t('registration_form.text_btn')"
      />
      <v-alert
          v-if="!!errorMessage"
          :title="$t('registration_form.alert_title')"
          :text="errorMessage"
          type="error"
          class="mt-4"
      />
    </v-form>
  </v-container>
</template>
<script lang="ts">
export default {
  name: "RegistrationForm",
  props: {
    loading: {
      type: Boolean,
      default: () => false,
    },
    errorMessage: {
      type: String,
      default: () => '',
    },
  },
  setup(props: any, {emit}: any) {
    function onRegister(data: any) {
      emit('onRegister', data);
    }
    function onGoNext() {
      emit('onGoNext', {});
    }

    return {
      onRegister,
      onGoNext,
    }
  },
  data(): any {
    const promocode = this.$services.authService.getPromoCode();

    return {
      promocodeStatus: {
        error: null,
        valid: false,
        loading: false,
      },
      promocodeTimer: null,
      promocode,
      nickname: '',
      email: '',
      rules: {
        nickname: (value: any) => value.match(/^[a-z0-9]+$/ui) || this.$t('registration_form.nickname_error'),
        email: (value: any) => value.match(/^[a-z0-9\._]+@[a-z0-9-\.]+\.[a-z]+$/ui) || this.$t('registration_form.email_error'),
        required: (value: any) => !!value || this.$t('registration_form.field_error'),
      },
    }
  },
  methods: {
    signUp() {
      this.$refs.form.validate().then(({valid}: any) => {
        if (!valid) {
          return;
        }

        if (this.promocode) {
          this.verifyPromoCode(this.promocode).then(() => {
            this.onRegister({
              email: this.email,
              nickname: this.nickname,
              secret: 'sfsdfsfdsfdkfkrnfsgninrtigni4ilgbvlirdtblrjdbnjdlrngjhdrbjldflgj',
              promoCode: this.promocode.toUpperCase(),
            })
          });
        } else {
          this.onRegister({
            email: this.email,
            nickname: this.nickname,
            secret: 'sfsdfsfdsfdkfkrnfsgninrtigni4ilgbvlirdtblrjdbnjdlrngjhdrbjldflgj',
          });
        }
      });
    },
    checkPromoCode(value: string) {
      this.promocodeStatus.valid = false;
      this.promocodeStatus.error = null;

      if (this.promocodeTimer) {
        clearTimeout(this.promocodeTimer);
      }

      if (!value) {
        return;
      }

      this.promocodeTimer = setTimeout(async () => {
        await this.verifyPromoCode(value);
      }, 2000);
    },
    verifyPromoCode(code: string): Promise<boolean> {
      this.promocodeStatus.loading = true;
      return this.$services.httpClient.publicGetClient(`/public/promocode/${code}/test`).then(() => {
        this.promocodeStatus.valid = true;
        this.promocodeStatus.error = null;
        return Promise.resolve(true);
      }).catch((error: Error) => {
        this.promocodeStatus.valid = false;
        this.promocodeStatus.error = error.message;
        return Promise.reject(false);
      }).finally(() => this.promocodeStatus.loading = false)
    },
  }
}
</script>

<style scoped>

</style>