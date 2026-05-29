package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	tickevent "github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/tickstore"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/service/exchange"
)

func freshTickWithCandles(symbol string) tickevent.MarketTick {
	now := time.Now().UTC()
	return tickevent.MarketTick{
		Exchange:   "binance",
		Symbol:     symbol,
		Source:     tickevent.SourceTrade,
		EventTime:  now,
		IngestedAt: now,
		Price:      decimal.NewFromFloat(60000),
		Candles: tickevent.Candles{
			Resolution: time.Minute,
			Series: []tickevent.Candle{
				{
					OpenTime: now.Add(-time.Minute),
					Open:     decimal.NewFromFloat(59999),
					Close:    decimal.NewFromFloat(60000),
					Volume:   decimal.NewFromInt(10),
				},
			},
		},
	}
}

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
		TickStore:           tickstore.NewInMemory(),
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
		TickStore:           tickstore.NewInMemory(),
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

// Phase D rewrite of the legacy TestCantGetCurrentKline. The invariant
// is preserved verbatim: when the strategy has no recent price view, the
// facade returns the kill-switch shape (Hold = DecisionHighestPriorityScore,
// no Buy/Sell). The symbol of "no recent price view" changed from
// ExchangeRepository.GetCurrentKline returning nil to TickStore.Latest
// returning a tick with empty Candles.
func TestCantGetCurrentKline(t *testing.T) {
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
		TickStore:           tickstore.NewInMemory(),
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

	result, err := strategyFacade.Decide("BTCUSDT")
	assertion.ErrorContains(err, "[BTCUSDT] no candles on tick")
	assertion.Equal(999.99, result.Hold)
}

// Phase D rewrite of the legacy TestPriceIsExpired. Same shape: a stale
// tick + a buy-leaning decision set must trip the freshness kill-switch.
func TestPriceIsExpired(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	decisionStorage := new(DecisionReadStorageMock)
	orderStorage := new(OrderStorageMock)
	botService := new(BotServiceMock)
	ticks := tickstore.NewInMemory()
	stale := freshTickWithCandles("BTCUSDT")
	stale.EventTime = time.Now().UTC().Add(-time.Hour)
	ticks.Set(stale)
	strategyFacade := exchange.StrategyFacade{
		DecisionReadStorage: decisionStorage,
		ExchangeRepository:  exchangeRepository,
		OrderRepository:     orderStorage,
		BotService:          botService,
		TickStore:           ticks,
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

	result, err := strategyFacade.Decide("BTCUSDT")
	assertion.ErrorContains(err, "[BTCUSDT] tick is stale")
	assertion.Equal(999.99, result.Hold)
}

func TestDropHoldForHighPriority(t *testing.T) {
	assertion := assert.New(t)

	exchangeRepository := new(ExchangeTradeInfoMock)
	decisionStorage := new(DecisionReadStorageMock)
	orderStorage := new(OrderStorageMock)
	botService := new(BotServiceMock)
	ticks := tickstore.NewInMemory()
	ticks.Set(freshTickWithCandles("BTCUSDT"))
	strategyFacade := exchange.StrategyFacade{
		DecisionReadStorage: decisionStorage,
		ExchangeRepository:  exchangeRepository,
		OrderRepository:     orderStorage,
		BotService:          botService,
		TickStore:           ticks,
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
