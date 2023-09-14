package model

type ChartPoint struct {
	XAxis int64   `json:"x"` // date (timestamp)
	YAxis float64 `json:"y"` // price
}

type FinancialPoint struct {
	XAxis int64   `json:"x"`
	Open  float64 `json:"o"`
	High  float64 `json:"h"`
	Low   float64 `json:"l"`
	Close float64 `json:"c"`
}
