package tests

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service/strategy"
	"testing"
	"time"
)

func TestNoTradeLimit(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  40000.00,
	}

	signalStorage.On("GetSignal", "BTCUSDT").Times(0)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(model.TradeLimit{}, errors.New("Test!!!"))

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(0.00, decision.Score)
	assertion.Equal("HOLD", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(40000.00, decision.Price)
}

func TestHasBinanceBuyOrder(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  40000.00,
	}

	signalStorage.On("GetSignal", "BTCUSDT").Times(0)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(model.TradeLimit{
		Symbol: "BTCUSDT",
	}, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(&model.BinanceOrder{
		Price: 40001.00,
	})

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(999.99, decision.Score)
	assertion.Equal("BUY", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(40001.00, decision.Price)
}

func TestNoOrderAndCanBuyHasManualBuy(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  40000.00,
	}
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}

	signalStorage.On("GetSignal", "BTCUSDT").Times(0)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetManualOrder", "BTCUSDT").Return(&model.ManualOrder{
		Operation: "BUY",
		Price:     40002.00,
	})

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(999.99, decision.Score)
	assertion.Equal("BUY", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(40002.00, decision.Price)
}

func TestNoOrderAndCanBuyNoManual(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  40000.00,
	}
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}

	signalStorage.On("GetSignal", "BTCUSDT").Return(nil).Times(1)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(15.00, decision.Score)
	assertion.Equal("BUY", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(40000.00, decision.Price)
}

func TestNoOrderAndCanBuyNoManualHasSignal(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  40000.00,
	}
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}
	signal := model.Signal{
		Symbol:          "BTCUSDT",
		BuyPrice:        39000.00,
		ExpireTimestamp: time.Now().Add(time.Minute).UnixMilli(),
	}

	signalStorage.On("GetSignal", "BTCUSDT").Return(&signal).Times(1)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(999.99, decision.Score)
	assertion.Equal("BUY", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(39000.00, decision.Price)
}

func TestHasOrderAndHasBinanceSellOrder(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  40000.00,
	}
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}

	signalStorage.On("GetSignal", "BTCUSDT").Times(0)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "SELL").Return(&model.BinanceOrder{
		Side:  "SELL",
		Price: 40005.00,
	})
	orderStorage.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(&model.Order{
		Price: 38000.00,
	})

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(999.99, decision.Score)
	assertion.Equal("SELL", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(40005.00, decision.Price)
}

func TestHasOrderAndHasManualSell(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  40000.00,
	}
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}

	signalStorage.On("GetSignal", "BTCUSDT").Times(0)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "SELL").Return(nil)
	botService.On("UseSwapCapital").Return(false)
	orderStorage.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(&model.Order{
		Price:            38000.00,
		ExecutedQuantity: 1.00,
	})
	orderStorage.On("GetManualOrder", "BTCUSDT").Return(&model.ManualOrder{
		Price:     40003.00,
		Operation: "SELL",
	})

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(999.99, decision.Score)
	assertion.Equal("SELL", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(40003.00, decision.Price)
}

func TestHasOrderAndTimeToExtraBuy(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  98.00,
	}
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}

	signalStorage.On("GetSignal", "BTCUSDT").Times(0)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "SELL").Return(nil)
	orderStorage.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(&model.Order{
		Price:            100.00,
		ExecutedQuantity: 1,
		ExtraChargeOptions: []model.ExtraChargeOption{
			{
				Index:      0,
				Percent:    -2.00,
				AmountUsdt: 50,
			},
		},
	})
	botService.On("UseSwapCapital").Return(false)

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(999.99, decision.Score)
	assertion.Equal("BUY", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(98.00, decision.Price)
	orderStorage.AssertNumberOfCalls(t, "GetManualOrder", 0)
}

func TestHasOrderAndProfitPercentReached(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  102.00,
	}
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}

	signalStorage.On("GetSignal", "BTCUSDT").Times(0)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "SELL").Return(nil)
	order := model.Order{
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		Price:            100.00,
		ExecutedQuantity: 1,
		ExtraChargeOptions: []model.ExtraChargeOption{
			{
				Index:      0,
				Percent:    -2.00,
				AmountUsdt: 50,
			},
		},
		ProfitOptions: []model.ProfitOption{
			{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 2.00,
			},
		},
	}
	orderStorage.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(&order)
	profitService.On("GetMinProfitPercent", &order).Return(model.Percent(2.00))
	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)
	botService.On("UseSwapCapital").Return(false)

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(999.99, decision.Score)
	assertion.Equal("SELL", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(102.00, decision.Price)
}

func TestHasOrderAndHalfOfProfitPercentReached(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  101.00,
	}
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}

	signalStorage.On("GetSignal", "BTCUSDT").Times(0)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "SELL").Return(nil)
	order := model.Order{
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		Price:            100.00,
		ExecutedQuantity: 1,
		ExtraChargeOptions: []model.ExtraChargeOption{
			{
				Index:      0,
				Percent:    -2.00,
				AmountUsdt: 50,
			},
		},
		ProfitOptions: []model.ProfitOption{
			{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 2.00,
			},
		},
	}
	orderStorage.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(&order)
	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)
	botService.On("UseSwapCapital").Return(false)
	profitService.On("GetMinProfitPercent", &order).Return(model.Percent(2.00))

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(50.00, decision.Score)
	assertion.Equal("SELL", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(101.00, decision.Price)
}

func TestHasOrderAndCurrentPriceIsGreaterThenOrderPrice(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  100.01,
	}
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}

	signalStorage.On("GetSignal", "BTCUSDT").Times(0)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "SELL").Return(nil)
	order := model.Order{
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		Price:            100.00,
		ExecutedQuantity: 1,
		ExtraChargeOptions: []model.ExtraChargeOption{
			{
				Index:      0,
				Percent:    -2.00,
				AmountUsdt: 50,
			},
		},
		ProfitOptions: []model.ProfitOption{
			{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 2.00,
			},
		},
	}
	orderStorage.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(&order)
	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)
	botService.On("UseSwapCapital").Return(false)
	profitService.On("GetMinProfitPercent", &order).Return(model.Percent(2.00))
	profitService.On("GetMinClosePrice", &order, 100.00).Return(100.02)

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(30.00, decision.Score)
	assertion.Equal("SELL", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(100.02, decision.Price)
}

func TestHasOrderAndCurrentPriceIsLessOrEqualOrderPrice(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	orderStorage := new(OrderStorageMock)
	profitService := new(ProfitServiceMock)
	tradeStack := new(BuyOrderStackMock)
	botService := new(BotServiceMock)
	signalStorage := new(SignalStorageMock)

	orderBasedStrategy := strategy.OrderBasedStrategy{
		ExchangeRepository: exchangeRepository,
		OrderRepository:    orderStorage,
		ProfitService:      profitService,
		BotService:         botService,
		SignalStorage:      signalStorage,
	}

	kline := model.KLine{
		Symbol: "BTCUSDT",
		Close:  99.99,
	}
	tradeLimit := model.TradeLimit{
		Symbol: "BTCUSDT",
	}

	signalStorage.On("GetSignal", "BTCUSDT").Times(0)
	tradeStack.On("CanBuy", tradeLimit).Return(true)
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(tradeLimit, nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "BUY").Return(nil)
	orderStorage.On("GetBinanceOrder", "BTCUSDT", "SELL").Return(nil)
	order := model.Order{
		CreatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		Price:            100.00,
		ExecutedQuantity: 1,
		ExtraChargeOptions: []model.ExtraChargeOption{
			{
				Index:      0,
				Percent:    -2.00,
				AmountUsdt: 50,
			},
		},
		ProfitOptions: []model.ProfitOption{
			{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 2.00,
			},
		},
	}
	orderStorage.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(&order)
	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)
	botService.On("UseSwapCapital").Return(false)
	profitService.On("GetMinProfitPercent", &order).Return(model.Percent(2.00))

	decision := orderBasedStrategy.Decide(kline)
	assertion.Equal(99.99, decision.Score)
	assertion.Equal("HOLD", decision.Operation)
	assertion.Equal(model.OrderBasedStrategyName, decision.StrategyName)
	assertion.Equal(99.99, decision.Price)
}
