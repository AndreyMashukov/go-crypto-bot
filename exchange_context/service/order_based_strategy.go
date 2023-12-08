package service

import (
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"time"
)

type OrderBasedStrategy struct {
	ExchangeRepository ExchangeRepository.ExchangeRepository
	OrderRepository    ExchangeRepository.OrderRepository
}

func (o *OrderBasedStrategy) Decide(kLine ExchangeModel.KLine) ExchangeModel.Decision {
	order, err := o.OrderRepository.GetOpenedOrderCached(kLine.Symbol, "BUY")

	if err != nil {
		return ExchangeModel.Decision{
			StrategyName: "order_based_strategy",
			Score:        0.00,
			Operation:    "HOLD",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	tradeLimit, err := o.ExchangeRepository.GetTradeLimit(order.Symbol)

	if err != nil {
		return ExchangeModel.Decision{
			StrategyName: "order_based_strategy",
			Score:        0.00,
			Operation:    "HOLD",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	profitPercent := order.GetProfitPercent(kLine.Close)

	if profitPercent.Gte(tradeLimit.GetMinProfitPercent()) {
		return ExchangeModel.Decision{
			StrategyName: "order_based_strategy",
			Score:        30.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	periodMinPrice := o.ExchangeRepository.GetPeriodMinPrice(kLine.Symbol, 200)

	// If time to extra buy and price is near Low (Low + 0.5%)
	if tradeLimit.IsExtraChargeEnabled() && profitPercent.Lte(tradeLimit.GetBuyOnFallPercent()) && kLine.Close <= kLine.GetLowPercent(0.5) {
		return ExchangeModel.Decision{
			StrategyName: "order_based_strategy",
			Score:        999.99,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        periodMinPrice,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if kLine.Close > order.Price {
		sellPrice := order.GetMinClosePrice(tradeLimit)

		return ExchangeModel.Decision{
			StrategyName: "order_based_strategy",
			Score:        30.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        sellPrice,
			Params:       [3]float64{0, 0, 0},
		}
	}

	return ExchangeModel.Decision{
		StrategyName: "order_based_strategy",
		Score:        99.00,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        kLine.Close,
		Params:       [3]float64{0, 0, 0},
	}
}
