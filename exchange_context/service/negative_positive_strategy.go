package exchange_context

import (
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"time"
)

type NegativePositiveStrategy struct {
	LastKline map[string]ExchangeModel.KLine
}

func (n *NegativePositiveStrategy) Decide(kLine ExchangeModel.KLine) ExchangeModel.Decision {
	lastKLine, hasKline := n.LastKline[kLine.Symbol]
	if !hasKline || lastKLine.Timestamp.Time.Unix() != kLine.Timestamp.Time.Unix() {
		n.LastKline[kLine.Symbol] = kLine
	}

	if !hasKline {
		return ExchangeModel.Decision{
			StrategyName: "negative_positive_strategy",
			Score:        33.33,
			Operation:    "HOLD",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if lastKLine.IsNegative() && kLine.IsPositive() {
		return ExchangeModel.Decision{
			StrategyName: "negative_positive_strategy",
			Score:        33.33,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if lastKLine.IsPositive() && kLine.IsNegative() {
		return ExchangeModel.Decision{
			StrategyName: "negative_positive_strategy",
			Score:        33.33,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	return ExchangeModel.Decision{
		StrategyName: "negative_positive_strategy",
		Score:        33.33,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        kLine.Close,
		Params:       [3]float64{0, 0, 0},
	}
}
