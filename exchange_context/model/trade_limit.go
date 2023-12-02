package model

type TradeLimit struct {
	Id               int64   `json:"id"`
	Symbol           string  `json:"symbol"`
	USDTLimit        float64 `json:"USDTLimit"`
	MinPrice         float64 `json:"minPrice"`
	MinQuantity      float64 `json:"minQuantity"`
	MinProfitPercent float64 `json:"minProfitPercent"`
	IsEnabled        bool    `json:"isEnabled"`

	// Extra budget for market fall
	USDTExtraBudget  float64 `json:"USDTExtraBudget"`
	BuyOnFallPercent float64 `json:"buyOnFallPercent"`
}
