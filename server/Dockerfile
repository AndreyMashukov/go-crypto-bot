FROM amashukov/golang:1.21.1-ml

ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/go/src/app/src/service/ml/pkg-config
ENV GOTOOLCHAIN=auto

WORKDIR /go/src/app

COPY . /go/src/app
RUN mkdir /go/src/app/models /go/src/app/results /go/src/app/datasets

RUN apk add bash zip unzip

RUN go mod download

RUN go test ./tests

# Phase H: two role-specific binaries. The image carries both; the
# docker-compose service picks which one to run via `command:`.
RUN go build -o /usr/local/bin/market-watcher ./market-watcher \
 && go build -o /usr/local/bin/market-trader  ./market-trader

# Default to the trader so the legacy single-service composition still
# does something useful if anyone ignores the per-service command
# override in docker-compose.yaml.dist.
CMD ["/usr/local/bin/market-trader"]
