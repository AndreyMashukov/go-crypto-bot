package service

import (
	"github.com/AndreyMashukov/go-crypto-bot/server/src/event"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/repository"
	"log"
)

type KLineEventSubscriber struct {
	ExchangeRepository *repository.ExchangeRepository
	StatRepository     *repository.StatRepository
	BotService         BotServiceInterface
	StatService        *StatService
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

	tradeStat := k.StatService.GetTradeStat(*e.Previous, false, k.BotService.IsMasterBot())
	tradeVolume := model.TradeVolume{
		Symbol:     e.Previous.Symbol,
		Timestamp:  e.Previous.Timestamp,
		BuyQty:     tradeStat.BuyQty,
		SellQty:    tradeStat.SellQty,
		PeriodFrom: model.TimestampMilli(e.Previous.Timestamp.GetPeriodFromMinute()),
		PeriodTo:   model.TimestampMilli(e.Previous.Timestamp.GetPeriodToMinute()),
	}
	k.ExchangeRepository.SetTradeVolume(tradeVolume)

	if k.BotService.IsMasterBot() {
		_ = k.StatRepository.WriteTradeStat(tradeStat)
	}

	// todo: request kLines with 0 volume...???
	log.Printf(
		"[%s] new period kline event received %d -> %d | saved %d trades | buy = %.8f, sell = %.8f",
		e.Current.Symbol,
		e.Previous.Timestamp.GetPeriodToMinute(),
		e.Current.Timestamp.GetPeriodToMinute(),
		tradeStat.TradeCount,
		tradeStat.BuyQty,
		tradeStat.SellQty,
	)
}
