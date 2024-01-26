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
	lossSecurityMock := new(LossSecurityMock)

	priceCalculator := service.PriceCalculator{
		LossSecurity:       lossSecurityMock,
		ExchangeRepository: exchangeRepoMock,
		OrderRepository:    orderRepositoryMock,
		FrameService:       frameServiceMock,
		Binance:            binanceMock,
		Formatter:          &service.Formatter{},
	}

	tradeLimit := model.TradeLimit{
		Symbol:                       "ETHUSDT",
		MinPrice:                     0.01,
		MinQuantity:                  0.0001,
		MinProfitPercent:             2.50,
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
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
	lossSecurityMock.On("CheckBuyPriceOnHistory", tradeLimit, 900.00).Return(900.00)
	lossSecurityMock.On("BuyPriceCorrection", 900.00, tradeLimit).Return(900.00)

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
	lossSecurityMock := new(LossSecurityMock)

	priceCalculator := service.PriceCalculator{
		LossSecurity:       lossSecurityMock,
		ExchangeRepository: exchangeRepoMock,
		OrderRepository:    orderRepositoryMock,
		FrameService:       frameServiceMock,
		Binance:            binanceMock,
		Formatter:          &service.Formatter{},
	}

	tradeLimit := model.TradeLimit{
		Symbol:                       "ETHUSDT",
		MinPrice:                     0.01,
		MinQuantity:                  0.0001,
		MinProfitPercent:             2.50,
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
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
	lossSecurityMock.On("CheckBuyPriceOnHistory", tradeLimit, 1300.00).Return(1300.00)
	lossSecurityMock.On("BuyPriceCorrection", 1300.00, tradeLimit).Return(1200.00)

	price, err := priceCalculator.CalculateBuy(tradeLimit)
	assertion.Nil(err)
	assertion.Equal(1200.00, price)
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
	lossSecurityMock := new(LossSecurityMock)

	priceCalculator := service.PriceCalculator{
		LossSecurity:       lossSecurityMock,
		ExchangeRepository: exchangeRepoMock,
		OrderRepository:    orderRepositoryMock,
		FrameService:       frameServiceMock,
		Binance:            binanceMock,
		Formatter:          &service.Formatter{},
	}

	tradeLimit := model.TradeLimit{
		Symbol:                       "ETHUSDT",
		MinPrice:                     0.01,
		MinQuantity:                  0.0001,
		MinProfitPercent:             2.50,
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
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
	lossSecurityMock.On("CheckBuyPriceOnHistory", tradeLimit, 1365.850000000099).Return(1131.7)
	lossSecurityMock.On("BuyPriceCorrection", 1131.7, tradeLimit).Return(1131.1)

	price, err := priceCalculator.CalculateBuy(tradeLimit)
	assertion.Nil(err)
	assertion.Equal(1131.1, price)
}
