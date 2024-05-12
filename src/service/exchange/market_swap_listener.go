package exchange

import (
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"log"
	"slices"
)

type MarketSwapListener struct {
	TimeService        *utils.TimeHelper
	SwapManager        *SwapManager
	ExchangeRepository *repository.ExchangeRepository
	SwapStreamListener SwapStreamListenerInterface
}

func (m *MarketSwapListener) ListenAll() {
	go func() {
		for {
			baseAssets := make([]string, 0)
			for _, pair := range m.ExchangeRepository.GetSwapPairs() {
				if !slices.Contains(baseAssets, pair.BaseAsset) {
					baseAssets = append(baseAssets, pair.BaseAsset)
				}
			}

			for _, baseAsset := range baseAssets {
				m.SwapManager.CalculateSwapOptions(baseAsset)
			}

			m.TimeService.WaitMilliseconds(250)
		}
	}()

	m.SwapStreamListener.StartListening()
	log.Printf("WS Swap Price stream started.")

	runChannel := make(chan string)
	// just to keep running
	runChannel <- "run"
	log.Panic("Swap Listener Stopped")
}
