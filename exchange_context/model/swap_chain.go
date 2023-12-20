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

const SwapTransitionTypeSellBuySell = "SBS"
const SwapTransitionTypeSellBuyBuy = "SBB"
const SwapTransitionTypeSellSellBuy = "SSB"
const SwapTransitionOperationTypeSell = "SELL"
const SwapTransitionOperationTypeBuy = "BUY"

// SwapChainEntity (Entity)
type SwapChainEntity struct {
	Id                  int64                 `json:"id"`
	Title               string                `json:"title"`
	Type                string                `json:"type"`
	Hash                string                `json:"hash"`
	SwapOne             *SwapTransitionEntity `json:"swapOne"`
	SwapTwo             *SwapTransitionEntity `json:"swapTwo"`
	SwapThree           *SwapTransitionEntity `json:"swapThree"`
	Percent             Percent               `json:"percent"`
	Timestamp           int64                 `json:"timestamp"`
	MaxPercent          Percent               `json:"maxPercent"`
	MaxPercentTimestamp *int64                `json:"maxPercentTimestamp"`
}

func (s SwapChainEntity) IsSSB() bool {
	return s.Type == SwapTransitionTypeSellSellBuy
}

func (s SwapChainEntity) IsSBB() bool {
	return s.Type == SwapTransitionTypeSellBuyBuy
}

func (s SwapChainEntity) IsSBS() bool {
	return s.Type == SwapTransitionTypeSellBuySell
}

type SwapTransition struct {
	Symbol        string           `json:"symbol"`
	Type          string           `json:"type"`
	BaseAsset     string           `json:"baseAsset"`
	QuoteAsset    string           `json:"quoteAsset"`
	Operation     string           `json:"operation"`
	BaseQuantity  float64          `json:"baseQuantity"`
	QuoteQuantity float64          `json:"quoteQuantity"`
	Price         float64          `json:"price"`
	Balance       float64          `json:"balance"`
	Level         int64            `json:"level"`
	Transitions   []SwapTransition `json:"transitions,omitempty"`
}

type BestSwapChain struct {
	Title     string          `json:"title"`
	Type      string          `json:"type"`
	Hash      string          `json:"hash"`
	SwapOne   *SwapTransition `json:"swapOne"`
	SwapTwo   *SwapTransition `json:"swapTwo"`
	SwapThree *SwapTransition `json:"swapThree"`
	Percent   Percent         `json:"percent"`
	Timestamp int64           `json:"timestamp"`
}

type BBSArbitrageChain struct {
	Transitions []SwapTransition `json:"transitions"`
	BestChain   *BestSwapChain   `json:"bestChain"`
}
