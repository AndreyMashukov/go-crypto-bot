package tests

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"io/ioutil"
	"testing"
)

func TestCalculateBuyPriceByFrame1(t *testing.T) {
	assertion := assert.New(t)

	content, _ := ioutil.ReadFile("example/ethusdt@depth.json")
	var depth model.DepthEvent
	json.Unmarshal(content, &depth)

	exchangeRepoMock := new(ExchangePriceStorageMock)
	orderRepositoryMock := new(OrderCachedReaderMock)
	frameServiceMock := new(FrameServiceMock)
	binanceMock := new(ExchangePriceAPIMock)

	priceCalculator := service.PriceCalculator{
		ExchangeRepository:           exchangeRepoMock,
		OrderRepository:              orderRepositoryMock,
		FrameService:                 frameServiceMock,
		Binance:                      binanceMock,
		Formatter:                    &service.Formatter{},
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
	}

	tradeLimit := model.TradeLimit{
		Symbol:           "ETHUSDT",
		MinPrice:         0.01,
		MinQuantity:      0.0001,
		MinProfitPercent: 2.50,
	}
	exchangeRepoMock.On("GetDepth", "ETHUSDT").Return(depth.Depth)
	binanceMock.On("GetDepth", "ETHUSDT").Times(0)
	exchangeRepoMock.On("GetLastKLine", "ETHUSDT").Return(&model.KLine{
		Close: 1474.64,
	})
	exchangeRepoMock.On("GetPeriodMinPrice", "ETHUSDT", int64(200)).Return(900.00)
	orderRepositoryMock.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(model.Order{}, errors.New("Order is not found"))
	frameServiceMock.On("GetFrame", "ETHUSDT", "2h", int64(20)).Return(model.Frame{
		High:    1480.00,
		Low:     1250.30,
		AvgHigh: 1400.00,
		AvgLow:  1300.00,
	})
	binanceMock.On("GetKLinesCached", "ETHUSDT", "1d", int64(14)).Return([]model.KLineHistory{
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
	})

	price, err := priceCalculator.CalculateBuy(tradeLimit)
	assertion.Nil(err)
	assertion.Equal(900.00, price)
}

func TestCalculateBuyPriceByFrame2(t *testing.T) {
	assertion := assert.New(t)

	content, _ := ioutil.ReadFile("example/ethusdt@depth.json")
	var depth model.DepthEvent
	json.Unmarshal(content, &depth)

	exchangeRepoMock := new(ExchangePriceStorageMock)
	orderRepositoryMock := new(OrderCachedReaderMock)
	frameServiceMock := new(FrameServiceMock)
	binanceMock := new(ExchangePriceAPIMock)

	priceCalculator := service.PriceCalculator{
		ExchangeRepository:           exchangeRepoMock,
		OrderRepository:              orderRepositoryMock,
		FrameService:                 frameServiceMock,
		Binance:                      binanceMock,
		Formatter:                    &service.Formatter{},
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
	}

	tradeLimit := model.TradeLimit{
		Symbol:           "ETHUSDT",
		MinPrice:         0.01,
		MinQuantity:      0.0001,
		MinProfitPercent: 2.50,
	}
	exchangeRepoMock.On("GetDepth", "ETHUSDT").Return(depth.Depth)
	binanceMock.On("GetDepth", "ETHUSDT").Times(0)
	exchangeRepoMock.On("GetLastKLine", "ETHUSDT").Return(&model.KLine{
		Close: 1474.64,
	})
	exchangeRepoMock.On("GetPeriodMinPrice", "ETHUSDT", int64(200)).Return(1300.00)
	orderRepositoryMock.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(model.Order{}, errors.New("Order is not found"))
	frameServiceMock.On("GetFrame", "ETHUSDT", "2h", int64(20)).Return(model.Frame{
		High:    1480.00,
		Low:     1250.30,
		AvgHigh: 1400.00,
		AvgLow:  1300.00,
	})
	binanceMock.On("GetKLinesCached", "ETHUSDT", "1d", int64(14)).Return([]model.KLineHistory{
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
		{
			High: "1480.00",
		},
	})

	price, err := priceCalculator.CalculateBuy(tradeLimit)
	assertion.Nil(err)
	assertion.Equal(1300.00, price)
}

func TestCalculateBuyPriceByFrame3(t *testing.T) {
	assertion := assert.New(t)

	content, _ := ioutil.ReadFile("example/ethusdt@depth.json")
	var depth model.DepthEvent
	json.Unmarshal(content, &depth)

	exchangeRepoMock := new(ExchangePriceStorageMock)
	orderRepositoryMock := new(OrderCachedReaderMock)
	frameServiceMock := new(FrameServiceMock)
	binanceMock := new(ExchangePriceAPIMock)

	priceCalculator := service.PriceCalculator{
		ExchangeRepository:           exchangeRepoMock,
		OrderRepository:              orderRepositoryMock,
		FrameService:                 frameServiceMock,
		Binance:                      binanceMock,
		Formatter:                    &service.Formatter{},
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
	}

	tradeLimit := model.TradeLimit{
		Symbol:           "ETHUSDT",
		MinPrice:         0.01,
		MinQuantity:      0.0001,
		MinProfitPercent: 2.50,
	}
	exchangeRepoMock.On("GetDepth", "ETHUSDT").Return(depth.Depth)
	binanceMock.On("GetDepth", "ETHUSDT").Times(0)
	exchangeRepoMock.On("GetLastKLine", "ETHUSDT").Return(&model.KLine{
		Close: 1474.64,
	})
	exchangeRepoMock.On("GetPeriodMinPrice", "ETHUSDT", int64(200)).Return(1400.00)
	orderRepositoryMock.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(model.Order{}, errors.New("Order is not found"))
	frameServiceMock.On("GetFrame", "ETHUSDT", "2h", int64(20)).Return(model.Frame{
		High:    1480.00,
		Low:     1250.30,
		AvgHigh: 1400.00,
		AvgLow:  1300.00,
	})
	binanceMock.On("GetKLinesCached", "ETHUSDT", "1d", int64(14)).Return([]model.KLineHistory{
		{
			High: "1200.00",
		},
		{
			High: "1220.00",
		},
		{
			High: "1160.00",
		},
		{
			High: "1160.00",
		},
		{
			High: "1210.00",
		},
		{
			High: "1310.00",
		},
		{
			High: "1310.00",
		},
		{
			High: "1310.00",
		},
	})

	price, err := priceCalculator.CalculateBuy(tradeLimit)
	assertion.Nil(err)
	assertion.Equal(1131.7, price)
}
