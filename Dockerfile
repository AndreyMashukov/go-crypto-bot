FROM golang:1.16

WORKDIR /go/src/app

COPY . /go/src/app

RUN go get github.com/gorilla/websocket \
    && go get github.com/go-sql-driver/mysql

RUN go build main.go

CMD ["./main"]
