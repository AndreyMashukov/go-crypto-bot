package model

type OrderTrade struct {
	Open         string  `json:"open"`
	Close        string  `json:"close"`
	Buy          float64 `json:"buy"`
	Sell         float64 `json:"sell"`
	BuyQuantity  float64 `json:"buyQuantity"`
	SellQuantity float64 `json:"sellQuantity"`
	Profit       float64 `json:"profit"`
	Symbol       string  `json:"symbol"`
	HoursOpened  int64   `json:"hoursOpened"`
	Budget       float64 `json:"budget"`
	Percent      float64 `json:"percent"`
}
