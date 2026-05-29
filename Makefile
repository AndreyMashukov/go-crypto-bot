.PHONY: fmt vet lint ailint build test check hooks init-db-dev stack-up stack-down stack-status stack-logs stack-rebuild migrations

ENV_FILE ?= .env.local
COMPOSE = docker compose --env-file $(ENV_FILE)

# All Go targets delegate to server/. The Go module lives in server/go.mod.

fmt:
	$(MAKE) -C server fmt

vet:
	$(MAKE) -C server vet

lint:
	$(MAKE) -C server lint

ailint:
	$(MAKE) -C server ailint

build:
	$(MAKE) -C server build

test:
	$(MAKE) -C server test

check:
	$(MAKE) -C server check

hooks:
	git config core.hooksPath .githooks/
	@echo "pre-commit hook activated"

init-db-dev:
	$(MAKE) -C server init-db-dev

# Phase J: closed-perimeter full-stack lifecycle. `make stack-up` brings up
# postgres + redis + clickhouse + prometheus first, applies all SQL
# migrations to both Postgres and ClickHouse (idempotent), then builds and
# starts the bot (watcher + trader) and the UI (Symfony gateway + Nuxt
# admin). The UI lands on http://localhost:3000.

stack-up:
	@test -f docker-compose.yaml || cp docker-compose.yaml.dist docker-compose.yaml
	@test -f $(ENV_FILE) || (echo "create $(ENV_FILE) with BINANCE_API_KEY and BINANCE_API_SECRET first" && exit 1)
	@echo "▶ infra up..."
	$(COMPOSE) up -d postgres redis clickhouse prometheus
	@bash ./scripts/wait-healthy.sh postgres 90
	@bash ./scripts/apply-migrations.sh
	@echo "▶ bot + ui up..."
	$(COMPOSE) up -d --build market-watcher market-trader ui-server ui-app
	@echo
	@echo "✓ stack up"
	@echo "  dashboard:   http://localhost:13000"
	@echo "  bot admin:   http://localhost:18090"
	@echo "  ui server:   http://localhost:18000/api"
	@echo "  prometheus:  http://localhost:19090"
	@echo "  clickhouse:  http://localhost:19123"

stack-down:
	$(COMPOSE) down

stack-rebuild:
	$(COMPOSE) build --no-cache market-watcher ui-server ui-app
	$(COMPOSE) up -d --force-recreate market-watcher market-trader ui-server ui-app

stack-status:
	$(COMPOSE) ps

stack-logs:
	$(COMPOSE) logs --tail 100 -f market-watcher market-trader ui-server ui-app

migrations:
	@bash ./scripts/apply-migrations.sh
