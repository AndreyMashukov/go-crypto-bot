package model

type DepthEvent struct {
	Stream string `json:"stream"`
	Depth  Depth  `json:"data"`
}
