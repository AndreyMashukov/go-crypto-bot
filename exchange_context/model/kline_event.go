package model

type KlineEvent struct {
	Stream    string    `json:"stream"`
	KlineData KlineData `json:"data"`
}
