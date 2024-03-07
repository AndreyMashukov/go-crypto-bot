package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

const MinProfitPercent = 0.50

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

type ProfitPositionInterface interface {
	GetPositionTime() PositionTime
	GetProfitOptions() ProfitOptions
	GetSymbol() string
	GetExecutedQuantity() float64
	GetPositionQuantityWithSwap() float64
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
	ProfitOptions      ProfitOptions      `json:"profitOptions"`
	ExtraChargeOptions ExtraChargeOptions `json:"extraChargeOptions"`
	SwapQuantity       *float64           `json:"swapQuantity"`
}

func (o *Order) CanExtraBuy(kLine KLine, withSwap bool) bool {
	if o.Swap {
		return false
	}

	if len(o.ExtraChargeOptions) == 0 {
		return false
	}

	return o.GetAvailableExtraBudget(kLine, withSwap) >= 10.00
}

func (o Order) GetExecutedQuantity() float64 {
	return o.GetRemainingToSellQuantity(false)
}

func (o Order) GetPositionQuantityWithSwap() float64 {
	return o.GetRemainingToSellQuantity(true)
}

func (o *Order) GetAvailableExtraBudget(kLine KLine, withSwap bool) float64 {
	availableExtraBudget := 0.00

	if len(o.ExtraChargeOptions) > 0 {
		availableExtraBudget = 0.00
		// sort DESC
		sort.SliceStable(o.ExtraChargeOptions, func(i int, j int) bool {
			return o.ExtraChargeOptions[i].Percent > o.ExtraChargeOptions[j].Percent
		})

		profit := o.GetProfitPercent(kLine.Close, withSwap)

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

func (o Order) GetPositionTime() PositionTime {
	date, _ := time.Parse("2006-01-02 15:04:05", o.CreatedAt)

	return PositionTime(time.Now().Unix() - date.Unix())
}

func (o *Order) GetProfitPercent(currentPrice float64, withSwap bool) Percent {
	return Percent(math.Round(((o.GetQuoteProfit(currentPrice, withSwap)*100.00)/(o.Price*o.GetRemainingToSellQuantity(false)))*100) / 100)
}

func (o *Order) GetQuoteProfit(sellPrice float64, withSwap bool) float64 {
	return (sellPrice * o.GetRemainingToSellQuantity(withSwap)) - (o.GetRemainingToSellQuantity(false) * o.Price)
}

func (o Order) GetProfitOptions() ProfitOptions {
	return o.ProfitOptions
}

func (o *Order) GetManualMinClosePrice() float64 {
	return o.Price * (100 + MinProfitPercent) / 100
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

func (o *Order) GetRemainingToSellQuantity(swap bool) float64 {
	executedQuantity := o.ExecutedQuantity

	if swap && o.SwapQuantity != nil {
		executedQuantity += *o.SwapQuantity
	}

	if o.SoldQuantity != nil {
		executedQuantity -= *o.SoldQuantity
	}

	return executedQuantity
}

func (o *Order) IsSwap() bool {
	return o.Swap
}
func (o Order) GetSymbol() string {
	return o.Symbol
}

type Position struct {
	Symbol            string                `json:"symbol"`
	KLine             KLine                 `json:"kLine"`
	Order             Order                 `json:"order"`
	Percent           Percent               `json:"percent"`
	SellPrice         float64               `json:"sellPrice"`
	PredictedPrice    float64               `json:"predictedPrice"`
	Profit            float64               `json:"profit"`
	TargetProfit      float64               `json:"targetProfit"`
	Interpolation     Interpolation         `json:"interpolation"`
	OrigQty           float64               `json:"origQty"`
	ExecutedQty       float64               `json:"executedQty"`
	ManualOrderConfig ManualOrderConfig     `json:"manualOrderConfig"`
	PositionTime      PositionTime          `json:"positionTime"`
	CloseStrategy     PositionCloseStrategy `json:"closeStrategy"`
	IsPriceExpired    bool                  `json:"isPriceExpired"`
}

type PositionCloseStrategy struct {
	MinClosePrice    float64 `json:"minClosePrice"`
	MinProfitPercent Percent `json:"minProfitPercent"`
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

type ProfitOptions []ProfitOption

const ProfitOptionUnitMinute = "i"
const ProfitOptionUnitHour = "h"
const ProfitOptionUnitDay = "d"
const ProfitOptionUnitMonth = "m"

type ProfitOption struct {
	Index           int64   `json:"index"`
	IsTriggerOption bool    `json:"isTriggerOption"`
	OptionValue     float64 `json:"optionValue"`
	OptionUnit      string  `json:"optionUnit"`
	OptionPercent   Percent `json:"optionPercent"`
}

func (p *ProfitOptions) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &p)
}
func (p ProfitOptions) Value() (driver.Value, error) {
	jsonV, err := json.Marshal(p)
	return string(jsonV), err
}
func (p ProfitOption) IsMinutely() bool {
	return p.OptionUnit == ProfitOptionUnitMinute
}
func (p ProfitOption) IsHourly() bool {
	return p.OptionUnit == ProfitOptionUnitHour
}
func (p ProfitOption) IsDaily() bool {
	return p.OptionUnit == ProfitOptionUnitDay
}
func (p ProfitOption) IsMonthly() bool {
	return p.OptionUnit == ProfitOptionUnitMonth
}
func (p ProfitOption) GetPositionTime() (PositionTime, error) {
	switch p.OptionUnit {
	case ProfitOptionUnitMinute:
		return PositionTime(p.OptionValue * 60), nil
	case ProfitOptionUnitHour:
		return PositionTime(p.OptionValue * 3600), nil
	case ProfitOptionUnitDay:
		return PositionTime(p.OptionValue * 3600 * 24), nil
	case ProfitOptionUnitMonth:
		return PositionTime(p.OptionValue * 3600 * 24 * 30), nil
	}

	return PositionTime(0.00), errors.New("position time is invalid")
}

type PositionTime int64

func (p PositionTime) GetMinutes() float64 {
	return float64(p) / float64(60)
}
func (p PositionTime) GetHours() float64 {
	return float64(p) / float64(3600)
}
func (p PositionTime) GetDays() float64 {
	return float64(p) / float64(3600*24)
}
func (p PositionTime) GetMonths() float64 {
	return float64(p) / float64(3600*24*30)
}
