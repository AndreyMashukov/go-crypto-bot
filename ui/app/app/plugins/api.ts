// $api — single $fetch instance every Pinia store + page uses.
// In closed perimeter: no Bearer auth, no token refresh. Every request
// hits Nuxt's own /api/* proxy which forwards to ui-server (Symfony),
// which in turn either serves /api/prometheus/* itself or proxies
// /api/bot/* into the market-trader admin HTTP server.
//
// The bot's HTTP controllers require ?botUuid=<uuid> on every call;
// the interceptor stamps it from runtimeConfig.public.botUuid so
// callers can pass clean paths like "/bot/health/check".

export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig()

  const api = $fetch.create({
    baseURL: config.public.apiBase,
    retry: false,
    onRequest: ({ request, options }) => {
      const url = typeof request === 'string' ? request : (request as URL).pathname
      if (url.startsWith('/bot/')) {
        const existing = options.query ?? {}
        options.query = { botUuid: config.public.botUuid, ...existing }
      }
    },
  })

  return {
    provide: { api },
  }
})
