package model

type TradeStackItem struct {
	Percent    Percent `json:"percent"`
	Symbol     string  `json:"symbol"`
	BudgetUsdt float64 `json:"budgetUsdt"`
}
