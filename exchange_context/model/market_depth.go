package model

type MarketDepth struct {
	LastUpdateId int64       `json:"lastUpdateId"`
	Bids         [][2]Number `json:"bids"`
	Asks         [][2]Number `json:"asks"`
}
