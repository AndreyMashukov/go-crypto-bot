#!/usr/bin/env bash
# Apply Postgres + ClickHouse migrations against the running compose stack.
#
# Postgres migrations 1-23 are pre-existing schema and are skipped when the
# `bots` table already exists. Phase J's migration_24 (and any future seed
# migrations) are idempotent on their own (INSERT ... ON CONFLICT / WHERE NOT
# EXISTS) so they re-run safely on every invocation.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ENV_FILE="${ENV_FILE:-.env.local}"
COMPOSE="docker compose --env-file ${ENV_FILE}"

bots_exists=$(${COMPOSE} exec -T -e PGPASSWORD=go_crypto_bot postgres \
    psql -tA -U go_crypto_bot -d go_crypto_bot \
    -c "SELECT to_regclass('public.bots') IS NOT NULL" 2>/dev/null | tr -d '[:space:]')

if [ "$bots_exists" = "t" ]; then
    echo "▶ Postgres schema present — applying seed migrations only..."
    seeds=$(ls server/migrations/migration_*.sql | sort -V | awk -F'migration_|\\.sql' '$2 >= 24')
else
    echo "▶ Postgres schema absent — applying full migration set..."
    seeds=$(ls server/migrations/migration_*.sql | sort -V)
fi

for n in $seeds; do
    printf "  · %s\n" "$(basename "$n")"
    ${COMPOSE} exec -T -e PGPASSWORD=go_crypto_bot postgres \
        psql -U go_crypto_bot -d go_crypto_bot -v ON_ERROR_STOP=1 \
        < "$n" >/dev/null
done

market_tick_exists=$(${COMPOSE} exec -T clickhouse \
    clickhouse-client --password 123456 -q "EXISTS TABLE default.market_tick" 2>/dev/null | tr -d '[:space:]')

if [ "$market_tick_exists" = "1" ]; then
    echo "▶ ClickHouse schema present — skipping migrations."
else
    echo "▶ ClickHouse migrations..."
    for n in $(ls server/migrations/clickhouse/migration_*.sql | sort -V); do
        printf "  · %s\n" "$(basename "$n")"
        ${COMPOSE} exec -T clickhouse \
            clickhouse-client --password 123456 --multiquery \
            --query "$(cat "$n")" >/dev/null
    done
fi

echo "✓ migrations applied"
