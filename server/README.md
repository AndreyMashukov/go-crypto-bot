# server — Go trading engine

The Go module that drives the actual trading. Module path
`github.com/AndreyMashukov/go-crypto-bot/server`, Go 1.21+, single binary entry point
in `main.go`.

This README focuses on the engine itself. For the top-level repo layout,
hosted version, and Docker image references, see [`../Readme.md`](../Readme.md).

---

## Layout

```
server/
├── Dockerfile               ← built from CGO base image (Python embedded for ML)
├── .dockerignore
├── Makefile                 ← fmt/vet/lint/ailint/build/test/check/init-db-dev
├── main.go                  ← entry point: env load → container → HTTP + listeners
├── go.mod / go.sum
├── migrations/              ← MySQL migrations (numbered .sql) + clickhouse/
├── src/
│   ├── client/              ← Binance + ByBit HTTP and WebSocket clients
│   ├── config/              ← service container, env wiring (single point that calls os.Getenv)
│   ├── controller/          ← HTTP handlers (gin)
│   ├── event/               ← event payloads (DTO-style)
│   ├── event_subscriber/    ← subscribers
│   ├── model/               ← domain types (Order, Trade, Bot, Signal, Swap…)
│   ├── repository/          ← MySQL / Redis / ClickHouse persistence
│   ├── service/             ← trading, ML, balance, health, swap, strategies
│   ├── utils/               ← formatter, math, time helpers
│   └── validator/           ← input validation
└── tests/                   ← integration tests + JSON fixtures
```

`main.go` boots a single `ServiceContainer` (`src/config/container.go`)
that wires every dependency. From there: ping DB, initialize Python ML
bridge, start HTTP server, balance check, then enter the trading loop
(`MakerService.StartTrade`) plus market/MC listeners.

---

## Configuration

All settings come from environment variables. `.env` is auto-loaded at
startup (via `godotenv`) if it exists alongside the binary; otherwise the
process expects the variables to already be set.

| Variable | Required | Description | Example |
|---|---|---|---|
| `BOT_UUID` | yes | Must match a row in MySQL `bot` table | `6c26e421-06fd-4c61-84d9-caf36b8966af` |
| `BOT_EXCHANGE` | yes | `binance` or `bybit` | `binance` |
| `DATABASE_DSN` | yes | MySQL DSN | `root:go_crypto_bot@tcp(mysql:3306)/go_crypto_bot` |
| `REDIS_DSN` | yes | Redis address | `redis:6379` |
| `REDIS_PASSWORD` | no | Empty if Redis has no auth | — |
| `CLICKHOUSE_DSN` | yes | ClickHouse `host:port` | `clickhouse:8123` |
| `CLICKHOUSE_PASSWORD` | no | ClickHouse password | — |
| `BINANCE_API_KEY` / `BINANCE_API_SECRET` | binance | Binance credentials, [docs](https://www.binance.com/en/support/faq/how-to-create-api-keys-on-binance-360002502072) | — |
| `BINANCE_API_DSN` | binance | REST endpoint | testnet `https://testnet.binance.vision`; prod `https://api.binance.com` |
| `BINANCE_WS_DSN` | binance | WS API endpoint | testnet `wss://testnet.binance.vision/ws-api/v3`; prod `wss://ws-api.binance.com:443/ws-api/v3` |
| `BINANCE_STREAM_DSN` | binance | WS market-data stream | `wss://stream.binance.com` |
| `BYBIT_API_KEY` / `BYBIT_API_SECRET` | bybit | ByBit credentials | — |
| `BYBIT_API_DSN` / `BYBIT_STREAM_DSN` | bybit | ByBit REST + stream | testnet `https://api-testnet.bybit.com` |
| `MC_DSN` | no | Capitalization service WS DSN (master-bot only) | — |

> **Never commit real keys.** `docker-compose.yaml` is `.gitignore`d; only
> `docker-compose.yaml.dist` (template with placeholders) is tracked.
> `.env*` is `.gitignore`d.

---

## HTTP API

Listens on `:8090`. Every endpoint requires `?botUuid=<UUID>`.

### Bot config

`PUT /bot/update?botUuid=<UUID>` — full bot configuration.

```json
{
  "isMasterBot": true,
  "isSwapEnabled": true,
  "tradeStackSorting": "percent",
  "swapConfig": {
    "swapMinPercent":          2.00,
    "swapOrderProfitTrigger": -5.00,
    "orderTimeTrigger":     36000,
    "useSwapCapital":        true,
    "historyInterval":      "1d",
    "historyPeriod":          14
  }
}
```

**Master bot:** with multiple bots on one host, designate exactly one as
master — it owns shared static state (symbol lists, market depth caches)
the others consume.

**SWAP (triangular arbitrage):**
- `swapMinPercent` — minimum profit % to engage a 3-leg swap.
- `swapOrderProfitTrigger` — engage on positions worse than this PnL %.
- `orderTimeTrigger` — minimum position age (seconds) before swap is allowed.
- `useSwapCapital` — credit swap-realised capital against position PnL.
- `historyInterval` / `historyPeriod` — candle interval and lookback for
  swap leg discovery.

### Trade limits

`POST /trade/limit/create?botUuid=<UUID>` — create per-symbol limit.

```json
{
  "symbol":                       "PERPUSDT",
  "USDTLimit":                    100,
  "minPrice":                     0.00001,
  "minQuantity":                  0.01,
  "minNotional":                  5,
  "isEnabled":                    true,
  "USDTExtraBudget":              80,
  "buyOnFallPercent":             -3.5,
  "minPriceMinutesPeriod":        200,
  "frameInterval":                "2h",
  "framePeriod":                  20,
  "buyPriceHistoryCheckInterval": "1d",
  "buyPriceHistoryCheckPeriod":   14,
  "profitOptions": [
    { "index": 0, "optionValue": 1,  "optionUnit": "h", "optionPercent": 2.40, "isTriggerOption": true  },
    { "index": 0, "optionValue": 30, "optionUnit": "i", "optionPercent": 1.50, "isTriggerOption": false }
  ],
  "extraChargeOptions": [
    { "index": 0, "percent": -4.50, "amountUsdt": 20.00 }
  ],
  "tradeFiltersBuy":          [],
  "tradeFiltersSell":         [],
  "tradeFiltersExtraCharge":  []
}
```

`PUT /trade/limit/update?botUuid=<UUID>` — same payload shape, updates an
existing limit. Filters use a tree of conditions:

```json
"tradeFiltersBuy": [
  {
    "symbol":    "BTCUSDT",
    "parameter": "price",
    "condition": "lt",         // lt | gt | lte | gte | eq | neq
    "value":     "50000.00",
    "type":      "or",         // or | and
    "children":  []
  }
]
```

`GET /trade/limit/list?botUuid=<UUID>` — list all configured symbols.

### Orders, positions, charts, health

| Method + path | What |
|---|---|
| `GET /trade/stack?botUuid=<UUID>` | Current decision stack |
| `GET /order/position/list?botUuid=<UUID>` | Open positions |
| `GET /order/pending/list?botUuid=<UUID>` | Pending limit-buy orders |
| `PUT /order/extra/charge/update?botUuid=<UUID>` | Update extra-charge ladder on a position |
| `GET /chart/list?botUuid=<UUID>` | Chart data per symbol |
| `GET /health/check?botUuid=<UUID>` | Health status |

`PUT /order/extra/charge/update` body:

```json
{
  "orderId": 92,
  "extraChargeOptions": [
    { "index": 3, "percent": -14.50, "amountUsdt": 120.00 },
    { "index": 2, "percent":  -4.00, "amountUsdt":  10.00 },
    { "index": 1, "percent":  -2.00, "amountUsdt":  30.00 },
    { "index": 0, "percent":  -1.00, "amountUsdt":  30.00 }
  ]
}
```

---

## Development

From the repo root:

```bash
make hooks        # activate the pre-commit gate (one-time)
make check        # fmt + vet + build + lint + ailint + test
make test         # go test -race
make lint         # golangci-lint
make ailint       # go-lint AI-bloat analyzers
```

Directly from `server/` for tighter loops:

```bash
go build ./...
go test ./src/utils                                # focused
golangci-lint run --config=../.golangci.yml --timeout=3m
go-lint ./...
```

### Tooling

- **golangci-lint** — configured in `../.golangci.yml`: errcheck, gosimple,
  govet (all), staticcheck, gofmt, goimports, revive (strict), gosec,
  gocritic, unconvert, unparam, misspell, prealloc, bodyclose, noctx,
  nilerr, nilnil, errorlint, exhaustive, forbidigo (no `fmt.Print*`, no
  `panic`), nakedret, funlen, gocyclo, lll, dupl, wastedassign,
  copyloopvar.
- **go-lint** — install via
  `go install github.com/AndreyMashukov/go-lint/cmd/go-lint@latest`.
  Blocks: inline narration in function bodies, `//nolint` directives,
  `os.Getenv` outside `src/config/`, env branching (`if env == "prod"`),
  `panic` in production code, direct `time.Now()`, type-only test
  assertions, raw `INSERT/UPDATE/DELETE` in tests, tautological godoc,
  banal `fmt.Errorf` wrappers, ownerless `TODO`s, redundant if-return,
  nil-checks on value types.

### Pre-commit checks (blocking)

1. `gofmt -l` on staged `.go` files
2. `go vet ./...`
3. `go build ./...`
4. `golangci-lint run --new-from-rev=HEAD --timeout=2m`
5. `go-lint ./<staged packages>`
6. Source-without-test pairing (model/, config/, event/ exempt)
7. `go test -count=1 -timeout=60s` on touched packages

Renames short-circuit content checks. There is no `--no-verify` bypass.

### Adding a migration

Drop `migrations/migration_NN.sql` (or `migrations/clickhouse/...`),
bump the number, then re-run `make init-db-dev` against your dev MySQL.

### Building the Docker image

`Dockerfile` in this directory; build context is also this directory:

```bash
docker build -t go-crypto-bot:dev .
```

Or via root compose (build context is `./server`):

```bash
docker-compose -f ../docker-compose.yaml build writer
```

---

## Why a separate `server/` subdirectory?

The repo is structured to keep deployment, dev-stack config, and tooling
on the root while the Go module lives in `server/`. Same pattern used by
the sibling monorepo `crypto-saas` (Nuxt `app/` + Symfony `server/`).

Module path is unchanged (`github.com/AndreyMashukov/go-crypto-bot/server`), so
imports keep resolving without any code modification — the relocation is
purely directory-level.
