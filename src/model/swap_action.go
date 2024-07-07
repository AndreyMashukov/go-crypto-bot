package model

import "strings"

const SwapActionStatusPending = "pending"
const SwapActionStatusProcess = "process"
const SwapActionStatusCanceled = "canceled"
const SwapActionStatusSuccess = "success"

type SwapContainer struct {
	SwapAction SwapActionExtended            `json:"action"`
	PriceMap   map[string]map[string]float64 `json:"priceMap"`
	Balance    map[string]Balance            `json:"balance"`
}

// SwapActionExtended Note: This is how to do class extension in Go
type SwapActionExtended struct {
	SwapAction
	PriceOneSell   float64 `json:"priceOneSell"`
	PriceOneBuy    float64 `json:"priceOneBuy"`
	PriceTwoSell   float64 `json:"priceTwoSell"`
	PriceTwoBuy    float64 `json:"priceTwoBuy"`
	PriceThreeSell float64 `json:"priceThreeSell"`
	PriceThreeBuy  float64 `json:"priceThreeBuy"`
}

type SwapAction struct {
	Id                      int64    `json:"id"`
	OrderId                 int64    `json:"orderId"`
	BotId                   int64    `json:"botId"`
	SwapChainId             int64    `json:"swapChainId"`
	Asset                   string   `json:"asset"`
	Status                  string   `json:"status"`
	StartTimestamp          int64    `json:"startTimestamp"`
	StartQuantity           float64  `json:"startQuantity"`
	EndTimestamp            *int64   `json:"endTimestamp"`
	EndQuantity             *float64 `json:"endQuantity"`
	SwapOneSide             *string  `json:"swapOneSide"`
	SwapOneQuantity         *float64 `json:"swapOneQuantity"`
	SwapOneSymbol           string   `json:"swapOneSymbol"`
	SwapOnePrice            float64  `json:"swapOnePrice"`
	SwapOneExternalId       *string  `json:"swapOneExternalId"`
	SwapOneExternalStatus   *string  `json:"swapOneExternalStatus"`
	SwapOneTimestamp        *int64   `json:"swapOneTimestamp"`
	SwapTwoSide             *string  `json:"swapTwoSide"`
	SwapTwoQuantity         *float64 `json:"swapTwoQuantity"`
	SwapTwoSymbol           string   `json:"swapTwoSymbol"`
	SwapTwoPrice            float64  `json:"swapTwoPrice"`
	SwapTwoExternalId       *string  `json:"swapTwoExternalId"`
	SwapTwoExternalStatus   *string  `json:"swapTwoExternalStatus"`
	SwapTwoTimestamp        *int64   `json:"swapTwoTimestamp"`
	SwapThreeSide           *string  `json:"swapThreeSide"`
	SwapThreeQuantity       *float64 `json:"swapThreeQuantity"`
	SwapThreeSymbol         string   `json:"swapThreeSymbol"`
	SwapThreePrice          float64  `json:"swapThreePrice"`
	SwapThreeExternalId     *string  `json:"swapThreeExternalId"`
	SwapThreeExternalStatus *string  `json:"swapThreeExternalStatus"`
	SwapThreeTimestamp      *int64   `json:"swapThreeTimestamp"`
}

func (a *SwapAction) IsPending() bool {
	return a.Status == SwapActionStatusPending
}

func (a *SwapAction) IsOneExpired() bool {
	return *a.SwapOneExternalStatus == "EXPIRED" || *a.SwapOneExternalStatus == "EXPIRED_IN_MATCH"
}

func (a *SwapAction) IsOneCanceled() bool {
	return *a.SwapOneExternalStatus == "CANCELED"
}

func (a *SwapAction) IsTwoExpired() bool {
	return *a.SwapTwoExternalStatus == "EXPIRED" || *a.SwapTwoExternalStatus == "EXPIRED_IN_MATCH"
}

func (a *SwapAction) IsTwoCanceled() bool {
	return *a.SwapTwoExternalStatus == "CANCELED"
}

func (a *SwapAction) IsThreeExpired() bool {
	return *a.SwapThreeExternalStatus == "EXPIRED" || *a.SwapThreeExternalStatus == "EXPIRED_IN_MATCH"
}

func (a *SwapAction) IsThreeCanceled() bool {
	return *a.SwapThreeExternalStatus == "CANCELED"
}

func (a *SwapAction) GetAssetTwo() string {
	return strings.ReplaceAll(a.SwapOneSymbol, a.Asset, "")
}

func (a *SwapAction) GetAssetThree() string {
	return strings.ReplaceAll(a.SwapTwoSymbol, a.GetAssetTwo(), "")
}
