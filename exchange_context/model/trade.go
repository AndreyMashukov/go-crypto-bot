package model

type Trade struct {
	Price        float64 `json:"p,string"`
	Symbol       string  `json:"s"`
	Quantity     float64 `json:"q,string"`
	IsBuyerMaker bool    `json:"m"` // IsBuyerMaker = true -> SELL / IsBuyerMaker = false -> BUY
	Timestamp    int     `json:"T,int"`
}

func (c *Trade) GetOperation() string {
	if c.IsBuyerMaker {
		return "SELL"
	}

	return "BUY"
}
