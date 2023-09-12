package model

type Trade struct {
	Price        float64  `json:"p,string"`
	Symbol       string   `json:"s"`
	Quantity     float64  `json:"q,string"`
	IsBuyerMaker bool     `json:"m"` // IsBuyerMaker = true -> SELL / IsBuyerMaker = false -> BUY
	Timestamp    UnixTime `json:"T"`
}

func (c *Trade) GetOperation() string {
	if c.IsBuyerMaker {
		return "SELL"
	}

	return "BUY"
}

func (c *Trade) GetDate() string {
	return c.Timestamp.Time.Format("2006-01-02 15:04:05")
}
