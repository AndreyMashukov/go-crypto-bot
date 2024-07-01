package tests

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"sync"
	"testing"
	"time"
)

func TestTradeFilterNoFiltersForBuy(t *testing.T) {
	assertion := assert.New(t)

	tradeLimit := model.TradeLimit{
		TradeFiltersBuy: make(model.TradeFilters, 0),
	}
	filterService := exchange.TradeFilterService{}
	assertion.True(filterService.CanBuy(tradeLimit))
}

func TestTradeFilterMatchingOrNoChildren(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepoMock := new(ExchangeTradeInfoMock)

	filterService := exchange.TradeFilterService{
		ExchangeTradeInfo: exchangeRepoMock,
	}
	exchangeRepoMock.On("GetCurrentKline", "BTCUSDT").Return(&model.KLine{
		Close: 60000.00,
	})
	tradeLimit := model.TradeLimit{
		TradeFiltersBuy: model.TradeFilters{
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterPrice,
				Condition: model.TradeFilterConditionLt,
				Value:     "70000.00",
				Type:      model.TradeFilterConditionTypeOr,
				Children:  make(model.TradeFilters, 0),
			},
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterPrice,
				Condition: model.TradeFilterConditionGt,
				Value:     "50000.00",
				Type:      model.TradeFilterConditionTypeOr,
				Children:  make(model.TradeFilters, 0),
			},
		},
	}
	assertion.True(filterService.CanBuy(tradeLimit))
}

func TestTradeFilterMatchingAndNoChildren(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepoMock := new(ExchangeTradeInfoMock)

	filterService := exchange.TradeFilterService{
		ExchangeTradeInfo: exchangeRepoMock,
	}
	exchangeRepoMock.On("GetCurrentKline", "BTCUSDT").Return(&model.KLine{
		Close: 60000.99,
	})
	tradeLimit := model.TradeLimit{
		TradeFiltersBuy: model.TradeFilters{
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterPrice,
				Condition: model.TradeFilterConditionGte,
				Value:     "50000.00",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterPrice,
				Condition: model.TradeFilterConditionLte,
				Value:     "70000.00",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterPrice,
				Condition: model.TradeFilterConditionGt,
				Value:     "50000.00",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterPrice,
				Condition: model.TradeFilterConditionLt,
				Value:     "70000.00",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterPrice,
				Condition: model.TradeFilterConditionEq,
				Value:     "60000.99",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterPrice,
				Condition: model.TradeFilterConditionNeq,
				Value:     "60000.98",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
		},
	}
	assertion.True(filterService.CanBuy(tradeLimit))
}

func TestTradeFilterMatchingChildrenTrue(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepoMock := new(ExchangeTradeInfoMock)
	binance := new(ExchangePriceAPIMock)

	filterService := exchange.TradeFilterService{
		ExchangeTradeInfo: exchangeRepoMock,
		ExchangePriceAPI:  binance,
	}
	exchangeRepoMock.On("GetCurrentKline", "BTCUSDT").Return(&model.KLine{
		Close: 60000.99,
	})
	binance.On("GetKLinesCached", "BTCUSDT", "1d", int64(1)).Return([]model.KLine{
		{
			Open:  100.00,
			Close: 95.00,
		},
	})
	tradeLimit := model.TradeLimit{
		TradeFiltersBuy: model.TradeFilters{
			{
				Children: model.TradeFilters{
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionGte,
						Value:     "50000.00",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionLte,
						Value:     "70000.00",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterDailyPercent,
						Condition: model.TradeFilterConditionEq,
						Value:     "-5.00",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionGt,
						Value:     "50000.00",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
				},
				Type: model.TradeFilterConditionTypeAnd,
			},
			{
				Children: model.TradeFilters{
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionLt,
						Value:     "70000.00",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionEq,
						Value:     "60000.99",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionNeq,
						Value:     "60000.98",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
				},
				Type: model.TradeFilterConditionTypeAnd,
			},
		},
	}
	assertion.True(filterService.CanBuy(tradeLimit))
}

func TestTradeFilterMatchingChildrenFalse(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepoMock := new(ExchangeTradeInfoMock)

	filterService := exchange.TradeFilterService{
		ExchangeTradeInfo: exchangeRepoMock,
	}
	exchangeRepoMock.On("GetCurrentKline", "BTCUSDT").Return(&model.KLine{
		Close: 60000.99,
	})
	tradeLimit := model.TradeLimit{
		TradeFiltersBuy: model.TradeFilters{
			{
				Children: model.TradeFilters{
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionGte,
						Value:     "50000.00",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionLte,
						Value:     "70000.00",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionGt,
						Value:     "50000.00",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
				},
				Type: model.TradeFilterConditionTypeAnd,
			},
			{
				Children: model.TradeFilters{
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionLt,
						Value:     "70000.00",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionEq,
						Value:     "60001.00",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
					{
						Symbol:    "BTCUSDT",
						Parameter: model.TradeFilterParameterPrice,
						Condition: model.TradeFilterConditionNeq,
						Value:     "60000.98",
						Type:      model.TradeFilterConditionTypeAnd,
						Children:  make(model.TradeFilters, 0),
					},
				},
				Type: model.TradeFilterConditionTypeAnd,
			},
		},
	}
	assertion.False(filterService.CanBuy(tradeLimit))
}

func TestTradeFilterMatchingPositionTimeMinutes(t *testing.T) {
	assertion := assert.New(t)

	orderStorageMock := new(OrderStorageMock)

	filterService := exchange.TradeFilterService{
		OrderRepository: orderStorageMock,
	}
	orderStorageMock.On("GetOpenedOrderCached", "BTCUSDT", "BUY").Return(&model.Order{
		CreatedAt: time.Now().Add(time.Minute * -5).Format("2006-01-02 15:04:05"),
	})
	tradeLimit := model.TradeLimit{
		TradeFiltersBuy: model.TradeFilters{
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterPositionTimeMinutes,
				Condition: model.TradeFilterConditionLt,
				Value:     "7",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterPositionTimeMinutes,
				Condition: model.TradeFilterConditionGt,
				Value:     "4",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
		},
	}
	assertion.True(filterService.CanBuy(tradeLimit))
}

func TestTradeFilterMatchingExtraOrdersCountToday(t *testing.T) {
	assertion := assert.New(t)

	orderStorageMock := new(OrderStorageMock)

	filterService := exchange.TradeFilterService{
		OrderRepository: orderStorageMock,
	}
	orderMap := sync.Map{}
	orderMap.Store("BTCUSDT", float64(5.00)) // attention, must be float64
	orderStorageMock.On("GetTodayExtraOrderMap").Return(&orderMap)
	tradeLimit := model.TradeLimit{
		TradeFiltersBuy: model.TradeFilters{
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterExtraOrdersToday,
				Condition: model.TradeFilterConditionEq,
				Value:     "5",
				Type:      model.TradeFilterConditionTypeOr,
				Children:  make(model.TradeFilters, 0),
			},
		},
	}
	assertion.True(filterService.CanBuy(tradeLimit))
}

func TestTradeFilterMatchingExtraOrdersCountTodayEmptyMap(t *testing.T) {
	assertion := assert.New(t)

	orderStorageMock := new(OrderStorageMock)

	filterService := exchange.TradeFilterService{
		OrderRepository: orderStorageMock,
	}
	orderMap := sync.Map{}
	orderStorageMock.On("GetTodayExtraOrderMap").Return(&orderMap)
	tradeLimit := model.TradeLimit{
		TradeFiltersBuy: model.TradeFilters{
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterExtraOrdersToday,
				Condition: model.TradeFilterConditionEq,
				Value:     "0",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
		},
	}
	assertion.True(filterService.CanBuy(tradeLimit))
}

func TestTradeFilterMatchingHasSignal(t *testing.T) {
	assertion := assert.New(t)

	orderStorageMock := new(OrderStorageMock)
	signalStorage := new(SignalStorageMock)

	filterService := exchange.TradeFilterService{
		OrderRepository: orderStorageMock,
		SignalStorage:   signalStorage,
	}

	signalStorage.On("GetSignal", "BTCUSDT").Return(&model.Signal{
		Symbol:          "BTCUSDT",
		ExpireTimestamp: time.Now().Add(time.Minute).UnixMilli(),
	})
	tradeLimit := model.TradeLimit{
		TradeFiltersBuy: model.TradeFilters{
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterHasSignal,
				Condition: model.TradeFilterConditionEq,
				Value:     "1",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterHasSignal,
				Condition: model.TradeFilterConditionEq,
				Value:     "true",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterHasSignal,
				Condition: model.TradeFilterConditionNeq,
				Value:     "false",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
		},
	}
	assertion.True(filterService.CanBuy(tradeLimit))
}

func TestTradeFilterMatchingSentimentData(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepoMock := new(ExchangeTradeInfoMock)

	filterService := exchange.TradeFilterService{
		ExchangeTradeInfo: exchangeRepoMock,
	}

	sentimentScore := 0.34
	sentimentLabel := "BULLISH"

	exchangeRepoMock.On("GetTradeLimitCached", "BTCUSDT").Return(&model.TradeLimit{
		Symbol:         "BTCUSDT",
		SentimentScore: &sentimentScore,
		SentimentLabel: &sentimentLabel,
	})
	tradeLimit := model.TradeLimit{
		TradeFiltersBuy: model.TradeFilters{
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterSentimentLabel,
				Condition: model.TradeFilterConditionEq,
				Value:     "BULLISH",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
			{
				Symbol:    "BTCUSDT",
				Parameter: model.TradeFilterParameterSentimentScore,
				Condition: model.TradeFilterConditionGte,
				Value:     "0.34",
				Type:      model.TradeFilterConditionTypeAnd,
				Children:  make(model.TradeFilters, 0),
			},
		},
	}
	assertion.True(filterService.CanBuy(tradeLimit))
}
