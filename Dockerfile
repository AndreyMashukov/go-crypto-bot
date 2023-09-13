FROM golang:1.21.1

WORKDIR /go/src/app

COPY . /go/src/app

RUN go get github.com/gorilla/websocket \
    && go get github.com/go-sql-driver/mysql \
    && go get github.com/redis/go-redis/v9

RUN go build main.go

CMD ["./main"]
