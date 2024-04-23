package exchange

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"sort"
	"time"
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
	PriceCalculator    PriceCalculatorInterface
	RDB                *redis.Client
	Ctx                *context.Context
	TradeFilterService TradeFilterServiceInterface
}

type TradeStackParams struct {
	SkipFiltered    bool
	SkipDisabled    bool
	SkipLocked      bool
	BalanceFilter   bool
	SkipPending     bool
	WithValidPrice  bool
	AttachDecisions bool
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

	result := t.GetTradeStack(TradeStackParams{
		SkipFiltered:    true,
		SkipLocked:      true,
		SkipDisabled:    true,
		BalanceFilter:   true,
		SkipPending:     true,
		WithValidPrice:  true,
		AttachDecisions: false,
	})

	if len(result) == 0 {
		return false
	}

	return limit.Symbol == result[0].Symbol
}

func (t *TradeStack) GetTradeStack(params TradeStackParams) []model.TradeStackItem {
	balanceUsdt, err := t.BalanceService.GetAssetBalance("USDT", true)
	stack := make([]model.TradeStackItem, 0)

	if err != nil {
		return stack
	}

	resultChannel := make(chan *model.TradeStackItem)
	defer close(resultChannel)
	count := 0

	for index, tradeLimit := range t.ExchangeRepository.GetTradeLimits() {
		go func(limit model.TradeLimit, index int64, params TradeStackParams) {
			resultChannel <- t.ProcessItem(
				index,
				limit,
				params,
			)
		}(tradeLimit, int64(index), params)
		count++
	}

	processed := 0

	for {
		stackItem := <-resultChannel
		if stackItem != nil {
			stack = append(stack, *stackItem)
		}
		processed++

		if processed == count {
			break
		}
	}

	switch t.BotService.GetTradeStackSorting() {
	case model.TradeStackSortingLessPercent:
		sort.SliceStable(stack, func(i int, j int) bool {
			return stack[i].Percent < stack[j].Percent
		})
		break
	case model.TradeStackSortingLessPriceDiff:
		sort.SliceStable(stack, func(i int, j int) bool {
			return stack[i].PricePointsDiff > stack[j].PricePointsDiff
		})
		break
	}

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
				Index:                   stackItem.Index,
				Symbol:                  stackItem.Symbol,
				Percent:                 stackItem.Percent,
				BudgetUsdt:              stackItem.BudgetUsdt,
				HasEnoughBalance:        true,
				BalanceAfter:            balanceUsdt,
				BinanceOrder:            stackItem.BinanceOrder,
				IsExtraCharge:           stackItem.IsExtraCharge,
				IsPriceValid:            stackItem.IsPriceValid,
				Price:                   stackItem.Price,
				StrategyDecisions:       stackItem.StrategyDecisions,
				IsBuyLocked:             stackItem.IsBuyLocked,
				IsEnabled:               stackItem.IsEnabled,
				BuyPrice:                stackItem.BuyPrice,
				PricePointsDiff:         stackItem.PricePointsDiff,
				IsFiltered:              stackItem.IsFiltered,
				TradeFiltersBuy:         stackItem.TradeFiltersBuy,
				TradeFiltersSell:        stackItem.TradeFiltersSell,
				TradeFiltersExtraCharge: stackItem.TradeFiltersExtraCharge,
				PriceChangeSpeedAvg:     stackItem.PriceChangeSpeedAvg,
				Capitalization:          stackItem.Capitalization,
			})
		} else {
			impossible = append(impossible, stackItem)
		}
	}

	if !params.BalanceFilter {
		for _, stackItem := range impossible {
			balanceUsdt -= stackItem.BudgetUsdt

			result = append(result, model.TradeStackItem{
				Index:                   stackItem.Index,
				Symbol:                  stackItem.Symbol,
				Percent:                 stackItem.Percent,
				BudgetUsdt:              stackItem.BudgetUsdt,
				HasEnoughBalance:        false,
				BalanceAfter:            balanceUsdt,
				BinanceOrder:            stackItem.BinanceOrder,
				IsExtraCharge:           stackItem.IsExtraCharge,
				IsPriceValid:            stackItem.IsPriceValid,
				Price:                   stackItem.Price,
				StrategyDecisions:       stackItem.StrategyDecisions,
				IsBuyLocked:             stackItem.IsBuyLocked,
				IsEnabled:               stackItem.IsEnabled,
				BuyPrice:                stackItem.BuyPrice,
				PricePointsDiff:         stackItem.PricePointsDiff,
				IsFiltered:              stackItem.IsFiltered,
				TradeFiltersBuy:         stackItem.TradeFiltersBuy,
				TradeFiltersSell:        stackItem.TradeFiltersSell,
				TradeFiltersExtraCharge: stackItem.TradeFiltersExtraCharge,
				PriceChangeSpeedAvg:     stackItem.PriceChangeSpeedAvg,
				Capitalization:          stackItem.Capitalization,
			})
		}
	}

	return result
}

func (t *TradeStack) ProcessItem(
	index int64,
	tradeLimit model.TradeLimit,
	params TradeStackParams,
) *model.TradeStackItem {
	isEnabled := tradeLimit.IsEnabled

	if !isEnabled && params.SkipDisabled {
		return nil
	}

	isBuyLocked := t.OrderRepository.HasBuyLock(tradeLimit.Symbol)

	if params.SkipLocked && isBuyLocked {
		return nil
	}

	lastKLine := t.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)
	lastPrice := 0.00
	isPriceValid := false
	priceChangeSpeedAvg := 0.00

	if lastKLine != nil && !lastKLine.IsPriceExpired() {
		isPriceValid = true
	}

	if lastKLine != nil {
		lastPrice = lastKLine.Close
		priceChangeSpeedAvg = t.Formatter.ToFixed(lastKLine.GetPriceChangeSpeedAvg(), 2)
	}

	if !isPriceValid && params.WithValidPrice {
		return nil
	}

	// Skip if order has already opened
	binanceOrder := t.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "BUY")
	if binanceOrder != nil && params.SkipPending {
		return nil
	}

	decisions := make([]model.Decision, 0)
	if params.AttachDecisions {
		decisions = t.ExchangeRepository.GetDecisions(tradeLimit.Symbol)
	}

	openedOrder, err := t.OrderRepository.GetOpenedOrderCached(tradeLimit.Symbol, "BUY")

	isFiltered := false
	if err != nil {
		isFiltered = !t.TradeFilterService.CanBuy(tradeLimit)
	} else {
		isFiltered = !t.TradeFilterService.CanExtraBuy(tradeLimit)
	}

	if isFiltered && params.SkipFiltered {
		return nil
	}

	buyPrice := 0.00
	capitalization := model.Capitalization{
		Capitalization: 0.00,
		MarketPrice:    0.00,
	}

	if binanceOrder != nil {
		buyPrice = binanceOrder.Price
	} else if lastKLine != nil {
		if !lastKLine.IsPriceExpired() {
			buyPrice = t.GetBuyPriceCached(tradeLimit)

			capitalizationValue := t.ExchangeRepository.GetCapitalization(tradeLimit.Symbol, lastKLine.Timestamp)
			if capitalizationValue != nil {
				capitalization = model.Capitalization{
					Capitalization: t.Formatter.ToFixed(capitalizationValue.Capitalization, 2),
					MarketPrice:    t.Formatter.FormatPrice(tradeLimit, capitalizationValue.Price),
				}
			}
		}
	}

	pricePointsDiff := int64((t.Formatter.FormatPrice(tradeLimit, buyPrice) - t.Formatter.FormatPrice(tradeLimit, lastPrice)) / tradeLimit.MinPrice)

	if err == nil {
		kline := t.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)
		if kline != nil && openedOrder.CanExtraBuy(*kline, t.BotService.UseSwapCapital()) {
			// todo: Add filter configuration (in TradeLimit database table) for profitPercent, example: profitPercent < 0
			// todo: Allow user to trade only after reaching specific daily price fall (or make it multi-step)
			profitPercent := openedOrder.GetProfitPercent(kline.Close, t.BotService.UseSwapCapital())
			if profitPercent.Lte(tradeLimit.GetBuyOnFallPercent(openedOrder, *kline, t.BotService.UseSwapCapital())) {
				return &model.TradeStackItem{
					Index:                   index,
					Symbol:                  tradeLimit.Symbol,
					Percent:                 profitPercent,
					BudgetUsdt:              openedOrder.GetAvailableExtraBudget(*kline, t.BotService.UseSwapCapital()),
					HasEnoughBalance:        false,
					BinanceOrder:            binanceOrder,
					IsExtraCharge:           true,
					Price:                   lastPrice,
					IsPriceValid:            isPriceValid,
					StrategyDecisions:       decisions,
					IsBuyLocked:             isBuyLocked,
					IsEnabled:               isEnabled,
					BuyPrice:                buyPrice,
					PricePointsDiff:         pricePointsDiff,
					IsFiltered:              isFiltered,
					TradeFiltersBuy:         tradeLimit.TradeFiltersBuy,
					TradeFiltersSell:        tradeLimit.TradeFiltersSell,
					TradeFiltersExtraCharge: tradeLimit.TradeFiltersExtraCharge,
					PriceChangeSpeedAvg:     priceChangeSpeedAvg,
					Capitalization:          capitalization,
				}
			}
		}

		return nil
	}

	kLines := t.Binance.GetKLinesCached(tradeLimit.Symbol, "1d", 1)
	if len(kLines) > 0 {
		kLine := kLines[0]
		return &model.TradeStackItem{
			Index:                   index,
			Symbol:                  tradeLimit.Symbol,
			Percent:                 model.Percent(t.Formatter.ToFixed((t.Formatter.ComparePercentage(kLine.Open, kLine.Close) - 100.00).Value(), 2)),
			BudgetUsdt:              tradeLimit.USDTLimit,
			HasEnoughBalance:        false,
			BinanceOrder:            binanceOrder,
			IsExtraCharge:           false,
			Price:                   lastPrice,
			IsPriceValid:            isPriceValid,
			StrategyDecisions:       decisions,
			IsBuyLocked:             isBuyLocked,
			IsEnabled:               isEnabled,
			BuyPrice:                buyPrice,
			PricePointsDiff:         pricePointsDiff,
			IsFiltered:              isFiltered,
			TradeFiltersBuy:         tradeLimit.TradeFiltersBuy,
			TradeFiltersSell:        tradeLimit.TradeFiltersSell,
			TradeFiltersExtraCharge: tradeLimit.TradeFiltersExtraCharge,
			PriceChangeSpeedAvg:     priceChangeSpeedAvg,
			Capitalization:          capitalization,
		}
	}

	return nil
}

func (t *TradeStack) GetBuyPriceCached(limit model.TradeLimit) float64 {
	var ticker model.WSTickerPrice

	// todo: invalidate on limit changes...
	cacheKey := fmt.Sprintf("buy-price-cached-%s-%d", limit.Symbol, t.BotService.GetBot().Id)
	buyPriceCached := t.RDB.Get(*t.Ctx, cacheKey).Val()

	if len(buyPriceCached) > 0 {
		err := json.Unmarshal([]byte(buyPriceCached), &ticker)
		if err == nil {
			return ticker.Price
		}
	}

	buyPriceCalc, buyPriceErr := t.PriceCalculator.CalculateBuy(limit)
	if buyPriceErr == nil {
		ticker = model.WSTickerPrice{
			Symbol: limit.Symbol,
			Price:  buyPriceCalc,
		}

		encoded, err := json.Marshal(ticker)
		if err == nil {
			t.RDB.Set(*t.Ctx, cacheKey, string(encoded), time.Minute)
		}

		return ticker.Price
	}

	return 0.00
}

func (t *TradeStack) GetBuyPricePoints(kLine model.KLine, tradeLimit model.TradeLimit) (int64, error) {
	if kLine.IsPriceExpired() {
		return 0, errors.New(fmt.Sprintf("[%s] current price is expired", tradeLimit.Symbol))
	}

	buyPrice := t.GetBuyPriceCached(tradeLimit)

	if buyPrice == 0.00 {
		return 0, errors.New(fmt.Sprintf("[%s] buy price is invalid", tradeLimit.Symbol))
	}

	return int64((t.Formatter.FormatPrice(tradeLimit, buyPrice) - t.Formatter.FormatPrice(tradeLimit, kLine.Close)) / tradeLimit.MinPrice), nil
}
