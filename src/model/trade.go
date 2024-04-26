package model

import "time"

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
