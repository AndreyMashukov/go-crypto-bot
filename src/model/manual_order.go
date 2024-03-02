package model

type ManualOrder struct {
	Operation string  `json:"operation"`
	Price     float64 `json:"price"`
	Symbol    string  `json:"symbol"`
	BotUuid   string  `json:"botUuid"`
}

type UpdateOrderExtraChargeOptions struct {
	OrderId            int64              `json:"orderId"`
	ExtraChargeOptions ExtraChargeOptions `json:"extraChargeOptions"`
}
