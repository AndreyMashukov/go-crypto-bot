package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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

	swapEnabled := currentBot.Id == 1

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
		DB:         db,
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
		ExchangeRepository: &exchangeRepository,
		ChartService:       &chartService,
		RDB:                rdb,
		Ctx:                &ctx,
		CurrentBot:         currentBot,
	}

	// Swap Settings
	swapMinPercentValid := 10.70
	swapOrderOnProfitPercent := -1.00
	var swapOpenedSellOrderFromHoursOpened int64 = 2
	//swapMinPercentValid := 0.1
	//swapOrderOnProfitPercent := 10.00
	//var swapOpenedSellOrderFromHoursOpened int64 = 0

	swapValidator := ExchangeService.SwapValidator{
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

	makerService := ExchangeService.MakerService{
		SwapValidator:      &swapValidator,
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		SwapRepository:     &swapRepository,
		Binance:            &binance,
		LockChannel:        &lockTradeChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		Formatter:          &formatter,
		FrameService:       &frameService,
		MinDecisions:       4.00,
		HoldScore:          75.00,
		SwapSellOrderDays:  swapOpenedSellOrderFromHoursOpened,
		SwapProfitPercent:  swapOrderOnProfitPercent,
		RDB:                rdb,
		Ctx:                &ctx,
		CurrentBot:         currentBot,
		BalanceService:     &balanceService,
		SwapEnabled:        swapEnabled,
	}

	orderController := controller.OrderController{
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
		MakerService:       &makerService,
		CurrentBot:         currentBot,
	}

	tradeController := controller.TradeController{
		CurrentBot:         currentBot,
		ExchangeRepository: &exchangeRepository,
	}

	http.HandleFunc("/kline/list/", exchangeController.GetKlineListAction)
	http.HandleFunc("/depth/", exchangeController.GetDepthAction)
	http.HandleFunc("/trade/list/", exchangeController.GetTradeListAction)
	http.HandleFunc("/chart/list", exchangeController.GetChartListAction)
	http.HandleFunc("/order/list", orderController.GetOrderListAction)
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

	if swapEnabled {
		makerService.UpdateSwapPairs()
	}

	// todo: BuyExtraOnMarketFallStrategy
	baseKLineStrategy := ExchangeService.BaseKLineStrategy{}
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

	if swapEnabled {
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

		// existing swaps real time monitoring
		go func() {
			for {
				swapPairs := exchangeRepository.GetSwapPairs()
				assets := make([]string, 0)
				for _, swapPair := range swapPairs {
					possibleSwap := swapRepository.GetSwapChainCache(swapPair.BaseAsset)
					if possibleSwap != nil {
						if !slices.Contains(assets, swapPair.BaseAsset) {
							assets = append(assets, swapPair.BaseAsset)
						}

						swapUpdater.UpdateSwapPair(swapPair, false)
					}
				}

				for _, asset := range assets {
					go func(asset string) {
						log.Printf("realtime swap check: %s", asset)
						swapManager.CalculateSwapOptions(asset)
					}(asset)
				}
				time.Sleep(time.Second)
			}
		}()

		// new possible swaps search algorithm
		go func() {
			iterator := 0
			for {
				time.Sleep(time.Second * 20)
				baseAssets := make([]string, 0)
				swapPairs := exchangeRepository.GetSwapPairs()
				for _, swapPair := range swapPairs {
					possibleSwap := swapRepository.GetSwapChainCache(swapPair.BaseAsset)
					if possibleSwap == nil {
						if !slices.Contains(baseAssets, swapPair.BaseAsset) {
							baseAssets = append(baseAssets, swapPair.BaseAsset)
						}

						swapUpdater.UpdateSwapPair(swapPair, iterator == 0)
					}
				}

				for _, baseAsset := range baseAssets {
					go func(baseAsset string) {
						swapManager.CalculateSwapOptions(baseAsset)
					}(baseAsset)
				}
				iterator++
				if iterator > 20 {
					iterator = 0
				}
			}
		}()
	}

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

	go func() {
		for {
			// Read the channel, todo -> better to use select: https://go.dev/tour/concurrency/5
			message := <-eventChannel

			switch true {
			case strings.Contains(string(message), "aggTrade"):
				var trade ExchangeModel.Trade
				json.Unmarshal(message, &trade)
				smaDecision := smaStrategy.Decide(trade)
				exchangeRepository.SetDecision(smaDecision)
				break
			case strings.Contains(string(message), "kline"):
				var klineData ExchangeModel.KlineData
				json.Unmarshal(message, &klineData)
				kLine := klineData.Kline

				exchangeRepository.AddKLine(kLine)
				baseKLineDecision := baseKLineStrategy.Decide(kLine)
				exchangeRepository.SetDecision(baseKLineDecision)
				orderBasedDecision := orderBasedStrategy.Decide(kLine)
				exchangeRepository.SetDecision(orderBasedDecision)

				break
			case strings.Contains(string(message), "depthUpdate"):
				var depth ExchangeModel.Depth
				json.Unmarshal(message, &depth)
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
			makerService.SetDepth(depth)
		}
	}()

	// todo: sync existed orders in Binance with bot database...

	streams := []string{}
	events := [3]string{"@aggTrade", "@kline_1m", "@depth20@100ms"}

	for _, tradeLimit := range tradeLimits {
		for i := 0; i < len(events); i++ {
			event := events[i]
			streams = append(streams, fmt.Sprintf("%s%s", strings.ToLower(tradeLimit.Symbol), event))
		}

		kLines := exchangeRepository.KLineList(tradeLimit.Symbol, false, 200)
		if len(kLines) < 200 {
			history := binance.GetKLines(tradeLimit.Symbol, "1m", 200)

			for _, kline := range history {
				dto := kline.ToKLine(tradeLimit.Symbol)
				exchangeRepository.AddKLine(kline.ToKLine(tradeLimit.Symbol))
				log.Printf("[%s] Added history for [%d] = %.8f", dto.Symbol, dto.Timestamp, dto.Close)
			}
		}
	}

	userDataStreamStart, err := binance.UserDataStreamStart()

	if err != nil {
		log.Println(err)
	}

	wsConnection := ExchangeClient.Listen(fmt.Sprintf(
		"%s/ws/%s",
		os.Getenv("BINANCE_STREAM_DSN"),
		userDataStreamStart.ListenKey,
	), eventChannel, streams, 1)
	defer wsConnection.Close()

	http.ListenAndServe(":8080", nil)
}
