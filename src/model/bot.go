package model

import (
	"database/sql/driver"
	"encoding/json"
)

type Bot struct {
	Id            int64      `json:"id"`
	BotUuid       string     `json:"botUuid"`
	IsMasterBot   bool       `json:"isMasterBot"`
	IsSwapEnabled bool       `json:"isSwapEnabled"`
	SwapConfig    SwapConfig `json:"swapConfig"`
}

type SwapConfig struct {
	MinValidPercent    float64      `json:"swapMinPercent"`
	FallPercentTrigger float64      `json:"swapOrderProfitTrigger"`
	OrderTimeTrigger   PositionTime `json:"orderTimeTrigger"`
	UseSwapCapital     bool         `json:"useSwapCapital"`
	HistoryInterval    string       `json:"historyInterval"`
	HistoryPeriod      int64        `json:"historyPeriod"`
}

func (s *SwapConfig) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &s)
}
func (s SwapConfig) Value() (driver.Value, error) {
	jsonV, err := json.Marshal(s)
	return string(jsonV), err
}

type BotConfigUpdate struct {
	IsMasterBot   bool       `json:"isMasterBot"`
	IsSwapEnabled bool       `json:"isSwapEnabled"`
	SwapConfig    SwapConfig `json:"swapConfig"`
}
