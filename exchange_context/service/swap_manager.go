package service

import (
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"time"
)

type SwapManager struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	BalanceService     *BalanceService
	Formatter          *Formatter
}

func (s *SwapManager) CalculateSwapOptions(symbol string) {
	buyBuySell := s.BuyBuySell(symbol)

	asset := symbol[:len(symbol)-3]

	if buyBuySell.BestChain != nil && buyBuySell.BestChain.Percent.Gte(0.70) {
		log.Printf(
			"[%s] Swap Chain Found! %s buy-> %s(%f) buy-> %s(%f) sell-> %s(%f) = %.2f percent profit",
			asset,
			asset,
			buyBuySell.BestChain.BuyOne.QuoteAsset,
			buyBuySell.BestChain.BuyOne.Price,
			buyBuySell.BestChain.BuyTwo.QuoteAsset,
			buyBuySell.BestChain.BuyTwo.Price,
			buyBuySell.BestChain.SellOne.BaseAsset,
			buyBuySell.BestChain.SellOne.Price,
			buyBuySell.BestChain.Percent.Value(),
		)
	}
}

func (s *SwapManager) BuyBuySell(symbol string) BBSArbitrageChain {
	asset := symbol[:len(symbol)-3]

	balance, err := s.BalanceService.GetAssetBalance(asset)

	transitions := make([]SwapTransition, 0)
	chain := BBSArbitrageChain{
		Transitions: make([]SwapTransition, 0),
		BestChain:   nil,
	}

	if err != nil {
		return chain
	}

	if balance <= 0 {
		time.Sleep(time.Minute * 5)
		return chain
	}

	options0 := s.ExchangeRepository.GetSwapPairsByBaseAsset(asset)

	var bestChain *BuyBuySell = nil

	for _, option0 := range options0 {
		buy0 := SwapTransition{
			BaseAsset:     asset,
			QuoteAsset:    option0.QuoteAsset,
			Operation:     "BUY",
			BaseQuantity:  balance,
			QuoteQuantity: 0.00,
			Price:         option0.LastPrice,
			Balance:       (balance * option0.LastPrice) - (balance*option0.LastPrice)*0.002,
			Level:         0,
			Transitions:   make([]SwapTransition, 0),
		}
		// log.Printf("[%s] 1 BUY %s -> %s | %f x %f = %f", asset, buy0.BaseAsset, buy0.QuoteAsset, buy0.BaseQuantity, buy0.Price, buy0.Balance)
		options1 := s.ExchangeRepository.GetSwapPairsByBaseAsset(buy0.QuoteAsset)
		for _, option1 := range options1 {
			if option1.QuoteAsset == asset {
				continue
			}

			buy1 := SwapTransition{
				BaseAsset:     option1.BaseAsset,
				QuoteAsset:    option1.QuoteAsset,
				Operation:     "BUY",
				BaseQuantity:  buy0.Balance,
				QuoteQuantity: 0.00,
				Price:         option1.LastPrice,
				Balance:       (buy0.Balance * option1.LastPrice) - (buy0.Balance*option1.LastPrice)*0.002,
				Level:         1,
				Transitions:   make([]SwapTransition, 0),
			}
			// log.Printf("[%s] 2 BUY %s -> %s | %f x %f = %f", asset, buy1.BaseAsset, buy1.QuoteAsset, buy1.BaseQuantity, buy1.Price, buy1.Balance)
			options2 := s.ExchangeRepository.GetSwapPairsByBaseAsset(asset)
			for _, option2 := range options2 {
				if option2.QuoteAsset != buy1.QuoteAsset {
					continue
				}

				sellBalance := (buy1.Balance / option2.LastPrice) - (buy1.Balance/option2.LastPrice)*0.002

				sell0 := SwapTransition{
					BaseAsset:     asset,
					QuoteAsset:    buy1.QuoteAsset,
					Operation:     "SELL",
					BaseQuantity:  0.00,
					QuoteQuantity: buy1.Balance,
					Price:         option2.LastPrice,
					Balance:       sellBalance,
					Level:         2,
					Transitions:   make([]SwapTransition, 0),
				}

				profit := s.Formatter.ComparePercentage(buy0.BaseQuantity, sellBalance) - 100

				if bestChain == nil || profit.Gt(bestChain.Percent) {
					bbsChain := BuyBuySell{
						BuyOne:  &buy0,
						BuyTwo:  &buy1,
						SellOne: &sell0,
						Percent: profit,
					}
					bestChain = &bbsChain
				}

				// log.Printf("[%s] 3 SELL %s -> %s | %f x %f = %f", asset, sell0.QuoteAsset, sell0.BaseAsset, sell0.QuoteQuantity, buy1.Price, buy1.Balance)

				buy1.Transitions = append(buy1.Transitions, sell0)
			}

			if len(buy1.Transitions) == 0 {
				continue
			}

			buy0.Transitions = append(buy0.Transitions, buy1)
		}

		if len(buy0.Transitions) == 0 {
			continue
		}

		transitions = append(transitions, buy0)
	}

	if bestChain == nil {
		return chain
	}

	chain.Transitions = transitions
	chain.BestChain = bestChain

	return chain
}

type SwapTransition struct {
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

type BBSArbitrageChain struct {
	Transitions []SwapTransition `json:"transitions"`
	BestChain   *BuyBuySell      `json:"bestChain"`
}

type BuyBuySell struct {
	BuyOne  *SwapTransition `json:"buyOne"`
	BuyTwo  *SwapTransition `json:"buyTwo"`
	SellOne *SwapTransition `json:"sellOne"`
	Percent model.Percent
}
