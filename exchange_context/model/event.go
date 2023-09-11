package exchange_context

import (
	"encoding/json"
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
