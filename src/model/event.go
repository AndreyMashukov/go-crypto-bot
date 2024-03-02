package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Event struct {
	Stream string `json:"stream"`
}

type Number struct {
	Value float64
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

func (p *Number) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%.8f", p.Value))
}
