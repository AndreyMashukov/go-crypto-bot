package model

type Decision struct {
	Operation    string     `json:"operation"`
	Timestamp    int64      `json:"timestamp"`
	StrategyName string     `json:"strategyName"`
	Score        float64    `json:"score"`
	Price        float64    `json:"price"`
	Params       [3]float64 `json:"params"`
}
