package service

import (
	"crypto/md5"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"io"
	"log"
	"time"
)

type SwapManager struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	SwapRepository     *ExchangeRepository.SwapRepository
	Formatter          *Formatter
}

func (s *SwapManager) CalculateSwapOptions(symbol string) {
	buyBuySell := s.BuyBuySell(symbol)

	asset := symbol[:len(symbol)-3]

	if buyBuySell.BestChain != nil && buyBuySell.BestChain.Percent.Gte(0.40) {
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
		swapChainEntity, err := s.SwapRepository.GetSwapChain(buyBuySell.BestChain.Hash)
		var swapChainId int64 = 0
		var swapOneId int64 = 0
		var swapOneTwo int64 = 0
		var swapOneThree int64 = 0
		if err == nil {
			swapChainId = swapChainEntity.Id
			swapOneId = swapChainEntity.SwapOne.Id
			swapOneTwo = swapChainEntity.SwapTwo.Id
			swapOneThree = swapChainEntity.SwapThree.Id
		}

		swapChainEntity = model.SwapChainEntity{
			Id:        swapChainId,
			Title:     buyBuySell.BestChain.Title,
			Type:      buyBuySell.BestChain.Type,
			Hash:      buyBuySell.BestChain.Hash,
			Percent:   buyBuySell.BestChain.Percent,
			Timestamp: buyBuySell.BestChain.Timestamp,
			SwapOne: &model.SwapTransitionEntity{
				Id:   swapOneId,
				Type: buyBuySell.BestChain.BuyOne.Type,
				Symbol: fmt.Sprintf(
					"%s%s",
					buyBuySell.BestChain.BuyOne.BaseAsset,
					buyBuySell.BestChain.BuyOne.QuoteAsset,
				),
				BaseAsset:  buyBuySell.BestChain.BuyOne.BaseAsset,
				QuoteAsset: buyBuySell.BestChain.BuyOne.QuoteAsset,
				Operation:  buyBuySell.BestChain.BuyOne.Operation,
				Quantity:   buyBuySell.BestChain.BuyOne.BaseQuantity,
				Price:      buyBuySell.BestChain.BuyOne.Price,
				Level:      buyBuySell.BestChain.BuyOne.Level,
			},
			SwapTwo: &model.SwapTransitionEntity{
				Id:   swapOneTwo,
				Type: buyBuySell.BestChain.BuyTwo.Type,
				Symbol: fmt.Sprintf(
					"%s%s",
					buyBuySell.BestChain.BuyTwo.BaseAsset,
					buyBuySell.BestChain.BuyTwo.QuoteAsset,
				),
				BaseAsset:  buyBuySell.BestChain.BuyTwo.BaseAsset,
				QuoteAsset: buyBuySell.BestChain.BuyTwo.QuoteAsset,
				Operation:  buyBuySell.BestChain.BuyTwo.Operation,
				Quantity:   buyBuySell.BestChain.BuyTwo.BaseQuantity,
				Price:      buyBuySell.BestChain.BuyTwo.Price,
				Level:      buyBuySell.BestChain.BuyTwo.Level,
			},
			SwapThree: &model.SwapTransitionEntity{
				Id:   swapOneThree,
				Type: buyBuySell.BestChain.SellOne.Type,
				Symbol: fmt.Sprintf(
					"%s%s",
					buyBuySell.BestChain.SellOne.BaseAsset,
					buyBuySell.BestChain.SellOne.QuoteAsset,
				),
				BaseAsset:  buyBuySell.BestChain.SellOne.BaseAsset,
				QuoteAsset: buyBuySell.BestChain.SellOne.QuoteAsset,
				Operation:  buyBuySell.BestChain.SellOne.Operation,
				Quantity:   buyBuySell.BestChain.SellOne.QuoteQuantity,
				Price:      buyBuySell.BestChain.SellOne.Price,
				Level:      buyBuySell.BestChain.SellOne.Level,
			},
		}

		if swapChainId > 0 {
			_ = s.SwapRepository.UpdateSwapChain(swapChainEntity)
		} else {
			_, _ = s.SwapRepository.CreateSwapChain(swapChainEntity)
		}

		// Set to cache, will be read in MakerService
		s.SwapRepository.SaveSwapChainCache(swapChainEntity.SwapOne.BaseAsset, swapChainEntity)
	}
}

func (s *SwapManager) BuyBuySell(symbol string) BBSArbitrageChain {
	asset := symbol[:len(symbol)-3]
	balance := 100000.00

	transitions := make([]SwapTransition, 0)
	chain := BBSArbitrageChain{
		Transitions: make([]SwapTransition, 0),
		BestChain:   nil,
	}

	options0 := s.ExchangeRepository.GetSwapPairsByBaseAsset(asset)

	var bestChain *BuyBuySell = nil

	for _, option0 := range options0 {
		if option0.IsPriceExpired() {
			continue
		}

		option0Price := option0.LastPrice
		// sell two steps less
		option0Price -= option0.MinPrice
		option0Price = s.Formatter.FormatPrice(option0, option0Price)
		buy0Quantity := balance //s.Formatter.FormatQuantity(option0, balance)

		buy0 := SwapTransition{
			Type:          model.SwapTransitionTypeBuyBuySell,
			BaseAsset:     asset,
			QuoteAsset:    option0.QuoteAsset,
			Operation:     model.SwapTransitionOperationTypeBuy,
			BaseQuantity:  buy0Quantity,
			QuoteQuantity: 0.00,
			Price:         option0Price,
			Balance:       (buy0Quantity * option0Price) - (buy0Quantity*option0Price)*0.002,
			Level:         0,
			Transitions:   make([]SwapTransition, 0),
		}
		// log.Printf("[%s] 1 BUY %s -> %s | %f x %f = %f", asset, buy0.BaseAsset, buy0.QuoteAsset, buy0.BaseQuantity, buy0.Price, buy0.Balance)
		options1 := s.ExchangeRepository.GetSwapPairsByBaseAsset(buy0.QuoteAsset)
		for _, option1 := range options1 {
			if option1.IsPriceExpired() {
				continue
			}

			if option1.QuoteAsset == asset {
				continue
			}

			option1Price := option1.LastPrice
			// sell two steps less
			option1Price -= option1.MinPrice * 2
			option1Price = s.Formatter.FormatPrice(option1, option1Price)
			buy1Quantity := buy0.Balance //s.Formatter.FormatQuantity(option1, buy0.Balance)

			buy1 := SwapTransition{
				Type:          model.SwapTransitionTypeBuyBuySell,
				BaseAsset:     option1.BaseAsset,
				QuoteAsset:    option1.QuoteAsset,
				Operation:     model.SwapTransitionOperationTypeBuy,
				BaseQuantity:  buy1Quantity,
				QuoteQuantity: 0.00,
				Price:         option1Price,
				Balance:       (buy1Quantity * option1Price) - (buy1Quantity*option1Price)*0.002,
				Level:         1,
				Transitions:   make([]SwapTransition, 0),
			}
			// log.Printf("[%s] 2 BUY %s -> %s | %f x %f = %f", asset, buy1.BaseAsset, buy1.QuoteAsset, buy1.BaseQuantity, buy1.Price, buy1.Balance)
			options2 := s.ExchangeRepository.GetSwapPairsByBaseAsset(asset)
			for _, option2 := range options2 {
				if option2.IsPriceExpired() {
					continue
				}

				if option2.QuoteAsset != buy1.QuoteAsset {
					continue
				}

				option2Price := option2.LastPrice
				// buy two steps greater
				option2Price += option2.MinPrice * 2
				option2Price = s.Formatter.FormatPrice(option2, option2Price)
				sell1Quantity := buy1.Balance //s.Formatter.FormatQuantity(option2, buy1.Balance)

				sellBalance := (sell1Quantity / option2Price) - (sell1Quantity/option2Price)*0.002

				sell0 := SwapTransition{
					Type:          model.SwapTransitionTypeBuyBuySell,
					BaseAsset:     asset,
					QuoteAsset:    buy1.QuoteAsset,
					Operation:     model.SwapTransitionOperationTypeSell,
					BaseQuantity:  0.00,
					QuoteQuantity: sell1Quantity,
					Price:         option2Price,
					Balance:       sellBalance,
					Level:         2,
					Transitions:   make([]SwapTransition, 0),
				}

				profit := s.Formatter.ComparePercentage(buy0.BaseQuantity, sellBalance) - 100

				if bestChain == nil || profit.Gt(bestChain.Percent) {
					title := fmt.Sprintf(
						"%s buy-> %s buy-> %s sell-> %s",
						asset,
						buy0.QuoteAsset,
						buy1.QuoteAsset,
						sell0.BaseAsset,
					)

					h := md5.New()
					_, _ = io.WriteString(h, title)

					bbsChain := BuyBuySell{
						Type:      model.SwapTransitionTypeBuyBuySell,
						Title:     title,
						Hash:      fmt.Sprintf("%x", h.Sum(nil)),
						BuyOne:    &buy0,
						BuyTwo:    &buy1,
						SellOne:   &sell0,
						Percent:   profit,
						Timestamp: time.Now().Unix(),
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

type BuyBuySell struct {
	Title     string          `json:"title"`
	Type      string          `json:"type"`
	Hash      string          `json:"hash"`
	BuyOne    *SwapTransition `json:"buyOne"`
	BuyTwo    *SwapTransition `json:"buyTwo"`
	SellOne   *SwapTransition `json:"sellOne"`
	Percent   model.Percent   `json:"percent"`
	Timestamp int64           `json:"timestamp"`
}

type BBSArbitrageChain struct {
	Transitions []SwapTransition `json:"transitions"`
	BestChain   *BuyBuySell      `json:"bestChain"`
}