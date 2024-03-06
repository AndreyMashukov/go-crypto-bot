package tests

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"testing"
	"time"
)

func TestBuyPriceHistoryCorrection(t *testing.T) {
	assertion := assert.New(t)

	binance := new(ExchangePriceAPIMock)
	botService := new(BotServiceMock)

	profitService := exchange.ProfitService{
		Binance:    binance,
		BotService: botService,
	}

	limit := model.TradeLimit{
		Symbol:                       "BTCUSDT",
		BuyPriceHistoryCheckInterval: "1h",
		BuyPriceHistoryCheckPeriod:   10,
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:           0,
				OptionValue:     1,
				OptionUnit:      model.ProfitOptionUnitMinute,
				OptionPercent:   2.40,
				IsTriggerOption: true,
			},
			model.ProfitOption{
				Index:           1,
				OptionValue:     2,
				OptionUnit:      model.ProfitOptionUnitHour,
				OptionPercent:   2.80,
				IsTriggerOption: false,
			},
		},
		MinPrice: 0.001,
	}

	binance.On("GetKLinesCached", limit.Symbol, limit.BuyPriceHistoryCheckInterval, limit.BuyPriceHistoryCheckPeriod).Return([]model.KLine{
		{
			High: 26000.00,
		},
		{
			High: 25000.00,
		},
		{
			High: 23000.00,
		},
		{
			High: 23000.00,
		},
		{
			High: 23000.00,
		},
		{
			High: 24000.00,
		},
		{
			High: 21500.00,
		},
		{
			High: 22000.00,
		},
		{
			High: 23000.00,
		},
		{
			High: 23500.00,
		},
	})
	botService.On("UseSwapCapital").Return(true)

	price := profitService.CheckBuyPriceOnHistory(limit, 23500)
	assertion.Equal(21484.374999589363, price)
}

func TestShouldCalculateMultiStepMinProfitPercent(t *testing.T) {
	assertion := assert.New(t)

	binance := new(ExchangePriceAPIMock)
	botService := new(BotServiceMock)
	botService.On("UseSwapCapital").Return(true)

	profitService := exchange.ProfitService{
		Binance:    binance,
		BotService: botService,
	}

	orderNow := model.Order{
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 2.50,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 1.25,
			},
		},
	}
	assertion.Equal(model.Percent(2.5), profitService.GetMinProfitPercent(orderNow))

	orderNext := model.Order{
		CreatedAt: time.Now().Add(time.Minute * -1).Add(time.Second * -1).Format("2006-01-02 15:04:05"),
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 2.50,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 1.25,
			},
		},
	}
	assertion.Equal(model.Percent(1.25), profitService.GetMinProfitPercent(orderNext))

	orderLast := model.Order{
		CreatedAt: time.Now().Add(time.Hour * -3).Format("2006-01-02 15:04:05"),
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         0,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitMinute,
				OptionPercent: 2.50,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 1.25,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitDay,
				OptionPercent: 0.75,
			},
		},
	}
	assertion.Equal(model.Percent(0.75), profitService.GetMinProfitPercent(orderLast))

	orderUndefined := model.Order{
		CreatedAt: time.Now().Add(time.Hour * -48).Format("2006-01-02 15:04:05"),
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 1.25,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitDay,
				OptionPercent: 0.75,
			},
			model.ProfitOption{
				Index:         2,
				OptionValue:   1.5,
				OptionUnit:    model.ProfitOptionUnitDay,
				OptionPercent: 0.90,
			},
		},
	}
	assertion.Equal(model.Percent(0.90), profitService.GetMinProfitPercent(orderUndefined))

	orderEmpty := model.Order{
		CreatedAt:     time.Now().Add(time.Hour * -48).Format("2006-01-02 15:04:05"),
		ProfitOptions: model.ProfitOptions{},
	}
	assertion.Equal(model.Percent(0.50), profitService.GetMinProfitPercent(orderEmpty))
}

func TestShouldCalculateProfitWithSwapValue(t *testing.T) {
	assertion := assert.New(t)

	binance := new(ExchangePriceAPIMock)
	botService := new(BotServiceMock)
	botService.On("UseSwapCapital").Return(true)

	profitService := exchange.ProfitService{
		Binance:    binance,
		BotService: botService,
	}

	soldQuantity := 0.00
	swapQuantity := 0.00

	order := model.Order{
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 1.25,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitDay,
				OptionPercent: 0.75,
			},
		},
		ExecutedQuantity: 1000,
		SoldQuantity:     &soldQuantity,
		SwapQuantity:     &swapQuantity,
		Price:            100.00,
	}

	assertion.Equal(0.00, order.GetQuoteProfit(100.00, true))
	assertion.Equal(0.00, order.GetQuoteProfit(100.00, false))
	assertion.Equal(model.Percent(0.00), order.GetProfitPercent(100.00, true))
	assertion.Equal(model.Percent(0.00), order.GetProfitPercent(100.00, false))
	assertion.Equal(1000.00, order.GetRemainingToSellQuantity(true))
	assertion.Equal(1000.00, order.GetRemainingToSellQuantity(false))
	assertion.Equal(1000.00, order.GetPositionQuantityWithSwap())
	assertion.Equal(1000.00, order.GetExecutedQuantity())
	assertion.Equal(model.Percent(1.25), profitService.GetMinProfitPercent(order))
	assertion.Equal(101.25, profitService.GetMinClosePrice(order, 100.00))

	soldQuantity = 0.00
	swapQuantity = 1.25

	order = model.Order{
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 1.25,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   1,
				OptionUnit:    model.ProfitOptionUnitDay,
				OptionPercent: 0.75,
			},
		},
		ExecutedQuantity: 100,
		SoldQuantity:     &soldQuantity,
		SwapQuantity:     &swapQuantity,
		Price:            1.00,
	}

	assertion.Equal(1.25, order.GetQuoteProfit(1.00, true))
	assertion.Equal(model.Percent(1.25), order.GetProfitPercent(1.00, true))
	assertion.Equal(model.Percent(0.00), order.GetProfitPercent(1.00, false))
	assertion.Equal(0.00, order.GetQuoteProfit(1.00, false))
	assertion.Equal(101.25, order.GetRemainingToSellQuantity(true))
	assertion.Equal(100.00, order.GetRemainingToSellQuantity(false))
	assertion.Equal(101.25, order.GetPositionQuantityWithSwap())
	assertion.Equal(100.00, order.GetExecutedQuantity())
	assertion.Equal(model.Percent(1.25), profitService.GetMinProfitPercent(order))
	assertion.Equal(1.00, profitService.GetMinClosePrice(order, 1.00))
}
