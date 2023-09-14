package model

type ChartPoint struct {
	XAxis int64   `json:"x"` // date (timestamp)
	YAxis float64 `json:"y"` // price
}
