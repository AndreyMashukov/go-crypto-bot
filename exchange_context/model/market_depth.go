package exchange_context

import (
	"encoding/json"
	"strconv"
)

type Number struct {
	Value float64
}

type MarketDepth struct {
	LastUpdateId int64       `json:"lastUpdateId"`
	Bids         [][2]Number `json:"bids"`
	Asks         [][2]Number `json:"asks"`
}

func (p *Number) UnmarshalJSON(b []byte) error {
	var value string
	err := json.Unmarshal(b, &value)
	if err != nil {
		return err
	}

	result, err := strconv.ParseFloat(value, 64)

	if err != nil {
		return err
	}

	p.Value = result
	return nil
}
