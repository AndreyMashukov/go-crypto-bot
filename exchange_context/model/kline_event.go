package exchange_context

type KlineEvent struct {
	Stream    string    `json:"stream"`
	KlineData KlineData `json:"data"`
}
