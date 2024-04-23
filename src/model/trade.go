package model

type Trade struct {
	AggregateTradeId int64   `json:"a,int"`
	Price            float64 `json:"p,string"`
	Symbol           string  `json:"s"`
	Quantity         float64 `json:"q,string"`
	IsBuyerMaker     bool    `json:"m,bool"` // IsBuyerMaker = true -> SELL / IsBuyerMaker = false -> BUY
	Timestamp        int64   `json:"T,int"`
	Ignore           bool    `json:"M,bool"`
}

func (c *Trade) GetOperation() string {
	if c.IsSell() {
		return "SELL"
	}

	return "BUY"
}

func (c *Trade) IsSell() bool {
	return c.IsBuyerMaker == true
}

func (c *Trade) IsBuy() bool {
	return c.IsBuyerMaker == false
}

type TradeVolume struct {
	Symbol     string  `json:"symbol"`
	Timestamp  int64   `json:"timestamp"`
	PeriodFrom int64   `json:"periodFrom"`
	PeriodTo   int64   `json:"periodTo"`
	BuyQty     float64 `json:"buyQty"`
	SellQty    float64 `json:"sellQty"`
}
