# server — Symfony backend

Symfony 7.3 / PHP 8.3 SaaS backend for AutoTrade.cloud. User accounts,
bot configuration, billing (OxaPay / Telegram), signal orchestration,
analytics. Talks to the Nuxt frontend in [`../app/`](../app/README.md)
and to Go microservices under [`microservices/`](microservices/README.md).

For repo orientation see [`../../Readme.md`](../../Readme.md).

---

## Layout

```
server/
├── Dockerfile / Dockerfile.prod     ← PHP 8.3-fpm-alpine base
├── docker-compose.yml               ← dev stack (backend + mysql + redis)
├── Makefile                         ← all targets via docker compose exec
├── entrypoint.sh
├── phpunit.xml.dist
├── phpstan.neon                     ← level 6, src/ + bundles/ + tests/
├── rector.php                       ← amashukov/rector-php-rules + Symfony sets
├── .php-cs-fixer.dist.php
├── .env.dist / .env.template
├── composer.json / composer.lock
├── symfony.lock
├── bin/
├── config/
│   ├── bundles.php
│   ├── packages/
│   ├── routes/
│   └── services.yaml
├── public/                          ← index.php
├── migrations/                      ← Doctrine migrations
├── translations/                    ← i18n (XLIFF)
├── src/                             ← App\
│   ├── Controller/
│   ├── Entity/
│   ├── Repository/
│   ├── Service/
│   ├── Form/
│   ├── DataFixtures/
│   └── ...
├── bundles/                         ← Bundles\
│   ├── UserContext/
│   ├── CryptoBotContext/
│   ├── OxaPayContext/
│   └── TgBotContext/
├── League/OAuth2/                   ← local League\OAuth2 traits fork
├── microservices/                   ← Go services consumed by this backend
│   ├── sentiment/
│   └── signal-server/
└── tests/                           ← PHPUnit (App\Tests\)
```

PSR-4 mapping from `composer.json`:

| Namespace | Path |
|---|---|
| `App\` | `src/` |
| `Bundles\` | `bundles/` |
| `League\OAuth2\Server\Entities\Traits\` | `League/OAuth2/Server/Entities/Traits/` |
| `App\Tests\` | `tests/` (dev/test only) |

---

## Bundles in use

Internal (`bundles/`):

- `UserContext` — users, registration, OAuth2 server, JWT.
- `CryptoBotContext` — bot configuration, trade limits, swap config.
- `OxaPayContext` — crypto billing via OxaPay.
- `TgBotContext` — Telegram notifications and commands.

External (notable):

- Doctrine ORM 3 + Migrations + Fixtures
- TrikoderOAuth2Bundle 5 (OAuth2 server)
- FOSRestBundle, JMS SerializerBundle (REST + serialization)
- Flagception (feature flags)
- Stof Doctrine Extensions (tree, sluggable, timestampable)
- Gedmo Doctrine Extensions
- Symfony Messenger + KunicMarko JMS adapter
- Sentry monitoring
- KnpPaginator, PrestaSitemap, NelmioApiDocBundle

---

## Quality gates

PHP-CS-Fixer + PHPStan + Rector wired into `make`, and the project-level
pre-commit hook (`../../.githooks/pre-commit`) runs them on staged
files. Custom Rector rules come from
[`amashukov/rector-php-rules`](https://github.com/AndreyMashukov/rector-php-rules)
(packagist `amashukov/rector-php-rules`):

| Rector rule | Blocks |
|---|---|
| `NoCommentsOutsideInterfaceMethodDocBlockRector` | inline narration in functions |
| `NoPhpstanIgnoreRector` | `@phpstan-ignore` / `@psalm-suppress` |
| `NoSuperglobalAccessRector` | `$_ENV`, `$_GET`, `getenv()`, etc. outside config |
| `NoEnvironmentCheckInSrcRector` | `if env == "prod"` in src |
| `NoAssertCallInSrcRector` | bare `assert()` in src |
| `RequirePsrClockInterfaceRector` | `time()`, `new DateTime` — require PSR `ClockInterface` |
| `NoAssertInsideIfInFunctionalTestsRector` | conditional asserts in functional tests |
| `NoArrayAssertContainsInTestsRector` | `assertContains($x, [A, B])` with inline arrays |
| `NoTypeOnlyAssertionsInTestsRector` | `assertIsArray` / `assertIsString` |
| `NoExistenceOnlyAssertionsInTestsRector` | `assertNotNull` / `assertArrayHasKey` |
| `NoDirectDbMutationInFunctionalTestsRector` | raw ORM/DBAL mutations in tests |
| `NoDirectDispatchInFunctionalTestsRector` | direct event/message dispatch in tests |

PHPStan is at level 6, paths `src/`, `bundles/`, `tests/`. Suppression
directives are banned (per `NoPhpstanIgnoreRector`).

---

## Development

All commands run **inside the docker-compose stack** (service name
`backend`). No local PHP / composer / vendor.

```bash
make build              # docker compose build --no-cache (first run)
make up                 # start backend + mysql + redis
make composer-install   # composer install inside the container
make database           # drop + create + migrate + load fixtures
```

Quality gates:

```bash
make cs           # php-cs-fixer fix
make cs-dry       # php-cs-fixer fix --dry-run --diff
make phpstan      # phpstan analyse --memory-limit=1G
make rector       # rector process (writes changes)
make rector-dry   # rector process --dry-run (CI / pre-commit form)
make test         # phpunit
make check        # cs-dry + phpstan + rector-dry + test (CI entry point)
make shell        # sh inside backend container
```

---

## Configuration

`.env.dist` ships placeholders (real values arrive via Docker build args
in prod and via `.env` in dev — neither is committed).

| Variable | Description |
|---|---|
| `APP_ENV` | `dev` / `test` / `prod` |
| `APP_SECRET` | Symfony framework secret |
| `OAUTH2_ENCRYPTION_KEY` | OAuth2 token encryption key |
| `DATABASE_URL` | MySQL DSN |
| `REDIS_DSN` | Redis URL |
| `MESSENGER_TRANSPORT_DSN` | RabbitMQ / amqp DSN |
| `CLICKHOUSE_DSN` / `CLICKHOUSE_PASSWORD` | ClickHouse |
| `BREVO_API_KEY` | transactional email |
| `OXA_PAY_API_KEY` | crypto billing |
| `INTERNAL_SERVICE_TOKEN` | service-to-service auth |
| `ACCESS_TOKEN_TTL` | OAuth2 access-token lifetime |

Full list in `.env.dist`.

---

## Production deployment

Production image is built from `Dockerfile.prod` (PHP 8.3-fpm-alpine +
supervisord + dcron + nginx + composer 2 + Go 1.21 for the in-image
microservice build). Tagged and pushed to the GitLab registry by the
`build:backend` job in `.gitlab-ci.yml`. Deployed by
[`../deploy/cd-back.yml`](../deploy/cd-back.yml):

1. `docker login` to `registry.gitlab.com`.
2. Pull the latest image.
3. Recreate the container, map `127.0.0.1:8094 → :80`, networks
   `bridge` + `redis_net`.
4. `bin/console doctrine:migrations:migrate -n` inside the container.
5. `chown -R www-data /srv/www/var`.

Build-time secrets arrive as `--build-arg` (Brevo, OxaPay,
DATABASE_URL, REDIS_DSN, CLICKHOUSE_DSN, INTERNAL_SERVICE_TOKEN, etc.).
