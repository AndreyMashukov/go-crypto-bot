package exchange_context

type DepthEvent struct {
	Stream string `json:"stream"`
	Depth  Depth  `json:"data"`
}
