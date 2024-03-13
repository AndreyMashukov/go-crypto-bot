package model

type TradeStackItem struct {
	Index             int64         `json:"index"`
	Price             float64       `json:"price"`
	IsPriceValid      bool          `json:"isPriceValid"`
	Percent           Percent       `json:"percent"`
	Symbol            string        `json:"symbol"`
	BudgetUsdt        float64       `json:"budgetUsdt"`
	HasEnoughBalance  bool          `json:"hasEnoughBalance"`
	BalanceAfter      float64       `json:"balanceAfter"`
	BinanceOrder      *BinanceOrder `json:"binanceOrder"`
	IsExtraCharge     bool          `json:"isExtraCharge"`
	StrategyDecisions []Decision    `json:"strategyDecisions"`
	IsBuyLocked       bool          `json:"isBuyLocked"`
	IsEnabled         bool          `json:"isEnabled"`
	BuyPrice          float64       `json:"buyPrice"`
	PricePointsDiff   int64         `json:"pricePointsDiff"`
}
