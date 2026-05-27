import vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'
export default defineNuxtConfig({
  ssr: true,
  css: [
    '@mdi/font/css/materialdesignicons.min.css',
    '/public/fonts/Manrope.css',
    '/public/custom.css',
  ],
  app: {
    head: {
      script: [
        {
          src: "https://cdn.jsdelivr.net/npm/luxon@1.26.0",
        },
        {
          src: "https://cdn.jsdelivr.net/npm/chart.js@3.0.1/dist/chart.js",
        },
        {
          src: "/chartjs-adapter.js",
        },
        {
          src: '/chartjs-chart-financial.js',
        },
        {
          src: 'https://code.jivo.ru/widget/iVTuyV7xP2',
          async: true,
        },
      ],
    }
  },
  build: {
    transpile: ['vuetify', 'vue-flag-icon', 'rxjs'],
  },
  modules: [
    '@nuxt/eslint',
    'nuxt-tradingview',
    'nuxt-delay-hydration',
    (_options, nuxt) => {
      nuxt.hooks.hook('vite:extendConfig', (config) => {
        // @ts-expect-error
        config.plugins.push(vuetify({ autoImport: true }))
      })
    },
    '@nuxtjs/i18n',
    //...
  ],
  delayHydration: {
    // enables nuxt-delay-hydration in dev mode for testing
    debug: process.env.NODE_ENV === 'development',
    mode: 'init'
  },
  vite: {
    vue: {
      template: {
        transformAssetUrls,
      },
    },
  },
  nitro: {
    esbuild: {
      options: {
        target: 'esnext'
      }
    },
  },
  // @ts-ignore
  runtimeConfig: {
    public: {
      baseUrl: process.env.API_BASE_URL,
      s3Url: process.env.S3_DSN,
      clientId: process.env.CLIENT_ID,
      clientSecret: process.env.CLIENT_SECRET,
      gtagId: 'GTM-P78QN32D',
    }
  },
  i18n: {
    strategy: 'prefix_except_default',
    locales: [
      {code: 'en', name: 'English'},
      {code: 'ru', name: 'Русский'},
    ],
    defaultLocale: 'en',
    vueI18n: './i18n.config.ts',
    detectBrowserLanguage: {
      useCookie: true,
      alwaysRedirect: true,
      cookieKey: 'i18n_redirected',
      redirectOn: 'root',
      fallbackLocale: 'en'
    },
  },
})
