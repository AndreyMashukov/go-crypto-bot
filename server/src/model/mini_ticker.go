package model

type MiniTickerEvent struct {
	MiniTicker MiniTicker `json:"data"`
}

type MiniTicker struct {
	EventTime        TimestampMilli `json:"E"`
	Symbol           string         `json:"s"`
	Close            Price          `json:"c"`
	TotalVolumeAsset Volume         `json:"v"`
	TotalVolumeQuote Volume         `json:"q"`
}
