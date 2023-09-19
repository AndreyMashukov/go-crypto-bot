package model

type ManualOrder struct {
	Operation string  `json:"operation"`
	Price     float64 `json:"price"`
	Symbol    string  `json:"symbol"`
}
