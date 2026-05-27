# app — Nuxt frontend

SaaS frontend for AutoTrade.cloud. Nuxt 3.18+, Vue 3.6+, TypeScript, i18n,
Vuetify, embedded TradingView charts. Lives under the monorepo `ui/`
layer; for repo orientation see [`../../Readme.md`](../../Readme.md). The
Symfony backend this frontend talks to is in
[`../server/`](../server/README.md).

---

## Layout

```
app/
├── Dockerfile              ← multi-stage Node 22 build
├── docker-compose.yml      ← dev stack (service name: frontend)
├── Makefile                ← all targets run inside docker compose
├── package.json / yarn.lock
├── nuxt.config.ts
├── i18n.config.ts
├── tsconfig.json
├── components/             ← Vue components
├── pages/                  ← file-based routing
│   ├── index.vue
│   ├── account/
│   ├── bot/
│   └── dashboard/
├── layouts/                ← layout wrappers
├── model/                  ← class-transformer DTOs
├── plugins/                ← Nuxt plugins (GTM, Sentry, etc.)
├── services/               ← REST clients to the Symfony backend
├── server/                 ← Nuxt 3 server middleware (SSR API routes)
└── public/                 ← static assets
```

> `app/server/` is the Nuxt server middleware (SSR API routes), **not**
> the Symfony backend. Symfony lives at `../server/`. Different layers,
> different roles.

---

## Stack

- **Nuxt 3** (`^3.18`) — SSR + SSG, Vite bundler, file-based routing.
- **Vue 3** (`^3.6`) — Composition API.
- **TypeScript** (`^5.9`) — strict.
- **Vuetify 3** + **vite-plugin-vuetify** — Material Design components.
- **@nuxtjs/i18n** (`^9`) — multilingual (RU/EN/...).
- **rxjs** + **class-transformer** — reactive streams and DTO mapping.
- **nuxt-tradingview** — embedded TradingView charts.
- **@gtm-support/vue-gtm** — Google Tag Manager.
- **nuxt-delay-hydration** — TTI optimisation.
- **@nuxt/eslint** + ESLint 9 — flat config.
- **Sentry** — error tracking.
- **@mdi/font** — MDI icon set.

---

## Development

All commands run inside the docker-compose stack — there is no local
Node/yarn requirement, only Docker.

```bash
make build         # docker compose build --no-cache (first run)
make install       # yarn install --frozen-lockfile inside the container
make up            # start dev server (frontend service)
                   # → http://localhost:3000
```

Useful per-task targets:

```bash
make lint          # ESLint
make lint-fix      # ESLint --fix
make typecheck     # nuxt typecheck (vue-tsc)
make generate      # static site generation (SSG)
make preview       # local preview of the production build
make shell         # interactive shell in the frontend container
make check         # lint + typecheck (CI entry point)
```

---

## Production build

The production image is built by the second stage of `Dockerfile`. CI
tags it and pushes to the GitLab registry; `../deploy/cd-front.yml`
pulls and recreates the runtime container.

Build args expected at image build time:

| Arg | Purpose |
|---|---|
| `API_BASE_URL` | Symfony backend URL |
| `CLIENT_ID` | OAuth2 client id |
| `CLIENT_SECRET` | OAuth2 client secret |

Runtime env:

| Variable | Default |
|---|---|
| `HOST` | `0.0.0.0` |
| `PORT` | `3000` |
| `NODE_ENV` | `production` |

---

## See also

- [Nuxt 3 docs](https://nuxt.com/docs)
- [Vue 3 docs](https://vuejs.org)
- [Vuetify 3 docs](https://vuetifyjs.com)
