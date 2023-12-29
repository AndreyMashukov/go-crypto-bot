package tests

import (
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"sync"
	"testing"
	"time"
)

func TestSellAction(t *testing.T) {
	assertion := assert.New(t)

	balanceService := new(BalanceServiceMock)
	binance := new(ExchangeOrderAPIMock)
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(ExchangeTradeInfoMock)
	swapRepository := new(SwapRepositoryMock)
	priceCalculator := new(PriceCalculatorMock)
	swapExecutor := new(SwapExecutorMock)
	swapValidator := new(SwapValidatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

	swapRepository.On("GetSwapChainCache", "ETH").Return(nil)

	lockChannel := make(chan model.Lock)

	orderExecutor := service.OrderExecutor{
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:         timeService,
		BalanceService:      balanceService,
		Binance:             binance,
		OrderRepository:     orderRepository,
		ExchangeRepository:  exchangeRepository,
		SwapRepository:      swapRepository,
		PriceCalculator:     priceCalculator,
		SwapExecutor:        swapExecutor,
		SwapValidator:       swapValidator,
		Formatter:           &service.Formatter{},
		SwapSellOrderDays:   10,
		SwapEnabled:         true,
		SwapProfitPercent:   1.50,
		LockChannel:         &lockChannel,
		Lock:                make(map[string]bool),
		TradeLockMutex:      sync.RWMutex{},
		TelegramNotificator: telegramNotificatorMock,
	}

	go func(orderExecutor *service.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:     999,
		Symbol:      "ETHUSDT",
		Side:        "SELL",
		ExecutedQty: 0.00,
		OrigQty:     0.0089,
		Status:      "NEW",
		Price:       2212.92,
	}
	timeService.On("GetNowDateTimeString").Return("2023-12-28 00:52:00")
	orderRepository.On("GetBinanceOrder", "ETHUSDT", "SELL").Return(nil)
	binance.On("GetOpenedOrders").Return([]model.BinanceOrder{
		initialBinanceOrder,
	}, nil)
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(1)
	priceCalculator.On("GetDepth", "ETHUSDT").Return(model.Depth{
		Symbol: "ETHUSDT",
		Asks: [][2]model.Number{
			{
				{
					2212.92,
				},
				{
					0.009,
				},
			},
		},
		Bids: [][2]model.Number{
			{
				{
					2212.92,
				},
				{
					0.009,
				},
			},
		},
	})
	tradeLimit := model.TradeLimit{
		Symbol:           "ETHUSDT",
		MinProfitPercent: 3.1,
		MinPrice:         0.01,
		MinQuantity:      0.0001,
	}
	exchangeRepository.On("GetTradeLimit", "ETHUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetLastKLine", "ETHUSDT").Return(&model.KLine{
		Symbol: "ETHUSDT",
		Close:  2281.52,
	})
	openedExternalId := int64(988)
	openedOrder := model.Order{
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "ETHUSDT",
		Quantity:         0.009,
		Price:            2212.92,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 0.009,
	}
	orderRepository.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(openedOrder, nil)
	orderRepository.On("GetManualOrder", "ETHUSDT").Return(nil)
	timeService.On("WaitMilliseconds", int64(20)).Maybe()
	binance.On("QueryOrder", "ETHUSDT", int64(999)).Return(model.BinanceOrder{
		OrderId:             999,
		Symbol:              "ETHUSDT",
		Side:                "SELL",
		ExecutedQty:         0.0089,
		OrigQty:             0.0089,
		Status:              "FILLED",
		Price:               2212.92,
		CummulativeQuoteQty: 0.009 * 2212.92,
	}, nil)
	orderRepository.On("DeleteBinanceOrder", initialBinanceOrder).Times(1)
	orderId := int64(100)
	orderRepository.On("Create", mock.Anything).Return(&orderId, nil)
	orderRepository.On("DeleteManualOrder", "ETHUSDT").Times(1)
	orderRepository.On("Find", orderId).Times(1).Return(model.Order{}, nil)
	orderRepository.On("GetClosesOrderList", openedOrder).Times(1).Return([]model.Order{
		{
			Status:           "closed",
			ExecutedQuantity: 0.0089,
			Price:            2212.92,
		},
	})
	orderRepository.On("Update", mock.Anything).Times(1).Return(nil)
	balanceService.On("InvalidateBalanceCache", "USDT").Times(1)
	balanceService.On("InvalidateBalanceCache", "ETH").Times(1)

	telegramNotificatorMock.On("SellOrder", mock.Anything, mock.Anything, mock.Anything).Times(1)

	err := orderExecutor.Sell(tradeLimit, openedOrder, "ETHUSDT", 2281.52, 0.0089, 0, 0, 0)
	assertion.Nil(err)
	assertion.Equal("closed", orderRepository.Updated.Status)
	assertion.Equal(2212.92, orderRepository.Updated.Price)
	assertion.Equal(openedExternalId, *orderRepository.Updated.ExternalId)
}

func TestSellFoundFilled(t *testing.T) {
	assertion := assert.New(t)

	balanceService := new(BalanceServiceMock)
	binance := new(ExchangeOrderAPIMock)
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(ExchangeTradeInfoMock)
	swapRepository := new(SwapRepositoryMock)
	priceCalculator := new(PriceCalculatorMock)
	swapExecutor := new(SwapExecutorMock)
	swapValidator := new(SwapValidatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

	swapRepository.On("GetSwapChainCache", "ETH").Return(nil)

	lockChannel := make(chan model.Lock)

	orderExecutor := service.OrderExecutor{
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:         timeService,
		BalanceService:      balanceService,
		Binance:             binance,
		OrderRepository:     orderRepository,
		ExchangeRepository:  exchangeRepository,
		SwapRepository:      swapRepository,
		PriceCalculator:     priceCalculator,
		SwapExecutor:        swapExecutor,
		SwapValidator:       swapValidator,
		Formatter:           &service.Formatter{},
		SwapSellOrderDays:   10,
		SwapEnabled:         true,
		SwapProfitPercent:   1.50,
		LockChannel:         &lockChannel,
		Lock:                make(map[string]bool),
		TradeLockMutex:      sync.RWMutex{},
		TelegramNotificator: telegramNotificatorMock,
	}

	go func(orderExecutor *service.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:             999,
		Symbol:              "ETHUSDT",
		Side:                "SELL",
		ExecutedQty:         0.0089,
		OrigQty:             0.0089,
		Status:              "FILLED",
		Price:               2212.92,
		CummulativeQuoteQty: 0.009 * 2212.92,
	}
	timeService.On("GetNowDateTimeString").Return("2023-12-28 00:52:00")
	orderRepository.On("GetBinanceOrder", "ETHUSDT", "SELL").Return(nil)
	binance.On("GetOpenedOrders").Return([]model.BinanceOrder{
		initialBinanceOrder,
	}, nil)
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(1)
	priceCalculator.On("GetDepth", "ETHUSDT").Return(model.Depth{
		Symbol: "ETHUSDT",
		Asks: [][2]model.Number{
			{
				{
					2212.92,
				},
				{
					0.009,
				},
			},
		},
		Bids: [][2]model.Number{
			{
				{
					2212.92,
				},
				{
					0.009,
				},
			},
		},
	})
	tradeLimit := model.TradeLimit{
		Symbol:           "ETHUSDT",
		MinProfitPercent: 3.1,
		MinPrice:         0.01,
		MinQuantity:      0.0001,
	}
	exchangeRepository.On("GetTradeLimit", "ETHUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetLastKLine", "ETHUSDT").Return(&model.KLine{
		Symbol: "ETHUSDT",
		Close:  2281.52,
	})
	openedExternalId := int64(988)
	openedOrder := model.Order{
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "ETHUSDT",
		Quantity:         0.009,
		Price:            2212.92,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 0.009,
	}
	orderRepository.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(openedOrder, nil)
	orderRepository.On("GetManualOrder", "ETHUSDT").Unset()
	timeService.On("WaitMilliseconds", int64(20)).Unset()
	binance.On("QueryOrder", "ETHUSDT", int64(999)).Unset()
	orderRepository.On("DeleteBinanceOrder", initialBinanceOrder).Times(1)
	orderId := int64(100)
	orderRepository.On("Create", mock.Anything).Return(&orderId, nil)
	orderRepository.On("DeleteManualOrder", "ETHUSDT").Times(1)
	orderRepository.On("Find", orderId).Times(1).Return(model.Order{}, nil)
	orderRepository.On("GetClosesOrderList", openedOrder).Times(1).Return([]model.Order{
		{
			Status:           "closed",
			ExecutedQuantity: 0.0089,
			Price:            2212.92,
		},
	})
	orderRepository.On("Update", mock.Anything).Times(1).Return(nil)
	balanceService.On("InvalidateBalanceCache", "USDT").Times(1)
	balanceService.On("InvalidateBalanceCache", "ETH").Times(1)

	telegramNotificatorMock.On("SellOrder", mock.Anything, mock.Anything, mock.Anything).Times(1)

	err := orderExecutor.Sell(tradeLimit, openedOrder, "ETHUSDT", 2281.52, 0.0089, 0, 0, 0)
	assertion.Nil(err)
	assertion.Equal("closed", orderRepository.Updated.Status)
	assertion.Equal(2212.92, orderRepository.Updated.Price)
	assertion.Equal(openedExternalId, *orderRepository.Updated.ExternalId)
}

func TestSellCancelledInProcess(t *testing.T) {
	assertion := assert.New(t)

	balanceService := new(BalanceServiceMock)
	binance := new(ExchangeOrderAPIMock)
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(ExchangeTradeInfoMock)
	swapRepository := new(SwapRepositoryMock)
	priceCalculator := new(PriceCalculatorMock)
	swapExecutor := new(SwapExecutorMock)
	swapValidator := new(SwapValidatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

	swapRepository.On("GetSwapChainCache", "ETH").Return(nil)

	lockChannel := make(chan model.Lock)

	orderExecutor := service.OrderExecutor{
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:         timeService,
		BalanceService:      balanceService,
		Binance:             binance,
		OrderRepository:     orderRepository,
		ExchangeRepository:  exchangeRepository,
		SwapRepository:      swapRepository,
		PriceCalculator:     priceCalculator,
		SwapExecutor:        swapExecutor,
		SwapValidator:       swapValidator,
		Formatter:           &service.Formatter{},
		SwapSellOrderDays:   10,
		SwapEnabled:         true,
		SwapProfitPercent:   1.50,
		LockChannel:         &lockChannel,
		Lock:                make(map[string]bool),
		TradeLockMutex:      sync.RWMutex{},
		TelegramNotificator: telegramNotificatorMock,
	}

	go func(orderExecutor *service.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:     999,
		Symbol:      "ETHUSDT",
		Side:        "SELL",
		ExecutedQty: 0.00,
		OrigQty:     0.0089,
		Status:      "NEW",
		Price:       2212.92,
	}
	timeService.On("GetNowDateTimeString").Return("2023-12-28 00:52:00")
	orderRepository.On("GetBinanceOrder", "ETHUSDT", "SELL").Return(nil)
	binance.On("GetOpenedOrders").Return([]model.BinanceOrder{
		initialBinanceOrder,
	}, nil)
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(1)
	priceCalculator.On("GetDepth", "ETHUSDT").Return(model.Depth{
		Symbol: "ETHUSDT",
		Asks: [][2]model.Number{
			{
				{
					2212.92,
				},
				{
					0.009,
				},
			},
		},
		Bids: [][2]model.Number{
			{
				{
					2212.92,
				},
				{
					0.009,
				},
			},
		},
	})
	tradeLimit := model.TradeLimit{
		Symbol:           "ETHUSDT",
		MinProfitPercent: 3.1,
		MinPrice:         0.01,
		MinQuantity:      0.0001,
	}
	exchangeRepository.On("GetTradeLimit", "ETHUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetLastKLine", "ETHUSDT").Return(&model.KLine{
		Symbol: "ETHUSDT",
		Close:  2281.52,
	})
	openedExternalId := int64(988)
	openedOrder := model.Order{
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "ETHUSDT",
		Quantity:         0.009,
		Price:            2212.92,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 0.009,
	}
	orderRepository.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(openedOrder, nil)
	orderRepository.On("GetManualOrder", "ETHUSDT").Return(nil)
	timeService.On("WaitMilliseconds", int64(20)).Maybe()
	binance.On("QueryOrder", "ETHUSDT", int64(999)).Return(model.BinanceOrder{
		OrderId:             999,
		Symbol:              "ETHUSDT",
		Side:                "SELL",
		ExecutedQty:         0.0000,
		OrigQty:             0.0089,
		Status:              "CANCELED",
		Price:               2212.92,
		CummulativeQuoteQty: 0.00,
	}, nil)
	orderRepository.On("DeleteBinanceOrder", initialBinanceOrder).Times(1)
	orderId := int64(100)
	orderRepository.On("Create", mock.Anything).Return(&orderId, nil).Unset()
	orderRepository.On("DeleteManualOrder", "ETHUSDT").Unset()
	balanceService.On("InvalidateBalanceCache", "USDT").Times(1)
	balanceService.On("InvalidateBalanceCache", "ETH").Times(1)

	err := orderExecutor.Sell(tradeLimit, openedOrder, "ETHUSDT", 2281.52, 0.0089, 0, 0, 0)
	assertion.Error(errors.New("Order is cancelled"), err)
}

func TestSellQueryFail(t *testing.T) {
	assertion := assert.New(t)

	balanceService := new(BalanceServiceMock)
	binance := new(ExchangeOrderAPIMock)
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(ExchangeTradeInfoMock)
	swapRepository := new(SwapRepositoryMock)
	priceCalculator := new(PriceCalculatorMock)
	swapExecutor := new(SwapExecutorMock)
	swapValidator := new(SwapValidatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

	swapRepository.On("GetSwapChainCache", "ETH").Return(nil)

	lockChannel := make(chan model.Lock)

	orderExecutor := service.OrderExecutor{
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:         timeService,
		BalanceService:      balanceService,
		Binance:             binance,
		OrderRepository:     orderRepository,
		ExchangeRepository:  exchangeRepository,
		SwapRepository:      swapRepository,
		PriceCalculator:     priceCalculator,
		SwapExecutor:        swapExecutor,
		SwapValidator:       swapValidator,
		Formatter:           &service.Formatter{},
		SwapSellOrderDays:   10,
		SwapEnabled:         true,
		SwapProfitPercent:   1.50,
		LockChannel:         &lockChannel,
		Lock:                make(map[string]bool),
		TradeLockMutex:      sync.RWMutex{},
		TelegramNotificator: telegramNotificatorMock,
	}

	go func(orderExecutor *service.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:     999,
		Symbol:      "ETHUSDT",
		Side:        "SELL",
		ExecutedQty: 0.00,
		OrigQty:     0.0089,
		Status:      "NEW",
		Price:       2212.92,
	}
	timeService.On("GetNowDateTimeString").Return("2023-12-28 00:52:00")
	orderRepository.On("GetBinanceOrder", "ETHUSDT", "SELL").Return(nil)
	binance.On("GetOpenedOrders").Return([]model.BinanceOrder{
		initialBinanceOrder,
	}, nil)
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(1)
	priceCalculator.On("GetDepth", "ETHUSDT").Return(model.Depth{
		Symbol: "ETHUSDT",
		Asks: [][2]model.Number{
			{
				{
					2212.92,
				},
				{
					0.009,
				},
			},
		},
		Bids: [][2]model.Number{
			{
				{
					2212.92,
				},
				{
					0.009,
				},
			},
		},
	})
	tradeLimit := model.TradeLimit{
		Symbol:           "ETHUSDT",
		MinProfitPercent: 3.1,
		MinPrice:         0.01,
		MinQuantity:      0.0001,
	}
	exchangeRepository.On("GetTradeLimit", "ETHUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetLastKLine", "ETHUSDT").Return(&model.KLine{
		Symbol: "ETHUSDT",
		Close:  2281.52,
	})
	openedExternalId := int64(988)
	openedOrder := model.Order{
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "ETHUSDT",
		Quantity:         0.009,
		Price:            2212.92,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 0.009,
	}
	orderRepository.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(openedOrder, nil)
	orderRepository.On("GetManualOrder", "ETHUSDT").Return(nil)
	timeService.On("WaitMilliseconds", int64(20)).Maybe()
	binance.On("QueryOrder", "ETHUSDT", int64(999)).Return(model.BinanceOrder{}, errors.New("Order was canceled or expired"))
	orderRepository.On("DeleteBinanceOrder", initialBinanceOrder).Times(1)
	orderId := int64(100)
	orderRepository.On("Create", mock.Anything).Return(&orderId, nil).Unset()
	orderRepository.On("DeleteManualOrder", "ETHUSDT").Unset()
	balanceService.On("InvalidateBalanceCache", "USDT").Times(1)
	balanceService.On("InvalidateBalanceCache", "ETH").Times(1)

	err := orderExecutor.Sell(tradeLimit, openedOrder, "ETHUSDT", 2281.52, 0.0089, 0, 0, 0)
	assertion.Equal(errors.New("Order was canceled or expired"), err)
}

func TestSellClosingAction(t *testing.T) {
	assertion := assert.New(t)

	balanceService := new(BalanceServiceMock)
	binance := new(ExchangeOrderAPIMock)
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(ExchangeTradeInfoMock)
	swapRepository := new(SwapRepositoryMock)
	priceCalculator := new(PriceCalculatorMock)
	swapExecutor := new(SwapExecutorMock)
	swapValidator := new(SwapValidatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

	swapRepository.On("GetSwapChainCache", "BTC").Return(nil)

	lockChannel := make(chan model.Lock)

	orderExecutor := service.OrderExecutor{
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:         timeService,
		BalanceService:      balanceService,
		Binance:             binance,
		OrderRepository:     orderRepository,
		ExchangeRepository:  exchangeRepository,
		SwapRepository:      swapRepository,
		PriceCalculator:     priceCalculator,
		SwapExecutor:        swapExecutor,
		SwapValidator:       swapValidator,
		Formatter:           &service.Formatter{},
		SwapSellOrderDays:   10,
		SwapEnabled:         true,
		SwapProfitPercent:   1.50,
		LockChannel:         &lockChannel,
		Lock:                make(map[string]bool),
		TradeLockMutex:      sync.RWMutex{},
		TelegramNotificator: telegramNotificatorMock,
	}

	go func(orderExecutor *service.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:     999,
		Symbol:      "BTCUSDT",
		Side:        "SELL",
		ExecutedQty: 0.00,
		OrigQty:     0.00046,
		Status:      "NEW",
		Price:       43496.99,
	}
	timeService.On("GetNowDateTimeString").Return("2023-12-28 00:52:00")
	orderRepository.On("GetBinanceOrder", "BTCUSDT", "SELL").Return(nil)
	binance.On("GetOpenedOrders").Return([]model.BinanceOrder{
		initialBinanceOrder,
	}, nil)
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(1)
	priceCalculator.On("GetDepth", "BTCUSDT").Return(model.Depth{
		Symbol: "BTCUSDT",
		Asks: [][2]model.Number{
			{
				{
					43496.99,
				},
				{
					0.009,
				},
			},
		},
		Bids: [][2]model.Number{
			{
				{
					43496.99,
				},
				{
					0.009,
				},
			},
		},
	})
	tradeLimit := model.TradeLimit{
		Symbol:           "BTCUSDT",
		MinProfitPercent: 3.1,
		MinPrice:         0.01,
		MinQuantity:      0.00001,
	}
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetLastKLine", "BTCUSDT").Return(&model.KLine{
		Symbol: "BTCUSDT",
		Close:  43496.99,
	})
	openedExternalId := int64(988)
	openedOrder := model.Order{
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "BTCUSDT",
		Quantity:         0.00047,
		Price:            42026.08,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 0.00047,
	}
	orderRepository.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(openedOrder, nil)
	orderRepository.On("GetManualOrder", "BTCUSDT").Return(nil)
	timeService.On("WaitMilliseconds", int64(20)).Maybe()
	binance.On("QueryOrder", "BTCUSDT", int64(999)).Return(model.BinanceOrder{
		OrderId:             999,
		Symbol:              "BTCUSDT",
		Side:                "SELL",
		ExecutedQty:         0.00046,
		OrigQty:             0.00046,
		Status:              "FILLED",
		Price:               43496.99,
		CummulativeQuoteQty: 0.00046 * 43496.99,
	}, nil)
	orderRepository.On("DeleteBinanceOrder", initialBinanceOrder).Times(1)
	orderId := int64(100)
	orderRepository.On("Create", mock.Anything).Return(&orderId, nil)
	orderRepository.On("DeleteManualOrder", "BTCUSDT").Times(1)
	orderRepository.On("Find", orderId).Times(1).Return(model.Order{}, nil)
	orderRepository.On("GetClosesOrderList", openedOrder).Times(1).Return([]model.Order{
		{
			Status:           "closed",
			ExecutedQuantity: 0.00046,
			Price:            43496.99,
		},
	})
	orderRepository.On("Update", mock.Anything).Times(1).Return(nil)
	balanceService.On("InvalidateBalanceCache", "USDT").Times(1)
	balanceService.On("InvalidateBalanceCache", "BTC").Times(1)

	telegramNotificatorMock.On("SellOrder", mock.Anything, mock.Anything, mock.Anything).Times(1)

	err := orderExecutor.Sell(tradeLimit, openedOrder, "BTCUSDT", 43496.99, 0.00046, 0, 0, 0)
	assertion.Nil(err)
	assertion.Equal("closed", orderRepository.Updated.Status)
	assertion.Equal(42026.08, orderRepository.Updated.Price)
	assertion.Equal(openedExternalId, *orderRepository.Updated.ExternalId)
}
