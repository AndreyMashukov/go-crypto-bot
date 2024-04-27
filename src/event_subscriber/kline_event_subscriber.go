package event_subscriber

import (
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/event"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"log"
)

type KLineEventSubscriber struct {
	Binance            *client.Binance
	ExchangeRepository *repository.ExchangeRepository
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
	for _, trade := range tradesPastPeriod {
		if trade.IsSell() {
			sellQty += trade.Quantity
		}
		if trade.IsBuy() {
			buyQty += trade.Quantity
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
}
