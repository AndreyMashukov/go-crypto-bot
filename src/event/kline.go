package event

import "github.com/AndreyMashukov/go-crypto-bot/server/src/model"

const EventNewKLineReceived = "event_new_kline_received"

type NewKlineReceived struct {
	Previous *model.KLine
	Current  *model.KLine
}
