package config

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/controller"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func InitServiceContainer() Container {
	db, err := sql.Open("mysql", os.Getenv("DATABASE_DSN")) // root:go_crypto_bot@tcp(mysql:3306)/go_crypto_bot

	db.SetMaxIdleConns(64)
	db.SetMaxOpenConns(64)
	db.SetConnMaxLifetime(time.Minute)

	swapDb, err := sql.Open("mysql", os.Getenv("DATABASE_DSN")) // root:go_crypto_bot@tcp(mysql:3306)/go_crypto_bot

	swapDb.SetMaxIdleConns(64)
	swapDb.SetMaxOpenConns(64)
	swapDb.SetConnMaxLifetime(time.Minute)

	if err != nil {
		log.Fatal(fmt.Sprintf("MySQL can't connect: %s", err.Error()))
	}

	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_DSN"),      // redis:6379,
		Password: os.Getenv("REDIS_PASSWORD"), // redis password
		DB:       0,                           // use default DB
	})

	httpClient := http.Client{}
	binance := client.Binance{
		ApiKey:               os.Getenv("BINANCE_API_KEY"),    // "0XVVs5VRWyjJH1fMReQyVUS614C8FlF1rnmvCZN2iK3UDhwncqpGYzF1jgV8KPLM",
		ApiSecret:            os.Getenv("BINANCE_API_SECRET"), // "tg5Ak5LoTFSCIadQLn5LkcnWHEPYSiA6wpY3rEqx89GG2aj9ZWsDyMl17S5TjTHM",
		DestinationURI:       os.Getenv("BINANCE_API_DSN"),    // "https://testnet.binance.vision",
		HttpClient:           &httpClient,
		Channel:              make(chan []byte),
		SocketWriter:         make(chan []byte),
		RDB:                  rdb,
		Ctx:                  &ctx,
		WaitMode:             false,
		APIKeyCheckCompleted: false,
		Connected:            false,
	}

	frameService := service.FrameService{
		RDB:     rdb,
		Ctx:     &ctx,
		Binance: &binance,
	}

	botRepository := repository.BotRepository{
		DB:  db,
		RDB: rdb,
		Ctx: &ctx,
	}

	currentBot := botRepository.GetCurrentBot()
	if currentBot == nil {
		botUuid := os.Getenv("BOT_UUID")
		currentBot := &model.Bot{
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

	balanceService := service.BalanceService{
		Binance:    &binance,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}

	callbackManager := service.CallbackManager{
		AutoTradeHost: "https://api.autotrade.cloud",
	}

	isMasterBot := currentBot.BotUuid == "5b51a35f-76a6-4747-8461-850fff9f7c18"
	swapEnabled := currentBot.BotUuid == "5b51a35f-76a6-4747-8461-850fff9f7c18"

	orderRepository := repository.OrderRepository{
		DB:         db,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}
	exchangeRepository := repository.ExchangeRepository{
		DB:         db,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}
	swapRepository := repository.SwapRepository{
		DB:         swapDb,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}

	formatter := service.Formatter{}
	chartService := service.ChartService{
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

	swapValidator := service.SwapValidator{
		Binance:        &binance,
		SwapRepository: &swapRepository,
		Formatter:      &formatter,
		SwapMinPercent: swapMinPercentValid,
	}

	lockTradeChannel := make(chan model.Lock)

	// own net: ATOM, XMR, XLM, DOT, ADA, XRP
	btcDependent := []string{"LTC", "ZEC", "ATOM", "XMR", "DOT", "XRP", "BCH", "ADA", "ETH", "DOGE", "PERP"}
	etcDependent := []string{"SHIB", "LINK", "UNI", "NEAR", "XLM", "ETC", "MATIC", "SOL", "BNB", "AVAX", "TRX", "NEO"}

	pythonMLBridge := service.PythonMLBridge{
		DataSetBuilder: &service.DataSetBuilder{
			ExcludeDependedDataset: []string{"SHIBUSDT", "BTCUSDT"},
			BtcDependent:           btcDependent,
			EthDependent:           etcDependent,
		},
		ExchangeRepository: &exchangeRepository,
		SwapRepository:     &swapRepository,
		CurrentBot:         currentBot,
		RDB:                rdb,
		Ctx:                &ctx,
		Learning:           true,
	}

	timeService := service.TimeService{}

	lossSecurity := service.LossSecurity{
		MlEnabled:            true,
		InterpolationEnabled: true,
		Formatter:            &formatter,
		ExchangeRepository:   &exchangeRepository,
		Binance:              &binance,
	}

	priceCalculator := service.PriceCalculator{
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Binance:            &binance,
		Formatter:          &formatter,
		FrameService:       &frameService,
		LossSecurity:       &lossSecurity,
	}

	orderExecutor := service.OrderExecutor{
		LossSecurity:       &lossSecurity,
		CurrentBot:         currentBot,
		TimeService:        &timeService,
		BalanceService:     &balanceService,
		Binance:            &binance,
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		PriceCalculator:    &priceCalculator,
		CallbackManager:    &callbackManager,
		SwapRepository:     &swapRepository,
		SwapExecutor: &service.SwapExecutor{
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

	tradeStack := service.TradeStack{
		OrderRepository:    &orderRepository,
		Binance:            &binance,
		ExchangeRepository: &exchangeRepository,
		BalanceService:     &balanceService,
		Formatter:          &formatter,
	}

	makerService := service.MakerService{
		TradeStack:         &tradeStack,
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

	swapManager := service.SwapManager{
		SwapChainBuilder: &service.SwapChainBuilder{},
		SwapRepository:   &swapRepository,
		Formatter:        &formatter,
		SBSSwapFinder: &service.SBSSwapFinder{
			ExchangeRepository: &exchangeRepository,
			Formatter:          &formatter,
		},
		SSBSwapFinder: &service.SSBSwapFinder{
			ExchangeRepository: &exchangeRepository,
			Formatter:          &formatter,
		},
		SBBSwapFinder: &service.SBBSwapFinder{
			ExchangeRepository: &exchangeRepository,
			Formatter:          &formatter,
		},
	}

	baseKLineStrategy := service.BaseKLineStrategy{
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
		MlEnabled:          true,
	}
	orderBasedStrategy := service.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderRepository,
		TradeStack:         &tradeStack,
	}
	marketDepthStrategy := service.MarketDepthStrategy{}
	smaStrategy := service.SmaTradeStrategy{
		ExchangeRepository: exchangeRepository,
	}

	swapUpdater := service.SwapUpdater{
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
		Binance:            &binance,
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
		ExchangeRepository: &exchangeRepository,
		PythonMLBridge:     &pythonMLBridge,
		Binance:            &binance,
		CurrentBot:         currentBot,
		DB:                 swapDb,
		RDB:                rdb,
		Ctx:                &ctx,
	}

	botController := controller.BotController{
		HealthService: &healthService,
		CurrentBot:    currentBot,
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
		Binance:             &binance,
		PythonMLBridge:      &pythonMLBridge,
		SwapRepository:      &swapRepository,
		ExchangeRepository:  &exchangeRepository,
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
		IsMasterBot:         isMasterBot,
	}
}

type Container struct {
	PriceCalculator     *service.PriceCalculator
	BotController       *controller.BotController
	HealthService       *service.HealthService
	Db                  *sql.DB
	DbSwap              *sql.DB
	CurrentBot          *model.Bot
	CallbackManager     *service.CallbackManager
	BalanceService      *service.BalanceService
	TimeService         *service.TimeService
	Binance             *client.Binance
	PythonMLBridge      *service.PythonMLBridge
	SwapRepository      *repository.SwapRepository
	ExchangeRepository  *repository.ExchangeRepository
	ExchangeController  *controller.ExchangeController
	TradeController     *controller.TradeController
	OrderController     *controller.OrderController
	MakerService        *service.MakerService
	OrderExecutor       *service.OrderExecutor
	SwapManager         *service.SwapManager
	SwapUpdater         *service.SwapUpdater
	SmaTradeStrategy    *service.SmaTradeStrategy
	MarketDepthStrategy *service.MarketDepthStrategy
	BaseKLineStrategy   *service.BaseKLineStrategy
	OrderBasedStrategy  *service.OrderBasedStrategy
	IsMasterBot         bool
}

func (c *Container) StartHttpServer() {
	// configure controllers
	http.HandleFunc("/kline/list/", c.ExchangeController.GetKlineListAction)
	http.HandleFunc("/depth/", c.ExchangeController.GetDepthAction)
	http.HandleFunc("/trade/list/", c.ExchangeController.GetTradeListAction)
	http.HandleFunc("/swap/list", c.ExchangeController.GetSwapListAction)
	http.HandleFunc("/chart/list", c.ExchangeController.GetChartListAction)
	http.HandleFunc("/order/list", c.OrderController.GetOrderListAction)
	http.HandleFunc("/order/position/list", c.OrderController.GetPositionListAction)
	http.HandleFunc("/order", c.OrderController.PostManualOrderAction)
	http.HandleFunc("/order/trade/list", c.OrderController.GetOrderTradeListAction)
	http.HandleFunc("/trade/limit/list", c.TradeController.GetTradeLimitsAction)
	http.HandleFunc("/trade/limit/create", c.TradeController.CreateTradeLimitAction)
	http.HandleFunc("/trade/limit/update", c.TradeController.UpdateTradeLimitAction)
	http.HandleFunc("/health/check", c.BotController.GetHealthCheck)

	// todo: add health check method!

	// Start HTTP server!
	go func() {
		_ = http.ListenAndServe(":8080", nil)
	}()
}
