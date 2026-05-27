package model

type TradeStat struct {
	Close              float64 `json:"close"`
	Low                float64 `json:"low"`
	High               float64 `json:"high"`
	Pivot              float64 `json:"pivot"`
	S1                 float64 `json:"s1"`
	S2                 float64 `json:"s2"`
	S3                 float64 `json:"s3"`
	R1                 float64 `json:"r1"`
	R2                 float64 `json:"r2"`
	R3                 float64 `json:"r3"`
	Symbol             string  `json:"symbol"`
	BuyPrice           float64 `json:"buyPrice"`
	SellPrice          float64 `json:"sellPrice"`
	HalfSellPrice      float64 `json:"halfSellPrice"`
	SellPriceNegative  float64 `json:"sellPriceNegative"`
	ExtraChargePrice   float64 `json:"extraChargePrice"`
	AvgPrice           float64 `json:"avgPrice"`
	Percent            Percent `json:"percent"`
	PercentHalf        Percent `json:"percentHalf"`
	PercentNegative    Percent `json:"percentNegative"`
	ExtraChargePercent Percent `json:"extraChargePercent"`
	IntervalMinutes    int64   `json:"intervalMinutes"`
	PeriodDays         int64   `json:"periodDays"`
}
