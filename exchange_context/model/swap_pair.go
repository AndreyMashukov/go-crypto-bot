package model

type SwapPair struct {
	Id             int64
	SourceSymbol   string
	Symbol         string
	BaseAsset      string
	QuoteAsset     string
	LastPrice      float64
	PriceTimestamp int64
	MinNotional    float64
	MinQuantity    float64
	MinPrice       float64
}
