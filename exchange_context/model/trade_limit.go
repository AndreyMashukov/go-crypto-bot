package model

import "strings"

type TradeLimitInterface interface {
	GetMinPrice() float64
	GetBaseAsset() string
	GetMinNotional() float64
	GetMinQuantity() float64
}

type TradeLimit struct {
	Id               int64   `json:"id"`
	Symbol           string  `json:"symbol"`
	USDTLimit        float64 `json:"USDTLimit"`
	MinPrice         float64 `json:"minPrice"`
	MinQuantity      float64 `json:"minQuantity"`
	MinNotional      float64 `json:"minNotional"`
	MinProfitPercent float64 `json:"minProfitPercent"`
	IsEnabled        bool    `json:"isEnabled"`

	// Extra budget for market fall
	USDTExtraBudget  float64 `json:"USDTExtraBudget"`
	BuyOnFallPercent float64 `json:"buyOnFallPercent"`

	MinPriceMinutesPeriod        int64  `json:"minPriceMinutesPeriod"`        //200,
	FrameInterval                string `json:"frameInterval"`                //"2h",
	FramePeriod                  int64  `json:"framePeriod"`                  //20,
	BuyPriceHistoryCheckInterval string `json:"buyPriceHistoryCheckInterval"` //"1d",
	BuyPriceHistoryCheckPeriod   int64  `json:"buyPriceHistoryCheckPeriod"`   //14,
}

func (t TradeLimit) GetMinPrice() float64 {
	return t.MinPrice
}

func (t TradeLimit) GetMinNotional() float64 {
	return t.MinNotional
}

func (t TradeLimit) GetMinQuantity() float64 {
	return t.MinQuantity
}

func (t TradeLimit) GetBaseAsset() string {
	return strings.ReplaceAll(t.Symbol, "USDT", "")
}

func (t *TradeLimit) GetMinProfitPercent() Percent {
	if t.MinProfitPercent < 0 {
		return Percent(t.MinProfitPercent * -1)
	} else {
		return Percent(t.MinProfitPercent)
	}
}

func (t *TradeLimit) GetBuyOnFallPercent() Percent {
	if t.BuyOnFallPercent > 0 {
		return Percent(t.BuyOnFallPercent * -1)
	} else {
		return Percent(t.BuyOnFallPercent)
	}
}

func (t *TradeLimit) IsExtraChargeEnabled() bool {
	return t.BuyOnFallPercent != 0.00
}

func (t *TradeLimit) GetClosePrice(buyPrice float64) float64 {
	return buyPrice * (100 + t.GetMinProfitPercent().Value()) / 100
}
