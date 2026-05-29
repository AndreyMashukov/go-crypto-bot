<script setup lang="ts">
import { computed } from 'vue'

const { locale, locales } = useI18n()
const switchLocalePath = useSwitchLocalePath()

const currentLocaleName = computed(() => {
  const entry = locales.value.find(l => l.code === locale.value)
  return entry?.name ?? locale.value
})
</script>

<template>
  <v-menu transition="slide-y-transition">
    <template #activator="{ props }">
      <v-btn color="primary" v-bind="props" rounded="0" class="language-selector">
        <span class="language-desktop">{{ currentLocaleName }}</span>
        <v-icon end>mdi-translate</v-icon>
      </v-btn>
    </template>
    <v-list density="comfortable">
      <v-list-item
        v-for="lang in locales"
        :key="lang.code"
        :to="switchLocalePath(lang.code)"
        :active="lang.code === locale"
      >
        <v-list-item-title>{{ lang.name }}</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<style scoped>
.language-desktop {
  color: #fff !important;
  margin-right: 6px;
}
@media (max-width: 700px) {
  .language-desktop { display: none }
}
</style>
