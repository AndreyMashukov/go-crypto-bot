package tests

import (
	"errors"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"testing"
	"time"
)

func TestDecisionError(t *testing.T) {
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(BaseTradeStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := new(StrategyFacadeMock)
	priceCalculator := new(PriceCalculatorMock)
	tradeStack := new(BuyOrderStackMock)
	orderExecutor := new(OrderExecutorMock)
	binance := new(ExchangePriceAPIMock)

	maker := exchange.MakerService{
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		BotService:         botService,
		StrategyFacade:     strategyFacade,
		PriceCalculator:    priceCalculator,
		TradeStack:         tradeStack,
		OrderExecutor:      orderExecutor,
		Binance:            binance,
		Formatter:          &utils.Formatter{},
		CurrentBot: &model.Bot{
			Id: 1,
		},
		HoldScore: 80.00,
	}

	strategyFacade.On("Decide", "BTCUSDT").Return(model.FacadeResponse{
		Hold: 0.00,
		Sell: 0.00,
		Buy:  0.00,
	}, errors.New("Test Facade!!!"))
	maker.Make("BTCUSDT")
	orderRepository.AssertNumberOfCalls(t, "GetOpenedOrderCached", 0)
}

func TestProcessSwap(t *testing.T) {
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(BaseTradeStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := new(StrategyFacadeMock)
	priceCalculator := new(PriceCalculatorMock)
	tradeStack := new(BuyOrderStackMock)
	orderExecutor := new(OrderExecutorMock)
	binance := new(ExchangePriceAPIMock)

	maker := exchange.MakerService{
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		BotService:         botService,
		StrategyFacade:     strategyFacade,
		PriceCalculator:    priceCalculator,
		TradeStack:         tradeStack,
		OrderExecutor:      orderExecutor,
		Binance:            binance,
		Formatter:          &utils.Formatter{},
		CurrentBot: &model.Bot{
			Id: 1,
		},
		HoldScore: 80.00,
	}

	strategyFacade.On("Decide", "BTCUSDT").Return(model.FacadeResponse{
		Hold: 0.00,
		Sell: 0.00,
		Buy:  0.00,
	}, nil)
	order := model.Order{
		Symbol: "BTCUSDT",
	}
	orderRepository.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(order, nil)
	orderExecutor.On("ProcessSwap", order).Return(true)

	maker.Make("BTCUSDT")
	exchangeRepository.AssertNumberOfCalls(t, "GetTradeLimit", 0)
	orderRepository.AssertNumberOfCalls(t, "GetOpenedOrderCached", 1)
	orderExecutor.AssertNumberOfCalls(t, "ProcessSwap", 1)
}

func TestHoldDecision(t *testing.T) {
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(BaseTradeStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := new(StrategyFacadeMock)
	priceCalculator := new(PriceCalculatorMock)
	tradeStack := new(BuyOrderStackMock)
	orderExecutor := new(OrderExecutorMock)
	binance := new(ExchangePriceAPIMock)

	maker := exchange.MakerService{
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		BotService:         botService,
		StrategyFacade:     strategyFacade,
		PriceCalculator:    priceCalculator,
		TradeStack:         tradeStack,
		OrderExecutor:      orderExecutor,
		Binance:            binance,
		Formatter:          &utils.Formatter{},
		CurrentBot: &model.Bot{
			Id: 1,
		},
		HoldScore: 80.00,
	}

	strategyFacade.On("Decide", "BTCUSDT").Return(model.FacadeResponse{
		Hold: 80.00,
		Sell: 0.00,
		Buy:  0.00,
	}, nil)
	order := model.Order{
		Symbol: "BTCUSDT",
	}
	orderRepository.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(order, nil)
	orderExecutor.On("ProcessSwap", order).Return(false)

	maker.Make("BTCUSDT")
	exchangeRepository.AssertNumberOfCalls(t, "GetTradeLimit", 0)
	orderRepository.AssertNumberOfCalls(t, "GetOpenedOrderCached", 1)
	orderExecutor.AssertNumberOfCalls(t, "ProcessSwap", 1)
}

func TestSellOperation(t *testing.T) {
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(BaseTradeStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := new(StrategyFacadeMock)
	priceCalculator := new(PriceCalculatorMock)
	tradeStack := new(BuyOrderStackMock)
	orderExecutor := new(OrderExecutorMock)
	binance := new(ExchangePriceAPIMock)

	tradeFilterService := new(TradeFilterServiceMock)

	maker := exchange.MakerService{
		TradeFilterService: tradeFilterService,
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		BotService:         botService,
		StrategyFacade:     strategyFacade,
		PriceCalculator:    priceCalculator,
		TradeStack:         tradeStack,
		OrderExecutor:      orderExecutor,
		Binance:            binance,
		Formatter:          &utils.Formatter{},
		CurrentBot: &model.Bot{
			Id: 1,
		},
		HoldScore: 80.00,
	}

	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}
	tradeFilterService.On("CanSell", tradeLimit).Return(true)
	strategyFacade.On("Decide", "BTCUSDT").Return(model.FacadeResponse{
		Hold: 40.00,
		Sell: 50.00,
		Buy:  40.00,
	}, nil)
	order := model.Order{
		Symbol:   "BTCUSDT",
		Price:    50000.00,
		Quantity: 1.00,
	}
	orderRepository.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(order, nil)
	orderExecutor.On("ProcessSwap", order).Return(false)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  65000.00,
	}
	exchangeRepository.On("GetCurrentKline", "BTCUSDT").Return(&kline)
	orderRepository.On("GetManualOrder", "BTCUSDT").Return(nil)
	priceCalculator.On("GetDepth", "BTCUSDT", int64(20)).Return(model.OrderBookModel{
		Asks: [][2]model.Number{
			{
				{
					Value: 65000.00,
				},
				{
					Value: 1.00,
				},
			},
			{
				{
					Value: 65000.00,
				},
				{
					Value: 1.00,
				},
			},
			{
				{
					Value: 65000.00,
				},
				{
					Value: 1.00,
				},
			},
			{
				{
					Value: 65000.00,
				},
				{
					Value: 1.00,
				},
			},
		},
	})
	priceCalculator.On("CalculateSell", tradeLimit, order).Return(65005.00, nil)
	orderExecutor.On("CalculateSellQuantity", order).Return(1.00)
	orderExecutor.On("Sell", tradeLimit, order, 65005.00, 1.00, false).Return(nil)

	maker.Make("BTCUSDT")
	orderExecutor.AssertNumberOfCalls(t, "Sell", 1)
	orderExecutor.AssertNumberOfCalls(t, "Buy", 0)
	orderExecutor.AssertNumberOfCalls(t, "BuyExtra", 0)
	exchangeRepository.AssertNumberOfCalls(t, "GetTradeLimit", 1)
	orderRepository.AssertNumberOfCalls(t, "GetOpenedOrderCached", 1)
	orderExecutor.AssertNumberOfCalls(t, "ProcessSwap", 1)
}

func TestBuyOperation(t *testing.T) {
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(BaseTradeStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := new(StrategyFacadeMock)
	priceCalculator := new(PriceCalculatorMock)
	tradeStack := new(BuyOrderStackMock)
	orderExecutor := new(OrderExecutorMock)
	binance := new(ExchangePriceAPIMock)

	maker := exchange.MakerService{
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		BotService:         botService,
		StrategyFacade:     strategyFacade,
		PriceCalculator:    priceCalculator,
		TradeStack:         tradeStack,
		OrderExecutor:      orderExecutor,
		Binance:            binance,
		Formatter:          &utils.Formatter{},
		CurrentBot: &model.Bot{
			Id: 1,
		},
		HoldScore: 80.00,
	}

	strategyFacade.On("Decide", "BTCUSDT").Return(model.FacadeResponse{
		Hold: 40.00,
		Sell: 40.00,
		Buy:  50.00,
	}, nil)
	orderRepository.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(model.Order{}, errors.New("No order!"))
	tradeLimit := model.TradeLimit{
		Symbol:      "BTCUSDT",
		IsEnabled:   true,
		USDTLimit:   100.00,
		MinNotional: 5.00,
		MinQuantity: 0.001,
		MinPrice:    10.00,
	}
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	tradeStack.On("CanBuy", tradeLimit).Return(true)
	kline := model.KLine{
		Symbol:    "BTCUSDT",
		Close:     50000.00,
		UpdatedAt: time.Now().Unix(),
	}
	exchangeRepository.On("GetCurrentKline", "BTCUSDT").Return(&kline)
	orderExecutor.On("CheckMinBalance", tradeLimit, kline).Return(nil)
	orderRepository.On("GetManualOrder", "BTCUSDT").Return(nil)
	priceCalculator.On("GetDepth", "BTCUSDT", int64(20)).Return(model.OrderBookModel{
		Bids: [][2]model.Number{
			{
				{
					Value: 50000.00,
				},
				{
					Value: 1.00,
				},
			},
			{
				{
					Value: 50000.00,
				},
				{
					Value: 1.00,
				},
			},
			{
				{
					Value: 50000.00,
				},
				{
					Value: 1.00,
				},
			},
			{
				{
					Value: 50000.00,
				},
				{
					Value: 1.00,
				},
			},
		},
	})
	priceCalculator.On("CalculateBuy", tradeLimit).Return(45000.00, nil)
	orderExecutor.On("Buy", tradeLimit, 45000.00, 0.002).Return(nil)

	maker.Make("BTCUSDT")
	orderExecutor.AssertNumberOfCalls(t, "Sell", 0)
	orderExecutor.AssertNumberOfCalls(t, "Buy", 1)
	orderExecutor.AssertNumberOfCalls(t, "BuyExtra", 0)
	exchangeRepository.AssertNumberOfCalls(t, "GetTradeLimit", 1)
	orderRepository.AssertNumberOfCalls(t, "GetOpenedOrderCached", 1)
	orderExecutor.AssertNumberOfCalls(t, "ProcessSwap", 0)
}

func TestExtraBuyOperation(t *testing.T) {
	orderRepository := new(OrderStorageMock)
	exchangeRepository := new(BaseTradeStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := new(StrategyFacadeMock)
	priceCalculator := new(PriceCalculatorMock)
	tradeStack := new(BuyOrderStackMock)
	orderExecutor := new(OrderExecutorMock)
	binance := new(ExchangePriceAPIMock)

	maker := exchange.MakerService{
		OrderRepository:    orderRepository,
		ExchangeRepository: exchangeRepository,
		BotService:         botService,
		StrategyFacade:     strategyFacade,
		PriceCalculator:    priceCalculator,
		TradeStack:         tradeStack,
		OrderExecutor:      orderExecutor,
		Binance:            binance,
		Formatter:          &utils.Formatter{},
		CurrentBot: &model.Bot{
			Id: 1,
		},
		HoldScore: 80.00,
	}

	strategyFacade.On("Decide", "BTCUSDT").Return(model.FacadeResponse{
		Hold: 40.00,
		Sell: 40.00,
		Buy:  50.00,
	}, nil)
	order := model.Order{
		Symbol:           "BTCUSDT",
		Price:            50000.00,
		ExecutedQuantity: 1.00,
		UsedExtraBudget:  0.00,
		ExtraChargeOptions: []model.ExtraChargeOption{
			{
				Index:      0,
				Percent:    -2.00,
				AmountUsdt: 100.00,
			},
		},
	}
	orderRepository.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(order, nil)
	orderExecutor.On("ProcessSwap", order).Return(false)
	tradeLimit := model.TradeLimit{
		Symbol:      "BTCUSDT",
		IsEnabled:   true,
		USDTLimit:   100.00,
		MinNotional: 5.00,
		MinQuantity: 0.001,
		MinPrice:    10.00,
	}
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	tradeStack.On("CanBuy", tradeLimit).Return(true)
	kline := model.KLine{
		Symbol:    "BTCUSDT",
		Close:     40000.00,
		UpdatedAt: time.Now().Unix(),
	}
	exchangeRepository.On("GetCurrentKline", "BTCUSDT").Return(&kline)
	orderExecutor.On("CheckMinBalance", tradeLimit, kline).Return(nil)
	orderRepository.On("GetManualOrder", "BTCUSDT").Return(nil)
	priceCalculator.On("GetDepth", "BTCUSDT", int64(20)).Return(model.OrderBookModel{
		Bids: [][2]model.Number{
			{
				{
					Value: 50000.00,
				},
				{
					Value: 1.00,
				},
			},
			{
				{
					Value: 50000.00,
				},
				{
					Value: 1.00,
				},
			},
			{
				{
					Value: 50000.00,
				},
				{
					Value: 1.00,
				},
			},
			{
				{
					Value: 50000.00,
				},
				{
					Value: 1.00,
				},
			},
		},
	})
	priceCalculator.On("CalculateBuy", tradeLimit).Return(40000.00, nil)
	botService.On("UseSwapCapital").Return(false)
	orderExecutor.On("BuyExtra", tradeLimit, order, 40000.00).Return(nil)

	maker.Make("BTCUSDT")
	orderExecutor.AssertNumberOfCalls(t, "Sell", 0)
	orderExecutor.AssertNumberOfCalls(t, "Buy", 0)
	orderExecutor.AssertNumberOfCalls(t, "BuyExtra", 1)
	exchangeRepository.AssertNumberOfCalls(t, "GetTradeLimit", 1)
	orderRepository.AssertNumberOfCalls(t, "GetOpenedOrderCached", 1)
	orderExecutor.AssertNumberOfCalls(t, "ProcessSwap", 1)
}
