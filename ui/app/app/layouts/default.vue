<template>
  <v-app>
    <v-app-bar :elevation="2" color="teal-darken-4">
      <template #prepend>
        <v-app-bar-nav-icon @click="drawer = !drawer" />
      </template>
      <template #title>
        <nuxt-link class="bot-title" :to="localePath('/dashboard')">
          <span class="site-title">{{ t('app.title') }}</span>
          <span class="site-subtitle">{{ t('app.subtitle') }}</span>
        </nuxt-link>
      </template>
      <template #append>
        <LanguageSelector />
      </template>
    </v-app-bar>

    <v-navigation-drawer v-model="drawer" :width="240">
      <v-list density="comfortable" nav>
        <v-list-item :to="localePath('/dashboard')" prepend-icon="mdi-view-dashboard" :title="t('nav.dashboard')" />
        <v-list-item :to="localePath('/orders')"    prepend-icon="mdi-format-list-bulleted" :title="t('nav.orders')" />
        <v-list-item :to="localePath('/config')"    prepend-icon="mdi-cog" :title="t('nav.config')" />
        <v-list-item :to="localePath('/charts')"    prepend-icon="mdi-chart-line" :title="t('nav.charts')" />
      </v-list>
    </v-navigation-drawer>

    <v-main>
      <slot />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const { t } = useI18n()
const localePath = useLocalePath()
const drawer = ref(true)
</script>

<style>
.bot-title {
  text-decoration: none;
  color: #fff;
  display: flex;
  flex-direction: column;
  line-height: 1.1;
}
.site-title {
  font-size: 16px;
  font-weight: 600;
}
.site-subtitle {
  font-size: 10px;
  text-transform: uppercase;
  opacity: 0.85;
}
</style>
