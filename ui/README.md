# ui — SaaS layer

User-facing SaaS for the AutoTrade.cloud platform: accounts, dashboards,
bot configuration, billing, signals, and documentation. Sits on top of
the trading engine in [`../server/`](../server/README.md).

For the repo orientation see [`../Readme.md`](../Readme.md).

---

## Layout

| Layer | Path | Stack | README |
|---|---|---|---|
| Frontend | `app/` | Nuxt 3 + Vue 3 + TS | [app/README.md](app/README.md) |
| Backend (REST / admin / OAuth2) | `server/` | Symfony 5.4 + PHP 7.4 | [server/README.md](server/README.md) |
| Go microservices | `server/microservices/` | Go 1.21 + gin | [server/microservices/README.md](server/microservices/README.md) |
| End-user documentation | `docs/` | Docsify (markdown) | [docs/README.md](docs/README.md) |
| Deployment playbooks | `deploy/cd-*.yml` | Ansible | — |

> Note on the two `server/` directories in this repo: `../server/` is
> the Go trading engine; `ui/server/` here is the Symfony web backend.
> Different roles, different layers — keep them straight when running
> commands.

---

## Quick start (dev)

Each layer runs independently. Minimal path:

```bash
# Symfony backend
cd ui/server
cp .env.dist .env       # fill credentials; .env is gitignored
make build
make run
docker exec -it <container> bash -c 'make database'

# Nuxt frontend (in another terminal)
cd ui/app
yarn install
yarn dev                # http://localhost:3000
```

See the per-layer README for full instructions, env variables, and
production deployment.

---

## Stack overview

- **Symfony 5.4** + PHP 7.4 — REST API, OAuth2 server, Doctrine ORM,
  Messenger (RabbitMQ), Flagception (feature flags), FOSRestBundle,
  Gedmo Doctrine extensions, AWS SDK, Sentry.
- **Nuxt 3** + Vue 3 + TypeScript — SSR/SSG frontend with TradingView
  charts, i18n, GTM, Sentry, rxjs streams.
- **Go 1.21 microservices** — `sentiment` (news analysis via embedded
  Python + scikit-learn) and `signal-server` (aggregator for trading
  signals).
- **Storage** — MySQL (state), ClickHouse (history), Redis (cache and
  locks), RabbitMQ (Messenger transport).

---

## Deployment

Four Ansible playbooks in `deploy/`:

| Component | Playbook | Notes |
|---|---|---|
| Symfony backend (`server/`) | `deploy/cd-back.yml` | container on `127.0.0.1:8094` → `:80` |
| Nuxt frontend (`app/`) | `deploy/cd-front.yml` | |
| User docs (`docs/`) | `deploy/cd-docs.yml` | docsify static |
| Sentiment microservice | `deploy/cd-sentiment.yml` | |

Each playbook does `docker login` → pull image → recreate container.
Build pipelines push images to the GitLab container registry; the
GitHub side of this repo runs only PR-level checks (build + tests) via
GitHub Actions.
