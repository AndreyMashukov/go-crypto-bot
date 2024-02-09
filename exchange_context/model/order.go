package model

import (
	"database/sql/driver"
	"encoding/json"
	"math"
	"sort"
	"strings"
	"time"
)

type Percent float64

func (p Percent) IsPositive() bool {
	return float64(p) > 0
}

func (p Percent) Value() float64 {
	return float64(p)
}

func (p Percent) Half() Percent {
	return Percent(float64(p) / 2)
}

func (p Percent) Gt(percent Percent) bool {
	return p.Value() > percent.Value()
}

func (p Percent) Gte(percent Percent) bool {
	return p.Value() >= percent.Value()
}

func (p Percent) Lte(percent Percent) bool {
	return p.Value() <= percent.Value()
}

func (p Percent) Lt(percent Percent) bool {
	return p.Value() < percent.Value()
}

type ErrorNotification struct {
	BotId        int64  `json:"bot"`
	Stop         bool   `json:"stop"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type TgOrderNotification struct {
	BotId     int64   `json:"bot"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"amount"`
	Symbol    string  `json:"symbol"`
	Operation string  `json:"operation"`
	DateTime  string  `json:"dateTime"`
	Details   string  `json:"details"`
}

type Order struct {
	Id                 int64              `json:"id"`
	Symbol             string             `json:"symbol"`
	Price              float64            `json:"price"`
	Quantity           float64            `json:"quantity"`
	ExecutedQuantity   float64            `json:"executedQuantity"`
	CreatedAt          string             `json:"createdAt"`
	SellVolume         float64            `json:"sellVolume"`
	BuyVolume          float64            `json:"buyVolume"`
	SmaValue           float64            `json:"smaValue"`
	Operation          string             `json:"operation"`
	Status             string             `json:"status"`
	ExternalId         *int64             `json:"externalId"`
	ClosesOrder        *int64             `json:"closesOrder"` // sell order here
	UsedExtraBudget    float64            `json:"usedExtraBudget"`
	Commission         *float64           `json:"commission"`
	CommissionAsset    *string            `json:"commissionAsset"`
	SoldQuantity       *float64           `json:"soldQuantity"`
	Swap               bool               `json:"swap"`
	ExtraChargeOptions ExtraChargeOptions `json:"extraChargeOptions"`
}

func (o *Order) CanExtraBuy(tradeLimit TradeLimit, kLine KLine) bool {
	if !tradeLimit.IsExtraChargeEnabled() && len(o.ExtraChargeOptions) == 0 {
		return false
	}

	return o.GetAvailableExtraBudget(tradeLimit, kLine) >= 10.00
}

func (o *Order) GetAvailableExtraBudget(tradeLimit TradeLimit, kLine KLine) float64 {
	availableExtraBudget := tradeLimit.USDTExtraBudget

	if len(o.ExtraChargeOptions) > 0 {
		availableExtraBudget = 0.00
		// sort DESC
		sort.SliceStable(o.ExtraChargeOptions, func(i int, j int) bool {
			return o.ExtraChargeOptions[i].Percent > o.ExtraChargeOptions[j].Percent
		})

		profit := o.GetProfitPercent(kLine.Close)

		for _, option := range o.ExtraChargeOptions {
			if profit.Lte(option.Percent) {
				availableExtraBudget += option.AmountUsdt
			}
		}
	}

	availableExtraBudget -= o.UsedExtraBudget

	return availableExtraBudget
}

func (o *Order) GetBaseAsset() string {
	return strings.ReplaceAll(o.Symbol, "USDT", "")
}

func (o *Order) GetHoursOpened() int64 {
	date, _ := time.Parse("2006-01-02 15:04:05", o.CreatedAt)

	return (time.Now().Unix() - date.Unix()) / 3600
}

func (o *Order) GetProfitPercent(currentPrice float64) Percent {
	return Percent(math.Round((currentPrice-o.Price)*100/o.Price*100) / 100)
}

func (o *Order) GetQuoteProfit(sellPrice float64) float64 {
	return (sellPrice - o.Price) * o.GetRemainingToSellQuantity()
}

func (o *Order) GetMinClosePrice(limit TradeLimit) float64 {
	return o.Price * (100 + limit.GetMinProfitPercent().Value()) / 100
}

func (o *Order) GetManualMinClosePrice() float64 {
	return o.Price * (100 + 0.50) / 100
}

func (o *Order) IsSell() bool {
	return o.Operation == "SELL"
}

func (o *Order) IsBuy() bool {
	return o.Operation == "BUY"
}

func (o *Order) IsClosed() bool {
	return o.Status == "closed"
}

func (o *Order) GetRemainingToSellQuantity() float64 {
	if o.SoldQuantity != nil {
		return o.ExecutedQuantity - *o.SoldQuantity
	}

	return o.ExecutedQuantity
}

func (o *Order) IsSwap() bool {
	return o.Swap
}

type Position struct {
	Symbol            string            `json:"symbol"`
	KLine             KLine             `json:"kLine"`
	Order             Order             `json:"order"`
	Percent           Percent           `json:"percent"`
	SellPrice         float64           `json:"sellPrice"`
	PredictedPrice    float64           `json:"predictedPrice"`
	Profit            float64           `json:"profit"`
	TargetProfit      float64           `json:"targetProfit"`
	Interpolation     Interpolation     `json:"interpolation"`
	OrigQty           float64           `json:"origQty"`
	ExecutedQty       float64           `json:"executedQty"`
	ManualOrderConfig ManualOrderConfig `json:"manualOrderConfig"`
}

type ManualOrderConfig struct {
	PriceStep     float64 `json:"priceStep"`
	MinClosePrice float64 `json:"minSellPrice"`
}

type PendingOrder struct {
	Symbol         string        `json:"symbol"`
	KLine          KLine         `json:"kLine"`
	BinanceOrder   BinanceOrder  `json:"binanceOrder"`
	PredictedPrice float64       `json:"predictedPrice"`
	Interpolation  Interpolation `json:"interpolation"`
	IsRisky        bool          `json:"isRisky"`
}

type ExtraChargeOptions []ExtraChargeOption

type ExtraChargeOption struct {
	Index      int64   `json:"index"`
	Percent    Percent `json:"percent"`
	AmountUsdt float64 `json:"amountUsdt"`
}

func (e *ExtraChargeOptions) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &e)
}

func (e ExtraChargeOptions) Value() (driver.Value, error) {
	jsonV, err := json.Marshal(e)
	return string(jsonV), err
}
