package exchange

import (
	"github.com/AndreyMashukov/go-crypto-bot/server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/repository"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/utils"
	"time"
)

type SwapUpdater struct {
	Binance            *client.Binance
	ExchangeRepository *repository.ExchangeRepository
	Formatter          *utils.Formatter
}

func (s SwapUpdater) UpdateSwapPair(swapPair model.SwapPair) {
	orderDepth := s.ExchangeRepository.GetDepth(swapPair.Symbol)
	// save support + resistance levels
	if len(orderDepth.Asks) >= 10 && len(orderDepth.Bids) >= 10 {
		kLines := s.Binance.GetKLinesCached(swapPair.Symbol, "1d", 1)
		if len(kLines) > 0 {
			kline := kLines[0]
			swapPair.DailyPercent = s.Formatter.ToFixed(
				(s.Formatter.ComparePercentage(kline.Open, kline.Close) - 100.00).Value(),
				2,
			)

			swapPair.BuyPrice = orderDepth.Bids[0][0].Value
			swapPair.SellPrice = orderDepth.Asks[0][0].Value
			swapPair.SellVolume = s.Formatter.ToFixed(orderDepth.GetAskVolume(), 2)
			swapPair.BuyVolume = s.Formatter.ToFixed(orderDepth.GetBidVolume(), 2)
			swapPair.PriceTimestamp = time.Now().Unix()
			_ = s.ExchangeRepository.UpdateSwapPair(swapPair)
		}
	}
}
