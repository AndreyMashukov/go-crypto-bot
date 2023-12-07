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

func (b *BinanceOrder) GetProfitPercent(currentPrice float64) float64 {
	return math.Round((currentPrice-b.Price)*100/b.Price*100) / 100
}
