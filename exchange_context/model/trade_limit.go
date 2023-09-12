package exchange_context

type TradeLimit struct {
	Id               int64
	Symbol           string
	USDTLimit        float64
	MinPrice         float64
	MinQuantity      float64
	MinProfitPercent float64
	IsEnabled        bool

	// Extra budget for market fall
	USDTExtraBudget  float64
	BuyOnFallPercent float64
}
