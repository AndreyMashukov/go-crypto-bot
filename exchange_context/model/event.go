package exchange_context

type Event struct {
	Stream string `json:"stream"`
	Trade  Trade  `json:"data"`
}
