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

type Candle struct {
	Price        float64  `json:"p,string"`
	Symbol       string   `json:"s"`
	Quantity     float64  `json:"q,string"`
	IsBuyerMaker bool     `json:"m"` // IsBuyerMaker = true -> SELL / IsBuyerMaker = false -> BUY
	Timestamp    UnixTime `json:"T"`
}

func (c *Candle) GetOperation() string {
	if c.IsBuyerMaker {
		return "SELL"
	}

	return "BUY"
}

func (c *Candle) GetDate() string {
	return c.Timestamp.Time.Format("2006-01-02 15:04:05")
}
