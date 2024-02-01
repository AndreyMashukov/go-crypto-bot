package service

import (
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"sort"
)

type TradeStack struct {
	OrderRepository    repository.OrderStorageInterface
	Binance            client.ExchangePriceAPIInterface
	ExchangeRepository repository.ExchangeRepositoryInterface
	BalanceService     BalanceServiceInterface
	Formatter          *Formatter
}

func (t *TradeStack) CanBuy(limit model.TradeLimit) bool {
	// Allow to process existing order
	binanceOrder := t.OrderRepository.GetBinanceOrder(limit.Symbol, "BUY")
	if binanceOrder != nil {
		return true
	}

	result := t.GetTradeStack(true, true, true, false)

	if len(result) == 0 {
		return false
	}

	for index, stackItem := range result {
		log.Printf("Stack [%d] %s = %.2f", index, stackItem.Symbol, stackItem.Percent)
		if index >= 1 {
			break
		}
	}

	return limit.Symbol == result[0].Symbol
}

func (t *TradeStack) GetTradeStack(balanceFilter bool, skipPending bool, withValidPrice bool, attachDecisions bool) []model.TradeStackItem {
	balanceUsdt, err := t.BalanceService.GetAssetBalance("USDT", true)
	stack := make([]model.TradeStackItem, 0)

	if err != nil {
		return stack
	}

	for _, tradeLimit := range t.ExchangeRepository.GetTradeLimits() {
		if !tradeLimit.IsEnabled {
			continue
		}

		lastKLine := t.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)
		lastPrice := 0.00
		isPriceValid := false

		if lastKLine != nil && !lastKLine.IsPriceExpired() {
			isPriceValid = true
		}

		if lastKLine != nil {
			lastPrice = lastKLine.Close
		}

		if !isPriceValid && withValidPrice {
			continue
		}

		// Skip if order has already opened
		binanceOrder := t.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "BUY")
		if binanceOrder != nil && skipPending {
			continue
		}

		decisions := make([]model.Decision, 0)
		if attachDecisions {
			decisions = t.ExchangeRepository.GetDecisions()
		}

		openedOrder, err := t.OrderRepository.GetOpenedOrderCached(tradeLimit.Symbol, "BUY")
		if err == nil {
			kline := t.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)
			if kline != nil && tradeLimit.IsExtraChargeEnabled() && openedOrder.CanExtraBuy(tradeLimit) {
				profitPercent := openedOrder.GetProfitPercent(kline.Close)
				if profitPercent.Lte(tradeLimit.GetBuyOnFallPercent()) {
					stack = append(stack, model.TradeStackItem{
						Symbol:            tradeLimit.Symbol,
						Percent:           profitPercent,
						BudgetUsdt:        tradeLimit.USDTExtraBudget - openedOrder.UsedExtraBudget,
						HasEnoughBalance:  false,
						BinanceOrder:      binanceOrder,
						IsExtraCharge:     true,
						Price:             lastPrice,
						IsPriceValid:      isPriceValid,
						StrategyDecisions: decisions,
					})
				}
			}
		} else {
			kLines := t.Binance.GetKLinesCached(tradeLimit.Symbol, "1d", 1)
			if len(kLines) > 0 {
				kLine := kLines[0]
				stack = append(stack, model.TradeStackItem{
					Symbol:            tradeLimit.Symbol,
					Percent:           model.Percent(t.Formatter.ToFixed((t.Formatter.ComparePercentage(kLine.Open, kLine.Close) - 100.00).Value(), 2)),
					BudgetUsdt:        tradeLimit.USDTLimit,
					HasEnoughBalance:  false,
					BinanceOrder:      binanceOrder,
					IsExtraCharge:     false,
					Price:             lastPrice,
					IsPriceValid:      isPriceValid,
					StrategyDecisions: decisions,
				})
			}
		}
	}

	sort.SliceStable(stack, func(i int, j int) bool {
		return stack[i].Percent < stack[j].Percent
	})

	result := make([]model.TradeStackItem, 0)

	for _, stackItem := range stack {
		if balanceFilter {
			if balanceUsdt <= 0.00 {
				break
			}
		} else {
			if balanceUsdt < stackItem.BudgetUsdt {
				if stackItem.BinanceOrder == nil {
					balanceUsdt -= stackItem.BudgetUsdt
				}

				result = append(result, model.TradeStackItem{
					Symbol:            stackItem.Symbol,
					Percent:           stackItem.Percent,
					BudgetUsdt:        stackItem.BudgetUsdt,
					HasEnoughBalance:  false,
					BalanceAfter:      balanceUsdt,
					BinanceOrder:      stackItem.BinanceOrder,
					IsExtraCharge:     stackItem.IsExtraCharge,
					IsPriceValid:      stackItem.IsPriceValid,
					Price:             stackItem.Price,
					StrategyDecisions: stackItem.StrategyDecisions,
				})

				continue
			}
		}

		if balanceUsdt >= stackItem.BudgetUsdt {
			if stackItem.BinanceOrder == nil {
				balanceUsdt -= stackItem.BudgetUsdt
			}

			result = append(result, model.TradeStackItem{
				Symbol:            stackItem.Symbol,
				Percent:           stackItem.Percent,
				BudgetUsdt:        stackItem.BudgetUsdt,
				HasEnoughBalance:  true,
				BalanceAfter:      balanceUsdt,
				BinanceOrder:      stackItem.BinanceOrder,
				IsExtraCharge:     stackItem.IsExtraCharge,
				IsPriceValid:      stackItem.IsPriceValid,
				Price:             stackItem.Price,
				StrategyDecisions: stackItem.StrategyDecisions,
			})
		}
	}

	return result
}
