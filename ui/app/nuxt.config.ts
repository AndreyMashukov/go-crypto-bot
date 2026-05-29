import vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  ssr: true,
  devtools: { enabled: process.env.NODE_ENV !== 'production' },
  css: [
    '@mdi/font/css/materialdesignicons.min.css',
  ],
  app: {
    head: {
      title: 'go-crypto-bot — admin',
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'robots', content: 'noindex, nofollow' },
      ],
    },
  },
  build: {
    transpile: ['vuetify'],
  },
  modules: [
    '@pinia/nuxt',
    '@nuxtjs/i18n',
    (_options, nuxt) => {
      nuxt.hooks.hook('vite:extendConfig', (config) => {
        config.plugins ||= []
        // @ts-expect-error vite plugin typing
        config.plugins.push(vuetify({ autoImport: true }))
      })
    },
  ],
  i18n: {
    defaultLocale: 'ru',
    strategy: 'prefix_except_default',
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: 'i18n_locale',
      fallbackLocale: 'ru',
      redirectOn: 'root',
    },
    lazy: true,
    locales: [
      { code: 'ru',      language: 'ru-RU', name: 'Русский',  file: 'ru.json' },
      { code: 'en',      language: 'en-US', name: 'English',  file: 'en.json' },
      { code: 'de',      language: 'de-DE', name: 'Deutsch',  file: 'de.json' },
      { code: 'fr',      language: 'fr-FR', name: 'Français', file: 'fr.json' },
      { code: 'es',      language: 'es-ES', name: 'Español', file: 'es.json' },
      { code: 'ja',      language: 'ja-JP', name: '日本語',    file: 'ja.json' },
      { code: 'zh-hans', language: 'zh-CN', name: '简体中文',   file: 'zh-hans.json' },
      { code: 'ko',      language: 'ko-KR', name: '한국어',    file: 'ko.json' },
    ],
  },
  vite: {
    vue: {
      template: {
        transformAssetUrls,
      },
    },
  },
  runtimeConfig: {
    backend: {
      apiBase: process.env.BACKEND_API_BASE ?? 'http://ui-server/api',
    },
    public: {
      apiBase: '/api',
      botUuid: process.env.BOT_UUID ?? '00000000-0000-0000-0000-000000000001',
    },
  },
})
