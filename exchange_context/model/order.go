package model

type Order struct {
	Id              int64
	Symbol          string
	Price           float64
	Quantity        float64
	CreatedAt       string
	SellVolume      float64
	BuyVolume       float64
	SmaValue        float64
	Operation       string
	Status          string
	ExternalId      *int64
	ClosedBy        *int64 // order here
	UsedExtraBudget float64
}
