package service

import (
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"time"
)

type BaseKLineStrategy struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Formatter          *Formatter
}

func (k *BaseKLineStrategy) Decide(kLine ExchangeModel.KLine) ExchangeModel.Decision {
	predict, _ := k.ExchangeRepository.GetPredict(kLine.Symbol)

	if kLine.IsPositive() && predict > 0.00 && k.Formatter.ComparePercentage(kLine.Close, predict).Lte(99.5) {
		return ExchangeModel.Decision{
			StrategyName: "base_kline_strategy",
			Score:        50.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

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

	if kLine.IsNegative() && predict > 0.00 && k.Formatter.ComparePercentage(kLine.Close, predict).Gt(100.5) {
		return ExchangeModel.Decision{
			StrategyName: "base_kline_strategy",
			Score:        50.00,
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
