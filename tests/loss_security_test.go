package tests

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"testing"
)

func TestBuyPriceCorrection(t *testing.T) {
	exchangeRepo := new(ExchangeTradeInfoMock)
	binance := new(ExchangePriceAPIMock)

	lossSecurity := service.LossSecurity{
		MlEnabled:            true,
		InterpolationEnabled: true,
		Formatter:            &service.Formatter{},
		ExchangeRepository:   exchangeRepo,
		Binance:              binance,
	}

	assertion := assert.New(t)

	limit := model.TradeLimit{
		Symbol:                       "BTCUSDT",
		BuyPriceHistoryCheckInterval: "1h",
		BuyPriceHistoryCheckPeriod:   10,
		MinProfitPercent:             2.40,
		MinPrice:                     0.001,
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
	exchangeRepo.On("GetLastKLine", "BTCUSDT").Return(&kline)
	exchangeRepo.On("GetPredict", "BTCUSDT").Return(22450.00, nil)
	exchangeRepo.On("GetInterpolation", kline).Return(model.Interpolation{
		Asset:                "BTC",
		BtcInterpolationUsdt: 21450.00,
		EthInterpolationUsdt: 21425.00,
	}, nil)

	price := lossSecurity.CheckBuyPriceOnHistory(limit, 25000.00)
	assertion.Equal(21484.374999283773, price)

	price = lossSecurity.BuyPriceCorrection(price, limit)
	assertion.Equal(21425.00, price)
}
