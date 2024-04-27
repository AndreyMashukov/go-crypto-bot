package model

import (
	"github.com/google/uuid"
	"time"
)

type TimestampMilli int64

func (t TimestampMilli) Value() int64 {
	return int64(t)
}

func (t TimestampMilli) GetPeriodFromMinute() int64 {
	dateTime := time.Unix(0, t.Value()*int64(time.Millisecond))
	newDate := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), dateTime.Hour(), dateTime.Minute(), 0, 0, dateTime.Location())
	return newDate.UnixMilli()
}

func (t TimestampMilli) GetPeriodToMinute() int64 {
	dateTime := time.Unix(0, t.Value()*int64(time.Millisecond))
	newDate := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), dateTime.Hour(), dateTime.Minute(), 59, 0, dateTime.Location())
	return newDate.UnixMilli() + 999
}

func (t TimestampMilli) Neq(milli TimestampMilli) bool {
	return t.Value() != milli.Value()
}

func (t TimestampMilli) Eq(milli TimestampMilli) bool {
	return t.Value() == milli.Value()
}

func (t TimestampMilli) PeriodToEq(milli TimestampMilli) bool {
	return t.GetPeriodToMinute() == milli.GetPeriodToMinute()
}

func (t TimestampMilli) Gt(milli TimestampMilli) bool {
	return t.Value() > milli.Value()
}

func (t TimestampMilli) Gte(milli TimestampMilli) bool {
	return t.Value() >= milli.Value()
}

func (t TimestampMilli) Lt(milli TimestampMilli) bool {
	return t.Value() < milli.Value()
}

func (t TimestampMilli) Lte(milli TimestampMilli) bool {
	return t.Value() <= milli.Value()
}

type Trade struct {
	AggregateTradeId int64          `json:"a,int"`
	Price            float64        `json:"p,string"`
	Symbol           string         `json:"s"`
	Quantity         float64        `json:"q,string"`
	IsBuyerMaker     bool           `json:"m,bool"` // IsBuyerMaker = true -> SELL / IsBuyerMaker = false -> BUY
	Timestamp        TimestampMilli `json:"T,int"`
	Ignore           bool           `json:"M,bool"`
}

func (c *Trade) GetOperation() string {
	if c.IsSell() {
		return "SELL"
	}

	return "BUY"
}

func (c *Trade) IsSell() bool {
	return c.IsBuyerMaker == true
}

func (c *Trade) IsBuy() bool {
	return c.IsBuyerMaker == false
}

type TradeVolume struct {
	Symbol     string         `json:"symbol"`
	Timestamp  TimestampMilli `json:"timestamp"`
	PeriodFrom TimestampMilli `json:"periodFrom"`
	PeriodTo   TimestampMilli `json:"periodTo"`
	BuyQty     float64        `json:"buyQty"`
	SellQty    float64        `json:"sellQty"`
}

type TradeStat struct {
	Symbol        string         `json:"symbol"`
	BotId         uuid.UUID      `json:"botId"`
	Timestamp     TimestampMilli `json:"timestamp"`
	Price         float64        `json:"price"`
	BuyQty        float64        `json:"buyQty"`
	SellQty       float64        `json:"sellQty"`
	AvgSellPrice  float64        `json:"avgSellPrice"`
	AvgBuyPrice   float64        `json:"avgBuyPrice"`
	BuyVolume     float64        `json:"buyVolume"`
	SellVolume    float64        `json:"sellVolume"`
	SellCount     float64        `json:"sellCount"`
	BuyCount      float64        `json:"buyCount"`
	TradeCount    int64          `json:"tradeCount"`
	MaxPSC        float64        `json:"maxPSC"`
	MinPCS        float64        `json:"minPCS"`
	Open          float64        `json:"open"`
	Close         float64        `json:"close"`
	High          float64        `json:"high"`
	Low           float64        `json:"low"`
	Volume        float64        `json:"volume"`
	OrderBookStat OrderBookStat  `json:"orderBookStat"`
}
