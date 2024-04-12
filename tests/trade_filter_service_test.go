package tests

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"testing"
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
	exchangeRepoMock.On("GetLastKLine", "BTCUSDT").Return(&model.KLine{
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
	exchangeRepoMock.On("GetLastKLine", "BTCUSDT").Return(&model.KLine{
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
	exchangeRepoMock.On("GetLastKLine", "BTCUSDT").Return(&model.KLine{
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
	exchangeRepoMock.On("GetLastKLine", "BTCUSDT").Return(&model.KLine{
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
