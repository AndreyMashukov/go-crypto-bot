package model

import "math"

type BinanceOrder struct {
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
	return b.Status == "NEW"
}

func (b *BinanceOrder) IsExpired() bool {
	return b.Status == "EXPIRED"
}

func (b *BinanceOrder) IsFilled() bool {
	return b.Status == "FILLED"
}

func (b *BinanceOrder) IsCancelled() bool {
	return b.Status == "CANCELED"
}

func (b *BinanceOrder) IsPartiallyFilled() bool {
	return b.Status == "PARTIALLY_FILLED"
}

func (b *BinanceOrder) HasExecutedQuantity() bool {
	return b.ExecutedQty > 0
}

func (b *BinanceOrder) GetExecutedQuantity() float64 {
	return b.ExecutedQty
}
