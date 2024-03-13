package config

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/controller"
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
	"sync"
	"time"
)

func InitServiceContainer() Container {
	db, err := sql.Open("mysql", os.Getenv("DATABASE_DSN"))

	db.SetMaxIdleConns(64)
	db.SetMaxOpenConns(64)
	db.SetConnMaxLifetime(time.Minute)

	swapDb, err := sql.Open("mysql", os.Getenv("DATABASE_DSN"))

	swapDb.SetMaxIdleConns(64)
	swapDb.SetMaxOpenConns(64)
	swapDb.SetConnMaxLifetime(time.Minute)

	if err != nil {
		log.Fatal(fmt.Sprintf("MySQL can't connect: %s", err.Error()))
	}

	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_DSN"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	httpClient := http.Client{}
	binance := client.Binance{
		ApiKey:               os.Getenv("BINANCE_API_KEY"),
		ApiSecret:            os.Getenv("BINANCE_API_SECRET"),
		HttpClient:           &httpClient,
		Channel:              make(chan []byte),
		SocketWriter:         make(chan []byte),
		RDB:                  rdb,
		Ctx:                  &ctx,
		WaitMode:             false,
		APIKeyCheckCompleted: false,
		Connected:            false,
	}

	frameService := exchange.FrameService{
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
			BotUuid:       botUuid,
			IsMasterBot:   false,
			IsSwapEnabled: false,
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

	balanceService := exchange.BalanceService{
		Binance:    &binance,
		RDB:        rdb,
		Ctx:        &ctx,
		CurrentBot: currentBot,
	}

	callbackManager := service.CallbackManager{
		AutoTradeHost: "https://api.autotrade.cloud",
	}

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

	formatter := utils.Formatter{}
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

	botService := service.BotService{
		CurrentBot:    currentBot,
		BotRepository: &botRepository,
	}

	swapValidator := validator.SwapValidator{
		Binance:        &binance,
		SwapRepository: &swapRepository,
		Formatter:      &formatter,
		BotService:     &botService,
	}

	lockTradeChannel := make(chan model.Lock)

	// own net: ATOM, XMR, XLM, DOT, ADA, XRP
	btcDependent := []string{"LTC", "ZEC", "ATOM", "XMR", "DOT", "XRP", "BCH", "ADA", "ETH", "DOGE", "PERP", "NEO"}
	etcDependent := []string{"SHIB", "LINK", "UNI", "NEAR", "XLM", "ETC", "MATIC", "SOL", "BNB", "AVAX", "TRX"}

	timeService := utils.TimeHelper{}

	pythonMLBridge := ml.PythonMLBridge{
		DataSetBuilder: &ml.DataSetBuilder{
			ExcludeDependedDataset: []string{"SHIBUSDT", "BTCUSDT"},
			BtcDependent:           btcDependent,
			EthDependent:           etcDependent,
		},
		ExchangeRepository: &exchangeRepository,
		SwapRepository:     &swapRepository,
		TimeService:        &timeService,
		CurrentBot:         currentBot,
		RDB:                rdb,
		Ctx:                &ctx,
		Learning:           true,
	}

	profitService := exchange.ProfitService{
		Binance:    &binance,
		BotService: &botService,
	}

	lossSecurity := exchange.LossSecurity{
		MlEnabled:            true,
		InterpolationEnabled: true,
		Formatter:            &formatter,
		ExchangeRepository:   &exchangeRepository,
		ProfitService:        &profitService,
	}

	priceCalculator := exchange.PriceCalculator{
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Binance:            &binance,
		Formatter:          &formatter,
		FrameService:       &frameService,
		LossSecurity:       &lossSecurity,
		ProfitService:      &profitService,
		BotService:         &botService,
	}

	tradeStack := exchange.TradeStack{
		OrderRepository:    &orderRepository,
		Binance:            &binance,
		ExchangeRepository: &exchangeRepository,
		BalanceService:     &balanceService,
		Formatter:          &formatter,
		BotService:         &botService,
		PriceCalculator:    &priceCalculator,
	}

	orderExecutor := exchange.OrderExecutor{
		TradeStack:         &tradeStack,
		LossSecurity:       &lossSecurity,
		CurrentBot:         currentBot,
		TimeService:        &timeService,
		BalanceService:     &balanceService,
		Binance:            &binance,
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
			Binance:         &binance,
			Formatter:       &formatter,
			TimeService:     &timeService,
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
		TradeStack:         &tradeStack,
		OrderExecutor:      &orderExecutor,
		OrderRepository:    &orderRepository,
		ExchangeRepository: &exchangeRepository,
		Binance:            &binance,
		Formatter:          &formatter,
		HoldScore:          75.00,
		CurrentBot:         currentBot,
		PriceCalculator:    &priceCalculator,
		BotService:         &botService,
		StrategyFacade: &exchange.StrategyFacade{
			MinDecisions:        4.00,
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
	}

	tradeController := controller.TradeController{
		CurrentBot:          currentBot,
		ExchangeRepository:  &exchangeRepository,
		TradeStack:          &tradeStack,
		TradeLimitValidator: &tradeLimitValidator,
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

	baseKLineStrategy := strategy.BaseKLineStrategy{
		ExchangeRepository: &exchangeRepository,
		Formatter:          &formatter,
		MlEnabled:          true,
	}
	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: &exchangeRepository,
		OrderRepository:    &orderRepository,
		TradeStack:         &tradeStack,
		ProfitService:      &profitService,
		BotService:         &botService,
	}
	marketDepthStrategy := strategy.MarketDepthStrategy{}
	smaStrategy := strategy.SmaTradeStrategy{
		ExchangeRepository: exchangeRepository,
	}

	swapUpdater := exchange.SwapUpdater{
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
		BotRepository:      &botRepository,
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
		BotRepository: &botRepository,
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
			Binance:             &binance,
			PythonMLBridge:      &pythonMLBridge,
			PriceCalculator:     &priceCalculator,
		},
		MarketSwapListener: &exchange.MarketSwapListener{
			ExchangeRepository: &exchangeRepository,
			TimeService:        &timeService,
			SwapManager:        &swapManager,
			SwapUpdater:        &swapUpdater,
			SwapRepository:     &swapRepository,
			SwapKlineChannel:   make(chan []byte),
		},
	}
}

type Container struct {
	PriceCalculator     *exchange.PriceCalculator
	BotController       *controller.BotController
	HealthService       *service.HealthService
	Db                  *sql.DB
	DbSwap              *sql.DB
	CurrentBot          *model.Bot
	CallbackManager     *service.CallbackManager
	BalanceService      *exchange.BalanceService
	TimeService         *utils.TimeHelper
	Binance             *client.Binance
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
	http.HandleFunc("/order/trade/list", c.OrderController.GetOrderTradeListAction)
	http.HandleFunc("/trade/limit/list", c.TradeController.GetTradeLimitsAction)
	http.HandleFunc("/trade/stack", c.TradeController.GetTradeStackAction)
	http.HandleFunc("/trade/limit/create", c.TradeController.CreateTradeLimitAction)
	http.HandleFunc("/trade/limit/update", c.TradeController.UpdateTradeLimitAction)
	http.HandleFunc("/trade/limit/switch/", c.TradeController.SwitchTradeLimitAction)
	http.HandleFunc("/health/check", c.BotController.GetHealthCheckAction)
	http.HandleFunc("/bot/update", c.BotController.PutConfigAction)

	// Start HTTP server!
	go func() {
		_ = http.ListenAndServe(":8080", nil)
	}()
}
