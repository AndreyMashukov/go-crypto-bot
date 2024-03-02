package model

type TradeEvent struct {
	Stream string `json:"stream"`
	Trade  Trade  `json:"data"`
}
