package model

import "time"

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

func (s SwapPair) IsPriceExpired() bool {
	return (time.Now().Unix() - (s.PriceTimestamp / 1000)) > 60
}

func (s SwapPair) GetMinPrice() float64 {
	return s.MinPrice
}

func (s SwapPair) GetMinNotional() float64 {
	return s.MinNotional
}

func (s SwapPair) GetMinQuantity() float64 {
	return s.MinQuantity
}

func (s SwapPair) GetBaseAsset() string {
	return s.BaseAsset
}
