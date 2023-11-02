package model

type SocketRequest struct {
	Id     string         `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"msg"`
}

type BinanceOrderResponse struct {
	Id     string       `json:"id"`
	Status int64        `json:"status"`
	Result BinanceOrder `json:"result"`
	Error  *Error       `json:"error"`
}
