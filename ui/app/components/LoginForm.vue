<template>
  <v-container fluid>
    <v-form ref="form" @submit.prevent="signIn">
      <v-row>
        <v-col cols="12" lg="6" md="6" sm="6" xs="12" class="mt-0 pt-0">
          <v-text-field
              v-model="email"
              :label="$t('login_form.email_label')"
              variant="solo-filled"
              :rules="[rules.required, rules.email]"
          />
        </v-col>
        <v-col cols="12" lg="6" md="6" sm="6" xs="12" class="mt-0 pt-0">
          <v-text-field
              v-model="password"
              :label="$t('login_form.code_label')"
              variant="solo-filled"
              :rules="[rules.required]"
              type="password"
          />
        </v-col>
      </v-row>
      <div class="text-center">
        {{$t('login_form.forgotten_code')}} <a class="text-blue text-decoration-underline" @click="onGoBack">{{$t('login_form.get_new_code')}}</a>
      </div>
      <v-btn
          :disabled="loading"
          :loading="loading"
          type="submit"
          class="mt-4"
          width="100%"
          variant="flat"
          color="primary"
          :text="$t('login_form.text_btn')"
      />
      <v-alert
          v-if="!!errorMessage"
          :title="$t('login_form.alert_title')"
          :text="errorMessage"
          type="error"
          class="mt-4"
      />
    </v-form>
  </v-container>
</template>
<script lang="ts">
export default {
  name: "LoginForm",
  props: {
    defaultEmail: {
      type: String,
      default: () => '',
    },
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
    function onLogin(data: any) {
      emit('onLogin', data);
    }
    function onGoBack() {
      emit('onGoBack', {});
    }

    return {
      onLogin,
      onGoBack,
    }
  },
  data () {
    return {
      email: this.defaultEmail,
      password: '',
      rules: {
        email: (value: any) => value.match(/^[a-z0-9\._]+@[a-z0-9-\.]+\.[a-z]+$/ui) || this.$t('login_form.email_error'),
        required: (value: any) => !!value || this.$t('login_form.field_error'),
      },
    }
  },
  methods: {
    signIn() {
      this.$refs.form.validate().then(({valid}: any) => {
        if (!valid) {
          return;
        }

        this.onLogin({
          email: this.email,
          password: this.password,
        })
      });
    }
  }
}
</script>

<style scoped>

</style>