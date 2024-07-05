package strategy

import (
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"time"
)

type OrderBasedStrategy struct {
	ExchangeRepository repository.ExchangeTradeInfoInterface
	OrderRepository    repository.OrderStorageInterface
	ProfitService      exchange.ProfitServiceInterface
	BotService         service.BotServiceInterface
	SignalStorage      repository.SignalStorageInterface
}

func (o *OrderBasedStrategy) Decide(kLine model.KLine) model.Decision {
	tradeLimit, err := o.ExchangeRepository.GetTradeLimit(kLine.Symbol)

	if err != nil {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        0.00,
			Operation:    "HOLD",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close.Value(),
			Params:       [3]float64{0, 0, 0},
		}
	}

	order := o.OrderRepository.GetOpenedOrderCached(kLine.Symbol, "BUY")
	hasBuyOrder := order != nil

	if !hasBuyOrder {
		binanceBuyOrder := o.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "BUY")
		if binanceBuyOrder != nil {
			return model.Decision{
				StrategyName: model.OrderBasedStrategyName,
				Score:        model.DecisionHighestPriorityScore,
				Operation:    "BUY",
				Timestamp:    time.Now().Unix(),
				Price:        binanceBuyOrder.Price,
				Params:       [3]float64{0, 0, 0},
			}
		}

		manualOrder := o.OrderRepository.GetManualOrder(tradeLimit.Symbol)

		if manualOrder != nil && manualOrder.IsBuy() {
			return model.Decision{
				StrategyName: model.OrderBasedStrategyName,
				Score:        model.DecisionHighestPriorityScore,
				Operation:    "BUY",
				Timestamp:    time.Now().Unix(),
				Price:        manualOrder.Price,
				Params:       [3]float64{0, 0, 0},
			}
		}

		signal := o.SignalStorage.GetSignal(tradeLimit.Symbol)

		if signal != nil && !signal.IsExpired() {
			return model.Decision{
				StrategyName: model.OrderBasedStrategyName,
				Score:        model.DecisionHighestPriorityScore,
				Operation:    "BUY",
				Timestamp:    time.Now().Unix(),
				Price:        signal.BuyPrice,
				Params:       [3]float64{0, 0, 0},
			}
		}

		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        15.00,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close.Value(),
			Params:       [3]float64{0, 0, 0},
		}
	}

	binanceSellOrder := o.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "SELL")
	if binanceSellOrder != nil {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        model.DecisionHighestPriorityScore,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        binanceSellOrder.Price,
			Params:       [3]float64{0, 0, 0},
		}
	}

	profitPercent := order.GetProfitPercent(kLine.Close.Value(), o.BotService.UseSwapCapital())
	extraChargePercent := tradeLimit.GetBuyOnFallPercent(*order, kLine, o.BotService.UseSwapCapital())

	// ATTENTION: We can not do extra buy if CanBuy() is false
	// It can be the reason of active SELL orders, cancel SELL order when extra buy is possible
	if profitPercent.Lte(extraChargePercent) {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        model.DecisionHighestPriorityScore,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close.Value(),
			Params:       [3]float64{0, 0, 0},
		}
	}

	manualOrder := o.OrderRepository.GetManualOrder(tradeLimit.Symbol)

	if manualOrder != nil && manualOrder.IsSell() && manualOrder.CanSell(*order, o.BotService.UseSwapCapital()) {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        model.DecisionHighestPriorityScore,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        manualOrder.Price,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if profitPercent.Gte(o.ProfitService.GetMinProfitPercent(order)) {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        model.DecisionHighestPriorityScore,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close.Value(),
			Params:       [3]float64{0, 0, 0},
		}
	}

	if profitPercent.Gte(o.ProfitService.GetMinProfitPercent(order) / 2) {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        50.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close.Value(),
			Params:       [3]float64{0, 0, 0},
		}
	}

	if kLine.Close.Value() > order.Price {
		sellPrice := o.ProfitService.GetMinClosePrice(order, order.Price)

		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        30.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        sellPrice,
			Params:       [3]float64{0, 0, 0},
		}
	}

	return model.Decision{
		StrategyName: model.OrderBasedStrategyName,
		Score:        99.99,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        kLine.Close.Value(),
		Params:       [3]float64{0, 0, 0},
	}
}
