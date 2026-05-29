// Closed-perimeter Nitro proxy: every browser request to /api/* falls
// here and is forwarded to the ui-server (Symfony) inside bot-net.
//
// proxyRequest takes the target URL verbatim — it does NOT append the
// original request's query string. The query is parsed off the event
// here and re-attached so paths like /api/bot/health/check?botUuid=...
// reach ui-server with the same query the browser sent.

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const path = event.context.params?.path ?? ''
  const flat = Array.isArray(path) ? path.join('/') : path

  const url = getRequestURL(event)
  const query = url.search

  const target = `${config.backend.apiBase}/${flat}${query}`
  return proxyRequest(event, target)
})
