package tests

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"io/ioutil"
	"testing"
	"time"
)

func TestCalculateBuyPriceByFrame1(t *testing.T) {
	assertion := assert.New(t)

	content, _ := ioutil.ReadFile("example/ethusdt@depth.json")
	var depth model.DepthEvent
	json.Unmarshal(content, &depth)

	exchangeRepoMock := new(ExchangePriceStorageMock)
	orderRepositoryMock := new(OrderCachedReaderMock)
	frameServiceMock := new(FrameServiceMock)
	lossSecurityMock := new(LossSecurityMock)
	profitService := new(ProfitServiceMock)
	signalStorage := new(SignalStorageMock)

	priceCalculator := exchange.PriceCalculator{
		LossSecurity:       lossSecurityMock,
		ExchangeRepository: exchangeRepoMock,
		OrderRepository:    orderRepositoryMock,
		FrameService:       frameServiceMock,
		Formatter:          &utils.Formatter{},
		ProfitService:      profitService,
		SignalStorage:      signalStorage,
	}

	signalStorage.On("GetSignal", "ETHUSDT").Return(nil)
	tradeLimit := model.TradeLimit{
		Symbol:      "ETHUSDT",
		MinPrice:    0.01,
		MinQuantity: 0.0001,
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:           0,
				OptionValue:     1,
				OptionUnit:      model.ProfitOptionUnitMinute,
				OptionPercent:   2.50,
				IsTriggerOption: true,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 1.25,
			},
		},
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
	}
	exchangeRepoMock.On("GetDepth", "ETHUSDT", int64(20)).Return(depth.Depth)
	exchangeRepoMock.On("GetCurrentKline", "ETHUSDT").Return(&model.KLine{
		Close:     1474.64,
		UpdatedAt: time.Now().Unix(),
	})
	exchangeRepoMock.On("GetPeriodMinPrice", "ETHUSDT", int64(200)).Return(900.00)
	orderRepositoryMock.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(nil)
	frameServiceMock.On("GetFrame", "ETHUSDT", "2h", int64(20)).Return(model.Frame{
		High:    1480.00,
		Low:     1250.30,
		AvgHigh: 1400.00,
		AvgLow:  1300.00,
	})
	profitService.On("GetMinClosePrice", tradeLimit, 1474.64).Return(900.00)
	lossSecurityMock.On("CheckBuyPriceOnHistory", tradeLimit, 900.00).Return(900.00)
	lossSecurityMock.On("BuyPriceCorrection", 900.00, tradeLimit).Return(900.00)

	priceModel := priceCalculator.CalculateBuy(tradeLimit)
	assertion.Nil(priceModel.Error)
	assertion.Equal(900.00, priceModel.Price)
}

func TestCalculateBuyPriceByFrame2(t *testing.T) {
	assertion := assert.New(t)

	content, _ := ioutil.ReadFile("example/ethusdt@depth.json")
	var depth model.DepthEvent
	json.Unmarshal(content, &depth)

	exchangeRepoMock := new(ExchangePriceStorageMock)
	orderRepositoryMock := new(OrderCachedReaderMock)
	frameServiceMock := new(FrameServiceMock)
	lossSecurityMock := new(LossSecurityMock)
	profitService := new(ProfitServiceMock)
	signalStorage := new(SignalStorageMock)

	priceCalculator := exchange.PriceCalculator{
		LossSecurity:       lossSecurityMock,
		ExchangeRepository: exchangeRepoMock,
		OrderRepository:    orderRepositoryMock,
		FrameService:       frameServiceMock,
		Formatter:          &utils.Formatter{},
		ProfitService:      profitService,
		SignalStorage:      signalStorage,
	}

	signalStorage.On("GetSignal", "ETHUSDT").Return(nil)
	tradeLimit := model.TradeLimit{
		Symbol:      "ETHUSDT",
		MinPrice:    0.01,
		MinQuantity: 0.0001,
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:           0,
				OptionValue:     1,
				OptionUnit:      model.ProfitOptionUnitMinute,
				OptionPercent:   2.50,
				IsTriggerOption: true,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 1.25,
			},
		},
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
	}
	exchangeRepoMock.On("GetDepth", "ETHUSDT", int64(20)).Return(depth.Depth)
	exchangeRepoMock.On("GetCurrentKline", "ETHUSDT").Return(&model.KLine{
		Close:     1474.64,
		UpdatedAt: time.Now().Unix(),
	})
	exchangeRepoMock.On("GetPeriodMinPrice", "ETHUSDT", int64(200)).Return(1300.00)
	orderRepositoryMock.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(nil)
	frameServiceMock.On("GetFrame", "ETHUSDT", "2h", int64(20)).Return(model.Frame{
		High:    1480.00,
		Low:     1250.30,
		AvgHigh: 1400.00,
		AvgLow:  1300.00,
	})
	profitService.On("GetMinClosePrice", tradeLimit, 1474.64).Return(1300.00)
	lossSecurityMock.On("CheckBuyPriceOnHistory", tradeLimit, 1300.00).Return(1300.00)
	lossSecurityMock.On("BuyPriceCorrection", 1300.00, tradeLimit).Return(1200.00)

	priceModel := priceCalculator.CalculateBuy(tradeLimit)
	assertion.Nil(priceModel.Error)
	assertion.Equal(1200.00, priceModel.Price)
}

func TestCalculateBuyPriceByFrame3(t *testing.T) {
	assertion := assert.New(t)

	content, _ := ioutil.ReadFile("example/ethusdt@depth.json")
	var depth model.DepthEvent
	json.Unmarshal(content, &depth)

	exchangeRepoMock := new(ExchangePriceStorageMock)
	orderRepositoryMock := new(OrderCachedReaderMock)
	frameServiceMock := new(FrameServiceMock)
	lossSecurityMock := new(LossSecurityMock)
	profitService := new(ProfitServiceMock)
	signalStorage := new(SignalStorageMock)

	priceCalculator := exchange.PriceCalculator{
		LossSecurity:       lossSecurityMock,
		ExchangeRepository: exchangeRepoMock,
		OrderRepository:    orderRepositoryMock,
		FrameService:       frameServiceMock,
		Formatter:          &utils.Formatter{},
		ProfitService:      profitService,
		SignalStorage:      signalStorage,
	}

	signalStorage.On("GetSignal", "ETHUSDT").Return(nil)
	tradeLimit := model.TradeLimit{
		Symbol:      "ETHUSDT",
		MinPrice:    0.01,
		MinQuantity: 0.0001,
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:           0,
				OptionValue:     1,
				OptionUnit:      model.ProfitOptionUnitMinute,
				OptionPercent:   2.50,
				IsTriggerOption: true,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 1.25,
			},
		},
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
	}
	exchangeRepoMock.On("GetDepth", "ETHUSDT", int64(20)).Return(depth.Depth)
	exchangeRepoMock.On("GetCurrentKline", "ETHUSDT").Return(&model.KLine{
		Close:     1474.64,
		UpdatedAt: time.Now().Unix(),
	})
	exchangeRepoMock.On("GetPeriodMinPrice", "ETHUSDT", int64(200)).Return(1400.00)
	orderRepositoryMock.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(nil)
	frameServiceMock.On("GetFrame", "ETHUSDT", "2h", int64(20)).Return(model.Frame{
		High:    1480.00,
		Low:     1250.30,
		AvgHigh: 1400.00,
		AvgLow:  1300.00,
	})
	profitService.On("GetMinClosePrice", tradeLimit, 1474.64).Return(1131.7)
	lossSecurityMock.On("CheckBuyPriceOnHistory", tradeLimit, 1400.00).Return(1131.7)
	lossSecurityMock.On("CheckBuyPriceOnHistory", tradeLimit, 1365.850000000099).Return(1131.7)
	lossSecurityMock.On("BuyPriceCorrection", 1131.7, tradeLimit).Return(1131.1)

	priceModel := priceCalculator.CalculateBuy(tradeLimit)
	assertion.Nil(priceModel.Error)
	assertion.Equal(1131.1, priceModel.Price)
}

func TestCalculateSell(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepoMock := new(ExchangePriceStorageMock)
	orderRepositoryMock := new(OrderCachedReaderMock)
	frameServiceMock := new(FrameServiceMock)
	lossSecurityMock := new(LossSecurityMock)
	profitService := new(ProfitServiceMock)
	signalStorage := new(SignalStorageMock)

	priceCalculator := exchange.PriceCalculator{
		LossSecurity:       lossSecurityMock,
		ExchangeRepository: exchangeRepoMock,
		OrderRepository:    orderRepositoryMock,
		FrameService:       frameServiceMock,
		Formatter:          &utils.Formatter{},
		ProfitService:      profitService,
		SignalStorage:      signalStorage,
	}

	signalStorage.On("GetSignal", "ETHUSDT").Return(nil)
	tradeLimit := model.TradeLimit{
		Symbol:                       "ETHUSDT",
		MinPrice:                     0.01,
		MinQuantity:                  0.0001,
		ProfitOptions:                model.ProfitOptions{},
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
	}

	order := model.Order{
		Price: 1552.26,
	}

	exchangeRepoMock.On("GetCurrentKline", "ETHUSDT").Times(1).Return(nil)
	sellPrice, err := priceCalculator.CalculateSell(tradeLimit, order)
	assertion.ErrorContains(err, "price is unknown")
	assertion.Equal(0.00, sellPrice)

	exchangeRepoMock.On("GetCurrentKline", "ETHUSDT").Times(2).Return(&model.KLine{
		Close:     1474.64,
		UpdatedAt: time.Now().Unix(),
	})
	profitService.On("GetMinClosePrice", order, 1552.26).Times(1).Return(1400.00)

	sellPrice, err = priceCalculator.CalculateSell(tradeLimit, order)
	assertion.ErrorContains(err, "price is less than order price")
	assertion.Equal(0.00, sellPrice)

	profitService.On("GetMinClosePrice", order, 1552.26).Times(2).Return(1600.00)
	exchangeRepoMock.On("GetCurrentKline", "ETHUSDT").Times(3).Return(&model.KLine{
		Close:     1474.64,
		UpdatedAt: time.Now().Unix(),
	})

	sellPrice, err = priceCalculator.CalculateSell(tradeLimit, order)
	assertion.Nil(err)
	assertion.Equal(1600.00, sellPrice)
}

func TestCalculateBuyPriceBySignal(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepoMock := new(ExchangePriceStorageMock)
	orderRepositoryMock := new(OrderCachedReaderMock)
	frameServiceMock := new(FrameServiceMock)
	lossSecurityMock := new(LossSecurityMock)
	profitService := new(ProfitServiceMock)
	signalStorage := new(SignalStorageMock)

	priceCalculator := exchange.PriceCalculator{
		LossSecurity:       lossSecurityMock,
		ExchangeRepository: exchangeRepoMock,
		OrderRepository:    orderRepositoryMock,
		FrameService:       frameServiceMock,
		Formatter:          &utils.Formatter{},
		ProfitService:      profitService,
		SignalStorage:      signalStorage,
	}

	tradeLimit := model.TradeLimit{
		Symbol:      "ETHUSDT",
		MinPrice:    0.01,
		MinQuantity: 0.0001,
		ProfitOptions: model.ProfitOptions{
			model.ProfitOption{
				Index:           0,
				OptionValue:     1,
				OptionUnit:      model.ProfitOptionUnitMinute,
				OptionPercent:   2.50,
				IsTriggerOption: true,
			},
			model.ProfitOption{
				Index:         1,
				OptionValue:   2,
				OptionUnit:    model.ProfitOptionUnitHour,
				OptionPercent: 1.25,
			},
		},
		MinPriceMinutesPeriod:        200,
		FrameInterval:                "2h",
		FramePeriod:                  20,
		BuyPriceHistoryCheckInterval: "1d",
		BuyPriceHistoryCheckPeriod:   14,
	}

	exchangeRepoMock.On("GetCurrentKline", "ETHUSDT").Times(1).Return(&model.KLine{
		Close:     3000.00,
		UpdatedAt: time.Now().Unix(),
	})
	signalStorage.On("GetSignal", "ETHUSDT").Return(&model.Signal{
		Symbol:          "ETHUSDT",
		BuyPrice:        2090.00,
		ExpireTimestamp: time.Now().Add(time.Minute).UnixMilli(),
	})
	orderRepositoryMock.On("GetOpenedOrderCached", "ETHUSDT", "BUY").Return(nil)
	lossSecurityMock.On("BuyPriceCorrection", 2090.00, tradeLimit).Return(2080.00)

	priceModel := priceCalculator.CalculateBuy(tradeLimit)
	assertion.Equal(2080.00, priceModel.Price)
	assertion.Nil(priceModel.Error)
}
