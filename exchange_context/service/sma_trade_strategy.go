package service

import (
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"math"
	"sync"
	"time"
)

type SmaTradeStrategy struct {
	Trades         map[string][]ExchangeModel.Trade
	TradesMapMutex sync.RWMutex
}

func (s *SmaTradeStrategy) Decide(trade ExchangeModel.Trade) ExchangeModel.Decision {
	sellPeriod := 15
	buyPeriod := 60
	maxPeriod := int(math.Max(float64(sellPeriod), float64(buyPeriod)))

	s.TradesMapMutex.Lock()
	s.Trades[trade.Symbol] = append(s.Trades[trade.Symbol], trade)

	if len(s.Trades[trade.Symbol]) < maxPeriod {
		s.TradesMapMutex.Unlock()

		return ExchangeModel.Decision{
			StrategyName: "sma_trade_strategy",
			Score:        50.00,
			Operation:    "HOLD",
			Timestamp:    time.Now().Unix(),
			Price:        trade.Price,
			Params:       [3]float64{0, 0, 0},
		}
	}

	s.TradesMapMutex.Unlock()

	s.TradesMapMutex.Lock()
	tradeSlice := s.Trades[trade.Symbol][len(s.Trades[trade.Symbol])-maxPeriod:]
	s.Trades[trade.Symbol] = tradeSlice // override to avoid memory leaks
	s.TradesMapMutex.Unlock()

	sellSma := s.calculateSMA(tradeSlice[len(tradeSlice)-sellPeriod:])
	buySma := s.calculateSMA(tradeSlice[len(tradeSlice)-buyPeriod:])

	buyVolumeS, sellVolumeS := s.getByAndSellVolume(tradeSlice[len(tradeSlice)-sellPeriod:])
	buyVolumeB, sellVolumeB := s.getByAndSellVolume(tradeSlice[len(tradeSlice)-buyPeriod:])

	buyIndicator := buyVolumeB / sellVolumeB

	if buyIndicator > 150 && buySma < trade.Price {
		return ExchangeModel.Decision{
			StrategyName: "sma_trade_strategy",
			Score:        50.00,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        trade.Price,
			Params:       [3]float64{buyVolumeB, sellVolumeB, buySma},
		}
	}

	sellIndicator := sellVolumeS / buyVolumeS

	if sellIndicator > 50 && sellSma > trade.Price {
		return ExchangeModel.Decision{
			StrategyName: "sma_trade_strategy",
			Score:        50.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        trade.Price,
			Params:       [3]float64{buyVolumeS, sellVolumeS, sellSma},
		}
	}

	return ExchangeModel.Decision{
		StrategyName: "sma_trade_strategy",
		Score:        50.00,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        trade.Price,
		Params:       [3]float64{buyVolumeS, sellVolumeS, sellSma},
	}
}

func (s *SmaTradeStrategy) calculateSMA(trades []ExchangeModel.Trade) float64 {
	var sum float64

	slice := trades

	for _, trade := range slice {
		sum += trade.Price
	}

	return sum / float64(len(slice))
}

func (s *SmaTradeStrategy) getByAndSellVolume(trades []ExchangeModel.Trade) (float64, float64) {
	var buyVolume float64
	var sellVolume float64

	for _, trade := range trades {
		switch trade.GetOperation() {
		case "BUY":
			buyVolume += trade.Price * trade.Quantity
			break
		case "SELL":
			sellVolume += trade.Price * trade.Quantity
			break
		}
	}

	return buyVolume, sellVolume
}
