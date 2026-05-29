#!/usr/bin/env bash
# Block until the named compose service reports healthy.
# Usage: scripts/wait-healthy.sh <service> [timeout_seconds]

set -euo pipefail

SERVICE="${1:?service name required}"
TIMEOUT="${2:-60}"
ENV_FILE="${ENV_FILE:-.env.local}"
COMPOSE="docker compose --env-file ${ENV_FILE}"

end=$(( $(date +%s) + TIMEOUT ))
while [ "$(date +%s)" -lt "$end" ]; do
    state=$(${COMPOSE} ps --format json 2>/dev/null | grep -oE "\"Service\":\"${SERVICE}\"[^}]*\"Health\":\"[a-z]+\"" | grep -oE 'Health":"[a-z]+' | cut -d'"' -f3 || true)
    if [ "$state" = "healthy" ]; then
        echo "✓ ${SERVICE} healthy"
        exit 0
    fi
    sleep 2
done

echo "✗ ${SERVICE} did not become healthy within ${TIMEOUT}s" >&2
exit 1
