package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	ExchangeClient "gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/controller"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	ExchangeService "gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

func getStreamBatch(tradeLimits []ExchangeModel.SymbolInterface, events []string) [][]string {
	streamBatch := make([][]string, 0)

	streams := make([]string, 0)

	for _, tradeLimit := range tradeLimits {
		for i := 0; i < len(events); i++ {
			event := events[i]
			streams = append(streams, fmt.Sprintf("%s%s", strings.ToLower(tradeLimit.GetSymbol()), event))
		}

		if len(streams) >= 24 {
			streamBatch = append(streamBatch, streams)
			streams = make([]string, 0)
		}
	}

	if len(streams) > 0 {
		streamBatch = append(streamBatch, streams)
	}

	return streamBatch
}

func main() {
	pwd, _ := os.Getwd()
	if _, err := os.Stat(fmt.Sprintf("%s/.env", pwd)); err == nil {
		log.Println(".env is found, loading variables...")
		err = godotenv.Load()
		if err != nil {
			log.Println(err)
		}
	}

	db, err := sql.Open("mysql", os.Getenv("DATABASE_DSN")) // root:go_crypto_bot@tcp(mysql:3306)/go_crypto_bot
	defer db.Close()

	db.SetMaxIdleConns(64)
	db.SetMaxOpenConns(64)
	db.SetConnMaxLifetime(time.Minute)

	swapDb, err := sql.Open("mysql", os.Getenv("DATABASE_DSN")) // root:go_crypto_bot@tcp(mysql:3306)/go_crypto_bot
	defer swapDb.Close()

	swapDb.SetMaxIdleConns(64)
	swapDb.SetMaxOpenConns(64)
	swapDb.SetConnMaxLifetime(time.Minute)

	if err != nil {
		log.Fatal(fmt.Sprintf("MySQL can't connect: %s", err.Error()))
	}

	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_DSN"), //"redis:6379",
		Password: "",                     // no password set
		DB:       0,                      // use default DB
	})

	httpClient := http.Client{}
	binance := ExchangeClient.Binance{
		ApiKey:         os.Getenv("BINANCE_API_KEY"),    // "0XVVs5VRWyjJH1fMReQyVUS614C8FlF1rnmvCZN2iK3UDhwncqpGYzF1jgV8KPLM",
		ApiSecret:      os.Getenv("BINANCE_API_SECRET"), // "tg5Ak5LoTFSCIadQLn5LkcnWHEPYSiA6wpY3rEqx89GG2aj9ZWsDyMl17S5TjTHM",
		DestinationURI: os.Getenv("BINANCE_API_DSN"),    // "https://testnet.binance.vision",
		HttpClient:     &httpClient,
		Channel:        make(chan []byte),
		SocketWriter:   make(chan []byte),
		RDB:            rdb,
		Ctx:            &ctx,
	}
	binance.Connect(os.Getenv("BINANCE_WS_DSN")) // "wss://testnet.binance.vision/ws-api/v3"

	frameService := ExchangeService.FrameService{
		RDB:     rdb,
		Ctx:     &ctx,
		Binance: &binance,
	}

	botRepository := ExchangeRepository.BotRepository{
		DB:  db,
		RDB: rdb,
		Ctx: &ctx,
	}

	currentBot := botRepository.GetCurrentBot()
	if currentBot == nil {
		botUuid := os.Getenv("BOT_UUID")
		currentBot := &ExchangeModel.Bot{
			BotUuid: botUuid,
		}
		err := botRepository.Create(*currentBot)
		if err != nil {
			panic(err)
		}

		currentBot = botRepository.GetCurrentBot()
		if currentBot == nil {
			panic(fmt.Sprintf("Can't initialize bot: %s", botUuid))
		}
	}

	isMasterBot := currentBot.BotUuid == "5b51a35f-76a6-4747-8461-850fff9f7c18"
	swapEnabled := currentBot.BotUuid == "5b51a35f-76a6-4747-8461-850fff9f7c18"

	log.Printf("Bot [%s] is initialized successfully", currentBot.BotUuid)

	orderRepository := ExchangeRepository.OrderRepository{
		DB:         db,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}
	exchangeRepository := ExchangeRepository.ExchangeRepository{
		DB:         db,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}
	swapRepository := ExchangeRepository.SwapRepository{
		DB:         swapDb,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}

	formatter := ExchangeService.Formatter{}
	chartService := ExchangeService.ChartService{
		ExchangeRepository: &exchangeRepository,
		OrderRepository:    &orderRepository,
	}
	exchangeController := controller.ExchangeController{
		SwapRepository:     &swapRepository,
		ExchangeRepository: &exchangeRepository,
		ChartService:       &chartService,
		RDB:                rdb,
		Ctx:                &ctx,
		CurrentBot:         currentBot,
	}

	// Swap Settings
	swapMinPercentValid := 1.15
	swapOrderOnProfitPercent := -1.00
	var swapOpenedSellOrderFromHoursOpened int64 = 2
	//swapMinPercentValid := 0.1
	//swapOrderOnProfitPercent := 10.00
	//var swapOpenedSellOrderFromHoursOpened int64 = 0

	swapValidator := ExchangeService.SwapValidator{
		Binance:        &binance,
		SwapRepository: &swapRepository,
		Formatter:      &formatter,
		SwapMinPercent: swapMinPercentValid,
	}

	eventChannel := make(chan []byte)
	lockTradeChannel := make(chan ExchangeModel.Lock)
	depthChannel := make(chan ExchangeModel.Depth)

	balanceService := ExchangeService.BalanceService{
		Binance:    &binance,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}

	pythonMLBridge := ExchangeService.PythonMLBridge{
		DataSetBuilder:     &ExchangeService.DataSetBuilder{},
		ExchangeRepository: &exchangeRepository,
		SwapRepository:     &swapRepository,
		CurrentBot:         currentBot,
		RDB:                rdb,
		Ctx:                &ctx,
		Learning:           true,
	}
	pythonMLBridge.Initialize()
	defer pythonMLBridge.Finalize()

	priceCalculator := ExchangeService.PriceCalculator{
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Binance:            &binance,
		Formatter:          &formatter,
		FrameService:       &frameService,
	}

	timeService := ExchangeService.TimeService{}

	telegramNotificator := ExchangeService.TelegramNotificator{
		AutoTradeHost: "https://api.autotrade.cloud",
	}

	orderExecutor := ExchangeService.OrderExecutor{
		CurrentBot:          currentBot,
		TimeService:         &timeService,
		BalanceService:      &balanceService,
		Binance:             &binance,
		OrderRepository:     &orderRepository,
		ExchangeRepository:  &exchangeRepository,
		PriceCalculator:     &priceCalculator,
		TelegramNotificator: &telegramNotificator,
		SwapRepository:      &swapRepository,
		SwapExecutor: &ExchangeService.SwapExecutor{
			BalanceService:  &balanceService,
			SwapRepository:  &swapRepository,
			OrderRepository: &orderRepository,
			Binance:         &binance,
			Formatter:       &formatter,
			TimeService:     &timeService,
		},
		SwapValidator:          &swapValidator,
		Formatter:              &formatter,
		SwapSellOrderDays:      swapOpenedSellOrderFromHoursOpened,
		SwapEnabled:            swapEnabled,
		SwapProfitPercent:      swapOrderOnProfitPercent,
		TurboSwapProfitPercent: 20.00,
		Lock:                   make(map[string]bool),
		TradeLockMutex:         sync.RWMutex{},
		LockChannel:            &lockTradeChannel,
	}

	makerService := ExchangeService.MakerService{
		OrderExecutor:      &orderExecutor,
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Binance:            &binance,
		Formatter:          &formatter,
		MinDecisions:       4.00,
		HoldScore:          75.00,
		CurrentBot:         currentBot,
		PriceCalculator:    &priceCalculator,
	}

	orderController := controller.OrderController{
		RDB:                rdb,
		Ctx:                &ctx,
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
		PriceCalculator:    &priceCalculator,
		CurrentBot:         currentBot,
	}

	tradeController := controller.TradeController{
		CurrentBot:         currentBot,
		ExchangeRepository: &exchangeRepository,
	}

	http.HandleFunc("/kline/list/", exchangeController.GetKlineListAction)
	http.HandleFunc("/depth/", exchangeController.GetDepthAction)
	http.HandleFunc("/trade/list/", exchangeController.GetTradeListAction)
	http.HandleFunc("/swap/list", exchangeController.GetSwapListAction)
	http.HandleFunc("/chart/list", exchangeController.GetChartListAction)
	http.HandleFunc("/order/list", orderController.GetOrderListAction)
	http.HandleFunc("/order/position/list", orderController.GetPositionListAction)
	http.HandleFunc("/order", orderController.PostManualOrderAction)
	http.HandleFunc("/order/trade/list", orderController.GetOrderTradeListAction)
	http.HandleFunc("/trade/limit/list", tradeController.GetTradeLimitsAction)
	http.HandleFunc("/trade/limit/create", tradeController.CreateTradeLimitAction)
	http.HandleFunc("/trade/limit/update", tradeController.UpdateTradeLimitAction)

	go func() {
		for {
			makerService.UpdateLimits()
			time.Sleep(time.Minute * 5)
		}
	}()

	if isMasterBot {
		makerService.UpdateSwapPairs()
	}

	baseKLineStrategy := ExchangeService.BaseKLineStrategy{
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
	}
	orderBasedStrategy := ExchangeService.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderRepository,
	}
	marketDepthStrategy := ExchangeService.MarketDepthStrategy{}
	smaStrategy := ExchangeService.SmaTradeStrategy{
		ExchangeRepository: exchangeRepository,
	}

	swapUpdater := ExchangeService.SwapUpdater{
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
		Binance:            &binance,
	}

	if isMasterBot {
		swapManager := ExchangeService.SwapManager{
			SwapChainBuilder: &ExchangeService.SwapChainBuilder{},
			SwapRepository:   &swapRepository,
			Formatter:        &formatter,
			SBSSwapFinder: &ExchangeService.SBSSwapFinder{
				ExchangeRepository: &exchangeRepository,
				Formatter:          &formatter,
			},
			SSBSwapFinder: &ExchangeService.SSBSwapFinder{
				ExchangeRepository: &exchangeRepository,
				Formatter:          &formatter,
			},
			SBBSwapFinder: &ExchangeService.SBBSwapFinder{
				ExchangeRepository: &exchangeRepository,
				Formatter:          &formatter,
			},
		}

		swapKlineChannel := make(chan []byte)
		defer close(swapKlineChannel)

		go func() {
			for {
				baseAssets := make([]string, 0)
				for _, pair := range exchangeRepository.GetSwapPairs() {
					if !slices.Contains(baseAssets, pair.BaseAsset) {
						baseAssets = append(baseAssets, pair.BaseAsset)
					}
				}

				for _, baseAsset := range baseAssets {
					swapManager.CalculateSwapOptions(baseAsset)
				}

				time.Sleep(time.Millisecond * 250)
			}
		}()

		// existing swaps real time monitoring
		go func() {
			for {
				swapMsg := <-swapKlineChannel
				swapSymbol := ""

				if strings.Contains(string(swapMsg), "kline") {
					var event ExchangeModel.KlineEvent
					json.Unmarshal(swapMsg, &event)
					kLine := event.KlineData.Kline
					exchangeRepository.AddKLine(kLine)
					swapSymbol = kLine.Symbol
				}

				if strings.Contains(string(swapMsg), "@depth20") {
					var event ExchangeModel.OrderBookEvent
					json.Unmarshal(swapMsg, &event)

					depth := event.Depth.ToDepth(strings.ToUpper(strings.ReplaceAll(event.Stream, "@depth20@1000ms", "")))
					exchangeRepository.SetDepth(depth)
					swapSymbol = depth.Symbol
				}

				if swapSymbol == "" {
					continue
				}

				swapPair, err := exchangeRepository.GetSwapPair(swapSymbol)
				if err == nil {
					swapUpdater.UpdateSwapPair(swapPair)

					possibleSwap := swapRepository.GetSwapChainCache(swapPair.BaseAsset)
					if possibleSwap != nil {
						go func(asset string) {
							swapManager.CalculateSwapOptions(asset)
						}(swapPair.BaseAsset)
					}
				}
			}
		}()

		swapWebsockets := make([]*websocket.Conn, 0)

		swapPairCollection := make([]ExchangeModel.SymbolInterface, 0)
		for _, swapPair := range exchangeRepository.GetSwapPairs() {
			swapPairCollection = append(swapPairCollection, swapPair)
		}

		for index, streamBatchItem := range getStreamBatch(swapPairCollection, []string{"@kline_1m", "@depth20@1000ms"}) {
			swapWebsockets = append(swapWebsockets, ExchangeClient.Listen(fmt.Sprintf(
				"%s/stream?streams=%s",
				"wss://stream.binance.com:9443",
				strings.Join(streamBatchItem, "/"),
			), swapKlineChannel, []string{}, 10000+int64(index)))

			log.Printf("Swap batch %d websocket: %s", index, strings.Join(streamBatchItem, ", "))

			defer swapWebsockets[index].Close()
		}
	}

	go func(orderExecutor *ExchangeService.OrderExecutor) {
		for {
			lock := <-lockTradeChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	tradeLimits := exchangeRepository.GetTradeLimits()

	for _, limit := range tradeLimits {
		go func(limit ExchangeModel.TradeLimit) {
			for {
				// todo: write to database and read from database
				err := pythonMLBridge.LearnModel(limit.Symbol)
				if err != nil {
					log.Printf("[%s] %s", limit.Symbol, err.Error())
					timeService.WaitSeconds(60)
					continue
				}

				timeService.WaitSeconds(3600 * 6)
			}
		}(limit)
		// learn every 1000 minutes

		go func(symbol string) {
			for {
				currentDecisions := make([]ExchangeModel.Decision, 0)
				smaDecision := exchangeRepository.GetDecision("sma_trade_strategy")
				kLineDecision := exchangeRepository.GetDecision("base_kline_strategy")
				marketDepthDecision := exchangeRepository.GetDecision("market_depth_strategy")
				orderBasedDecision := exchangeRepository.GetDecision("order_based_strategy")

				if smaDecision != nil {
					currentDecisions = append(currentDecisions, *smaDecision)
				}
				if kLineDecision != nil {
					currentDecisions = append(currentDecisions, *kLineDecision)
				}
				if marketDepthDecision != nil {
					currentDecisions = append(currentDecisions, *marketDepthDecision)
				}
				if orderBasedDecision != nil {
					currentDecisions = append(currentDecisions, *orderBasedDecision)
				}

				if len(currentDecisions) > 0 {
					makerService.Make(symbol, currentDecisions)
				}
				time.Sleep(time.Millisecond * 500)
			}
		}(limit.Symbol)
	}

	predictChannel := make(chan string)
	go func(channel chan string) {
		for {
			symbol := <-channel
			predicted, err := pythonMLBridge.Predict(symbol)
			if err == nil && predicted > 0.00 {
				kLine := exchangeRepository.GetLastKLine(symbol)
				if kLine != nil {
					exchangeRepository.SaveKLinePredict(predicted, *kLine)
				}
				exchangeRepository.SavePredict(predicted, symbol)
			}
		}
	}(predictChannel)

	go func() {
		for {
			message := <-eventChannel

			switch true {
			case strings.Contains(string(message), "aggTrade"):
				var tradeEvent ExchangeModel.TradeEvent
				json.Unmarshal(message, &tradeEvent)
				smaDecision := smaStrategy.Decide(tradeEvent.Trade)
				exchangeRepository.SetDecision(smaDecision)

				go func(channel chan string, symbol string) {
					predictChannel <- symbol
				}(predictChannel, tradeEvent.Trade.Symbol)
				break
			case strings.Contains(string(message), "kline"):
				var event ExchangeModel.KlineEvent
				json.Unmarshal(message, &event)
				kLine := event.KlineData.Kline
				exchangeRepository.AddKLine(kLine)

				go func(channel chan string, symbol string) {
					predictChannel <- symbol
				}(predictChannel, kLine.Symbol)

				baseKLineDecision := baseKLineStrategy.Decide(kLine)
				exchangeRepository.SetDecision(baseKLineDecision)
				orderBasedDecision := orderBasedStrategy.Decide(kLine)
				exchangeRepository.SetDecision(orderBasedDecision)

				break
			case strings.Contains(string(message), "depth20"):
				var event ExchangeModel.OrderBookEvent
				json.Unmarshal(message, &event)

				depth := event.Depth.ToDepth(strings.ToUpper(strings.ReplaceAll(event.Stream, "@depth20@100ms", "")))
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
			depth := <-depthChannel
			exchangeRepository.SetDepth(depth)
		}
	}()

	if err != nil {
		log.Println(err)
	}

	websockets := make([]*websocket.Conn, 0)

	tradeLimitCollection := make([]ExchangeModel.SymbolInterface, 0)
	hasBtcUsdt := false
	for _, limit := range tradeLimits {
		tradeLimitCollection = append(tradeLimitCollection, limit)

		history := binance.GetKLines(limit.GetSymbol(), "1m", 200)
		for _, kline := range history {
			exchangeRepository.AddKLine(kline.ToKLine(limit.GetSymbol()))
		}

		if "BTCUSDT" == limit.GetSymbol() {
			hasBtcUsdt = true
		}
	}

	if !hasBtcUsdt {
		tradeLimitCollection = append(tradeLimitCollection, ExchangeModel.DummySymbol{Symbol: "BTCUSDT"})
	}

	for index, streamBatchItem := range getStreamBatch(tradeLimitCollection, []string{"@aggTrade", "@kline_1m", "@depth20@100ms"}) {
		websockets = append(websockets, ExchangeClient.Listen(fmt.Sprintf(
			"%s/stream?streams=%s",
			os.Getenv("BINANCE_STREAM_DSN"),
			strings.Join(streamBatchItem, "/"),
		), eventChannel, []string{}, int64(index)))

		log.Printf("Batch %d websocket: %s", index, strings.Join(streamBatchItem, ", "))

		defer websockets[index].Close()
	}

	http.ListenAndServe(":8080", nil)
}
