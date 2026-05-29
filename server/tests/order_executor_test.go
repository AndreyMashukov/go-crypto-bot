package tests

import (
	"errors"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/service/exchange"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	priceCalculator := new(PriceCalculatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

	profitServiceMock := new(ProfitServiceMock)

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
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
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
				{Value: 2212.92},
				{Value: 0.009},
			},
		},
		Bids: [][2]model.Number{
			{
				{Value: 2212.92},
				{Value: 0.009},
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
	priceCalculator := new(PriceCalculatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

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
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
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
				{Value: 2212.92},
				{Value: 0.009},
			},
		},
		Bids: [][2]model.Number{
			{
				{Value: 2212.92},
				{Value: 0.009},
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
	priceCalculator := new(PriceCalculatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

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
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
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
				{Value: 2212.92},
				{Value: 0.009},
			},
		},
		Bids: [][2]model.Number{
			{
				{Value: 2212.92},
				{Value: 0.009},
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
	priceCalculator := new(PriceCalculatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

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
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
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
				{Value: 2212.92},
				{Value: 0.009},
			},
		},
		Bids: [][2]model.Number{
			{
				{Value: 2212.92},
				{Value: 0.009},
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
	priceCalculator := new(PriceCalculatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

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
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
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
				{Value: 43496.99},
				{Value: 0.009},
			},
		},
		Bids: [][2]model.Number{
			{
				{Value: 43496.99},
				{Value: 0.009},
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
	priceCalculator := new(PriceCalculatorMock)
	timeService := new(TimeServiceMock)
	telegramNotificatorMock := new(TelegramNotificatorMock)

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
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
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
		OrigQty:     382.1,
		Price:       0.10692,
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
				{Value: 0.10692},
				{Value: 0.009},
			},
		},
		Bids: [][2]model.Number{
			{
				{Value: 0.10692},
				{Value: 0.009},
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
		OrigQty:             382.1,
		Price:               0.10692,
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

func TestCheckIsTimeToCancel(t *testing.T) {
	assertion := assert.New(t)

	profitServiceMock := new(ProfitServiceMock)
	balanceService := new(BalanceServiceMock)
	binance := new(ExchangeOrderAPIMock)
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(ExchangeTradeInfoMock)
	priceCalculator := new(PriceCalculatorMock)
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
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
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
	priceCalculator := new(PriceCalculatorMock)
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
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
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
	priceCalculator := new(PriceCalculatorMock)
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
		PriceCalculator:    priceCalculator,
		ProfitService:      profitServiceMock,
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
