package model

import "fmt"

// SwapTransitionEntity (Entity)
type SwapTransitionEntity struct {
	Id         int64   `json:"id"`
	Type       string  `json:"type"`
	Symbol     string  `json:"symbol"`
	BaseAsset  string  `json:"baseAsset"`
	QuoteAsset string  `json:"quoteAsset"`
	Operation  string  `json:"operation"`
	Quantity   float64 `json:"quoteQuantity"`
	Price      float64 `json:"price"`
	Level      int64   `json:"level"`
}

func (s *SwapTransitionEntity) GetSymbol() string {
	return fmt.Sprintf("%s%s", s.BaseAsset, s.QuoteAsset)
}

func (s *SwapTransitionEntity) IsBuy() bool {
	return s.Type == SwapTransitionOperationTypeBuy
}

func (s *SwapTransitionEntity) IsSell() bool {
	return s.Type == SwapTransitionOperationTypeSell
}

const SwapTransitionTypeBuyBuySell = "BBS"
const SwapTransitionOperationTypeSell = "SELL"
const SwapTransitionOperationTypeBuy = "BUY"

// SwapChainEntity (Entity)
type SwapChainEntity struct {
	Id        int64                 `json:"id"`
	Title     string                `json:"title"`
	Type      string                `json:"type"`
	Hash      string                `json:"hash"`
	SwapOne   *SwapTransitionEntity `json:"swapOne"`
	SwapTwo   *SwapTransitionEntity `json:"swapTwo"`
	SwapThree *SwapTransitionEntity `json:"swapThree"`
	Percent   Percent               `json:"percent"`
	Timestamp int64                 `json:"timestamp"`
}

func (s SwapChainEntity) IsBBS() bool {
	return s.Type == SwapTransitionTypeBuyBuySell
}
