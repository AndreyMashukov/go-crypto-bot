FROM golang:1.21.1

WORKDIR /go/src/app

COPY . /go/src/app

RUN go mod download

RUN go test ./tests

RUN go build main.go

CMD ["./main"]
