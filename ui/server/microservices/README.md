# microservices — Go services under the Symfony backend

Two Go 1.21 services that sit next to the Symfony backend and provide
specialised functions consumed by it.

| Service | Path | Purpose | Port |
|---|---|---|---|
| `sentiment` | `sentiment/` | News-text sentiment classifier — embeds Python + scikit-learn via cgo | `:8080` |
| `signal-server` | `signal-server/` | Trading-signal aggregator and stats API | `:8080` |

Both built on `gin`. The production images live in the GitLab container
registry; `../../deploy/cd-sentiment.yml` deploys the sentiment service.

---

## sentiment

```
sentiment/
├── Dockerfile
├── go.mod / go.sum         ← module: sentiment
├── main.go                 ← entry point
└── src/
    ├── form/
    ├── http/
    ├── model/
    └── service/
        └── ml/             ← cgo bridge to CPython + sklearn
```

Routes:

| Method + Path | Description |
|---|---|
| `POST /sentiment/predict` | Score a piece of text; returns coin mentions + sentiment vector |

**Why cgo + CPython:** the linear model is trained and serialised in
Python (`pickle`). Calling the model from Go directly through cgo lets
the binary serve predictions without spawning a subprocess per request.

---

## signal-server

```
signal-server/
├── go.mod / go.sum         ← module: github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server (rename to .../server/... pending)
├── main.go                 ← entry point: env, container, gin, background container.Start()
└── src/
    ├── client/
    ├── config/             ← service container, env wiring
    ├── http/               ← controllers + CORS
    ├── model/
    ├── repository/         ← MySQL / ClickHouse access
    ├── service/
    └── utils/
```

Routes:

| Method + Path | Description |
|---|---|
| `GET /stats/pivot/:symbol` | Stats for one symbol |
| `GET /stats/pivot/grid` | Multi-symbol pivot grid |

`container.Start()` runs background workers (pollers, aggregators) on
top of the gin HTTP server.

---

## Tooling

Both services are linted with
[`github.com/AndreyMashukov/go-lint`](https://github.com/AndreyMashukov/go-lint)
as part of the project-level pre-commit hook
(`../../../.githooks/pre-commit`). The same `go-lint` AST analyzers that
guard the trading engine (`server/`) apply here.

### Build inside docker

Both services have their own `Dockerfile` and are built by the
production pipeline of the parent Symfony backend (the parent `Dockerfile`
copies the Go toolchain in and runs `go build` for signal-server during
image build, see `../Dockerfile`).

Locally, from inside the backend container shell:

```bash
make shell                         # from ui/server/
cd /srv/www/microservices/signal-server
go build ./...
go test ./...
```
