package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	ExchangeClient "gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	ExchangeController "gitlab.com/open-soft/go-crypto-bot/exchange_context/controller"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	ExchangeService "gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"log"
	"net/http"
	"os"
	"sync"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(mysql:3306)/go_crypto_bot")
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/hello", ExchangeController.Hello)
	httpClient := http.Client{}
	binance := ExchangeClient.Binance{
		ApiKey:         "0XVVs5VRWyjJH1fMReQyVUS614C8FlF1rnmvCZN2iK3UDhwncqpGYzF1jgV8KPLM",
		ApiSecret:      "tg5Ak5LoTFSCIadQLn5LkcnWHEPYSiA6wpY3rEqx89GG2aj9ZWsDyMl17S5TjTHM",
		DestinationURI: "https://testnet.binance.vision",
		HttpClient:     &httpClient,
	}
	orderRepository := ExchangeRepository.OrderRepository{
		DB: db,
	}
	exchangeRepository := ExchangeRepository.ExchangeRepository{
		DB: db,
	}

	tradeChannel := make(chan ExchangeModel.Trade)
	logChannel := make(chan ExchangeModel.Trade)
	lockTradeChannel := make(chan ExchangeModel.Lock)

	traderService := ExchangeService.TraderService{
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Binance:            &binance,
		LockChannel:        &lockTradeChannel,
		BuyLowestOnly:      false,
		SellHighestOnly:    false,
		Trades:             make(map[string][]ExchangeModel.Trade),
		Lock:               make(map[string]bool),
		TradesMapMutex:     sync.RWMutex{},
		TradeLockMutex:     sync.RWMutex{},
	}

	file, _ := os.OpenFile("trade.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	go func() {
		for {
			lock := <-lockTradeChannel
			traderService.TradeLockMutex.Lock()
			traderService.Lock[lock.Symbol] = lock.IsLocked
			traderService.TradeLockMutex.Unlock()
		}
	}()

	go func() {
		for {
			// Read the channel
			trade := <-tradeChannel

			go func() {
				// log.Printf("Trade [%s]: S:%s, P:%f, Q:%f, O:%s\n", symbolTrade.GetDate(), symbolTrade.Symbol, symbolTrade.Price, symbolTrade.Quantity, symbolTrade.GetOperation())
				traderService.Trade(trade)
			}()

			go func() {
				logChannel <- trade
			}()
		}
	}()

	go func() {
		for {
			trade := <-logChannel
			encoded, err := json.Marshal(trade)

			if err == nil {
				file.WriteString(fmt.Sprintf("%s\n", encoded))
			}
		}
	}()

	// todo: concatenate streams...
	for _, symbol := range exchangeRepository.GetSubscribedSymbols() {
		fmt.Println(symbol)
	}

	// todo: sync existed orders in Binance with bot database...

	// /perpusdt@aggTrade
	wsConnection := ExchangeClient.Listen("wss://fstream.binance.com/stream?streams=btcusdt@aggTrade/ltcusdt@aggTrade/ethusdt@aggTrade/solusdt@aggTrade", tradeChannel)
	defer wsConnection.Close()

	http.ListenAndServe(":8080", nil)
}
