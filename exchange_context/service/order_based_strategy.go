package service

import (
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"math"
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

	diff := kLine.Close - order.Price
	profitPercent := math.Round(diff*100/order.Price*100) / 100

	if profitPercent >= tradeLimit.MinProfitPercent {
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

	if tradeLimit.BuyOnFallPercent != 0.00 && profitPercent <= tradeLimit.BuyOnFallPercent && periodMinPrice != 0.00 && kLine.Close <= periodMinPrice {
		return ExchangeModel.Decision{
			StrategyName: "order_based_strategy",
			Score:        30.00,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if kLine.Close > order.Price {
		sellPrice := order.Price + (order.Price * tradeLimit.MinProfitPercent / 100)

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
		Score:        0.00,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        kLine.Close,
		Params:       [3]float64{0, 0, 0},
	}
}
