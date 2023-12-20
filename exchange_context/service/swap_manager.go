package service

import (
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"math"
	"time"
)

type SwapManager struct {
	SwapRepository ExchangeRepository.SwapBasicRepositoryInterface
	Formatter      *Formatter
	SBBSwapFinder  *SBBSwapFinder
	SSBSwapFinder  *SSBSwapFinder
	SBSSwapFinder  *SBSSwapFinder
}

func (s *SwapManager) CalculateSwapOptions(asset string) {
	sellBuyBuy := s.SBBSwapFinder.Find(asset)

	if sellBuyBuy.BestChain != nil && sellBuyBuy.BestChain.Percent.Gte(0.10) {
		log.Printf(
			"[%s] Swap Chain Found! %s sell-> %s(%f) buy-> %s(%f) buy-> %s(%f) = %.2f percent profit",
			asset,
			asset,
			sellBuyBuy.BestChain.SwapOne.QuoteAsset,
			sellBuyBuy.BestChain.SwapOne.Price,
			sellBuyBuy.BestChain.SwapTwo.BaseAsset,
			sellBuyBuy.BestChain.SwapTwo.Price,
			sellBuyBuy.BestChain.SwapThree.BaseAsset,
			sellBuyBuy.BestChain.SwapThree.Price,
			sellBuyBuy.BestChain.Percent.Value(),
		)

		swapChainEntity := s.UpdateSwapChain(*sellBuyBuy.BestChain)

		// Set to cache, will be read in MakerService
		s.SwapRepository.SaveSwapChainCache(swapChainEntity.SwapOne.BaseAsset, swapChainEntity)
	}

	sellBuySell := s.SBSSwapFinder.Find(asset)

	if sellBuySell.BestChain != nil && sellBuySell.BestChain.Percent.Gte(0.10) {
		log.Printf(
			"[%s] Swap Chain Found! %s sell-> %s(%f) buy-> %s(%f) sell-> %s(%f) = %.2f percent profit",
			asset,
			asset,
			sellBuySell.BestChain.SwapOne.QuoteAsset,
			sellBuySell.BestChain.SwapOne.Price,
			sellBuySell.BestChain.SwapTwo.BaseAsset,
			sellBuySell.BestChain.SwapTwo.Price,
			sellBuySell.BestChain.SwapThree.BaseAsset,
			sellBuySell.BestChain.SwapThree.Price,
			sellBuySell.BestChain.Percent.Value(),
		)

		swapChainEntity := s.UpdateSwapChain(*sellBuySell.BestChain)

		// Set to cache, will be read in MakerService
		s.SwapRepository.SaveSwapChainCache(swapChainEntity.SwapOne.BaseAsset, swapChainEntity)
	}

	sellSellBuy := s.SSBSwapFinder.Find(asset)

	if sellSellBuy.BestChain != nil && sellSellBuy.BestChain.Percent.Gte(0.10) {
		log.Printf(
			"[%s] Swap Chain Found! %s sell-> %s(%f) sell-> %s(%f) buy-> %s(%f) = %.2f percent profit",
			asset,
			asset,
			sellSellBuy.BestChain.SwapOne.QuoteAsset,
			sellSellBuy.BestChain.SwapOne.Price,
			sellSellBuy.BestChain.SwapTwo.QuoteAsset,
			sellSellBuy.BestChain.SwapTwo.Price,
			sellSellBuy.BestChain.SwapThree.BaseAsset,
			sellSellBuy.BestChain.SwapThree.Price,
			sellSellBuy.BestChain.Percent.Value(),
		)

		swapChainEntity := s.UpdateSwapChain(*sellSellBuy.BestChain)

		// Set to cache, will be read in MakerService
		s.SwapRepository.SaveSwapChainCache(swapChainEntity.SwapOne.BaseAsset, swapChainEntity)
	}
}

func (s *SwapManager) UpdateSwapChain(BestChain model.BestSwapChain) model.SwapChainEntity {
	swapChainEntity, err := s.SwapRepository.GetSwapChain(BestChain.Hash)
	var swapChainId int64 = 0
	var swapOneId int64 = 0
	var swapOneTwo int64 = 0
	var swapOneThree int64 = 0
	var maxPercent model.Percent
	var maxPercentTimestamp *int64 = nil
	nowTimestamp := time.Now().Unix()

	if err == nil {
		swapChainId = swapChainEntity.Id
		swapOneId = swapChainEntity.SwapOne.Id
		swapOneTwo = swapChainEntity.SwapTwo.Id
		swapOneThree = swapChainEntity.SwapThree.Id
		maxPercentTimestamp = swapChainEntity.MaxPercentTimestamp
		if swapChainEntity.MaxPercent.Lt(BestChain.Percent) || swapChainEntity.MaxPercentTimestamp == nil {
			maxPercentTimestamp = &nowTimestamp
		}

		maxPercent = model.Percent(math.Max(swapChainEntity.MaxPercent.Value(), BestChain.Percent.Value()))
	} else {
		maxPercent = BestChain.Percent
		maxPercentTimestamp = &nowTimestamp
	}

	swapChainEntity = model.SwapChainEntity{
		Id:                  swapChainId,
		Title:               BestChain.Title,
		Type:                BestChain.Type,
		Hash:                BestChain.Hash,
		Percent:             BestChain.Percent,
		MaxPercent:          maxPercent,
		Timestamp:           nowTimestamp,
		MaxPercentTimestamp: maxPercentTimestamp,
		SwapOne: &model.SwapTransitionEntity{
			Id:         swapOneId,
			Type:       BestChain.SwapOne.Type,
			Symbol:     BestChain.SwapOne.Symbol,
			BaseAsset:  BestChain.SwapOne.BaseAsset,
			QuoteAsset: BestChain.SwapOne.QuoteAsset,
			Operation:  BestChain.SwapOne.Operation,
			Quantity:   BestChain.SwapOne.BaseQuantity,
			Price:      BestChain.SwapOne.Price,
			Level:      BestChain.SwapOne.Level,
		},
		SwapTwo: &model.SwapTransitionEntity{
			Id:         swapOneTwo,
			Type:       BestChain.SwapTwo.Type,
			Symbol:     BestChain.SwapTwo.Symbol,
			BaseAsset:  BestChain.SwapTwo.BaseAsset,
			QuoteAsset: BestChain.SwapTwo.QuoteAsset,
			Operation:  BestChain.SwapTwo.Operation,
			Quantity:   BestChain.SwapTwo.BaseQuantity,
			Price:      BestChain.SwapTwo.Price,
			Level:      BestChain.SwapTwo.Level,
		},
		SwapThree: &model.SwapTransitionEntity{
			Id:         swapOneThree,
			Type:       BestChain.SwapThree.Type,
			Symbol:     BestChain.SwapThree.Symbol,
			BaseAsset:  BestChain.SwapThree.BaseAsset,
			QuoteAsset: BestChain.SwapThree.QuoteAsset,
			Operation:  BestChain.SwapThree.Operation,
			Quantity:   BestChain.SwapThree.QuoteQuantity,
			Price:      BestChain.SwapThree.Price,
			Level:      BestChain.SwapThree.Level,
		},
	}

	if swapChainId > 0 {
		_ = s.SwapRepository.UpdateSwapChain(swapChainEntity)
	} else {
		_, _ = s.SwapRepository.CreateSwapChain(swapChainEntity)
	}

	return swapChainEntity
}
