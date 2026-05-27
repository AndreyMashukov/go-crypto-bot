FROM amashukov/golang:1.21.1-pytorch

ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/go/src/app/src/service/ml/pkg-config

WORKDIR /go/src/app

COPY . /go/src/app
COPY model /go/src/app/model
RUN mkdir /go/src/app/models
RUN mkdir /go/src/app/results

RUN go mod download

RUN go build main.go

EXPOSE 8080

CMD ["./main"]
