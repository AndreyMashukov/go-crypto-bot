package model

import (
	"github.com/google/uuid"
)

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
	Exchange      string         `json:"exchange"`
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

type TradeLearnDataset struct {
	OrderBookBuyFirstQty   float64
	OrderBookSellFirstQty  float64
	OrderBookBuyQtySum     float64
	OrderBookSellQtySum    float64
	OrderBookBuyVolumeSum  float64
	OrderBookSellVolumeSum float64
	SecondaryPrice         float64
	PrimaryPrice           float64
}

type TradePricePredictParams struct {
	Symbol                 string
	OrderBookBuyFirstQty   float64
	OrderBookSellFirstQty  float64
	OrderBookBuyQtySum     float64
	OrderBookSellQtySum    float64
	OrderBookBuyVolumeSum  float64
	OrderBookSellVolumeSum float64
	SecondaryPrice         float64
}
