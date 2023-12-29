package service

import (
	"crypto/md5"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"io"
	"time"
)

type BBSSwapFinder struct {
	ExchangeRepository ExchangeRepository.SwapPairRepositoryInterface
	Formatter          *Formatter
}

func (s *BBSSwapFinder) Find(quote string) model.BBSArbitrageChain {
	initialBalance := 100.00

	transitions := make([]model.SwapTransition, 0)
	chain := model.BBSArbitrageChain{
		Transitions: make([]model.SwapTransition, 0),
		BestChain:   nil,
	}

	options0 := s.ExchangeRepository.GetSwapPairsByQuoteAsset(quote, "")

	var bestChain *model.BestSwapChain = nil

	for _, option0 := range options0 {
		if option0.IsPriceExpired() {
			continue
		}

		// Do not validate first order for gainer/looser and bull/bear

		option0Price := option0.BuyPrice - (option0.MinPrice * 2)
		option0Price = s.Formatter.FormatPrice(option0, option0Price)
		//log.Printf("[%s] formatted [1] %f -> %f", option0.Symbol, option0.BuyPrice, option0Price)
		buy0Quantity := initialBalance //s.Formatter.FormatQuantity(option0, initialBalance)

		buy0 := model.SwapTransition{
			Symbol:        option0.Symbol,
			Type:          model.SwapTransitionTypeBuyBuySell,
			BaseAsset:     option0.BaseAsset,
			QuoteAsset:    quote,
			Operation:     model.SwapTransitionOperationTypeBuy,
			BaseQuantity:  buy0Quantity,
			QuoteQuantity: 0.00,
			Price:         option0Price,
			Balance:       (buy0Quantity / option0Price) - (buy0Quantity/option0Price)*0.002,
			Level:         0,
			Transitions:   make([]model.SwapTransition, 0),
		}
		//log.Printf("[%s] 1 BUY %s -> %s | %f x %f = %f", quote, buy0.BaseAsset, buy0.QuoteAsset, buy0.BaseQuantity, buy0.Price, buy0.Balance)
		options1 := s.ExchangeRepository.GetSwapPairsByQuoteAsset(buy0.BaseAsset, "USDT")
		for _, option1 := range options1 {
			if option1.IsPriceExpired() {
				continue
			}

			if option1.BaseAsset == quote {
				continue
			}

			option1Price := option1.BuyPrice - (option1.MinPrice * 2)
			option1Price = s.Formatter.FormatPrice(option1, option1Price)
			//log.Printf("[%s] formatted [1] %f -> %f", option1.Symbol, option1.BuyPrice, option1Price)
			buy1Quantity := buy0.Balance //s.Formatter.FormatQuantity(option1, buy0.Balance)

			buy1 := model.SwapTransition{
				Symbol:        option1.Symbol,
				Type:          model.SwapTransitionTypeBuyBuySell,
				BaseAsset:     option1.BaseAsset,
				QuoteAsset:    option1.QuoteAsset,
				Operation:     model.SwapTransitionOperationTypeBuy,
				BaseQuantity:  buy1Quantity,
				QuoteQuantity: 0.00,
				Price:         option1Price,
				Balance:       (buy1Quantity / option1Price) - (buy1Quantity/option1Price)*0.002,
				Level:         1,
				Transitions:   make([]model.SwapTransition, 0),
			}
			//log.Printf("[%s] 2 BUY %s -> %s | %f x %f = %f", quote, buy1.BaseAsset, buy1.QuoteAsset, buy1.BaseQuantity, buy1.Price, buy1.Balance)
			options2 := s.ExchangeRepository.GetSwapPairsByBaseAsset(option1.BaseAsset, "")
			for _, option2 := range options2 {
				if option2.IsPriceExpired() {
					continue
				}

				if option2.QuoteAsset != quote {
					continue
				}

				//if !option2.IsBearMarket() && !option2.IsLooser() {
				//	continue
				//}

				option2Price := option2.SellPrice - (option2.MinPrice * 10)
				option2Price = s.Formatter.FormatPrice(option2, option2Price)
				//log.Printf("[%s] formatted [2] %f -> %f", option2.Symbol, option2.BuyPrice, option2Price)
				sell1Quantity := buy1.Balance //s.Formatter.FormatQuantity(option2, buy1.Balance)

				sellBalance := (sell1Quantity * option2Price) - (sell1Quantity*option2Price)*0.002

				sell0 := model.SwapTransition{
					Symbol:        option2.Symbol,
					Type:          model.SwapTransitionTypeBuyBuySell,
					BaseAsset:     option2.BaseAsset,
					QuoteAsset:    quote,
					Operation:     model.SwapTransitionOperationTypeSell,
					BaseQuantity:  0.00,
					QuoteQuantity: sell1Quantity,
					Price:         option2Price,
					Balance:       sellBalance,
					Level:         2,
					Transitions:   make([]model.SwapTransition, 0),
				}

				profit := model.Percent(s.Formatter.ToFixed(s.Formatter.ComparePercentage(buy0.BaseQuantity, sellBalance).Value()-100, 2))

				if bestChain == nil || profit.Gt(bestChain.Percent) {
					title := fmt.Sprintf(
						"%s buy-> %s buy-> %s sell-> %s",
						quote,
						buy0.BaseAsset,
						buy1.BaseAsset,
						sell0.QuoteAsset,
					)

					//log.Printf(
					//	"[%s] Swap chain statistics: before = %.8f after = %.8f, percent = %.2f",
					//	title,
					//	initialBalance,
					//	sell0.Balance,
					//	profit,
					//)

					h := md5.New()
					_, _ = io.WriteString(h, title)

					bssChain := model.BestSwapChain{
						Type:      model.SwapTransitionTypeBuyBuySell,
						Title:     title,
						Hash:      fmt.Sprintf("%x", h.Sum(nil)),
						SwapOne:   &buy0,
						SwapTwo:   &buy1,
						SwapThree: &sell0,
						Percent:   profit,
						Timestamp: time.Now().Unix(),
					}
					bestChain = &bssChain
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
