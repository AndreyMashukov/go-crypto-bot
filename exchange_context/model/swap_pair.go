package model

import "time"

type SwapPair struct {
	Id             int64   `json:"id"`
	SourceSymbol   string  `json:"sourceSymbol"`
	Symbol         string  `json:"symbol"`
	BaseAsset      string  `json:"baseAsset"`
	QuoteAsset     string  `json:"quoteAsset"`
	BuyPrice       float64 `json:"buyPrice"`
	SellPrice      float64 `json:"sellPrice"`
	PriceTimestamp int64   `json:"priceTimestamp"`
	MinNotional    float64 `json:"minNotional"`
	MinQuantity    float64 `json:"minQuantity"`
	MinPrice       float64 `json:"minPrice"`
	BuyVolume      float64 `json:"buyVolume"`
	SellVolume     float64 `json:"sellVolume"`
	DailyPercent   float64 `json:"dailyPercent"`
}

func (s SwapPair) IsGainer() bool {
	return s.DailyPercent >= 0.5
}

func (s SwapPair) IsLooser() bool {
	return s.DailyPercent <= -0.5
}

func (s SwapPair) IsPriceExpired() bool {
	return (time.Now().Unix() - (s.PriceTimestamp)) > 60
}

func (s SwapPair) IsBullMarket() bool {
	return s.BuyVolume/s.SellVolume >= 1.60
}

func (s SwapPair) IsBearMarket() bool {
	return s.SellVolume/s.BuyVolume >= 1.60
}

func (s SwapPair) IsQuietMarket() bool {
	return !s.IsBullMarket() && !s.IsBearMarket()
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
