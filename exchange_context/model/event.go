package model

import (
	"encoding/json"
	"strconv"
	"time"
)

type UnixTime struct {
	time.Time
}

// UnmarshalJSON is the method that satisfies the Unmarshaller interface
func (u *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	u.Time = time.Unix(timestamp/1000, 0)
	return nil
}

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
