package service

import (
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"time"
)

type BaseKLineStrategy struct {
}

func (k *BaseKLineStrategy) Decide(kLine ExchangeModel.KLine) ExchangeModel.Decision {
	if kLine.IsPositive() && kLine.Close < (kLine.High+kLine.Open)/2 {
		return ExchangeModel.Decision{
			StrategyName: "base_kline_strategy",
			Score:        25.00,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if kLine.IsNegative() {
		return ExchangeModel.Decision{
			StrategyName: "base_kline_strategy",
			Score:        25.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	return ExchangeModel.Decision{
		StrategyName: "base_kline_strategy",
		Score:        25.00,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        kLine.Close,
		Params:       [3]float64{0, 0, 0},
	}
}
