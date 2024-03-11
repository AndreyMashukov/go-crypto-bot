package model

import "strings"

type ManualOrder struct {
	Operation string  `json:"operation"`
	Price     float64 `json:"price"`
	Symbol    string  `json:"symbol"`
	BotUuid   string  `json:"botUuid"`
}

func (m *ManualOrder) IsBuy() bool {
	return strings.ToUpper(m.Operation) == "BUY"
}

func (m *ManualOrder) IsSell() bool {
	return strings.ToUpper(m.Operation) == "SELL"
}

func (m *ManualOrder) CanSell(order Order, withSwap bool) bool {
	profit := order.GetQuoteProfit(m.Price, withSwap)
	minProfit := order.GetQuoteProfit(order.GetManualMinClosePrice(), withSwap)

	return profit >= minProfit
}

type UpdateOrderExtraChargeOptions struct {
	OrderId            int64              `json:"orderId"`
	ExtraChargeOptions ExtraChargeOptions `json:"extraChargeOptions"`
}

type UpdateOrderProfitOptions struct {
	OrderId       int64         `json:"orderId"`
	ProfitOptions ProfitOptions `json:"profitOptions"`
}
