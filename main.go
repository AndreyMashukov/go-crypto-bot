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

	tradeChannel := make(chan ExchangeModel.Trade)

	go func() {
		for {
			// Read the channel
			trade := <-tradeChannel
			log.Printf("Trade [%s]: S:%s, P:%f, Q:%f, O:%s\n", trade.GetDate(), trade.Symbol, trade.Price, trade.Quantity, trade.GetOperation())
			traderService.Trade(trade)
		}
	}()

	// todo: concatenate streams...
	for _, symbol := range ExchangeRepository.GetSubscribedSymbols() {
		fmt.Println(symbol)
	}

	wsConnection := WebsocketClient.Listen("wss://fstream.binance.com/stream?streams=btcusdt@aggTrade/ltcusdt@aggTrade/ethusdt@aggTrade/perpusdt@aggTrade/solusdt@aggTrade", tradeChannel)
	defer wsConnection.Close()

	http.ListenAndServe(":8080", nil)
}
