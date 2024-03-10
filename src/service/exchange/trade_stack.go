package exchange

import (
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"sort"
)

type BuyOrderStackInterface interface {
	CanBuy(limit model.TradeLimit) bool
}

type TradeStack struct {
	OrderRepository    repository.OrderStorageInterface
	Binance            client.ExchangePriceAPIInterface
	ExchangeRepository repository.ExchangeRepositoryInterface
	BalanceService     BalanceServiceInterface
	Formatter          *utils.Formatter
	BotService         service.BotServiceInterface
}

func (t *TradeStack) CanBuy(limit model.TradeLimit) bool {
	// Allow to process existing order
	binanceOrder := t.OrderRepository.GetBinanceOrder(limit.Symbol, "BUY")
	if binanceOrder != nil {
		return true
	}

	manualOrder := t.OrderRepository.GetManualOrder(limit.Symbol)

	if manualOrder != nil && manualOrder.IsBuy() {
		return true
	}

	result := t.GetTradeStack(true, true, true, true, false)

	if len(result) == 0 {
		return false
	}

	return limit.Symbol == result[0].Symbol
}

func (t *TradeStack) GetTradeStack(skipLocked bool, balanceFilter bool, skipPending bool, withValidPrice bool, attachDecisions bool) []model.TradeStackItem {
	balanceUsdt, err := t.BalanceService.GetAssetBalance("USDT", true)
	stack := make([]model.TradeStackItem, 0)

	if err != nil {
		return stack
	}

	for index, tradeLimit := range t.ExchangeRepository.GetTradeLimits() {
		if !tradeLimit.IsEnabled {
			continue
		}

		isBuyLocked := t.OrderRepository.HasBuyLock(tradeLimit.Symbol)

		if skipLocked && isBuyLocked {
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
			decisions = t.ExchangeRepository.GetDecisions(tradeLimit.Symbol)
		}

		openedOrder, err := t.OrderRepository.GetOpenedOrderCached(tradeLimit.Symbol, "BUY")
		if err == nil {
			kline := t.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)
			if kline != nil && openedOrder.CanExtraBuy(*kline, t.BotService.UseSwapCapital()) {
				// todo: Add filter configuration (in TradeLimit database table) for profitPercent, example: profitPercent < 0
				// todo: Allow user to trade only after reaching specific daily price fall (or make it multi-step)
				profitPercent := openedOrder.GetProfitPercent(kline.Close, t.BotService.UseSwapCapital())
				if profitPercent.Lte(tradeLimit.GetBuyOnFallPercent(openedOrder, *kline, t.BotService.UseSwapCapital())) {
					stack = append(stack, model.TradeStackItem{
						Index:             int64(index),
						Symbol:            tradeLimit.Symbol,
						Percent:           profitPercent,
						BudgetUsdt:        openedOrder.GetAvailableExtraBudget(*kline, t.BotService.UseSwapCapital()),
						HasEnoughBalance:  false,
						BinanceOrder:      binanceOrder,
						IsExtraCharge:     true,
						Price:             lastPrice,
						IsPriceValid:      isPriceValid,
						StrategyDecisions: decisions,
						IsBuyLocked:       isBuyLocked,
					})
				}
			}
		} else {
			kLines := t.Binance.GetKLinesCached(tradeLimit.Symbol, "1d", 1)
			if len(kLines) > 0 {
				kLine := kLines[0]
				stack = append(stack, model.TradeStackItem{
					Index:             int64(index),
					Symbol:            tradeLimit.Symbol,
					Percent:           model.Percent(t.Formatter.ToFixed((t.Formatter.ComparePercentage(kLine.Open, kLine.Close) - 100.00).Value(), 2)),
					BudgetUsdt:        tradeLimit.USDTLimit,
					HasEnoughBalance:  false,
					BinanceOrder:      binanceOrder,
					IsExtraCharge:     false,
					Price:             lastPrice,
					IsPriceValid:      isPriceValid,
					StrategyDecisions: decisions,
					IsBuyLocked:       isBuyLocked,
				})
			}
		}
	}

	sort.SliceStable(stack, func(i int, j int) bool {
		return stack[i].Percent < stack[j].Percent
	})

	result := make([]model.TradeStackItem, 0)
	impossible := make([]model.TradeStackItem, 0)

	for index, stackItem := range stack {
		if stackItem.BinanceOrder != nil {
			balanceUsdt += stackItem.BinanceOrder.OrigQty * stackItem.BinanceOrder.Price
		}

		stack[index].Index = int64(index)
	}

	for _, stackItem := range stack {
		if balanceUsdt >= stackItem.BudgetUsdt {
			balanceUsdt -= stackItem.BudgetUsdt

			result = append(result, model.TradeStackItem{
				Index:             stackItem.Index,
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
				IsBuyLocked:       stackItem.IsBuyLocked,
			})
		} else {
			impossible = append(impossible, stackItem)
		}
	}

	if !balanceFilter {
		for _, stackItem := range impossible {
			balanceUsdt -= stackItem.BudgetUsdt

			result = append(result, model.TradeStackItem{
				Index:             stackItem.Index,
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
				IsBuyLocked:       stackItem.IsBuyLocked,
			})
		}
	}

	return result
}
