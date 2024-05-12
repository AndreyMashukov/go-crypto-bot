package model

import (
	"encoding/json"
	"time"
)

type ByBitOrderResultList struct {
	List []ByBitOrder `json:"list"`
}

type ByBitOrderListResponse struct {
	Code    int64                `json:"retCode"`
	Message string               `json:"retMsg"`
	Result  ByBitOrderResultList `json:"result"`
}

type ByBitKeyValueResult struct {
	Code    int64          `json:"retCode"`
	Message string         `json:"retMsg"`
	Result  map[string]any `json:"result"`
}

type ByBitOrderBookResponse struct {
	Code    int64               `json:"retCode"`
	Message string              `json:"retMsg"`
	Result  ByBitOrderBookModel `json:"result"`
}

type ByBitKLineResultList struct {
	List []ByBitKLineHistory `json:"list"`
}

type ByBitKLineHistory struct {
	OpenTime string `json:"openTime"`
	Open     string `json:"open"`
	High     string `json:"high"`
	Low      string `json:"low"`
	Close    string `json:"close"`
	Volume   string `json:"volume"`
	Turnover string `json:"turnover"`
}

func (k *ByBitKLineHistory) UnmarshalJSON(data []byte) error {
	var s []json.RawMessage
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	dest := []interface{}{
		&k.OpenTime,
		&k.Open,
		&k.High,
		&k.Low,
		&k.Close,
		&k.Volume,
		&k.Turnover,
	}

	for i := 0; i < len(s); i++ {
		if err := json.Unmarshal(s[i], dest[i]); err != nil {
			return err
		}
	}

	return nil
}

type ByBitKLineHistoryResponse struct {
	Code    int64                `json:"retCode"`
	Message string               `json:"retMsg"`
	Result  ByBitKLineResultList `json:"result"`
}

type ByBitTradeHistoryList struct {
	List []ByBitTrade `json:"list"`
}

type ByBitTradeHistoryResponse struct {
	Code    int64                 `json:"retCode"`
	Message string                `json:"retMsg"`
	Result  ByBitTradeHistoryList `json:"result"`
}

const ByBitTradeSideBuy = "Buy"
const ByBitTradeSideSell = "Sell"

type ByBitTrade struct {
	ExecId       int64          `json:"execId,string"`
	Symbol       string         `json:"symbol"`
	Price        float64        `json:"price,string"`
	Size         float64        `json:"size,string"`
	Side         string         `json:"side"`
	Time         TimestampMilli `json:"time,string"`
	IsBlockTrade bool           `json:"isBlockTrade"`
}

type ByBitExchangeLotSizeFilter struct {
	BasePrecision  string  `json:"basePrecision"`
	QuotePrecision string  `json:"quotePrecision"`
	MinOrderQty    float64 `json:"minOrderQty,string"`
	MaxOrderQty    float64 `json:"maxOrderQty,string"`
	MinOrderAmt    float64 `json:"minOrderAmt,string"`
	MaxOrderAmt    float64 `json:"maxOrderAmt,string"`
}

type ByBitPriceFilter struct {
	TickSize float64 `json:"tickSize,string"`
}

type ByBitRiskParameters struct {
	LimitParameter  string `json:"limitParameter"`
	MarketParameter string `json:"marketParameter"`
}

type ByBitExchangeInfoList struct {
	List []ByBitExchangeSymbol `json:"list"`
}

type ByBitExchangeInfoResponse struct {
	Code    int64                 `json:"retCode"`
	Message string                `json:"retMsg"`
	Result  ByBitExchangeInfoList `json:"result"`
}

type ByBitExchangeSymbol struct {
	Symbol         string                     `json:"symbol"`
	BaseCoin       string                     `json:"baseCoin"`
	QuoteCoin      string                     `json:"quoteCoin"`
	Innovation     string                     `json:"innovation"`
	Status         string                     `json:"status"`
	MarginTrading  string                     `json:"marginTrading"`
	LotSizeFilter  ByBitExchangeLotSizeFilter `json:"lotSizeFilter"`
	PriceFilter    ByBitPriceFilter           `json:"priceFilter"`
	RiskParameters ByBitRiskParameters        `json:"riskParameters"`
}

type ByBitCoin struct {
	AvailableToBorrow   string  `json:"availableToBorrow"`
	Bonus               string  `json:"bonus"`
	AccruedInterest     string  `json:"accruedInterest"`
	AvailableToWithdraw float64 `json:"availableToWithdraw,string"`
	TotalOrderIM        string  `json:"totalOrderIM"`
	Equity              float64 `json:"equity,string"`
	Free                float64 `json:"free,string"`
	TotalPositionMM     string  `json:"totalPositionMM"`
	UsdValue            string  `json:"usdValue"`
	SpotHedgingQty      string  `json:"spotHedgingQty"`
	UnrealisedPnl       string  `json:"unrealisedPnl"`
	CollateralSwitch    bool    `json:"collateralSwitch"`
	BorrowAmount        string  `json:"borrowAmount"`
	TotalPositionIM     string  `json:"totalPositionIM"`
	WalletBalance       string  `json:"walletBalance"`
	CumRealisedPnl      string  `json:"cumRealisedPnl"`
	Locked              float64 `json:"locked,string"`
	MarginCollateral    bool    `json:"marginCollateral"`
	Coin                string  `json:"coin"`
}

const ByBitAccountTypeUnified = "UNIFIED"
const ByBitAccountTypeContract = "CONTRACT"
const ByBitAccountTypeSpot = "SPOT"

type ByBitBalanceList struct {
	List []ByBitBalance `json:"list"`
}

type ByBitBalanceResponse struct {
	Code    int64            `json:"retCode"`
	Message string           `json:"retMsg"`
	Result  ByBitBalanceList `json:"result"`
}

type ByBitBalance struct {
	TotalEquity            string      `json:"totalEquity"`
	AccountIMRate          string      `json:"accountIMRate"`
	TotalMarginBalance     string      `json:"totalMarginBalance"`
	TotalInitialMargin     string      `json:"totalInitialMargin"`
	AccountType            string      `json:"accountType"`
	TotalAvailableBalance  string      `json:"totalAvailableBalance"`
	AccountMMRate          string      `json:"accountMMRate"`
	TotalPerpUPL           string      `json:"totalPerpUPL"`
	TotalWalletBalance     string      `json:"totalWalletBalance"`
	AccountLTV             string      `json:"accountLTV"`
	TotalMaintenanceMargin string      `json:"totalMaintenanceMargin"`
	Coin                   []ByBitCoin `json:"coin"`
}

type ByBitTickerList struct {
	List []ByBitTicker `json:"list"`
}

type ByBitTickerResponse struct {
	Code    int64           `json:"retCode"`
	Message string          `json:"retMsg"`
	Result  ByBitTickerList `json:"result"`
}

type ByBitTicker struct {
	Symbol        string  `json:"symbol"`
	Bid1Price     string  `json:"bid1Price"`
	Bid1Size      string  `json:"bid1Size"`
	Ask1Price     string  `json:"ask1Price"`
	Ask1Size      string  `json:"ask1Size"`
	LastPrice     string  `json:"lastPrice"`
	PrevPrice24H  string  `json:"prevPrice24h"`
	Price24HPcnt  string  `json:"price24hPcnt"`
	HighPrice24H  string  `json:"highPrice24h"`
	LowPrice24H   string  `json:"lowPrice24h"`
	Turnover24H   string  `json:"turnover24h"`
	Volume24H     string  `json:"volume24h"`
	UsdIndexPrice float64 `json:"usdIndexPrice,string"`
}

type ByBitWsTicker struct {
	Symbol        string  `json:"symbol"`
	LastPrice     string  `json:"lastPrice"`
	HighPrice24H  string  `json:"highPrice24h"`
	LowPrice24H   string  `json:"lowPrice24h"`
	PrevPrice24H  string  `json:"prevPrice24h"`
	Volume24H     float64 `json:"volume24h,string"`
	Turnover24H   float64 `json:"turnover24h,string"`
	Price24HPcnt  string  `json:"price24hPcnt"`
	UsdIndexPrice float64 `json:"usdIndexPrice,string"`
}

func (t *ByBitWsTicker) ToBinanceMiniTicker(time TimestampMilli) MiniTicker {
	return MiniTicker{
		EventTime:        time,
		Symbol:           t.Symbol,
		Close:            t.UsdIndexPrice,
		TotalVolumeQuote: t.Volume24H,
		TotalVolumeAsset: t.Turnover24H,
	}
}

type ByBitWsTickerEvent struct {
	Topic string         `json:"topic"`
	Ts    TimestampMilli `json:"ts"`
	Type  string         `json:"type"`
	Cs    int            `json:"cs"`
	Data  ByBitWsTicker  `json:"data"`
}

type ByBitWsPublicTrade struct {
	I         int64          `json:"i,string"`
	Timestamp TimestampMilli `json:"T"`
	Price     float64        `json:"p,string"`
	Quantity  float64        `json:"v,string"`
	Side      string         `json:"S"`
	Symbol    string         `json:"s"`
	BT        bool           `json:"BT"`
}

func (b *ByBitWsPublicTrade) ToBinanceTrade() Trade {
	return Trade{
		AggregateTradeId: b.I,
		Price:            b.Price,
		Symbol:           b.Symbol,
		Quantity:         b.Quantity,
		IsBuyerMaker:     b.Side == ByBitTradeSideSell,
		Timestamp:        b.Timestamp,
		Ignore:           false,
	}
}

type ByBitWsPublicTradeEvent struct {
	Topic string               `json:"topic"`
	Ts    int64                `json:"ts"`
	Type  string               `json:"type"`
	Data  []ByBitWsPublicTrade `json:"data"`
}

type ByBitWsKline struct {
	Start     TimestampMilli `json:"start"`
	End       TimestampMilli `json:"end"`
	Interval  string         `json:"interval"`
	Open      float64        `json:"open,string"`
	Close     float64        `json:"close,string"`
	High      float64        `json:"high,string"`
	Low       float64        `json:"low,string"`
	Volume    float64        `json:"volume,string"`
	Turnover  float64        `json:"turnover,string"`
	Confirm   bool           `json:"confirm"`
	Timestamp TimestampMilli `json:"timestamp"`
}

func (b *ByBitWsKline) ToBinanceKline(symbol string, interval string) KLine {
	return KLine{
		Symbol:           symbol,
		Open:             b.Open,
		Close:            b.Close,
		Low:              b.Low,
		High:             b.High,
		Volume:           b.Volume,
		Interval:         interval,
		Timestamp:        b.Timestamp,
		OpenTime:         b.Start,
		PriceChangeSpeed: nil,
		TradeVolume:      nil,
	}
}

type ByBitWsKLineEvent struct {
	Type  string         `json:"type"`
	Topic string         `json:"topic"`
	Data  []ByBitWsKline `json:"data"`
	Ts    int64          `json:"ts"`
}

type ByBitWsOrderBook struct {
	Symbol string      `json:"s"`
	Bids   [][2]Number `json:"b"`
	Asks   [][2]Number `json:"a"`
	U      int         `json:"u"`
	Seq    int         `json:"seq"`
}

func (b *ByBitWsOrderBook) ToOrderBookModel() OrderBookModel {
	return OrderBookModel{
		Symbol:    b.Symbol,
		Timestamp: time.Now().UnixMilli(),
		Bids:      b.Bids,
		Asks:      b.Asks,
	}
}

type ByBitWsOrderBookEvent struct {
	Topic string           `json:"topic"`
	Ts    int64            `json:"ts"`
	Type  string           `json:"type"`
	Data  ByBitWsOrderBook `json:"data"`
	Cts   int64            `json:"cts"`
}
