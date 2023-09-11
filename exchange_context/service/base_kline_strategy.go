package exchange_context

import (
	ExchangeModel "github.com/AndreyMashukov/go-crypto-bot/server/exchange_context/model"
	"time"
)

type BaseKLineStrategy struct {
}

func (k *BaseKLineStrategy) Decide(kLine ExchangeModel.KLine) ExchangeModel.Decision {
	if kLine.Close > kLine.Open && kLine.Close < (kLine.High+kLine.Open)/2 {
		return ExchangeModel.Decision{
			StrategyName: "base_kline_strategy",
			Score:        50.00,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if kLine.Close < kLine.Open {
		return ExchangeModel.Decision{
			StrategyName: "base_kline_strategy",
			Score:        50.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	return ExchangeModel.Decision{
		StrategyName: "base_kline_strategy",
		Score:        50.00,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        kLine.Close,
		Params:       [3]float64{0, 0, 0},
	}
}
