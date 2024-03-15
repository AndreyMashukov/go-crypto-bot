package strategy

import (
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"time"
)

type BaseKLineStrategy struct {
	TradeStack         *exchange.TradeStack
	ExchangeRepository *repository.ExchangeRepository
	OrderRepository    *repository.OrderRepository
	Formatter          *utils.Formatter
	MlEnabled          bool
}

func (k *BaseKLineStrategy) Decide(kLine model.KLine) model.Decision {
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

	// todo: Add BTC/ETH interpolation...
	// todo: Cover By UNIT test!
	isPositivePredict := false

	if k.MlEnabled {
		predict, predictErr := k.ExchangeRepository.GetPredict(kLine.Symbol)

		if predictErr == nil && predict > kLine.Close {
			isPositivePredict = true
		}
	}

	if kLine.IsNegative() {
		if k.MlEnabled && isPositivePredict {
			_, orderErr := k.OrderRepository.GetOpenedOrderCached(kLine.Symbol, "BUY")

			if orderErr != nil {
				tradeLimit, err := k.ExchangeRepository.GetTradeLimit(kLine.Symbol)

				if err == nil {
					points, err := k.TradeStack.GetBuyPricePoints(kLine, tradeLimit)
					if err == nil {
						switch true {
						case points >= 50:
							return model.Decision{
								StrategyName: model.BaseKlineStrategyName,
								Score:        float64(points),
								Operation:    "BUY",
								Timestamp:    time.Now().Unix(),
								Price:        kLine.Close,
								Params:       [3]float64{0, 0, 0},
							}
						case points >= -5:
							return model.Decision{
								StrategyName: model.BaseKlineStrategyName,
								Score:        20,
								Operation:    "BUY",
								Timestamp:    time.Now().Unix(),
								Price:        kLine.Close,
								Params:       [3]float64{0, 0, 0},
							}
						}
					}
				}
			}
		}

		return model.Decision{
			StrategyName: model.BaseKlineStrategyName,
			Score:        25.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if k.MlEnabled && isPositivePredict {
		return model.Decision{
			StrategyName: model.BaseKlineStrategyName,
			Score:        50.00,
			Operation:    "BUY",
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
