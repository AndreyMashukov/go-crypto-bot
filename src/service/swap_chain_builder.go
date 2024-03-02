package service

import "gitlab.com/open-soft/go-crypto-bot/src/model"

type SwapChainBuilder struct {
}

func (s SwapChainBuilder) BuildEntity(
	chain model.BestSwapChain,
	maxPercent model.Percent,
	swapChainId int64,
	nowTimestamp int64,
	maxPercentTimestamp int64,
	swapOneId int64,
	swapTwoId int64,
	swapThreeId int64,
) model.SwapChainEntity {
	return model.SwapChainEntity{
		Id:                  swapChainId,
		Title:               chain.Title,
		Type:                chain.Type,
		Hash:                chain.Hash,
		Percent:             chain.Percent,
		MaxPercent:          maxPercent,
		Timestamp:           nowTimestamp,
		MaxPercentTimestamp: &maxPercentTimestamp,
		SwapOne: &model.SwapTransitionEntity{
			Id:         swapOneId,
			Type:       chain.SwapOne.Type,
			Symbol:     chain.SwapOne.Symbol,
			BaseAsset:  chain.SwapOne.BaseAsset,
			QuoteAsset: chain.SwapOne.QuoteAsset,
			Operation:  chain.SwapOne.Operation,
			Quantity:   chain.SwapOne.BaseQuantity,
			Price:      chain.SwapOne.Price,
			Level:      chain.SwapOne.Level,
		},
		SwapTwo: &model.SwapTransitionEntity{
			Id:         swapTwoId,
			Type:       chain.SwapTwo.Type,
			Symbol:     chain.SwapTwo.Symbol,
			BaseAsset:  chain.SwapTwo.BaseAsset,
			QuoteAsset: chain.SwapTwo.QuoteAsset,
			Operation:  chain.SwapTwo.Operation,
			Quantity:   chain.SwapTwo.BaseQuantity,
			Price:      chain.SwapTwo.Price,
			Level:      chain.SwapTwo.Level,
		},
		SwapThree: &model.SwapTransitionEntity{
			Id:         swapThreeId,
			Type:       chain.SwapThree.Type,
			Symbol:     chain.SwapThree.Symbol,
			BaseAsset:  chain.SwapThree.BaseAsset,
			QuoteAsset: chain.SwapThree.QuoteAsset,
			Operation:  chain.SwapThree.Operation,
			Quantity:   chain.SwapThree.QuoteQuantity,
			Price:      chain.SwapThree.Price,
			Level:      chain.SwapThree.Level,
		},
	}
}
