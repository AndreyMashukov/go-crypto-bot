package strategy

import (
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"math"
	"time"
)

type SmaTradeStrategy struct {
	ExchangeRepository *repository.ExchangeRepository
}

func (s *SmaTradeStrategy) Decide(trade model.Trade) model.Decision {
	sellPeriod := 15
	buyPeriod := 60
	maxPeriod := int(math.Max(float64(sellPeriod), float64(buyPeriod)))

	list := s.ExchangeRepository.TradeList(trade.Symbol)

	if len(list) < maxPeriod {
		return model.Decision{
			StrategyName: model.SmaTradeStrategyName,
			Score:        30.00,
			Operation:    "HOLD",
			Timestamp:    time.Now().Unix(),
			Price:        trade.Price,
			Params:       [3]float64{0, 0, 0},
		}
	}

	tradeSlice := list[0:maxPeriod]

	sellSma := s.calculateSMA(tradeSlice[0:sellPeriod])
	buySma := s.calculateSMA(tradeSlice[0:buyPeriod])

	buyVolumeS, sellVolumeS := s.getBuyAndSellVolume(tradeSlice[len(tradeSlice)-sellPeriod:])
	buyVolumeB, sellVolumeB := s.getBuyAndSellVolume(tradeSlice[len(tradeSlice)-buyPeriod:])

	buyIndicator := buyVolumeB / sellVolumeB

	// todo: buy operation is disabled
	if buyIndicator > 150 && buySma < trade.Price {
		return model.Decision{
			StrategyName: model.SmaTradeStrategyName,
			Score:        50.00,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        trade.Price,
			Params:       [3]float64{buyVolumeB, sellVolumeB, buySma},
		}
	}

	sellIndicator := sellVolumeS / buyVolumeS

	if sellIndicator > 50 && sellSma > trade.Price {
		return model.Decision{
			StrategyName: model.SmaTradeStrategyName,
			Score:        50.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        trade.Price,
			Params:       [3]float64{buyVolumeS, sellVolumeS, sellSma},
		}
	}

	return model.Decision{
		StrategyName: model.SmaTradeStrategyName,
		Score:        50.00,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        trade.Price,
		Params:       [3]float64{buyVolumeS, sellVolumeS, sellSma},
	}
}

func (s *SmaTradeStrategy) calculateSMA(trades []model.Trade) float64 {
	var sum float64

	slice := trades

	for _, trade := range slice {
		sum += trade.Price
	}

	return sum / float64(len(slice))
}

func (s *SmaTradeStrategy) getBuyAndSellVolume(trades []model.Trade) (float64, float64) {
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
