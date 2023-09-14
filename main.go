package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	ExchangeClient "gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/controller"
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

	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	httpClient := http.Client{}
	binance := ExchangeClient.Binance{
		ApiKey:         "0XVVs5VRWyjJH1fMReQyVUS614C8FlF1rnmvCZN2iK3UDhwncqpGYzF1jgV8KPLM",
		ApiSecret:      "tg5Ak5LoTFSCIadQLn5LkcnWHEPYSiA6wpY3rEqx89GG2aj9ZWsDyMl17S5TjTHM",
		DestinationURI: "https://testnet.binance.vision",
		HttpClient:     &httpClient,
	}
	orderRepository := ExchangeRepository.OrderRepository{
		DB:  db,
		RDB: rdb,
		Ctx: &ctx,
	}
	exchangeRepository := ExchangeRepository.ExchangeRepository{
		DB:  db,
		RDB: rdb,
		Ctx: &ctx,
	}

	chartService := ExchangeService.ChartService{
		ExchangeRepository: &exchangeRepository,
		OrderRepository:    &orderRepository,
	}
	exchangeController := controller.ExchangeController{
		ExchangeRepository: &exchangeRepository,
		ChartService:       &chartService,
	}
	orderController := controller.OrderController{
		OrderRepository: &orderRepository,
	}

	http.HandleFunc("/kline/list/", exchangeController.GetKlineListAction)
	http.HandleFunc("/depth/", exchangeController.GetDepthAction)
	http.HandleFunc("/trade/list/", exchangeController.GetTradeListAction)
	http.HandleFunc("/order/list", orderController.GetOrderListAction)
	http.HandleFunc("/chart/list", exchangeController.GetChartListAction)

	eventChannel := make(chan []byte)
	tradeLogChannel := make(chan ExchangeModel.Trade)
	kLineLogChannel := make(chan ExchangeModel.KLine)
	lockTradeChannel := make(chan ExchangeModel.Lock)
	depthChannel := make(chan ExchangeModel.Depth)

	makerService := ExchangeService.MakerService{
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Binance:            &binance,
		LockChannel:        &lockTradeChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		MinDecisions:       3.00,
		HoldScore:          75.00,
	}

	// todo: BuyExtraOnMarketFallStrategy
	baseKLineStrategy := ExchangeService.BaseKLineStrategy{}
	marketDepthStrategy := ExchangeService.MarketDepthStrategy{}
	smaStrategy := ExchangeService.SmaTradeStrategy{
		ExchangeRepository: exchangeRepository,
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

	tradeLimits := exchangeRepository.GetTradeLimits()

	for _, limit := range tradeLimits {
		go func(symbol string) {
			for {
				currentDecisions := make([]ExchangeModel.Decision, 0)
				smaDecision := exchangeRepository.GetDecision("sma_trade_strategy")
				kLineDecision := exchangeRepository.GetDecision("base_kline_strategy")
				marketDepthDecision := exchangeRepository.GetDecision("market_depth_strategy")

				if smaDecision != nil {
					currentDecisions = append(currentDecisions, *smaDecision)
				}
				if kLineDecision != nil {
					currentDecisions = append(currentDecisions, *kLineDecision)
				}
				if marketDepthDecision != nil {
					currentDecisions = append(currentDecisions, *marketDepthDecision)
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

			switch true {
			case strings.Contains(eventModel.Stream, "@aggTrade"):
				var tradeEvent ExchangeModel.TradeEvent
				json.Unmarshal(message, &tradeEvent)
				trade := tradeEvent.Trade
				smaDecision := smaStrategy.Decide(trade)
				exchangeRepository.SetDecision(smaDecision)

				go func() {
					tradeLogChannel <- tradeEvent.Trade
				}()
				break
			case strings.Contains(eventModel.Stream, "@kline_1m"):
				var klineEvent ExchangeModel.KlineEvent
				json.Unmarshal(message, &klineEvent)
				kLine := klineEvent.KlineData.Kline
				exchangeRepository.AddKLine(kLine)
				baseKLineDecision := baseKLineStrategy.Decide(kLine)
				exchangeRepository.SetDecision(baseKLineDecision)
				go func() {
					kLineLogChannel <- kLine
				}()
				break
			case strings.Contains(eventModel.Stream, "@depth"):
				var depthEvent ExchangeModel.DepthEvent
				json.Unmarshal(message, &depthEvent)
				depth := depthEvent.Depth
				depthDecision := marketDepthStrategy.Decide(depth)
				exchangeRepository.SetDecision(depthDecision)
				go func() {
					depthChannel <- depth
				}()
				break
			}
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
			depth := <-depthChannel
			makerService.SetDepth(depth)
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
	events := [3]string{"@aggTrade", "@kline_1m", "@depth"}

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
