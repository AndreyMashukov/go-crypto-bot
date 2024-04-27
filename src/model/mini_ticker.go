package model

type MiniTickerEvent struct {
	MiniTicker MiniTicker `json:"data"`
}

type MiniTicker struct {
	EventTime        TimestampMilli `json:"E"`
	Symbol           string         `json:"s"`
	Close            float64        `json:"c,string"`
	TotalVolumeAsset float64        `json:"v,string"`
	TotalVolumeQuote float64        `json:"q,string"`
}
