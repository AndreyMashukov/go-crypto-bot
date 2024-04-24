package event

import "gitlab.com/open-soft/go-crypto-bot/src/model"

const EventNewKLineReceived = "event_new_kline_received"

type NewKlineReceived struct {
	Previous *model.KLine
	Current  *model.KLine
}
