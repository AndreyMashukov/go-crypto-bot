package model

type Decision struct {
	Operation    string
	Timestamp    int64
	StrategyName string
	Score        float64
	Price        float64
	Params       [3]float64
}
