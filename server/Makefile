.PHONY: fmt vet lint ailint build test check init-db-dev init-ch-dev

# Go targets — module root is this directory.

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	golangci-lint run --config=../.golangci.yml --timeout=3m

ailint:
	go-lint ./...

build:
	go build ./...

test:
	go test -count=1 -race -timeout=120s ./...

check: fmt vet build lint ailint test

# Postgres init: applies every migrations/migration_*.sql in numeric order
# against the dev cluster wired up by docker-compose.yaml.dist (postgres
# service exposed on host port 5444). Uses PGPASSWORD so the prompt
# doesn't break a non-interactive `make`.
init-db-dev:
	@set -e; \
	export PGPASSWORD=go_crypto_bot; \
	for n in $$(ls migrations/migration_*.sql | sed -E 's|.*migration_([0-9]+)\.sql|\1|' | sort -n); do \
		echo "applying migrations/migration_$$n.sql"; \
		psql -h 127.0.0.1 -p 5444 -U go_crypto_bot -d go_crypto_bot -v ON_ERROR_STOP=1 -f migrations/migration_$$n.sql; \
	done

# ClickHouse init: applies the analytics schema (default.trades,
# default.market_tick) against the clickhouse service exposed by
# docker-compose.yaml.dist on host port 9123.
init-ch-dev:
	@set -e; \
	for f in migrations/clickhouse/*.sql; do \
		echo "applying $$f"; \
		curl -sS --data-binary @$$f 'http://localhost:9123/?user=default&password=123456&database=default'; \
		echo; \
	done
