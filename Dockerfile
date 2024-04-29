FROM amashukov/golang:1.21.1-ml

ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/go/src/app/src/service/ml/pkg-config

WORKDIR /go/src/app

COPY . /go/src/app
RUN mkdir /go/src/app/models
RUN mkdir /go/src/app/results
RUN mkdir /go/src/app/datasets
RUN apk add bash zip unzip

RUN go mod download

RUN go test ./tests

RUN go build main.go

CMD ["./main"]
