package tests

import (
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
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

	profitServiceMock := new(ProfitServiceMock)

	swapRepository.On("GetSwapChainCache", "ETH").Return(nil)

	lockChannel := make(chan model.Lock)

	lossSecurityMock := new(LossSecurityMock)
	tradeLimit := model.TradeLimit{
		Symbol: "ETHUSDT",
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 3.10,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 2.40,
			},
		},
		MinPrice:    0.01,
		MinQuantity: 0.0001,
	}
	lossSecurityMock.On("IsRiskyBuy", mock.Anything, tradeLimit).Return(false)

	botServiceMock := new(BotServiceMock)
	botServiceMock.On("IsSwapEnabled").Return(true)
	botServiceMock.On("UseSwapCapital").Return(true)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
	}

	go func(orderExecutor *exchange.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:     "999",
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
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(2)
	priceCalculator.On("GetDepth", "ETHUSDT", int64(20)).Return(model.OrderBookModel{
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
	exchangeRepository.On("GetTradeLimit", "ETHUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetCurrentKline", "ETHUSDT").Return(&model.KLine{
		Symbol: "ETHUSDT",
		Close:  2281.52,
	})
	openedExternalId := "988"
	openedOrder := model.Order{
		Id:               8889,
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "ETHUSDT",
		Quantity:         0.009,
		Price:            2212.92,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 0.009,
	}
	orderRepository.On("Find", int64(8889)).Return(openedOrder, nil)
	orderRepository.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(&openedOrder)
	orderRepository.On("GetManualOrder", "ETHUSDT").Return(nil)
	timeService.On("WaitMilliseconds", int64(20)).Maybe()
	filledOrder := model.BinanceOrder{
		OrderId:             "999",
		Symbol:              "ETHUSDT",
		Side:                "SELL",
		ExecutedQty:         0.0089,
		OrigQty:             0.0089,
		Status:              "FILLED",
		Price:               2212.92,
		CummulativeQuoteQty: 0.009 * 2212.92,
	}
	binance.On("QueryOrder", "ETHUSDT", "999").Return(filledOrder, nil)
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

	profitServiceMock.On("GetMinClosePrice", openedOrder, openedOrder.Price).Return(openedOrder.Price * (100 + 3.1) / 100)
	priceCalculator.On("CalculateSell", tradeLimit, openedOrder).Return(2281.52, nil)
	orderRepository.On("DeleteBinanceOrder", filledOrder).Times(1)

	err := orderExecutor.Sell(tradeLimit, openedOrder, 2281.52, 0.0089, false)
	assertion.Nil(err)
	assertion.Equal("closed", orderRepository.Updated.Status)
	assertion.Equal(2212.92, orderRepository.Updated.Price)
	assertion.Equal(openedExternalId, *orderRepository.Updated.ExternalId)
}

func TestSellFoundFilled(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
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
	lossSecurityMock := new(LossSecurityMock)
	tradeLimit := model.TradeLimit{
		Symbol: "ETHUSDT",
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 3.10,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 2.40,
			},
		},
		MinPrice:    0.01,
		MinQuantity: 0.0001,
	}
	lossSecurityMock.On("IsRiskyBuy", mock.Anything, tradeLimit).Return(false)

	botServiceMock := new(BotServiceMock)
	botServiceMock.On("IsSwapEnabled").Return(true)
	botServiceMock.On("UseSwapCapital").Return(true)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
	}

	go func(orderExecutor *exchange.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:             "999",
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
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(2)
	priceCalculator.On("GetDepth", "ETHUSDT", int64(20)).Return(model.OrderBookModel{
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
	exchangeRepository.On("GetTradeLimit", "ETHUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetCurrentKline", "ETHUSDT").Return(&model.KLine{
		Symbol: "ETHUSDT",
		Close:  2281.52,
	})
	openedExternalId := "988"
	openedOrder := model.Order{
		Id:               9998,
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "ETHUSDT",
		Quantity:         0.009,
		Price:            2212.92,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 0.009,
	}
	orderRepository.On("Find", int64(9998)).Return(openedOrder, nil)
	orderRepository.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(&openedOrder)
	orderRepository.On("GetManualOrder", "ETHUSDT").Return(nil)
	timeService.On("WaitMilliseconds", int64(20)).Unset()
	binance.On("QueryOrder", "ETHUSDT", "999").Unset()
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
	profitServiceMock.On("GetMinClosePrice", openedOrder, openedOrder.Price).Return(openedOrder.Price * (100 + 3.1) / 100)
	priceCalculator.On("CalculateSell", tradeLimit, openedOrder).Return(2281.52, nil)

	err := orderExecutor.Sell(tradeLimit, openedOrder, 2281.52, 0.0089, false)
	assertion.Nil(err)
	assertion.Equal("closed", orderRepository.Updated.Status)
	assertion.Equal(2212.92, orderRepository.Updated.Price)
	assertion.Equal(openedExternalId, *orderRepository.Updated.ExternalId)
}

func TestSellCancelledInProcess(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
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
	lossSecurityMock := new(LossSecurityMock)
	tradeLimit := model.TradeLimit{
		Symbol: "ETHUSDT",
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 3.10,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 2.40,
			},
		},
		MinPrice:    0.01,
		MinQuantity: 0.0001,
	}
	lossSecurityMock.On("IsRiskyBuy", mock.Anything, tradeLimit).Return(false)

	botServiceMock := new(BotServiceMock)
	botServiceMock.On("IsSwapEnabled").Return(true)
	botServiceMock.On("UseSwapCapital").Return(true)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
	}

	go func(orderExecutor *exchange.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:     "999",
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
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(2)
	priceCalculator.On("GetDepth", "ETHUSDT", int64(20)).Return(model.OrderBookModel{
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
	exchangeRepository.On("GetTradeLimit", "ETHUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetCurrentKline", "ETHUSDT").Return(&model.KLine{
		Symbol: "ETHUSDT",
		Close:  2281.52,
	})
	openedExternalId := "988"
	openedOrder := model.Order{
		Id:               8877,
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "ETHUSDT",
		Quantity:         0.009,
		Price:            2212.92,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 0.009,
	}
	orderRepository.On("Find", int64(8877)).Return(openedOrder, nil)
	orderRepository.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(&openedOrder)
	orderRepository.On("GetManualOrder", "ETHUSDT").Return(nil)
	timeService.On("WaitMilliseconds", int64(20)).Maybe()
	canceled := model.BinanceOrder{
		OrderId:             "999",
		Symbol:              "ETHUSDT",
		Side:                "SELL",
		ExecutedQty:         0.0000,
		OrigQty:             0.0089,
		Status:              "CANCELED",
		Price:               2212.92,
		CummulativeQuoteQty: 0.00,
	}
	binance.On("QueryOrder", "ETHUSDT", "999").Return(canceled, nil)
	orderRepository.On("DeleteBinanceOrder", canceled).Times(1)
	orderId := int64(100)
	orderRepository.On("Create", mock.Anything).Return(&orderId, nil).Unset()
	orderRepository.On("DeleteManualOrder", "ETHUSDT").Unset()
	balanceService.On("InvalidateBalanceCache", "USDT").Times(1)
	balanceService.On("InvalidateBalanceCache", "ETH").Times(1)

	profitServiceMock.On("GetMinClosePrice", openedOrder, openedOrder.Price).Return(openedOrder.Price * (100 + 3.1) / 100)
	priceCalculator.On("CalculateSell", tradeLimit, openedOrder).Return(2281.52, nil)

	err := orderExecutor.Sell(tradeLimit, openedOrder, 2281.52, 0.0089, false)
	assertion.Error(errors.New("Order was CANCELED"), err)
}

func TestSellQueryFail(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
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
	lossSecurityMock := new(LossSecurityMock)
	tradeLimit := model.TradeLimit{
		Symbol: "ETHUSDT",
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 3.10,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 2.40,
			},
		},
		MinPrice:    0.01,
		MinQuantity: 0.0001,
	}
	lossSecurityMock.On("IsRiskyBuy", mock.Anything, tradeLimit).Return(false)

	botServiceMock := new(BotServiceMock)
	botServiceMock.On("IsSwapEnabled").Return(true)
	botServiceMock.On("UseSwapCapital").Return(true)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
	}

	go func(orderExecutor *exchange.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:     "999",
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
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(2)
	priceCalculator.On("GetDepth", "ETHUSDT", int64(20)).Return(model.OrderBookModel{
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
	exchangeRepository.On("GetTradeLimit", "ETHUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetCurrentKline", "ETHUSDT").Return(&model.KLine{
		Symbol: "ETHUSDT",
		Close:  2281.52,
	})
	openedExternalId := "988"
	openedOrder := model.Order{
		Id:               88811,
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "ETHUSDT",
		Quantity:         0.009,
		Price:            2212.92,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 0.009,
	}
	orderRepository.On("Find", int64(88811)).Return(openedOrder, nil)
	orderRepository.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(&openedOrder)
	orderRepository.On("GetManualOrder", "ETHUSDT").Return(nil)
	timeService.On("WaitMilliseconds", int64(20)).Maybe()
	binance.On("QueryOrder", "ETHUSDT", "999").Return(model.BinanceOrder{}, errors.New("Order was canceled or expired"))
	orderRepository.On("DeleteBinanceOrder", initialBinanceOrder).Times(1)
	orderId := int64(100)
	orderRepository.On("Create", mock.Anything).Return(&orderId, nil).Unset()
	orderRepository.On("DeleteManualOrder", "ETHUSDT").Unset()
	balanceService.On("InvalidateBalanceCache", "USDT").Times(1)
	balanceService.On("InvalidateBalanceCache", "ETH").Times(1)

	profitServiceMock.On("GetMinClosePrice", openedOrder, openedOrder.Price).Return(openedOrder.Price * (100 + 3.1) / 100)
	priceCalculator.On("CalculateSell", tradeLimit, openedOrder).Return(2281.52, nil)

	err := orderExecutor.Sell(tradeLimit, openedOrder, 2281.52, 0.0089, false)
	assertion.Equal(errors.New("Order 999 was CANCELED or EXPIRED"), err)
}

func TestSellClosingAction(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
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
	lossSecurityMock := new(LossSecurityMock)
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 3.10,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 2.40,
			},
		},
		MinPrice:    0.01,
		MinQuantity: 0.00001,
	}
	lossSecurityMock.On("IsRiskyBuy", mock.Anything, tradeLimit).Return(false)

	botServiceMock := new(BotServiceMock)
	botServiceMock.On("IsSwapEnabled").Return(true)
	botServiceMock.On("UseSwapCapital").Return(true)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
	}

	go func(orderExecutor *exchange.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:     "999",
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
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(2)
	priceCalculator.On("GetDepth", "BTCUSDT", int64(20)).Return(model.OrderBookModel{
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
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetCurrentKline", "BTCUSDT").Return(&model.KLine{
		Symbol: "BTCUSDT",
		Close:  43496.99,
	})
	openedExternalId := "988"
	openedOrder := model.Order{
		Id:               11122,
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "BTCUSDT",
		Quantity:         0.00047,
		Price:            42026.08,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 0.00047,
	}
	orderRepository.On("Find", int64(11122)).Return(openedOrder, nil)
	orderRepository.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(&openedOrder)
	orderRepository.On("GetManualOrder", "BTCUSDT").Return(nil)
	timeService.On("WaitMilliseconds", int64(20)).Maybe()
	filledOrder := model.BinanceOrder{
		OrderId:             "999",
		Symbol:              "BTCUSDT",
		Side:                "SELL",
		ExecutedQty:         0.00046,
		OrigQty:             0.00046,
		Status:              "FILLED",
		Price:               43496.99,
		CummulativeQuoteQty: 0.00046 * 43496.99,
	}
	binance.On("QueryOrder", "BTCUSDT", "999").Return(filledOrder, nil)
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

	profitServiceMock.On("GetMinClosePrice", openedOrder, openedOrder.Price).Return(openedOrder.Price * (100 + 3.1) / 100)
	priceCalculator.On("CalculateSell", tradeLimit, openedOrder).Return(43496.99, nil)
	orderRepository.On("DeleteBinanceOrder", filledOrder).Times(1)

	err := orderExecutor.Sell(tradeLimit, openedOrder, 43496.99, 0.00046, false)
	assertion.Nil(err)
	assertion.Equal("closed", orderRepository.Updated.Status)
	assertion.Equal(42026.08, orderRepository.Updated.Price)
	assertion.Equal(openedExternalId, *orderRepository.Updated.ExternalId)
}

func TestSellClosingTrxAction(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
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

	swapRepository.On("GetSwapChainCache", "TRX").Return(nil)

	lockChannel := make(chan model.Lock)
	lossSecurityMock := new(LossSecurityMock)
	tradeLimit := model.TradeLimit{
		Symbol: "TRXUSDT",
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 2.25,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 2.40,
			},
		},
		MinPrice:    0.00001,
		MinQuantity: 0.1,
	}
	lossSecurityMock.On("IsRiskyBuy", mock.Anything, tradeLimit).Return(false)

	botServiceMock := new(BotServiceMock)
	botServiceMock.On("IsSwapEnabled").Return(true)
	botServiceMock.On("UseSwapCapital").Return(true)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
	}

	go func(orderExecutor *exchange.OrderExecutor) {
		for {
			lock := <-lockChannel
			orderExecutor.TradeLockMutex.Lock()
			orderExecutor.Lock[lock.Symbol] = lock.IsLocked
			orderExecutor.TradeLockMutex.Unlock()
		}
	}(&orderExecutor)

	initialBinanceOrder := model.BinanceOrder{
		OrderId:     "999",
		Symbol:      "TRXUSDT",
		Side:        "SELL",
		ExecutedQty: 0.00,
		OrigQty:     382.1,   //382.5
		Price:       0.10692, // 0.10457
		Status:      "NEW",
	}
	timeService.On("GetNowDateTimeString").Return("2023-12-28 00:52:00")
	orderRepository.On("GetBinanceOrder", "TRXUSDT", "SELL").Return(nil)
	binance.On("GetOpenedOrders").Return([]model.BinanceOrder{
		initialBinanceOrder,
	}, nil)
	orderRepository.On("SetBinanceOrder", mock.Anything).Times(2)
	priceCalculator.On("GetDepth", "TRXUSDT", int64(20)).Return(model.OrderBookModel{
		Symbol: "TRXUSDT",
		Asks: [][2]model.Number{
			{
				{
					0.10692,
				},
				{
					0.009,
				},
			},
		},
		Bids: [][2]model.Number{
			{
				{
					0.10692,
				},
				{
					0.009,
				},
			},
		},
	})
	exchangeRepository.On("GetTradeLimit", "TRXUSDT").Return(tradeLimit, nil)
	timeService.On("GetNowUnix").Times(1).Return(0)
	for i := 2; i < 1002; i++ {
		timeService.On("GetNowUnix").Times(i).Return(480)
	}
	exchangeRepository.On("GetCurrentKline", "TRXUSDT").Return(&model.KLine{
		Symbol: "TRXUSDT",
		Close:  0.10692,
	})
	openedExternalId := "988"
	openedOrder := model.Order{
		Id:               22235,
		ExternalId:       &openedExternalId,
		Status:           "opened",
		Symbol:           "TRXUSDT",
		Quantity:         382.5,
		Price:            0.10457,
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		ExecutedQuantity: 382.5,
	}
	orderRepository.On("Find", int64(22235)).Return(openedOrder, nil)
	orderRepository.On("GetOpenedOrderCached", "TRXUSDT", "BUY").Return(&openedOrder)
	orderRepository.On("GetManualOrder", "TRXUSDT").Return(nil)
	timeService.On("WaitMilliseconds", int64(20)).Maybe()
	filledOrder := model.BinanceOrder{
		OrderId:             "999",
		Symbol:              "TRXUSDT",
		Side:                "SELL",
		ExecutedQty:         382.1,
		OrigQty:             382.1,   //382.5
		Price:               0.10692, // 0.10457
		Status:              "FILLED",
		CummulativeQuoteQty: 382.1 * 0.10692,
	}
	binance.On("QueryOrder", "TRXUSDT", "999").Return(filledOrder, nil)
	orderId := int64(100)
	orderRepository.On("Create", mock.Anything).Return(&orderId, nil)
	orderRepository.On("DeleteManualOrder", "TRXUSDT").Times(1)
	orderRepository.On("Find", orderId).Times(1).Return(model.Order{}, nil)
	orderRepository.On("GetClosesOrderList", openedOrder).Times(1).Return([]model.Order{
		{
			Status:           "closed",
			ExecutedQuantity: 382.1,
			Price:            0.10692,
		},
	})
	orderRepository.On("Update", mock.Anything).Times(1).Return(nil)
	balanceService.On("InvalidateBalanceCache", "USDT").Times(1)
	balanceService.On("InvalidateBalanceCache", "TRX").Times(1)

	telegramNotificatorMock.On("SellOrder", mock.Anything, mock.Anything, mock.Anything).Times(1)

	profitServiceMock.On("GetMinClosePrice", openedOrder, openedOrder.Price).Return(openedOrder.Price * (100 + 2.25) / 100)
	priceCalculator.On("CalculateSell", tradeLimit, openedOrder).Return(0.10692, nil)
	orderRepository.On("DeleteBinanceOrder", filledOrder).Times(1)

	err := orderExecutor.Sell(tradeLimit, openedOrder, 0.10692, 382.1, false)
	assertion.Nil(err)
	assertion.Equal("closed", orderRepository.Updated.Status)
	assertion.Equal(0.10457, orderRepository.Updated.Price)
	assertion.Equal(openedExternalId, *orderRepository.Updated.ExternalId)
}

func TestCreateSwapAction(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
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
	lossSecurityMock := new(LossSecurityMock)
	botServiceMock := new(BotServiceMock)
	lockChannel := make(chan model.Lock)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
	}

	order := model.Order{
		Id:               777,
		Symbol:           "BTCUSDT",
		ExecutedQuantity: 1002.00,
	}

	swapChain := model.SwapChainEntity{
		Id: 888,
		SwapOne: &model.SwapTransitionEntity{
			BaseAsset:  "BTC",
			QuoteAsset: "ETH",
			Price:      1000.00, // fake
		},
		SwapTwo: &model.SwapTransitionEntity{
			BaseAsset:  "ETH",
			QuoteAsset: "SOL",
			Price:      100.00, // fake
		},
		SwapThree: &model.SwapTransitionEntity{
			BaseAsset:  "SOL",
			QuoteAsset: "BTC",
			Price:      4000.00, // fake
		},
	}

	balanceService.On("GetAssetBalance", "BTC", false).Return(2000.00, nil)
	swapRepository.On("GetActiveSwapAction", order).Return(model.SwapAction{}, errors.New("test!"))
	timeService.On("GetNowUnix").Return(9999)

	swapId := int64(1)
	swapRepository.On("CreateSwapAction", mock.Anything).Times(1).Return(&swapId, nil)
	orderRepository.On("Update", mock.Anything).Return(nil)

	orderExecutor.MakeSwap(order, swapChain)

	assertion.Equal(1002.00, swapRepository.swapAction.StartQuantity)
	assertion.Equal(1000.00, swapRepository.swapAction.SwapOnePrice)
	assertion.Equal(100.00, swapRepository.swapAction.SwapTwoPrice)
	assertion.Equal(4000.00, swapRepository.swapAction.SwapThreePrice)
	assertion.Equal("BTCETH", swapRepository.swapAction.SwapOneSymbol)
	assertion.Equal("ETHSOL", swapRepository.swapAction.SwapTwoSymbol)
	assertion.Equal("SOLBTC", swapRepository.swapAction.SwapThreeSymbol)
	assertion.Equal("BTC", swapRepository.swapAction.Asset)
	assertion.Equal(model.SwapActionStatusPending, swapRepository.swapAction.Status)
	assertion.Equal(order.Id, swapRepository.swapAction.OrderId)
	assertion.Equal(order.Id, orderRepository.Updated.Id)
	assertion.True(orderRepository.Updated.Swap)
}

func TestCreateSwapActionLessBalance(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
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
	lossSecurityMock := new(LossSecurityMock)
	botServiceMock := new(BotServiceMock)
	lockChannel := make(chan model.Lock)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
	}

	order := model.Order{
		Id:               777,
		Symbol:           "BTCUSDT",
		ExecutedQuantity: 1002.00,
	}

	swapChain := model.SwapChainEntity{
		Id: 888,
		SwapOne: &model.SwapTransitionEntity{
			BaseAsset:  "BTC",
			QuoteAsset: "ETH",
			Price:      1000.00, // fake
		},
		SwapTwo: &model.SwapTransitionEntity{
			BaseAsset:  "ETH",
			QuoteAsset: "SOL",
			Price:      100.00, // fake
		},
		SwapThree: &model.SwapTransitionEntity{
			BaseAsset:  "SOL",
			QuoteAsset: "BTC",
			Price:      4000.00, // fake
		},
	}

	balanceService.On("GetAssetBalance", "BTC", false).Return(900.00, nil)
	swapRepository.On("GetActiveSwapAction", order).Return(model.SwapAction{}, errors.New("test!"))
	timeService.On("GetNowUnix").Return(9999)

	swapId := int64(1)
	swapRepository.On("CreateSwapAction", mock.Anything).Times(1).Return(&swapId, nil)
	orderRepository.On("Update", mock.Anything).Return(nil)

	orderExecutor.MakeSwap(order, swapChain)

	assertion.Equal(900.00, swapRepository.swapAction.StartQuantity)
	assertion.Equal(1000.00, swapRepository.swapAction.SwapOnePrice)
	assertion.Equal(100.00, swapRepository.swapAction.SwapTwoPrice)
	assertion.Equal(4000.00, swapRepository.swapAction.SwapThreePrice)
	assertion.Equal("BTCETH", swapRepository.swapAction.SwapOneSymbol)
	assertion.Equal("ETHSOL", swapRepository.swapAction.SwapTwoSymbol)
	assertion.Equal("SOLBTC", swapRepository.swapAction.SwapThreeSymbol)
	assertion.Equal("BTC", swapRepository.swapAction.Asset)
	assertion.Equal(model.SwapActionStatusPending, swapRepository.swapAction.Status)
	assertion.Equal(order.Id, swapRepository.swapAction.OrderId)
	assertion.Equal(order.Id, orderRepository.Updated.Id)
	assertion.True(orderRepository.Updated.Swap)
}

func TestCheckIsTimeToCancel(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
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
	lossSecurityMock := new(LossSecurityMock)
	botServiceMock := new(BotServiceMock)
	lockChannel := make(chan model.Lock)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
		CancelRequestMap:   make(map[string]bool),
	}

	limit := model.TradeLimit{
		MinPrice: 0.01,
	}
	binanceOrder := model.BinanceOrder{
		Status: "NEW",
		Side:   "SELL",
		Price:  100.00,
		Symbol: "SOLUSDT",
	}
	orderManageChannel := make(chan string)
	control := make(chan string)

	openedPosition := model.Order{}

	go func() {
		request := <-orderManageChannel
		assertion.Equal("status", request)
		control <- "stop"
		request = <-orderManageChannel
		assertion.Equal("status", request)
		control <- "continue"
		request = <-orderManageChannel
		assertion.Equal("cancel", request)
		control <- "continue"
	}()

	exchangeRepository.On("GetCurrentKline", "SOLUSDT").Return(&model.KLine{
		Symbol: "SOLUSDT",
		Close:  95.00,
	})

	orderRepository.On("GetManualOrder", "SOLUSDT").Return(nil)
	orderRepository.On("GetOpenedOrderCached", "SOLUSDT", "BUY").Return(&openedPosition)
	priceCalculator.On("CalculateSell", limit, openedPosition).Return(99.00, nil)

	assertion.True(orderExecutor.CheckIsTimeToCancel(limit, &binanceOrder, orderManageChannel, control))
	assertion.False(orderExecutor.CheckIsTimeToCancel(limit, &binanceOrder, orderManageChannel, control))
}

func TestCheckIsTimeToCancelSamePrice(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
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
	lossSecurityMock := new(LossSecurityMock)
	botServiceMock := new(BotServiceMock)
	lockChannel := make(chan model.Lock)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
		CancelRequestMap:   make(map[string]bool),
	}

	limit := model.TradeLimit{
		MinPrice: 0.01,
	}
	binanceOrder := model.BinanceOrder{
		Status: "NEW",
		Side:   "SELL",
		Price:  100.00,
		Symbol: "SOLUSDT",
	}
	orderManageChannel := make(chan string)
	control := make(chan string)

	openedPosition := model.Order{}

	orderRepository.On("GetManualOrder", "SOLUSDT").Return(nil)
	exchangeRepository.On("GetCurrentKline", "SOLUSDT").Return(&model.KLine{
		Symbol: "SOLUSDT",
		Close:  95.00,
	})

	orderRepository.On("GetOpenedOrderCached", "SOLUSDT", "BUY").Return(&openedPosition)
	priceCalculator.On("CalculateSell", limit, openedPosition).Return(100.00, nil)

	assertion.False(orderExecutor.CheckIsTimeToCancel(limit, &binanceOrder, orderManageChannel, control))
}

func TestCheckIsTimeToCancelPriceIsMoreThanOrder(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
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
	lossSecurityMock := new(LossSecurityMock)
	botServiceMock := new(BotServiceMock)
	lockChannel := make(chan model.Lock)

	orderExecutor := exchange.OrderExecutor{
		TradeStack:   &exchange.TradeStack{},
		LossSecurity: lossSecurityMock,
		CurrentBot: &model.Bot{
			Id:      999,
			BotUuid: uuid.New().String(),
		},
		TimeService:        timeService,
		BalanceService:     balanceService,
		Binance:            binance,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		SwapRepository:     swapRepository,
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
		SwapExecutor:       swapExecutor,
		SwapValidator:      swapValidator,
		Formatter:          &utils.Formatter{},
		BotService:         botServiceMock,
		LockChannel:        &lockChannel,
		Lock:               make(map[string]bool),
		TradeLockMutex:     sync.RWMutex{},
		CallbackManager:    telegramNotificatorMock,
		CancelRequestMap:   make(map[string]bool),
	}

	limit := model.TradeLimit{
		MinPrice: 0.01,
	}
	binanceOrder := model.BinanceOrder{
		Status: "NEW",
		Side:   "SELL",
		Price:  100.00,
		Symbol: "SOLUSDT",
	}
	orderManageChannel := make(chan string)
	control := make(chan string)

	openedPosition := model.Order{}

	orderRepository.On("GetManualOrder", "SOLUSDT").Return(nil)
	exchangeRepository.On("GetCurrentKline", "SOLUSDT").Return(&model.KLine{
		Symbol: "SOLUSDT",
		Close:  101.00,
	})

	orderRepository.On("GetOpenedOrderCached", "SOLUSDT", "BUY").Return(&openedPosition)

	assertion.False(orderExecutor.CheckIsTimeToCancel(limit, &binanceOrder, orderManageChannel, control))
}

func TestAvgPriceCalculation(t *testing.T) {
	assertion := assert.New(t)

	opened := model.Order{
		ExecutedQuantity: 1.00,
		Price:            100.00,
	}
	extra := model.Order{
		ExecutedQuantity: 1.00,
		Price:            80.00,
	}

	orderExecutor := exchange.OrderExecutor{}
	avgPrice := orderExecutor.GetAvgPrice(opened, extra)
	assertion.Equal(90.00, avgPrice)
	opened.Price = avgPrice
	opened.ExecutedQuantity = 2.00
	avgPrice = orderExecutor.GetAvgPrice(opened, extra)
	assertion.Equal(86.66666666666667, avgPrice)
}
