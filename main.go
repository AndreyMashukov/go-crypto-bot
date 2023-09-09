package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	WebsocketClient "gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	ExchangeController "gitlab.com/open-soft/go-crypto-bot/exchange_context/controller"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	ExchangeService "gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"log"
	"net/http"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(mysql:3306)/go_crypto_bot")
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/hello", ExchangeController.Hello)
	orderRepository := ExchangeRepository.OrderRepository{
		DB: db,
	}

	traderService := ExchangeService.TraderService{
		OrderRepository: &orderRepository,
	}

	WebsocketClient.Listen("wss://fstream.binance.com/stream?streams=btcusdt@aggTrade", func(candle ExchangeModel.Candle) {
		log.Printf("Candle [%s]: S:%s, P:%f, Q:%f, O:%s\n", candle.GetDate(), candle.Symbol, candle.Price, candle.Quantity, candle.GetOperation())
		traderService.Trade(candle)
	})

	for _, symbol := range ExchangeRepository.GetSubscribedSymbols() {
		fmt.Println(symbol)
	}

	http.ListenAndServe(":8080", nil)
}
