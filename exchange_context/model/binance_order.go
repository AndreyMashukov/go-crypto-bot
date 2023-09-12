package model

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
}
