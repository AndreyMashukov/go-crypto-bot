package tests

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"testing"
)

func TestBuyPriceCorrection(t *testing.T) {
	exchangeRepo := new(ExchangeTradeInfoMock)
	binance := new(ExchangePriceAPIMock)

	profitServiceMock := new(ProfitServiceMock)

	lossSecurity := exchange.LossSecurity{
		MlEnabled:            true,
		InterpolationEnabled: true,
		Formatter:            &utils.Formatter{},
		ExchangeRepository:   exchangeRepo,
		ProfitService:        profitServiceMock,
	}

	assertion := assert.New(t)

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
	kline := model.KLine{
		Close:  23000.00,
		Low:    22500.00,
		Symbol: "BTCUSDT",
	}
	exchangeRepo.On("GetCurrentKline", "BTCUSDT").Return(&kline)
	exchangeRepo.On("GetPredict", "BTCUSDT").Return(22450.00, nil)
	exchangeRepo.On("GetInterpolation", kline).Return(model.Interpolation{
		Asset:                "BTC",
		BtcInterpolationUsdt: 21450.00,
		EthInterpolationUsdt: 21425.00,
	}, nil)

	price := lossSecurity.BuyPriceCorrection(21484.374999283773, limit)
	assertion.Equal(21425.00, price)
}
