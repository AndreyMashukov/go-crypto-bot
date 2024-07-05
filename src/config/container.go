package config

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/controller"
	"gitlab.com/open-soft/go-crypto-bot/src/event_subscriber"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/service/ml"
	"gitlab.com/open-soft/go-crypto-bot/src/service/strategy"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"gitlab.com/open-soft/go-crypto-bot/src/validator"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

const BotExchangeBinance = "binance"
const BotExchangeByBit = "bybit"

func InitServiceContainer() Container {
	if runtime.GOMAXPROCS(0) < 2 {
		procs := runtime.GOMAXPROCS(2)
		log.Printf("GOMAXPROCS is set to: %d", procs)
	}

	db, err := sql.Open("mysql", os.Getenv("DATABASE_DSN"))

	if err != nil {
		log.Fatal(fmt.Sprintf("[DB] MySQL can't connect: %s", err.Error()))
	}

	db.SetMaxIdleConns(64)
	db.SetMaxOpenConns(64)
	db.SetConnMaxLifetime(time.Minute)

	swapDb, swapErr := sql.Open("mysql", os.Getenv("DATABASE_DSN"))

	swapDb.SetMaxIdleConns(64)
	swapDb.SetMaxOpenConns(64)
	swapDb.SetConnMaxLifetime(time.Minute)

	if swapErr != nil {
		log.Fatal(fmt.Sprintf("[Swap DB] MySQL can't connect: %s", err.Error()))
	}

	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_DSN"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	clickhouseDb := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{os.Getenv("CLICKHOUSE_DSN")},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 30 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Protocol: clickhouse.HTTP,
	})
	chErr := clickhouseDb.Ping()

	if chErr != nil {
		log.Panic(fmt.Sprintf("[Stat DB] Clickhouse can't connect: %s", chErr.Error()))
	}
	clickhouseDb.SetMaxIdleConns(64)
	clickhouseDb.SetMaxOpenConns(64)
	clickhouseDb.SetConnMaxLifetime(time.Minute)

	botRepository := repository.BotRepository{
		DB:  db,
		RDB: rdb,
		Ctx: &ctx,
	}

	botExchange := os.Getenv("BOT_EXCHANGE")
	if botExchange == "" {
		botExchange = BotExchangeBinance
	}

	currentBot := botRepository.GetCurrentBot()
	if currentBot == nil {
		botUuid := os.Getenv("BOT_UUID")
		currentBot := &model.Bot{
			BotUuid:           botUuid,
			Exchange:          botExchange,
			IsMasterBot:       false,
			IsSwapEnabled:     false,
			TradeStackSorting: model.TradeStackSortingLessPriceDiff,
			SwapConfig: model.SwapConfig{
				MinValidPercent:    2.00,
				FallPercentTrigger: -5.00,
				OrderTimeTrigger:   3600,
				UseSwapCapital:     true,
				HistoryInterval:    "1d",
				HistoryPeriod:      14,
			},
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

	formatter := utils.Formatter{}
	var exchangeApi client.ExchangeAPIInterface
	var exchangeWSStreamer strategy.ExchangeWSStreamer

	switch botExchange {
	case BotExchangeBinance:
		binanceExchange := client.Binance{
			CurrentBot:           currentBot,
			ApiKey:               os.Getenv("BINANCE_API_KEY"),
			ApiSecret:            os.Getenv("BINANCE_API_SECRET"),
			Channel:              make(chan []byte, 500),
			SocketWriter:         make(chan []byte, 500),
			RDB:                  rdb,
			Ctx:                  &ctx,
			WaitMode:             false,
			APIKeyCheckCompleted: false,
			Connected:            false,
			Lock:                 &sync.Mutex{},
		}
		binanceExchange.Connect(os.Getenv("BINANCE_WS_DSN"))
		exchangeApi = &binanceExchange
		break
	case BotExchangeByBit:
		exchangeApi = &client.ByBit{
			CurrentBot:           currentBot,
			HttpClient:           &client.HttpClient{},
			ApiKey:               os.Getenv("BYBIT_API_KEY"),
			ApiSecret:            os.Getenv("BYBIT_API_SECRET"),
			DSN:                  os.Getenv("BYBIT_API_DSN"),
			Formatter:            &formatter,
			RDB:                  rdb,
			Ctx:                  &ctx,
			APIKeyCheckCompleted: false,
		}
		break
	default:
		log.Panic(fmt.Sprintf("Unsupported exchange: %s", botExchange))
	}

	exchangeRepository := repository.ExchangeRepository{
		DB:         db,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
		Formatter:  &formatter,
		Binance:    exchangeApi,
	}

	marketDepthStrategy := strategy.MarketDepthStrategy{}
	smaStrategy := strategy.SmaTradeStrategy{
		ExchangeRepository: &exchangeRepository,
	}

	var swapStreamListener exchange.SwapStreamListenerInterface
	swapUpdater := exchange.SwapUpdater{
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
		Binance:            exchangeApi,
	}
	objectRepository := repository.ObjectRepository{
		DB:         db,
		CurrentBot: currentBot,
		RDB:        rdb,
		Ctx:        &ctx,
	}
	swapRepository := repository.SwapRepository{
		DB:               swapDb,
		RDB:              rdb,
		Ctx:              &ctx,
		CurrentBot:       currentBot,
		ObjectRepository: &objectRepository,
	}

	swapManager := exchange.SwapManager{
		SwapChainBuilder: &exchange.SwapChainBuilder{},
		SwapRepository:   &swapRepository,
		Formatter:        &formatter,
		SBSSwapFinder: &exchange.SBSSwapFinder{
			ExchangeRepository: &exchangeRepository,
			Formatter:          &formatter,
		},
		SSBSwapFinder: &exchange.SSBSwapFinder{
			ExchangeRepository: &exchangeRepository,
			Formatter:          &formatter,
		},
		SBBSwapFinder: &exchange.SBBSwapFinder{
			ExchangeRepository: &exchangeRepository,
			Formatter:          &formatter,
		},
	}

	switch botExchange {
	case BotExchangeBinance:
		exchangeWSStreamer = &strategy.BinanceWSStreamer{
			SmaTradeStrategy:    &smaStrategy,
			MarketDepthStrategy: &marketDepthStrategy,
			ExchangeRepository:  &exchangeRepository,
		}
		swapStreamListener = &exchange.BinanceSwapStreamListener{
			ExchangeRepository: &exchangeRepository,
			SwapUpdater:        &swapUpdater,
			SwapRepository:     &swapRepository,
			SwapManager:        &swapManager,
		}
		break
	case BotExchangeByBit:
		exchangeWSStreamer = &strategy.ByBitWsStreamer{
			ExchangeRepository:  &exchangeRepository,
			SmaTradeStrategy:    &smaStrategy,
			MarketDepthStrategy: &marketDepthStrategy,
			Formatter:           &formatter,
		}
		swapStreamListener = &exchange.BybitSwapStreamListener{
			ExchangeRepository: &exchangeRepository,
			SwapUpdater:        &swapUpdater,
			SwapRepository:     &swapRepository,
			SwapManager:        &swapManager,
			Formatter:          &formatter,
		}
		break
	default:
		log.Panic(fmt.Sprintf("Unsupported exchange: %s", botExchange))
	}

	statRepository := repository.StatRepository{
		DB:         clickhouseDb,
		CurrentBot: currentBot,
	}

	frameService := exchange.FrameService{
		CurrentBot: currentBot,
		RDB:        rdb,
		Ctx:        &ctx,
		Binance:    exchangeApi,
	}

	balanceService := exchange.BalanceService{
		Binance:    exchangeApi,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}

	callbackManager := service.CallbackManager{
		AutoTradeHost: "https://api.autotrade.cloud",
	}
	orderRepository := repository.OrderRepository{
		DB:               db,
		RDB:              rdb,
		Ctx:              &ctx,
		CurrentBot:       currentBot,
		ObjectRepository: &objectRepository,
	}
	botService := service.BotService{
		CurrentBot:    currentBot,
		BotRepository: &botRepository,
	}
	swapValidator := validator.SwapValidator{
		Binance:        exchangeApi,
		SwapRepository: &swapRepository,
		Formatter:      &formatter,
		BotService:     &botService,
	}

	lockTradeChannel := make(chan model.Lock)

	timeService := utils.TimeHelper{}

	profitService := exchange.ProfitService{
		Binance:    exchangeApi,
		BotService: &botService,
	}

	lossSecurity := exchange.LossSecurity{
		MlEnabled:            true,
		InterpolationEnabled: true,
		Formatter:            &formatter,
		ExchangeRepository:   &exchangeRepository,
		ProfitService:        &profitService,
	}

	signalRepository := repository.SignalRepository{
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}

	priceCalculator := exchange.PriceCalculator{
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
		FrameService:       &frameService,
		LossSecurity:       &lossSecurity,
		ProfitService:      &profitService,
		BotService:         &botService,
		SignalStorage:      &signalRepository,
	}

	pythonMLBridge := ml.PythonMLBridge{
		DataSetBuilder: &ml.DataSetBuilder{
			StatRepository: &statRepository,
		},
		ExchangeRepository: &exchangeRepository,
		SwapRepository:     &swapRepository,
		TimeService:        &timeService,
		CurrentBot:         currentBot,
		RDB:                rdb,
		Ctx:                &ctx,
		Learning:           true,
	}

	statService := service.StatService{
		Binance:            exchangeApi,
		ExchangeRepository: &exchangeRepository,
	}

	chartService := service.ChartService{
		ExchangeRepository: &exchangeRepository,
		OrderRepository:    &orderRepository,
		Formatter:          &formatter,
		StatRepository:     &statRepository,
		StatService:        &statService,
	}

	exchangeController := controller.ExchangeController{
		SwapRepository:     &swapRepository,
		ExchangeRepository: &exchangeRepository,
		ChartService:       &chartService,
		RDB:                rdb,
		Ctx:                &ctx,
		CurrentBot:         currentBot,
		BotService:         &botService,
	}

	tradeFilterService := exchange.TradeFilterService{
		OrderRepository:   &orderRepository,
		ExchangeTradeInfo: &exchangeRepository,
		ExchangePriceAPI:  exchangeApi,
		Formatter:         &formatter,
		SignalStorage:     &signalRepository,
	}

	tradeStack := exchange.TradeStack{
		OrderRepository:    &orderRepository,
		Binance:            exchangeApi,
		ExchangeRepository: &exchangeRepository,
		BalanceService:     &balanceService,
		Formatter:          &formatter,
		BotService:         &botService,
		PriceCalculator:    &priceCalculator,
		RDB:                rdb,
		Ctx:                &ctx,
		TradeFilterService: &tradeFilterService,
		SignalStorage:      &signalRepository,
	}

	orderExecutor := exchange.OrderExecutor{
		TradeStack:         &tradeStack,
		LossSecurity:       &lossSecurity,
		CurrentBot:         currentBot,
		TimeService:        &timeService,
		BalanceService:     &balanceService,
		Binance:            exchangeApi,
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		PriceCalculator:    &priceCalculator,
		ProfitService:      &profitService,
		CallbackManager:    &callbackManager,
		SwapRepository:     &swapRepository,
		SwapExecutor: &exchange.SwapExecutor{
			BalanceService:  &balanceService,
			SwapRepository:  &swapRepository,
			OrderRepository: &orderRepository,
			Binance:         exchangeApi,
			Formatter:       &formatter,
			TimeService:     &timeService,
			CurrentBot:      currentBot,
		},
		SwapValidator:          &swapValidator,
		Formatter:              &formatter,
		BotService:             &botService,
		TurboSwapProfitPercent: 20.00,
		Lock:                   make(map[string]bool),
		TradeLockMutex:         sync.RWMutex{},
		LockChannel:            &lockTradeChannel,
		CancelRequestMap:       make(map[string]bool),
	}

	makerService := exchange.MakerService{
		TradeFilterService: &tradeFilterService,
		ExchangeApi:        exchangeApi,
		Binance:            exchangeApi,
		TradeStack:         &tradeStack,
		OrderExecutor:      &orderExecutor,
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
		HoldScore:          75.00,
		CurrentBot:         currentBot,
		PriceCalculator:    &priceCalculator,
		BotService:         &botService,
		StrategyFacade: &exchange.StrategyFacade{
			MinDecisions:        3.00,
			OrderRepository:     &orderRepository,
			DecisionReadStorage: &exchangeRepository,
			ExchangeRepository:  &exchangeRepository,
			BotService:          &botService,
		},
	}

	profitOptionsValidator := validator.ProfitOptionsValidator{}
	tradeLimitValidator := validator.TradeLimitValidator{
		ProfitOptionsValidator: &profitOptionsValidator,
	}

	orderController := controller.OrderController{
		RDB:                    rdb,
		Ctx:                    &ctx,
		OrderRepository:        &orderRepository,
		ExchangeRepository:     &exchangeRepository,
		Formatter:              &formatter,
		PriceCalculator:        &priceCalculator,
		CurrentBot:             currentBot,
		LossSecurity:           &lossSecurity,
		OrderExecutor:          &orderExecutor,
		ProfitOptionsValidator: &profitOptionsValidator,
		BotService:             &botService,
		ProfitService:          &profitService,
		TradeFilterService:     &tradeFilterService,
		ExchangeAPI:            exchangeApi,
	}

	tradeController := controller.TradeController{
		CurrentBot:          currentBot,
		ExchangeRepository:  &exchangeRepository,
		TradeStack:          &tradeStack,
		TradeLimitValidator: &tradeLimitValidator,
		SignalRepository:    &signalRepository,
	}

	baseKLineStrategy := strategy.BaseKLineStrategy{
		OrderRepository:    &orderRepository,
		TradeStack:         &tradeStack,
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
		MlEnabled:          true,
	}
	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: &exchangeRepository,
		OrderRepository:    &orderRepository,
		ProfitService:      &profitService,
		BotService:         &botService,
		SignalStorage:      &signalRepository,
	}

	go func() {
		for {
			lock := <-lockTradeChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}()

	healthService := service.HealthService{
		BotRepository:      &botRepository,
		ExchangeRepository: &exchangeRepository,
		PythonMLBridge:     &pythonMLBridge,
		Binance:            exchangeApi,
		CurrentBot:         currentBot,
		DB:                 db,
		SwapDb:             swapDb,
		RDB:                rdb,
		Ctx:                &ctx,
	}

	botController := controller.BotController{
		HealthService: &healthService,
		CurrentBot:    currentBot,
		BotRepository: &botRepository,
	}

	mcGatewayAddress := os.Getenv("MC_DSN")

	mcListener := exchange.MCListener{
		MSGatewayAddress:   mcGatewayAddress,
		ExchangeRepository: &exchangeRepository,
	}

	eventDispatcher := service.EventDispatcher{
		Subscribers: []event_subscriber.SubscriberInterface{
			&service.KLineEventSubscriber{
				ExchangeRepository: &exchangeRepository,
				StatRepository:     &statRepository,
				BotService:         &botService,
				StatService:        &statService,
			},
		},
		Enabled: false,
	}

	return Container{
		PriceCalculator:     &priceCalculator,
		BotController:       &botController,
		HealthService:       &healthService,
		Db:                  db,
		DbSwap:              swapDb,
		CurrentBot:          currentBot,
		CallbackManager:     &callbackManager,
		BalanceService:      &balanceService,
		TimeService:         &timeService,
		Binance:             exchangeApi,
		PythonMLBridge:      &pythonMLBridge,
		SwapRepository:      &swapRepository,
		ExchangeRepository:  &exchangeRepository,
		OrderRepository:     &orderRepository,
		ExchangeController:  &exchangeController,
		TradeController:     &tradeController,
		OrderController:     &orderController,
		MakerService:        &makerService,
		OrderExecutor:       &orderExecutor,
		SwapManager:         &swapManager,
		SwapUpdater:         &swapUpdater,
		SmaTradeStrategy:    &smaStrategy,
		MarketDepthStrategy: &marketDepthStrategy,
		OrderBasedStrategy:  &orderBasedStrategy,
		BaseKLineStrategy:   &baseKLineStrategy,
		IsMasterBot:         botService.IsMasterBot(),
		MarketTradeListener: &strategy.MarketTradeListener{
			SmaTradeStrategy:    &smaStrategy,
			MarketDepthStrategy: &marketDepthStrategy,
			OrderBasedStrategy:  &orderBasedStrategy,
			BaseKLineStrategy:   &baseKLineStrategy,
			ExchangeRepository:  &exchangeRepository,
			TimeService:         &timeService,
			Binance:             exchangeApi,
			PythonMLBridge:      &pythonMLBridge,
			PriceCalculator:     &priceCalculator,
			EventDispatcher:     &eventDispatcher,
			ExchangeWSStreamer:  exchangeWSStreamer,
		},
		MarketSwapListener: &exchange.MarketSwapListener{
			ExchangeRepository: &exchangeRepository,
			TimeService:        &timeService,
			SwapManager:        &swapManager,
			SwapStreamListener: swapStreamListener,
		},
		MCListener:      &mcListener,
		EventDispatcher: &eventDispatcher,
	}
}

type Container struct {
	EventDispatcher     *service.EventDispatcher
	MCListener          *exchange.MCListener
	PriceCalculator     *exchange.PriceCalculator
	BotController       *controller.BotController
	HealthService       *service.HealthService
	Db                  *sql.DB
	DbSwap              *sql.DB
	CurrentBot          *model.Bot
	CallbackManager     *service.CallbackManager
	BalanceService      *exchange.BalanceService
	TimeService         *utils.TimeHelper
	Binance             client.ExchangeAPIInterface
	PythonMLBridge      *ml.PythonMLBridge
	SwapRepository      *repository.SwapRepository
	ExchangeRepository  *repository.ExchangeRepository
	OrderRepository     *repository.OrderRepository
	ExchangeController  *controller.ExchangeController
	TradeController     *controller.TradeController
	OrderController     *controller.OrderController
	MakerService        *exchange.MakerService
	OrderExecutor       *exchange.OrderExecutor
	SwapManager         *exchange.SwapManager
	SwapUpdater         *exchange.SwapUpdater
	SmaTradeStrategy    *strategy.SmaTradeStrategy
	MarketDepthStrategy *strategy.MarketDepthStrategy
	BaseKLineStrategy   *strategy.BaseKLineStrategy
	OrderBasedStrategy  *strategy.OrderBasedStrategy
	MarketTradeListener *strategy.MarketTradeListener
	MarketSwapListener  *exchange.MarketSwapListener
	IsMasterBot         bool
}

func (c *Container) StartHttpServer() {
	// todo: use GIN http server
	// configure controllers
	http.HandleFunc("/kline/list/", c.ExchangeController.GetKlineListAction)
	http.HandleFunc("/depth/", c.ExchangeController.GetDepthAction)
	http.HandleFunc("/trade/list/", c.ExchangeController.GetTradeListAction)
	http.HandleFunc("/swap/list", c.ExchangeController.GetSwapListAction)
	http.HandleFunc("/chart/list", c.ExchangeController.GetChartListAction)
	http.HandleFunc("/order/list", c.OrderController.GetOrderListAction)
	http.HandleFunc("/order/extra/charge/update", c.OrderController.UpdateExtraChargeAction)
	http.HandleFunc("/order/profit/options/update", c.OrderController.UpdateProfitOptionsAction)
	http.HandleFunc("/order/pending/list", c.OrderController.GetPendingOrderListAction)
	http.HandleFunc("/order/position/list", c.OrderController.GetPositionListAction)
	http.HandleFunc("/order", c.OrderController.PostManualOrderAction)
	http.HandleFunc("/order/", c.OrderController.DeleteManualOrderAction)
	http.HandleFunc("/order/cancel/", c.OrderController.DeleteCancelExchangeOrderAction)
	http.HandleFunc("/order/trade/list", c.OrderController.GetOrderTradeListAction)
	http.HandleFunc("/trade/limit/list", c.TradeController.GetTradeLimitsAction)
	http.HandleFunc("/trade/stack", c.TradeController.GetTradeStackAction)
	http.HandleFunc("/trade/signal", c.TradeController.PostSignalAction)
	http.HandleFunc("/trade/limit/create", c.TradeController.CreateTradeLimitAction)
	http.HandleFunc("/trade/limit/update", c.TradeController.UpdateTradeLimitAction)
	http.HandleFunc("/trade/limit/switch/", c.TradeController.SwitchTradeLimitAction)
	http.HandleFunc("/trade/limit/sentiment/", c.TradeController.PatchSentimentAction)
	http.HandleFunc("/health/check", c.BotController.GetHealthCheckAction)
	http.HandleFunc("/bot/update", c.BotController.PutConfigAction)

	// Start HTTP server!
	go func() {
		_ = http.ListenAndServe(":8080", nil)
	}()
}

func (c *Container) PingDB() {
	go func() {
		for {
			err := c.Db.Ping()
			if err != nil {
				log.Printf("[DB] Connection ping error: %s", err.Error())
			}

			err = c.DbSwap.Ping()
			if err != nil {
				log.Printf("[Swap DB] Connection ping error: %s", err.Error())
			}

			time.Sleep(time.Second * 30)
		}
	}()
}
