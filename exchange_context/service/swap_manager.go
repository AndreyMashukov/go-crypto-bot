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
	sellBuyBuy := s.SellBuyBuy(symbol)
	asset := symbol[:len(symbol)-3]

	if sellBuyBuy.BestChain != nil && sellBuyBuy.BestChain.Percent.Gte(0.10) {
		log.Printf(
			"[%s] Swap Chain Found! %s sell-> %s(%f) buy-> %s(%f) buy-> %s(%f) = %.2f percent profit",
			asset,
			asset,
			sellBuyBuy.BestChain.BuyOne.QuoteAsset,
			sellBuyBuy.BestChain.BuyOne.Price,
			sellBuyBuy.BestChain.BuyTwo.BaseAsset,
			sellBuyBuy.BestChain.BuyTwo.Price,
			sellBuyBuy.BestChain.SellOne.BaseAsset,
			sellBuyBuy.BestChain.SellOne.Price,
			sellBuyBuy.BestChain.Percent.Value(),
		)

		_ = s.UpdateSwapChain(*sellBuyBuy.BestChain)
	}

	sellSellBuy := s.SellSellBuy(symbol)

	if sellSellBuy.BestChain != nil && sellSellBuy.BestChain.Percent.Gte(0.10) {
		log.Printf(
			"[%s] Swap Chain Found! %s sell-> %s(%f) sell-> %s(%f) buy-> %s(%f) = %.2f percent profit",
			asset,
			asset,
			sellSellBuy.BestChain.BuyOne.QuoteAsset,
			sellSellBuy.BestChain.BuyOne.Price,
			sellSellBuy.BestChain.BuyTwo.QuoteAsset,
			sellSellBuy.BestChain.BuyTwo.Price,
			sellSellBuy.BestChain.SellOne.BaseAsset,
			sellSellBuy.BestChain.SellOne.Price,
			sellSellBuy.BestChain.Percent.Value(),
		)

		swapChainEntity := s.UpdateSwapChain(*sellSellBuy.BestChain)

		// Set to cache, will be read in MakerService
		s.SwapRepository.SaveSwapChainCache(swapChainEntity.SwapOne.BaseAsset, swapChainEntity)
	}
}

func (s *SwapManager) UpdateSwapChain(BestChain BestSwapChain) model.SwapChainEntity {
	swapChainEntity, err := s.SwapRepository.GetSwapChain(BestChain.Hash)
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
		Title:     BestChain.Title,
		Type:      BestChain.Type,
		Hash:      BestChain.Hash,
		Percent:   BestChain.Percent,
		Timestamp: BestChain.Timestamp,
		SwapOne: &model.SwapTransitionEntity{
			Id:   swapOneId,
			Type: BestChain.BuyOne.Type,
			Symbol: fmt.Sprintf(
				"%s%s",
				BestChain.BuyOne.BaseAsset,
				BestChain.BuyOne.QuoteAsset,
			),
			BaseAsset:  BestChain.BuyOne.BaseAsset,
			QuoteAsset: BestChain.BuyOne.QuoteAsset,
			Operation:  BestChain.BuyOne.Operation,
			Quantity:   BestChain.BuyOne.BaseQuantity,
			Price:      BestChain.BuyOne.Price,
			Level:      BestChain.BuyOne.Level,
		},
		SwapTwo: &model.SwapTransitionEntity{
			Id:   swapOneTwo,
			Type: BestChain.BuyTwo.Type,
			Symbol: fmt.Sprintf(
				"%s%s",
				BestChain.BuyTwo.BaseAsset,
				BestChain.BuyTwo.QuoteAsset,
			),
			BaseAsset:  BestChain.BuyTwo.BaseAsset,
			QuoteAsset: BestChain.BuyTwo.QuoteAsset,
			Operation:  BestChain.BuyTwo.Operation,
			Quantity:   BestChain.BuyTwo.BaseQuantity,
			Price:      BestChain.BuyTwo.Price,
			Level:      BestChain.BuyTwo.Level,
		},
		SwapThree: &model.SwapTransitionEntity{
			Id:   swapOneThree,
			Type: BestChain.SellOne.Type,
			Symbol: fmt.Sprintf(
				"%s%s",
				BestChain.SellOne.BaseAsset,
				BestChain.SellOne.QuoteAsset,
			),
			BaseAsset:  BestChain.SellOne.BaseAsset,
			QuoteAsset: BestChain.SellOne.QuoteAsset,
			Operation:  BestChain.SellOne.Operation,
			Quantity:   BestChain.SellOne.QuoteQuantity,
			Price:      BestChain.SellOne.Price,
			Level:      BestChain.SellOne.Level,
		},
	}

	if swapChainId > 0 {
		_ = s.SwapRepository.UpdateSwapChain(swapChainEntity)
	} else {
		_, _ = s.SwapRepository.CreateSwapChain(swapChainEntity)
	}

	return swapChainEntity
}

func (s *SwapManager) SellSellBuy(symbol string) BBSArbitrageChain {
	asset := symbol[:len(symbol)-3]
	initialBalance := 100.00

	transitions := make([]SwapTransition, 0)
	chain := BBSArbitrageChain{
		Transitions: make([]SwapTransition, 0),
		BestChain:   nil,
	}

	options0 := s.ExchangeRepository.GetSwapPairsByBaseAsset(asset)

	var bestChain *BestSwapChain = nil

	for _, option0 := range options0 {
		if option0.IsPriceExpired() {
			continue
		}

		option0Price := option0.SellPrice
		// first sell is -2 points
		option0Price -= option0.MinPrice * 2
		option0Price = s.Formatter.FormatPrice(option0, option0Price)
		buy0Quantity := initialBalance //s.Formatter.FormatQuantity(option0, initialBalance)

		buy0 := SwapTransition{
			Type:          model.SwapTransitionTypeSellSellBuy,
			BaseAsset:     asset,
			QuoteAsset:    option0.QuoteAsset,
			Operation:     model.SwapTransitionOperationTypeSell,
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

			option1Price := option1.SellPrice
			// sell -3 points
			option1Price -= option1.MinPrice * 2
			option1Price = s.Formatter.FormatPrice(option1, option1Price)
			buy1Quantity := buy0.Balance //s.Formatter.FormatQuantity(option1, buy0.Balance)

			buy1 := SwapTransition{
				Type:          model.SwapTransitionTypeSellSellBuy,
				BaseAsset:     option1.BaseAsset,
				QuoteAsset:    option1.QuoteAsset,
				Operation:     model.SwapTransitionOperationTypeSell,
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

				option2Price := option2.BuyPrice
				// buy +4 points
				option2Price += option2.MinPrice * 4
				option2Price = s.Formatter.FormatPrice(option2, option2Price)
				sell1Quantity := buy1.Balance //s.Formatter.FormatQuantity(option2, buy1.Balance)

				sellBalance := (sell1Quantity / option2Price) - (sell1Quantity/option2Price)*0.002

				sell0 := SwapTransition{
					Type:          model.SwapTransitionTypeSellSellBuy,
					BaseAsset:     asset,
					QuoteAsset:    buy1.QuoteAsset,
					Operation:     model.SwapTransitionOperationTypeBuy,
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
						"%s sell-> %s sell-> %s buy-> %s",
						asset,
						buy0.QuoteAsset,
						buy1.QuoteAsset,
						sell0.BaseAsset,
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

					bbsChain := BestSwapChain{
						Type:      model.SwapTransitionTypeSellSellBuy,
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

func (s *SwapManager) SellBuyBuy(symbol string) BBSArbitrageChain {
	asset := symbol[:len(symbol)-3]
	initialBalance := 100.00

	transitions := make([]SwapTransition, 0)
	chain := BBSArbitrageChain{
		Transitions: make([]SwapTransition, 0),
		BestChain:   nil,
	}

	options0 := s.ExchangeRepository.GetSwapPairsByBaseAsset(asset)

	var bestChain *BestSwapChain = nil

	for _, option0 := range options0 {
		if option0.IsPriceExpired() {
			continue
		}

		option0Price := option0.SellPrice
		// first sell is -2 points
		option0Price -= option0.MinPrice * 2
		option0Price = s.Formatter.FormatPrice(option0, option0Price)
		buy0Quantity := initialBalance //s.Formatter.FormatQuantity(option0, initialBalance)

		buy0 := SwapTransition{
			Type:          model.SwapTransitionTypeSellBuyBuy,
			BaseAsset:     asset,
			QuoteAsset:    option0.QuoteAsset,
			Operation:     model.SwapTransitionOperationTypeSell,
			BaseQuantity:  buy0Quantity,
			QuoteQuantity: 0.00,
			Price:         option0Price,
			Balance:       (buy0Quantity * option0Price) - (buy0Quantity*option0Price)*0.002,
			Level:         0,
			Transitions:   make([]SwapTransition, 0),
		}
		// log.Printf("[%s] 1 BUY %s -> %s | %f x %f = %f", asset, buy0.BaseAsset, buy0.QuoteAsset, buy0.BaseQuantity, buy0.Price, buy0.Balance)
		options1 := s.ExchangeRepository.GetSwapPairsByQuoteAsset(buy0.QuoteAsset)
		for _, option1 := range options1 {
			if option1.IsPriceExpired() {
				continue
			}

			if option1.BaseAsset == asset {
				continue
			}

			option1Price := option1.BuyPrice
			// buy +2 points
			option1Price += option1.MinPrice * 2
			option1Price = s.Formatter.FormatPrice(option1, option1Price)
			buy1Quantity := buy0.Balance //s.Formatter.FormatQuantity(option1, buy0.Balance)

			buy1 := SwapTransition{
				Type:          model.SwapTransitionTypeSellBuyBuy,
				BaseAsset:     option1.BaseAsset,
				QuoteAsset:    option1.QuoteAsset,
				Operation:     model.SwapTransitionOperationTypeSell,
				BaseQuantity:  buy1Quantity,
				QuoteQuantity: 0.00,
				Price:         option1Price,
				Balance:       (buy1Quantity / option1Price) - (buy1Quantity/option1Price)*0.002,
				Level:         1,
				Transitions:   make([]SwapTransition, 0),
			}
			// log.Printf("[%s] 2 BUY %s -> %s | %f x %f = %f", asset, buy1.BaseAsset, buy1.QuoteAsset, buy1.BaseQuantity, buy1.Price, buy1.Balance)
			options2 := s.ExchangeRepository.GetSwapPairsByBaseAsset(asset)
			for _, option2 := range options2 {
				if option2.IsPriceExpired() {
					continue
				}

				if option2.QuoteAsset != buy1.BaseAsset {
					continue
				}

				option2Price := option2.BuyPrice
				// buy +4 points
				option2Price += option2.MinPrice * 4
				option2Price = s.Formatter.FormatPrice(option2, option2Price)
				sell1Quantity := buy1.Balance //s.Formatter.FormatQuantity(option2, buy1.Balance)

				sellBalance := (sell1Quantity / option2Price) - (sell1Quantity/option2Price)*0.002

				sell0 := SwapTransition{
					Type:          model.SwapTransitionTypeSellBuyBuy,
					BaseAsset:     asset,
					QuoteAsset:    buy1.QuoteAsset,
					Operation:     model.SwapTransitionOperationTypeBuy,
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
						"%s sell-> %s buy-> %s buy-> %s",
						asset,
						buy0.QuoteAsset,
						buy1.BaseAsset,
						sell0.BaseAsset,
					)

					h := md5.New()
					_, _ = io.WriteString(h, title)

					bbsChain := BestSwapChain{
						Type:      model.SwapTransitionTypeSellBuyBuy,
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

type BestSwapChain struct {
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
	BestChain   *BestSwapChain   `json:"bestChain"`
}
