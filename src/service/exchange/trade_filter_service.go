package exchange

import (
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"strconv"
	"strings"
)

type TradeFilterServiceInterface interface {
	CanBuy(limit model.TradeLimit) bool
	CanExtraBuy(limit model.TradeLimit) bool
	CanSell(limit model.TradeLimit) bool
}

type TradeFilterService struct {
	OrderRepository   repository.OrderStorageInterface
	ExchangeTradeInfo repository.ExchangeTradeInfoInterface
	ExchangePriceAPI  client.ExchangePriceAPIInterface
	Formatter         *utils.Formatter
	SignalStorage     repository.SignalStorageInterface
}

func (t *TradeFilterService) CanBuy(limit model.TradeLimit) bool {
	if len(limit.TradeFiltersBuy) == 0 {
		return true
	}

	return t.IsFilterMatched(limit.TradeFiltersBuy)
}

func (t *TradeFilterService) CanExtraBuy(limit model.TradeLimit) bool {
	if len(limit.TradeFiltersExtraCharge) == 0 {
		return true
	}

	return t.IsFilterMatched(limit.TradeFiltersExtraCharge)
}

func (t *TradeFilterService) CanSell(limit model.TradeLimit) bool {
	if len(limit.TradeFiltersSell) == 0 {
		return true
	}

	return t.IsFilterMatched(limit.TradeFiltersSell)
}

func (t *TradeFilterService) IsFilterMatched(filters []model.TradeFilter) bool {
	matchedAnd := 0
	matchedOr := 0

	for _, filter := range filters {
		matched := false
		if len(filter.Children) > 0 {
			matched = t.IsFilterMatched(filter.Children)
		} else {
			matched = t.IsValueMatched(filter)
		}

		if filter.And() && matched {
			matchedAnd++
		}

		if filter.Or() && matched {
			matchedOr++
		}
	}

	if matchedAnd == len(filters) {
		return true
	}

	if matchedOr > 0 {
		return true
	}

	// todo: return not matched filters...???
	return false
}

func (t *TradeFilterService) IsValueMatched(filter model.TradeFilter) bool {
	matched := false

	switch filter.Parameter {
	case model.TradeFilterParameterHasSignal:
		signal := t.SignalStorage.GetSignal(filter.Symbol)
		matched = t.CompareBool(
			signal != nil && !signal.IsExpired(),
			filter,
		)
		break
	case model.TradeFilterParameterPrice:
		kline := t.ExchangeTradeInfo.GetCurrentKline(filter.Symbol)
		if kline != nil {
			matched = t.CompareFloat(kline.Close, filter)
		}
		break
	case model.TradeFilterParameterPositionTimeMinutes:
		opened := t.OrderRepository.GetOpenedOrderCached(filter.Symbol, "BUY")
		if opened != nil {
			matched = t.CompareFloat(opened.GetPositionTime().GetMinutes(), filter)
		}
		break
	case model.TradeFilterParameterExtraOrdersToday:
		extraMap := t.OrderRepository.GetTodayExtraOrderMap()
		rawValue, _ := extraMap.LoadOrStore(filter.Symbol, float64(0.00))
		if value, ok := rawValue.(float64); ok {
			matched = t.CompareFloat(value, filter)
		}
		break
	case model.TradeFilterParameterDailyPercent:
		kLines := t.ExchangePriceAPI.GetKLinesCached(filter.Symbol, "1d", 1)
		if len(kLines) > 0 {
			kLine := kLines[0]
			percent := model.Percent(t.Formatter.ToFixed((t.Formatter.ComparePercentage(kLine.Open, kLine.Close) - 100.00).Value(), 2))
			matched = t.CompareFloat(percent.Value(), filter)
		}
		break
	case model.TradeFilterParameterSentimentScore:
		tradeLimit := t.ExchangeTradeInfo.GetTradeLimitCached(filter.Symbol)
		if tradeLimit != nil && tradeLimit.SentimentScore != nil {
			score := *tradeLimit.SentimentScore
			matched = t.CompareFloat(score, filter)
		}

		break
	case model.TradeFilterParameterSentimentLabel:
		tradeLimit := t.ExchangeTradeInfo.GetTradeLimitCached(filter.Symbol)
		if tradeLimit != nil && tradeLimit.SentimentLabel != nil {
			label := *tradeLimit.SentimentLabel
			matched = t.CompareString(label, filter)
		}
		break
	}

	return matched
}

func (t *TradeFilterService) CompareString(parameterValue string, filter model.TradeFilter) bool {
	matched := false

	stringValue := strings.ToUpper(filter.Value)
	switch filter.Condition {
	case model.TradeFilterConditionEq:
		matched = stringValue == strings.ToUpper(parameterValue)
		break
	case model.TradeFilterConditionNeq:
		matched = stringValue != strings.ToUpper(parameterValue)
		break
	}

	return matched
}

func (t *TradeFilterService) CompareFloat(parameterValue float64, filter model.TradeFilter) bool {
	matched := false

	value, err := strconv.ParseFloat(filter.Value, 64)
	if err == nil {
		switch filter.Condition {
		case model.TradeFilterConditionEq:
			matched = parameterValue == value
			break
		case model.TradeFilterConditionNeq:
			matched = parameterValue != value
			break
		case model.TradeFilterConditionGt:
			matched = parameterValue > value
			break
		case model.TradeFilterConditionGte:
			matched = parameterValue >= value
			break
		case model.TradeFilterConditionLt:
			matched = parameterValue < value
			break
		case model.TradeFilterConditionLte:
			matched = parameterValue <= value
			break
		}
	}

	return matched
}

func (t *TradeFilterService) CompareBool(parameterValue bool, filter model.TradeFilter) bool {
	matched := false

	boolValue, err := strconv.ParseBool(filter.Value)
	if err == nil {
		switch filter.Condition {
		case model.TradeFilterConditionEq:
			matched = parameterValue == boolValue
			break
		case model.TradeFilterConditionNeq:
			matched = parameterValue != boolValue
			break
		}
	}

	return matched
}
