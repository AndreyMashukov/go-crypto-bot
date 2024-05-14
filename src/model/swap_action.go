package model

const SwapActionStatusPending = "pending"
const SwapActionStatusProcess = "process"
const SwapActionStatusCanceled = "canceled"
const SwapActionStatusSuccess = "success"

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
	SwapOneSymbol           string   `json:"swapOneSymbol"`
	SwapOnePrice            float64  `json:"swapOnePrice"`
	SwapOneExternalId       *string  `json:"swapOneExternalId"`
	SwapOneExternalStatus   *string  `json:"swapOneExternalStatus"`
	SwapOneTimestamp        *int64   `json:"swapOneTimestamp"`
	SwapTwoSymbol           string   `json:"swapTwoSymbol"`
	SwapTwoPrice            float64  `json:"swapTwoPrice"`
	SwapTwoExternalId       *string  `json:"swapTwoExternalId"`
	SwapTwoExternalStatus   *string  `json:"swapTwoExternalStatus"`
	SwapTwoTimestamp        *int64   `json:"swapTwoTimestamp"`
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
