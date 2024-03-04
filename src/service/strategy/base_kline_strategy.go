package strategy

import (
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"time"
)

type BaseKLineStrategy struct {
	ExchangeRepository *repository.ExchangeRepository
	Formatter          *utils.Formatter
	MlEnabled          bool
}

func (k *BaseKLineStrategy) Decide(kLine model.KLine) model.Decision {
	//if kLine.IsPositive() && predict > 0.00 && k.Formatter.ComparePercentage(kLine.Close, predict).Lte(99.5) {
	//	return model.Decision{
	//		StrategyName: model.BaseKlineStrategyName,
	//		Score:        50.00,
	//		Operation:    "SELL",
	//		Timestamp:    time.Now().Unix(),
	//		Price:        kLine.Close,
	//		Params:       [3]float64{0, 0, 0},
	//	}
	//}

	// todo: buy operation is disabled
	if kLine.IsPositive() && kLine.Close < (kLine.High+kLine.Open)/2 {
		return model.Decision{
			StrategyName: model.BaseKlineStrategyName,
			Score:        25.00,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if k.MlEnabled {
		predict, predictErr := k.ExchangeRepository.GetPredict(kLine.Symbol)

		if predictErr == nil && predict > kLine.Close {
			return model.Decision{
				StrategyName: model.BaseKlineStrategyName,
				Score:        50.00,
				Operation:    "BUY",
				Timestamp:    time.Now().Unix(),
				Price:        kLine.Close,
				Params:       [3]float64{0, 0, 0},
			}
		}
	}

	if kLine.IsNegative() {
		return model.Decision{
			StrategyName: model.BaseKlineStrategyName,
			Score:        25.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	return model.Decision{
		StrategyName: model.BaseKlineStrategyName,
		Score:        25.00,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        kLine.Close,
		Params:       [3]float64{0, 0, 0},
	}
}
