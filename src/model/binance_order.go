package model

import (
	"math"
	"strconv"
)

type ByBitOrder struct {
	OrderId     string  `json:"orderId"`
	Symbol      string  `json:"symbol"`
	Price       float64 `json:"price,string"`
	OrigQty     float64 `json:"qty,string"`
	ExecutedQty float64 `json:"cumExecQty,string"`
	Status      string  `json:"orderStatus"`
	Type        string  `json:"orderType"`
	Side        string  `json:"side"`
	Timestamp   int64   `json:"createdTime,string"`
}

func (b ByBitOrder) GetOrderId() string {
	return b.OrderId
}

type ExchangeOrderInterface interface {
	GetOrderId() string
}

type BinanceOrderLegacy struct {
	OrderId             int64   `json:"orderId"`
	Symbol              string  `json:"symbol"`
	TransactTime        int64   `json:"transactTime"`
	Price               float64 `json:"price,string"`
	OrigQty             float64 `json:"origQty,string"`
	ExecutedQty         float64 `json:"executedQty,string"`
	CummulativeQuoteQty float64 `json:"cummulativeQuoteQty,string"`
	Status              string  `json:"status"`
	Type                string  `json:"type"`
	Side                string  `json:"side"`
	WorkingTime         int64   `json:"workingTime"`
	Timestamp           int64   `json:"time"`
}

func (b *BinanceOrderLegacy) ToModern() BinanceOrder {
	orderIdString := strconv.FormatInt(b.OrderId, 10)

	return BinanceOrder{
		OrderId:             orderIdString,
		Symbol:              b.Symbol,
		TransactTime:        b.TransactTime,
		Price:               b.Price,
		OrigQty:             b.OrigQty,
		ExecutedQty:         b.ExecutedQty,
		CummulativeQuoteQty: b.CummulativeQuoteQty,
		Status:              b.Status,
		Type:                b.Type,
		Side:                b.Side,
		WorkingTime:         b.WorkingTime,
		Timestamp:           b.Timestamp,
	}
}

const ExchangeOrderStatusNew = "NEW"

type BinanceOrder struct {
	OrderId             string  `json:"orderId"`
	Symbol              string  `json:"symbol"`
	TransactTime        int64   `json:"transactTime"`
	Price               float64 `json:"price,string"`
	OrigQty             float64 `json:"origQty,string"`
	ExecutedQty         float64 `json:"executedQty,string"`
	CummulativeQuoteQty float64 `json:"cummulativeQuoteQty,string"`
	Status              string  `json:"status"`
	Type                string  `json:"type"`
	Side                string  `json:"side"`
	WorkingTime         int64   `json:"workingTime"`
	Timestamp           int64   `json:"time"`
}

func (b *BinanceOrder) IsBuy() bool {
	return b.Side == "BUY"
}

func (b *BinanceOrder) IsSell() bool {
	return b.Side == "SELL"
}

func (b *BinanceOrder) GetProfitPercent(currentPrice float64) Percent {
	return Percent(math.Round((currentPrice-b.Price)*100/b.Price*100) / 100)
}

func (b *BinanceOrder) IsNew() bool {
	return b.Status == ExchangeOrderStatusNew
}

func (b *BinanceOrder) IsExpired() bool {
	return b.Status == "EXPIRED" || b.Status == "EXPIRED_IN_MATCH"
}

func (b *BinanceOrder) IsFilled() bool {
	return b.Status == "FILLED"
}

func (b *BinanceOrder) IsCanceled() bool {
	return b.Status == "CANCELED"
}

func (b *BinanceOrder) IsPartiallyFilled() bool {
	return b.Status == "PARTIALLY_FILLED"
}

func (b *BinanceOrder) IsNearlyFilled() bool {
	if b.IsFilled() {
		return true
	}

	if !b.IsPartiallyFilled() {
		return false
	}

	return (b.ExecutedQty * 100 / b.OrigQty) >= 99.5
}

func (b *BinanceOrder) HasExecutedQuantity() bool {
	return b.ExecutedQty > 0
}

func (b *BinanceOrder) GetExecutedQuantity() float64 {
	return b.ExecutedQty
}
