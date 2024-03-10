package tests

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"testing"
	"time"
)

func TestNotEnoughDecisions(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	decisionStorage := new(DecisionReadStorageMock)
	orderStorage := new(OrderStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := exchange.StrategyFacade{
		DecisionReadStorage: decisionStorage,
		ExchangeRepository:  exchangeRepository,
		OrderRepository:     orderStorage,
		BotService:          botService,
		MinDecisions:        2.00,
	}

	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)
	decisionStorage.On("GetDecisions", "BTCUSDT").Return([]model.Decision{
		{
			Score:     1.00,
			Operation: "HOLD",
		},
	})

	result, err := strategyFacade.Decide("BTCUSDT")
	assertion.ErrorContains(err, "[BTCUSDT] Not enough decision amount 1 of 2")
	assertion.Equal(999.99, result.Hold)
}

func TestCantGetTradeLimit(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	decisionStorage := new(DecisionReadStorageMock)
	orderStorage := new(OrderStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := exchange.StrategyFacade{
		DecisionReadStorage: decisionStorage,
		ExchangeRepository:  exchangeRepository,
		OrderRepository:     orderStorage,
		BotService:          botService,
		MinDecisions:        2.00,
	}

	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)
	decisionStorage.On("GetDecisions", "BTCUSDT").Return([]model.Decision{
		{
			Score:     1.00,
			Operation: "HOLD",
		},
		{
			Score:     1.00,
			Operation: "HOLD",
		},
	})
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(model.TradeLimit{}, errors.New("Test TL!!!"))

	result, err := strategyFacade.Decide("BTCUSDT")
	assertion.ErrorContains(err, "[BTCUSDT] Test TL!!!")
	assertion.Equal(999.99, result.Hold)
}

func TestCantGetLastKLine(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	decisionStorage := new(DecisionReadStorageMock)
	orderStorage := new(OrderStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := exchange.StrategyFacade{
		DecisionReadStorage: decisionStorage,
		ExchangeRepository:  exchangeRepository,
		OrderRepository:     orderStorage,
		BotService:          botService,
		MinDecisions:        2.00,
	}

	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)
	decisionStorage.On("GetDecisions", "BTCUSDT").Return([]model.Decision{
		{
			Score:     1.00,
			Operation: "HOLD",
		},
		{
			Score:     1.00,
			Operation: "HOLD",
		},
	})
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(model.TradeLimit{
		Symbol: "BTCUSDT",
	}, nil)
	exchangeRepository.On("GetLastKLine", "BTCUSDT").Return(nil)

	result, err := strategyFacade.Decide("BTCUSDT")
	assertion.ErrorContains(err, "[BTCUSDT] Last price is unknown")
	assertion.Equal(999.99, result.Hold)
}

func TestPriceIsExpired(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	decisionStorage := new(DecisionReadStorageMock)
	orderStorage := new(OrderStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := exchange.StrategyFacade{
		DecisionReadStorage: decisionStorage,
		ExchangeRepository:  exchangeRepository,
		OrderRepository:     orderStorage,
		BotService:          botService,
		MinDecisions:        2.00,
	}

	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)
	decisionStorage.On("GetDecisions", "BTCUSDT").Return([]model.Decision{
		{
			Score:     2.00,
			Operation: "BUY",
		},
		{
			Score:     1.00,
			Operation: "SELL",
		},
	})
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(model.TradeLimit{
		Symbol: "BTCUSDT",
	}, nil)
	exchangeRepository.On("GetLastKLine", "BTCUSDT").Return(&model.KLine{
		UpdatedAt: time.Now().Unix() - 60,
	})

	result, err := strategyFacade.Decide("BTCUSDT")
	assertion.ErrorContains(err, "[BTCUSDT] Last price is expired")
	assertion.Equal(999.99, result.Hold)
}

func TestDropHoldForHighPriority(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	decisionStorage := new(DecisionReadStorageMock)
	orderStorage := new(OrderStorageMock)
	botService := new(BotServiceMock)
	strategyFacade := exchange.StrategyFacade{
		DecisionReadStorage: decisionStorage,
		ExchangeRepository:  exchangeRepository,
		OrderRepository:     orderStorage,
		BotService:          botService,
		MinDecisions:        3.00,
	}

	orderStorage.On("GetManualOrder", "BTCUSDT").Return(nil)
	decisionStorage.On("GetDecisions", "BTCUSDT").Times(1).Return([]model.Decision{
		{
			Score:     999.99,
			Operation: "BUY",
		},
		{
			Score:     1.00,
			Operation: "SELL",
		},
		{
			Score:     888.00,
			Operation: "HOLD",
		},
	})
	decisionStorage.On("GetDecisions", "BTCUSDT").Times(1).Return([]model.Decision{
		{
			Score:     999.99,
			Operation: "SELL",
		},
		{
			Score:     1.00,
			Operation: "BUY",
		},
		{
			Score:     888.00,
			Operation: "HOLD",
		},
	})
	exchangeRepository.On("GetTradeLimit", "BTCUSDT").Return(model.TradeLimit{
		Symbol: "BTCUSDT",
	}, nil)
	exchangeRepository.On("GetLastKLine", "BTCUSDT").Return(&model.KLine{
		UpdatedAt: time.Now().Unix(),
	})

	result, err := strategyFacade.Decide("BTCUSDT")
	assertion.Nil(err)
	assertion.Equal(0.00, result.Hold)
	assertion.Equal(1.00, result.Sell)
	assertion.Equal(999.99, result.Buy)

	result, err = strategyFacade.Decide("BTCUSDT")
	assertion.Nil(err)
	assertion.Equal(0.00, result.Hold)
	assertion.Equal(999.99, result.Sell)
	assertion.Equal(1.00, result.Buy)
}
