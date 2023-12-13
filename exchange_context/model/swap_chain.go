package model

// SwapTransitionEntity (Entity)
type SwapTransitionEntity struct {
	Id         int64   `json:"id"`
	Type       string  `json:"type"`
	Symbol     string  `json:"symbol"`
	BaseAsset  string  `json:"baseAsset"`
	QuoteAsset string  `json:"quoteAsset"`
	Operation  string  `json:"operation"`
	Quantity   float64 `json:"quoteQuantity"`
	Price      float64 `json:"price"`
	Level      int64   `json:"level"`
}

// SwapChainEntity (Entity)
type SwapChainEntity struct {
	Id        int64                 `json:"id"`
	Title     string                `json:"title"`
	Type      string                `json:"type"`
	Hash      string                `json:"hash"`
	SwapOne   *SwapTransitionEntity `json:"swapOne"`
	SwapTwo   *SwapTransitionEntity `json:"swapTwo"`
	SwapThree *SwapTransitionEntity `json:"swapThree"`
	Percent   Percent               `json:"percent"`
	Timestamp int64                 `json:"timestamp"`
}
