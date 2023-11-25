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

type BinanceOrderListResponse struct {
	Id     string         `json:"id"`
	Status int64          `json:"status"`
	Result []BinanceOrder `json:"result"`
	Error  *Error         `json:"error"`
}

type RateLimit struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int64  `json:"intervalNum"`
	Limit         int64  `json:"limit"`
}

type ExchangeFilter struct {
	FilterType  string   `json:"filterType"`
	MinPrice    *float64 `json:"minPrice,string"`
	MaxPrice    *float64 `json:"maxPrice,string"`
	TickSize    *float64 `json:"tickSize,string"`
	MinQuantity *float64 `json:"minQty,string"`
	MaxQuantity *float64 `json:"maxQty,string"`
	StepSize    *float64 `json:"stepSize,string"`
}

type ExchangeSymbol struct {
	Symbol  string           `json:"symbol"`
	Status  string           `json:"status"`
	Filters []ExchangeFilter `json:"filters"`
}

type ExchangeInfo struct {
	Timezone   string           `json:"timezone"`
	ServerTime int64            `json:"serverTime"`
	RateLimits []RateLimit      `json:"rateLimits"`
	Symbols    []ExchangeSymbol `json:"symbols"`
}

type BinanceExchangeInfoResponse struct {
	Id     string       `json:"id"`
	Status int64        `json:"status"`
	Result ExchangeInfo `json:"result"`
	Error  *Error       `json:"error"`
}

type MyTrade struct {
	Id              int64   `json:"id"`
	OrderId         int64   `json:"orderId"`
	Price           float64 `json:"price,string"`
	Quantity        float64 `json:"qty,string"`
	QuoteQuantity   float64 `json:"quoteQty,string"`
	Commission      float64 `json:"commission,string"`
	CommissionAsset string  `json:"commissionAsset"`
	Time            int64   `json:"time"`
	IsBuyer         bool    `json:"isBuyer"`
	IsMaker         bool    `json:"isMaker"`
	IsBestMatch     bool    `json:"isBestMatch"`
}

type TradesResponse struct {
	Id     string    `json:"id"`
	Status int64     `json:"status"`
	Result []MyTrade `json:"result"`
	Error  *Error    `json:"error"`
}
