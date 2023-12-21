package service

import (
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"math"
	"time"
)

type SwapManager struct {
	SwapRepository   ExchangeRepository.SwapBasicRepositoryInterface
	Formatter        *Formatter
	SBBSwapFinder    *SBBSwapFinder
	SSBSwapFinder    *SSBSwapFinder
	SBSSwapFinder    *SBSSwapFinder
	SwapChainBuilder *SwapChainBuilder
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
	var swapTwoId int64 = 0
	var swapThreeId int64 = 0
	var maxPercent model.Percent
	var maxPercentTimestamp *int64 = nil
	nowTimestamp := time.Now().Unix()

	if err == nil {
		swapChainId = swapChainEntity.Id
		swapOneId = swapChainEntity.SwapOne.Id
		swapTwoId = swapChainEntity.SwapTwo.Id
		swapThreeId = swapChainEntity.SwapThree.Id
		maxPercentTimestamp = swapChainEntity.MaxPercentTimestamp
		if swapChainEntity.MaxPercent.Lt(BestChain.Percent) || swapChainEntity.MaxPercentTimestamp == nil {
			maxPercentTimestamp = &nowTimestamp
		}

		maxPercent = model.Percent(math.Max(swapChainEntity.MaxPercent.Value(), BestChain.Percent.Value()))
	} else {
		maxPercent = BestChain.Percent
		maxPercentTimestamp = &nowTimestamp
	}

	swapChainEntity = s.SwapChainBuilder.BuildEntity(
		BestChain,
		maxPercent,
		swapChainId,
		nowTimestamp,
		*maxPercentTimestamp,
		swapOneId,
		swapTwoId,
		swapThreeId,
	)

	if swapChainId > 0 {
		_ = s.SwapRepository.UpdateSwapChain(swapChainEntity)
	} else {
		_, _ = s.SwapRepository.CreateSwapChain(swapChainEntity)
	}

	return swapChainEntity
}
