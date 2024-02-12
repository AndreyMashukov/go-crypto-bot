package model

import (
	"sort"
	"strings"
)

type SymbolInterface interface {
	GetSymbol() string
}

type DummySymbol struct {
	Symbol string
}

func (d DummySymbol) GetSymbol() string {
	return d.Symbol
}

type TradeLimitInterface interface {
	GetMinPrice() float64
	GetBaseAsset() string
	GetMinNotional() float64
	GetMinQuantity() float64
	GetSymbol() string
}

type TradeLimit struct {
	Id                           int64              `json:"id"`
	Symbol                       string             `json:"symbol"`
	USDTLimit                    float64            `json:"USDTLimit"`
	MinPrice                     float64            `json:"minPrice"`
	MinQuantity                  float64            `json:"minQuantity"`
	MinNotional                  float64            `json:"minNotional"`
	MinProfitPercent             float64            `json:"minProfitPercent"`
	IsEnabled                    bool               `json:"isEnabled"`
	MinPriceMinutesPeriod        int64              `json:"minPriceMinutesPeriod"`        //200,
	FrameInterval                string             `json:"frameInterval"`                //"2h",
	FramePeriod                  int64              `json:"framePeriod"`                  //20,
	BuyPriceHistoryCheckInterval string             `json:"buyPriceHistoryCheckInterval"` //"1d",
	BuyPriceHistoryCheckPeriod   int64              `json:"buyPriceHistoryCheckPeriod"`   //14,
	ExtraChargeOptions           ExtraChargeOptions `json:"extraChargeOptions"`
}

func (t TradeLimit) GetMinPrice() float64 {
	return t.MinPrice
}

func (t TradeLimit) GetSymbol() string {
	return t.Symbol
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

func (t *TradeLimit) GetBuyOnFallPercent(order Order, kLine KLine) Percent {
	buyOnFallPercent := Percent(0.00)

	if len(order.ExtraChargeOptions) > 0 {
		// sort DESC
		sort.SliceStable(order.ExtraChargeOptions, func(i int, j int) bool {
			return order.ExtraChargeOptions[i].Percent > order.ExtraChargeOptions[j].Percent
		})

		profit := order.GetProfitPercent(kLine.Close)
		// set first step as default
		buyOnFallPercent = order.ExtraChargeOptions[0].Percent

		for _, option := range order.ExtraChargeOptions {
			if profit.Lte(option.Percent) {
				buyOnFallPercent = option.Percent
			}
		}
	}

	if buyOnFallPercent > 0 {
		return buyOnFallPercent * -1
	} else {
		return buyOnFallPercent
	}
}

func (t *TradeLimit) GetClosePrice(buyPrice float64) float64 {
	return buyPrice * (100 + t.GetMinProfitPercent().Value()) / 100
}
