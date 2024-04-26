package strategy

import (
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"time"
)

type MarketDepthStrategy struct {
}

func (m *MarketDepthStrategy) Decide(depth model.OrderBookModel) model.Decision {
	sellVolume := depth.GetAskVolume()
	buyVolume := depth.GetBidVolume()

	sellBuyDiff := sellVolume / buyVolume

	if sellBuyDiff > 10 {
		return model.Decision{
			StrategyName: model.MarketDepthStrategyName,
			Score:        30.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        depth.GetBestAsk(),
			Params:       [3]float64{buyVolume, sellVolume, 0},
		}
	}

	buySellDiff := buyVolume / sellVolume

	// todo: buy operation is disabled
	if buySellDiff > 10 {
		return model.Decision{
			StrategyName: model.MarketDepthStrategyName,
			Score:        30.00,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        depth.GetBestBid(),
			Params:       [3]float64{buyVolume, sellVolume, 0},
		}
	}

	return model.Decision{
		StrategyName: model.MarketDepthStrategyName,
		Score:        30.00,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        depth.GetBestBid(),
		Params:       [3]float64{buyVolume, sellVolume, 0},
	}
}
