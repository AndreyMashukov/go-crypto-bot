package exchange

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"log"
	"slices"
	"strings"
)

type MarketSwapListener struct {
	TimeService        *utils.TimeHelper
	SwapManager        *SwapManager
	ExchangeRepository *repository.ExchangeRepository
	SwapUpdater        *SwapUpdater
	SwapRepository     *repository.SwapRepository
}

func (m *MarketSwapListener) ListenAll() {
	swapKlineChannel := make(chan []byte)

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

	// existing swaps real time monitoring
	go func() {
		for {
			swapMsg := <-swapKlineChannel
			swapSymbol := ""

			if strings.Contains(string(swapMsg), "kline") {
				var event model.KlineEvent
				json.Unmarshal(swapMsg, &event)
				kLine := event.KlineData.Kline
				m.ExchangeRepository.AddKLine(kLine, false)
				swapSymbol = kLine.Symbol
			}

			if strings.Contains(string(swapMsg), "@depth20") {
				var event model.OrderBookEvent
				json.Unmarshal(swapMsg, &event)

				depth := event.Depth.ToDepth(strings.ToUpper(strings.ReplaceAll(event.Stream, "@depth20@1000ms", "")))
				m.ExchangeRepository.SetDepth(depth)
				swapSymbol = depth.Symbol
			}

			if swapSymbol == "" {
				continue
			}

			swapPair, err := m.ExchangeRepository.GetSwapPair(swapSymbol)
			if err == nil {
				m.SwapUpdater.UpdateSwapPair(swapPair)

				possibleSwap := m.SwapRepository.GetSwapChainCache(swapPair.BaseAsset)
				if possibleSwap != nil {
					go func(asset string) {
						m.SwapManager.CalculateSwapOptions(asset)
					}(swapPair.BaseAsset)
				}
			}
		}
	}()

	swapWebsockets := make([]*websocket.Conn, 0)

	swapPairCollection := make([]model.SymbolInterface, 0)
	for _, swapPair := range m.ExchangeRepository.GetSwapPairs() {
		swapPairCollection = append(swapPairCollection, swapPair)
	}

	for index, streamBatchItem := range client.GetStreamBatch(swapPairCollection, []string{"@kline_1m", "@depth20@1000ms"}) {
		swapWebsockets = append(swapWebsockets, client.Listen(fmt.Sprintf(
			"%s/stream?streams=%s",
			"wss://stream.binance.com:9443",
			strings.Join(streamBatchItem, "/"),
		), swapKlineChannel, []string{}, 10000+int64(index)))

		log.Printf("Swap batch %d websocket: %s", index, strings.Join(streamBatchItem, ", "))
	}

	runChannel := make(chan string)
	// just to keep running
	runChannel <- "run"
	log.Panic("Swap Listener Stopped")
}
