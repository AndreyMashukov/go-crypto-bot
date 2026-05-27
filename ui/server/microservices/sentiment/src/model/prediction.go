package model

type Prediction struct {
	Label string   `json:"label"`
	Score float64  `json:"score"`
	Coins []string `json:"coins"`
}
