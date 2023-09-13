package model

type Order struct {
	Id              int64   `json:"id"`
	Symbol          string  `json:"symbol"`
	Price           float64 `json:"price"`
	Quantity        float64 `json:"quantity"`
	CreatedAt       string  `json:"createdAt"`
	SellVolume      float64 `json:"sellVolume"`
	BuyVolume       float64 `json:"buyVolume"`
	SmaValue        float64 `json:"smaValue"`
	Operation       string  `json:"operation"`
	Status          string  `json:"status"`
	ExternalId      *int64  `json:"externalId"`
	ClosedBy        *int64  `json:"closedBy"` // sell order here
	UsedExtraBudget float64 `json:"usedExtraBudget"`
}
