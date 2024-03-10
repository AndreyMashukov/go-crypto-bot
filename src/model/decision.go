package model

const OrderBasedStrategyName = "order_based_strategy"
const MarketDepthStrategyName = "market_depth_strategy"
const BaseKlineStrategyName = "base_kline_strategy"
const SmaTradeStrategyName = "sma_trade_strategy"
const DecisionHighestPriorityScore = 999.99

type Decision struct {
	Operation    string     `json:"operation"`
	Timestamp    int64      `json:"timestamp"`
	StrategyName string     `json:"strategyName"`
	Score        float64    `json:"score"`
	Price        float64    `json:"price"`
	Params       [3]float64 `json:"params"`
}

type FacadeResponse struct {
	Hold float64
	Buy  float64
	Sell float64
}
