package model

import "strings"

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
}

func (t *TradeLimit) GetBaseAsset() string {
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
