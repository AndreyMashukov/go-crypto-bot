package exchange

import (
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"log"
	"math"
	"sort"
)

type ProfitServiceInterface interface {
	CheckBuyPriceOnHistory(limit model.TradeLimit, buyPrice float64) float64
	GetMinClosePrice(order model.ProfitPositionInterface, currentPrice float64) float64
	GetMinProfitPercent(order model.ProfitPositionInterface) model.Percent
}

type ProfitService struct {
	Binance    client.ExchangePriceAPIInterface
	BotService service.BotServiceInterface
}

func (p *ProfitService) CheckBuyPriceOnHistory(limit model.TradeLimit, buyPrice float64) float64 {
	kLines := p.Binance.GetKLinesCached(limit.Symbol, limit.BuyPriceHistoryCheckInterval, limit.BuyPriceHistoryCheckPeriod)

	//priceBefore := buyPrice

	for {
		closePrice := p.GetMinClosePrice(limit, buyPrice)
		var closePriceMetTimes int64 = 0
		for _, kline := range kLines {
			if kline.High >= closePrice {
				closePriceMetTimes++
			}
		}

		if float64(closePriceMetTimes) > (float64(limit.BuyPriceHistoryCheckPeriod) * 0.8) {
			break
		}

		buyPrice -= limit.MinPrice
	}

	// log.Printf("[%s] Price history check completed: %f -> %f", limit.Symbol, priceBefore, buyPrice)

	return buyPrice
}

func (p *ProfitService) GetMinProfitPercent(order model.ProfitPositionInterface) model.Percent {
	minAllowedValue := model.Percent(0.5)
	profitOptions := order.GetProfitOptions()
	sort.SliceStable(profitOptions, func(i int, j int) bool {
		return profitOptions[i].Index < profitOptions[j].Index
	})

	positionTime := order.GetPositionTime()

	for index, option := range profitOptions {
		optionPositionTime, err := option.GetPositionTime()

		if err != nil {
			log.Printf("[%s] Price Calculator: profit position [%d] time is invalid", order.GetSymbol(), index)
			continue
		}

		if optionPositionTime > 0.00 && positionTime <= optionPositionTime {
			return model.Percent(math.Max(minAllowedValue.Value(), option.OptionPercent.Value()))
		}
	}

	if len(profitOptions) > 0 {
		option := profitOptions[len(profitOptions)-1]

		return model.Percent(math.Max(minAllowedValue.Value(), option.OptionPercent.Value()))
	}

	return minAllowedValue
}

func (p *ProfitService) GetMinClosePrice(order model.ProfitPositionInterface, currentPrice float64) float64 {
	minProfitPercent := p.GetMinProfitPercent(order).Value()

	if p.BotService.UseSwapCapital() && order.GetExecutedQuantity() > 0.00 {
		executedValue := order.GetExecutedQuantity() * currentPrice
		targetValue := executedValue * (100 + minProfitPercent) / 100

		return targetValue / order.GetPositionQuantityWithSwap()
	}

	return currentPrice * (100 + minProfitPercent) / 100
}
