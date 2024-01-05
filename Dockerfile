FROM amashukov/golang:1.21.1-ml

ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/go/src/app/pkg-config

WORKDIR /go/src/app

COPY . /go/src/app
RUN mkdir /go/src/app/models
COPY .docker/datasets /go/src/app/datasets
RUN mkdir /go/src/app/results
RUN apk add bash zip unzip

RUN go mod download

RUN go test ./tests

RUN go build main.go

CMD ["./main"]
