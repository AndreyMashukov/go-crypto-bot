package model

type Bot struct {
	Id            int64  `json:"id"`
	BotUuid       string `json:"botUuid"`
	IsMasterBot   bool   `json:"isMasterBot"`
	IsSwapEnabled bool   `json:"isSwapEnabled"`
}

type BotConfigUpdate struct {
	IsMasterBot   bool `json:"isMasterBot"`
	IsSwapEnabled bool `json:"isSwapEnabled"`
}
