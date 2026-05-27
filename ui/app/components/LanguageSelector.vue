<template>
  <v-menu transition="slide-y-transition">
    <template v-slot:activator="{ props }">
      <v-btn color="primary" v-bind="props" :prepend-icon="getLanguageIcon" rounded="0" class="language-selector">
        <span class="language-desktop">{{ actualLocale }}</span>
      </v-btn>
    </template>
    <v-list>
      <v-list-item class="language-selector-item" @click="actualLocale = language.code" v-for="(language, name) in availableLocales" :key="name" :prepend-icon="flagMap[language.code]">
        <v-list-item-title class="language-desktop-selector">
          {{ language.name }}
        </v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script>
import England from "~/components/icons/England.vue";
import Russia from "~/components/icons/Russia.vue";
export default {
  name: "LanguageSelector",
  data() {
    return {
      availableLocales: this.locales,
      flagMap: {
        en: England,
        ru: Russia,
      }
    };
  },
  computed: {
    getLanguageIcon() {
      return this.flagMap[this.$i18n.locale];
    },
    actualLocale: {
      get() {
        return this.$i18n.locales.find(language => language.code === this.$i18n.locale)?.name;
      },
      set(locale) {
        this.$i18n.setLocale(locale);
      }
    },
  },
  props: {
    value: {
      type: String,
      default: () => '',
    },
    locales: {
      type: Array,
      default: () => [],
    },
  },
}
</script>

<style>
v-btn {
  display: flex;
  justify-content: center;
  align-items: center;
  text-align: center;
}
.language-desktop {
  color: #FFFFFF !important;
}
.language-desktop-selector {
  color: #000000 !important;
  font-weight: 500;
}
.language-selector {
}
@media only screen and (max-width: 700px) {
  .language-desktop {
    display: none;
  }
  .language-selector {
    max-width: 30px;
  }
  .language-selector-item {
    max-width: 30px;
  }
}
</style>
