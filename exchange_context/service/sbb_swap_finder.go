package service

import (
	"crypto/md5"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"io"
	"time"
)

type SBBSwapFinder struct {
	ExchangeRepository ExchangeRepository.SwapPairRepositoryInterface
	Formatter          *Formatter
}

func (s *SBBSwapFinder) Find(asset string) model.BBSArbitrageChain {
	initialBalance := 100.00

	transitions := make([]model.SwapTransition, 0)
	chain := model.BBSArbitrageChain{
		Transitions: make([]model.SwapTransition, 0),
		BestChain:   nil,
	}

	options0 := s.ExchangeRepository.GetSwapPairsByBaseAsset(asset)

	var bestChain *model.BestSwapChain = nil

	for _, option0 := range options0 {
		if option0.IsPriceExpired() {
			continue
		}

		// Do not validate first order for gainer/looser and bull/bear

		option0Price := option0.SellPrice
		option0Price = s.Formatter.FormatPrice(option0, option0Price)
		//log.Printf("[%s] formatted [3] %f -> %f", option0.Symbol, option0.SellPrice, option0Price)
		sell0Quantity := initialBalance //s.Formatter.FormatQuantity(option0, initialBalance)

		sell0 := model.SwapTransition{
			Symbol:        option0.Symbol,
			Type:          model.SwapTransitionTypeSellBuyBuy,
			BaseAsset:     asset,
			QuoteAsset:    option0.QuoteAsset,
			Operation:     model.SwapTransitionOperationTypeSell,
			BaseQuantity:  sell0Quantity,
			QuoteQuantity: 0.00,
			Price:         option0Price,
			Balance:       (sell0Quantity * option0Price) - (sell0Quantity*option0Price)*0.002,
			Level:         0,
			Transitions:   make([]model.SwapTransition, 0),
		}
		// log.Printf("[%s] 1 BUY %s -> %s | %f x %f = %f", asset, sell0.BaseAsset, sell0.QuoteAsset, sell0.BaseQuantity, sell0.Price, sell0.Balance)
		options1 := s.ExchangeRepository.GetSwapPairsByQuoteAsset(sell0.QuoteAsset)
		for _, option1 := range options1 {
			if option1.IsPriceExpired() {
				continue
			}

			if option1.BaseAsset == asset {
				continue
			}

			if !option1.IsBearMarket() && !option1.IsLooser() {
				continue
			}

			option1Price := option1.BuyPrice
			option1Price = s.Formatter.FormatPrice(option1, option1Price)
			//log.Printf("[%s] formatted [4] %f -> %f", option1.Symbol, option1.BuyPrice, option1Price)
			buy0Quantity := sell0.Balance //s.Formatter.FormatQuantity(option1, sell0.Balance)

			buy0 := model.SwapTransition{
				Symbol:        option1.Symbol,
				Type:          model.SwapTransitionTypeSellBuyBuy,
				BaseAsset:     option1.BaseAsset,
				QuoteAsset:    option1.QuoteAsset,
				Operation:     model.SwapTransitionOperationTypeBuy,
				BaseQuantity:  buy0Quantity,
				QuoteQuantity: 0.00,
				Price:         option1Price,
				Balance:       (buy0Quantity / option1Price) - (buy0Quantity/option1Price)*0.002,
				Level:         1,
				Transitions:   make([]model.SwapTransition, 0),
			}
			// log.Printf("[%s] 2 BUY %s -> %s | %f x %f = %f", asset, buy0.BaseAsset, buy0.QuoteAsset, buy0.BaseQuantity, buy0.Price, buy0.Balance)
			options2 := s.ExchangeRepository.GetSwapPairsByQuoteAsset(option1.BaseAsset)
			for _, option2 := range options2 {
				if option2.IsPriceExpired() {
					continue
				}

				if option2.BaseAsset != asset {
					continue
				}

				if !option2.IsBearMarket() && !option2.IsLooser() {
					continue
				}

				option2Price := option2.BuyPrice
				option2Price = s.Formatter.FormatPrice(option2, option2Price)
				//log.Printf("[%s] formatted [5] %f -> %f", option2.Symbol, option2.BuyPrice, option2Price)
				buy1Quantity := buy0.Balance //s.Formatter.FormatQuantity(option2, buy0.Balance)

				sellBalance := (buy1Quantity / option2Price) - (buy1Quantity/option2Price)*0.002

				buy1 := model.SwapTransition{
					Symbol:        option2.Symbol,
					Type:          model.SwapTransitionTypeSellBuyBuy,
					BaseAsset:     asset,
					QuoteAsset:    buy0.BaseAsset,
					Operation:     model.SwapTransitionOperationTypeBuy,
					BaseQuantity:  0.00,
					QuoteQuantity: buy1Quantity,
					Price:         option2Price,
					Balance:       sellBalance,
					Level:         2,
					Transitions:   make([]model.SwapTransition, 0),
				}

				profit := model.Percent(s.Formatter.ToFixed(s.Formatter.ComparePercentage(sell0.BaseQuantity, sellBalance).Value()-100, 2))

				if bestChain == nil || profit.Gt(bestChain.Percent) {
					title := fmt.Sprintf(
						"%s sell-> %s buy-> %s buy-> %s",
						asset,
						sell0.QuoteAsset,
						buy0.BaseAsset,
						buy1.BaseAsset,
					)

					h := md5.New()
					_, _ = io.WriteString(h, title)

					sbbChain := model.BestSwapChain{
						Type:      model.SwapTransitionTypeSellBuyBuy,
						Title:     title,
						Hash:      fmt.Sprintf("%x", h.Sum(nil)),
						SwapOne:   &sell0,
						SwapTwo:   &buy0,
						SwapThree: &buy1,
						Percent:   profit,
						Timestamp: time.Now().Unix(),
					}
					bestChain = &sbbChain
				}

				buy0.Transitions = append(buy0.Transitions, buy1)
			}

			if len(buy0.Transitions) == 0 {
				continue
			}

			sell0.Transitions = append(sell0.Transitions, buy0)
		}

		if len(sell0.Transitions) == 0 {
			continue
		}

		transitions = append(transitions, sell0)
	}

	if bestChain == nil {
		return chain
	}

	chain.Transitions = transitions
	chain.BestChain = bestChain

	return chain
}
