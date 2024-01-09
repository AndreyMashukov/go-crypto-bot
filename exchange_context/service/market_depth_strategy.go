package service

import (
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"time"
)

type MarketDepthStrategy struct {
}

func (m *MarketDepthStrategy) Decide(depth ExchangeModel.Depth) ExchangeModel.Decision {
	sellVolume := depth.GetAskVolume()
	buyVolume := depth.GetBidVolume()

	sellBuyDiff := sellVolume / buyVolume

	if sellBuyDiff > 10 {
		return ExchangeModel.Decision{
			StrategyName: "market_depth_strategy",
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
		return ExchangeModel.Decision{
			StrategyName: "market_depth_strategy",
			Score:        30.00,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        depth.GetBestBid(),
			Params:       [3]float64{buyVolume, sellVolume, 0},
		}
	}

	return ExchangeModel.Decision{
		StrategyName: "market_depth_strategy",
		Score:        30.00,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        depth.GetBestBid(),
		Params:       [3]float64{buyVolume, sellVolume, 0},
	}
}
