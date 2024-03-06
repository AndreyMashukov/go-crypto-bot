package strategy

import (
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"time"
)

type OrderBasedStrategy struct {
	ExchangeRepository repository.ExchangeRepository
	OrderRepository    repository.OrderRepository
	ProfitService      exchange.ProfitServiceInterface
	TradeStack         *exchange.TradeStack
	BotService         service.BotServiceInterface
}

func (o *OrderBasedStrategy) Decide(kLine model.KLine) model.Decision {
	tradeLimit, err := o.ExchangeRepository.GetTradeLimit(kLine.Symbol)

	if err != nil {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        0.00,
			Operation:    "HOLD",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	binanceBuyOrder := o.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "BUY")
	if binanceBuyOrder != nil {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        999.99,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	order, err := o.OrderRepository.GetOpenedOrderCached(kLine.Symbol, "BUY")

	if err != nil {
		if !o.TradeStack.CanBuy(tradeLimit) {
			return model.Decision{
				StrategyName: model.OrderBasedStrategyName,
				Score:        80.00,
				Operation:    "HOLD",
				Timestamp:    time.Now().Unix(),
				Price:        kLine.Close,
				Params:       [3]float64{0, 0, 0},
			}
		}

		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        15.00,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	binanceSellOrder := o.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "SELL")
	if binanceSellOrder != nil {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        999.99,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	profitPercent := order.GetProfitPercent(kLine.Close, o.BotService.UseSwapCapital())

	if profitPercent.Lte(tradeLimit.GetBuyOnFallPercent(order, kLine, o.BotService.UseSwapCapital())) && tradeLimit.IsEnabled && o.TradeStack.CanBuy(tradeLimit) {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        999.99,
			Operation:    "BUY",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if profitPercent.Gte(o.ProfitService.GetMinProfitPercent(order)) {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        999.99,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if profitPercent.Gte(o.ProfitService.GetMinProfitPercent(order) / 2) {
		return model.Decision{
			StrategyName: model.OrderBasedStrategyName,
			Score:        50.00,
			Operation:    "SELL",
			Timestamp:    time.Now().Unix(),
			Price:        kLine.Close,
			Params:       [3]float64{0, 0, 0},
		}
	}

	if kLine.Close > order.Price {
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
		Score:        99.00,
		Operation:    "HOLD",
		Timestamp:    time.Now().Unix(),
		Price:        kLine.Close,
		Params:       [3]float64{0, 0, 0},
	}
}
