package model

import (
	"encoding/json"
	"strconv"
)

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

type Balance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free,string"`
	Locked float64 `json:"locked,string"`
}

type AccountStatus struct {
	// "makerCommission": 15,
	//    "takerCommission": 15,
	//    "buyerCommission": 0,
	//    "sellerCommission": 0,
	//    "canTrade": true,
	//    "canWithdraw": true,
	//    "canDeposit": true,
	//    "commissionRates": {
	//      "maker": "0.00150000",
	//      "taker": "0.00150000",
	//      "buyer": "0.00000000",
	//      "seller":"0.00000000"
	//    },
	//    "brokered": false,
	//    "requireSelfTradePrevention": false,
	//    "preventSor": false,
	//    "updateTime": 1660801833000,
	//    "accountType": "SPOT",
	//    "balances": [
	//      {
	//        "asset": "BNB",
	//        "free": "0.00000000",
	//        "locked": "0.00000000"
	//      },
	//      {
	//        "asset": "BTC",
	//        "free": "1.3447112",
	//        "locked": "0.08600000"
	//      },
	//      {
	//        "asset": "USDT",
	//        "free": "1021.21000000",
	//        "locked": "0.00000000"
	//      }
	//    ],
	//    "permissions": [
	//      "SPOT"
	//    ],
	//    "uid": 354937868
	Balances []Balance `json:"balances"`
}

type AccountStatusResponse struct {
	Id     string        `json:"id"`
	Status int64         `json:"status"`
	Result AccountStatus `json:"result"`
	Error  *Error        `json:"error"`
}

type KLineHistory struct {
	OpenTime                 int64  `json:"openTime"`
	Open                     string `json:"open"`
	High                     string `json:"high"`
	Low                      string `json:"low"`
	Close                    string `json:"close"`
	Volume                   string `json:"volume"`
	CloseTime                int64  `json:"closeTime"`
	QuoteAssetVolume         string `json:"quoteAssetVolume"`
	TradesNumber             int64  `json:"tradesNumber"`
	TakerBuyBaseAssetVolume  string `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume string `json:"TakerBuyQuoteAssetVolume"`
	UnusedField              string `json:"_"`
}

func (k *KLineHistory) GetHighPrice() float64 {
	value, _ := strconv.ParseFloat(k.High, 64)

	return value
}

func (k *KLineHistory) GetLowPrice() float64 {
	value, _ := strconv.ParseFloat(k.Low, 64)

	return value
}

func (k *KLineHistory) UnmarshalJSON(data []byte) error {
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
		&k.CloseTime,
		&k.QuoteAssetVolume,
		&k.TradesNumber,
		&k.TakerBuyBaseAssetVolume,
		&k.TakerBuyQuoteAssetVolume,
		&k.UnusedField,
	}

	for i := 0; i < len(s); i++ {
		if err := json.Unmarshal(s[i], dest[i]); err != nil {
			return err
		}
	}

	return nil
}

type BinanceKLineResponse struct {
	Id     string         `json:"id"`
	Status int64          `json:"status"`
	Result []KLineHistory `json:"result"`
	Error  *Error         `json:"error"`
}
