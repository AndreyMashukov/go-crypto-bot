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
	"strings"
	"sync"
	"time"
)

func main() {
	db, err := sql.Open("mysql", "root:go_crypto_bot@tcp(mysql:3306)/go_crypto_bot")
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

	eventChannel := make(chan []byte)
	tradeLogChannel := make(chan ExchangeModel.Trade)
	kLineLogChannel := make(chan ExchangeModel.KLine)
	lockTradeChannel := make(chan ExchangeModel.Lock)

	makerService := ExchangeService.MakerService{
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Binance:            &binance,
		LockChannel:        &lockTradeChannel,
		BuyLowestOnly:      false,
		SellHighestOnly:    false,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		MinDecisions:       3.00,
		HoldScore:          75.00,
	}

	// todo: BuyExtraOnMarketFallStrategy
	baseKLineStrategy := ExchangeService.BaseKLineStrategy{}
	negativePositiveStrategy := ExchangeService.NegativePositiveStrategy{
		LastKline: make(map[string]ExchangeModel.KLine),
	}
	smaStrategy := ExchangeService.SmaTradeStrategy{
		Trades:         make(map[string][]ExchangeModel.Trade),
		TradesMapMutex: sync.RWMutex{},
	}

	tradeFile, _ := os.OpenFile("trade.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	klineFile, _ := os.OpenFile("kline.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	go func() {
		for {
			lock := <-lockTradeChannel
			makerService.TradeLockMutex.Lock()
			makerService.Lock[lock.Symbol] = lock.IsLocked
			makerService.TradeLockMutex.Unlock()
		}
	}()

	decisionLock := sync.RWMutex{}
	smaDecisions := make(map[string]ExchangeModel.Decision)
	baseKLineDecisions := make(map[string]ExchangeModel.Decision)
	negativePositiveDecisions := make(map[string]ExchangeModel.Decision)

	tradeLimits := exchangeRepository.GetTradeLimits()

	for _, limit := range tradeLimits {
		go func(symbol string) {
			for {
				currentDecisions := make([]ExchangeModel.Decision, 0)
				decisionLock.Lock()
				smaDecision, smaExists := smaDecisions[symbol]
				kLineDecision, klineExists := baseKLineDecisions[symbol]
				negPosDecision, negPosExists := negativePositiveDecisions[symbol]
				decisionLock.Unlock()

				if smaExists {
					currentDecisions = append(currentDecisions, smaDecision)
				}
				if klineExists {
					currentDecisions = append(currentDecisions, kLineDecision)
				}
				if negPosExists {
					currentDecisions = append(currentDecisions, negPosDecision)
				}

				if len(currentDecisions) > 0 {
					makerService.Make(symbol, currentDecisions)
				}

				time.Sleep(time.Millisecond * 500)
			}
		}(limit.Symbol)
	}

	go func() {
		for {
			// Read the channel, todo -> better to use select: https://go.dev/tour/concurrency/5
			message := <-eventChannel
			var eventModel ExchangeModel.Event
			json.Unmarshal(message, &eventModel)

			decisionLock.Lock()
			switch true {
			case strings.Contains(eventModel.Stream, "@aggTrade"):
				var tradeEvent ExchangeModel.TradeEvent
				json.Unmarshal(message, &tradeEvent)
				trade := tradeEvent.Trade
				smaDecision := smaStrategy.Decide(trade)
				smaDecisions[trade.Symbol] = smaDecision

				go func() {
					tradeLogChannel <- tradeEvent.Trade
				}()
				break
			case strings.Contains(eventModel.Stream, "@kline_1m"):
				var klineEvent ExchangeModel.KlineEvent
				json.Unmarshal(message, &klineEvent)
				kLine := klineEvent.KlineData.Kline
				baseKLineDecision := baseKLineStrategy.Decide(kLine)
				negPosDecision := negativePositiveStrategy.Decide(kLine)
				baseKLineDecisions[kLine.Symbol] = baseKLineDecision
				negativePositiveDecisions[kLine.Symbol] = negPosDecision
				go func() {
					kLineLogChannel <- kLine
				}()
				break
			}
			decisionLock.Unlock()
		}
	}()

	go func() {
		for {
			trade := <-tradeLogChannel
			encoded, err := json.Marshal(trade)

			if err == nil {
				tradeFile.WriteString(fmt.Sprintf("%s\n", encoded))
			}
		}
	}()

	go func() {
		for {
			kline := <-kLineLogChannel
			encoded, err := json.Marshal(kline)

			if err == nil {
				klineFile.WriteString(fmt.Sprintf("%s\n", encoded))
			}
		}
	}()

	// todo: concatenate streams...
	for _, symbol := range exchangeRepository.GetSubscribedSymbols() {
		fmt.Println(symbol)
	}

	// todo: sync existed orders in Binance with bot database...

	streams := []string{}
	events := [2]string{"@aggTrade", "@kline_1m"}

	for _, tradeLimit := range tradeLimits {
		for i := 0; i < len(events); i++ {
			event := events[i]
			streams = append(streams, fmt.Sprintf("%s%s", strings.ToLower(tradeLimit.Symbol), event))
		}
	}

	// ltcusdt@aggTrade/ethusdt@aggTrade/solusdt@aggTrade /perpusdt@aggTrade
	wsConnection := ExchangeClient.Listen(fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s", strings.Join(streams[:], "/")), eventChannel)
	defer wsConnection.Close()

	http.ListenAndServe(":8080", nil)
}
