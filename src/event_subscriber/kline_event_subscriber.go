package event_subscriber

import (
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/event"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"log"
)

type KLineEventSubscriber struct {
	Binance            *client.Binance
	ExchangeRepository *repository.ExchangeRepository
	StatRepository     *repository.StatRepository
	BotService         service.BotServiceInterface
}

func (k KLineEventSubscriber) GetSubscribedEvents() map[string]func(interface{}) {
	return map[string]func(interface{}){
		event.EventNewKLineReceived: k.OnNewKlineReceived,
	}
}

func (k KLineEventSubscriber) OnNewKlineReceived(eventModel interface{}) {
	e, ok := eventModel.(event.NewKlineReceived)
	if !ok {
		return
	}

	tradesPastPeriod := k.Binance.TradesAggregate(
		e.Previous.Symbol,
		1000,
		e.Previous.Timestamp.GetPeriodFromMinute(),
		e.Previous.Timestamp.GetPeriodToMinute(),
	)

	sellQty := 0.00
	buyQty := 0.00
	sellVolume := 0.00
	buyVolume := 0.00
	buyCount := 0.00
	sellCount := 0.00
	buyPriceSum := 0.00
	sellPriceSum := 0.00

	for _, trade := range tradesPastPeriod {
		if trade.IsSell() {
			sellCount += 1.0
			sellQty += trade.Quantity
			sellVolume += trade.Quantity * trade.Price
			sellPriceSum += trade.Price
		}
		if trade.IsBuy() {
			buyCount += 1.0
			buyQty += trade.Quantity
			buyVolume += trade.Quantity * trade.Price
			buyPriceSum += trade.Price
		}
	}

	tradeVolume := model.TradeVolume{
		Symbol:     e.Previous.Symbol,
		Timestamp:  e.Previous.Timestamp,
		BuyQty:     buyQty,
		SellQty:    sellQty,
		PeriodFrom: model.TimestampMilli(e.Previous.Timestamp.GetPeriodFromMinute()),
		PeriodTo:   model.TimestampMilli(e.Previous.Timestamp.GetPeriodToMinute()),
	}

	// todo: request kLines with 0 volume...???
	k.ExchangeRepository.SetTradeVolume(tradeVolume)

	log.Printf(
		"[%s] new period kline event received %d -> %d | saved %d trades | buy = %.8f, sell = %.8f",
		e.Current.Symbol,
		e.Previous.Timestamp.GetPeriodToMinute(),
		e.Current.Timestamp.GetPeriodToMinute(),
		len(tradesPastPeriod),
		buyQty,
		sellQty,
	)
	k.ExchangeRepository.SaveKlineHistory(*e.Previous)

	// todo: move to separate service in future
	if !k.BotService.IsMasterBot() {
		return
	}

	maxPcs := 0.00
	minPcs := 0.00
	priceChangeSpeed := k.ExchangeRepository.GetPriceChangeSpeed(e.Previous.Symbol, e.Previous.Timestamp)

	if priceChangeSpeed != nil {
		maxPcs = priceChangeSpeed.MaxChange
		minPcs = priceChangeSpeed.MinChange
	}

	orderBookModel := model.OrderBookModel{
		Symbol: e.Previous.Symbol,
		Asks:   make([][2]model.Number, 0),
		Bids:   make([][2]model.Number, 0),
	}

	orderBook := k.Binance.GetDepth(e.Previous.Symbol, 500)
	if orderBook != nil {
		orderBookModel = orderBook.ToOrderBookModel(e.Previous.Symbol)
	}
	orderBookStat := orderBookModel.GetStat()
	stat := model.TradeStat{
		Symbol:        e.Previous.Symbol,
		Timestamp:     e.Previous.Timestamp,
		Price:         e.Previous.Close,
		BuyQty:        buyQty,
		SellQty:       sellQty,
		BuyVolume:     buyVolume,
		SellVolume:    sellVolume,
		AvgBuyPrice:   buyPriceSum / buyCount,
		AvgSellPrice:  sellPriceSum / sellCount,
		BuyCount:      buyCount,
		SellCount:     sellCount,
		TradeCount:    int64(len(tradesPastPeriod)),
		MaxPSC:        maxPcs,
		MinPCS:        minPcs,
		Open:          e.Previous.Open,
		Close:         e.Previous.Close,
		High:          e.Previous.High,
		Low:           e.Previous.Low,
		Volume:        e.Previous.Volume,
		OrderBookStat: orderBookStat,
	}
	_ = k.StatRepository.WriteTradeStat(stat)
}
