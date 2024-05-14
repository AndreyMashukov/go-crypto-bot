package exchange

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"log"
	"strings"
	"sync"
	"time"
)

type SwapStreamListenerInterface interface {
	StartListening()
}

type BinanceSwapStreamListener struct {
	ExchangeRepository *repository.ExchangeRepository
	SwapUpdater        *SwapUpdater
	SwapRepository     *repository.SwapRepository
	SwapManager        *SwapManager
}

func (s *BinanceSwapStreamListener) StartListening() {
	swapKlineChannel := make(chan []byte)
	// existing swaps real time monitoring
	go func() {
		for {
			swapMsg := <-swapKlineChannel
			swapSymbol := ""

			if strings.Contains(string(swapMsg), "kline") {
				var event model.KlineEvent
				json.Unmarshal(swapMsg, &event)
				kLine := event.KlineData.Kline
				kLine.UpdatedAt = time.Now().Unix()
				// todo: track price timestamp...
				s.ExchangeRepository.SetCurrentKline(kLine)
				swapSymbol = kLine.Symbol
			}

			if strings.Contains(string(swapMsg), "@depth20") {
				var event model.OrderBookEvent
				json.Unmarshal(swapMsg, &event)

				depth := event.Depth.ToOrderBookModel(strings.ToUpper(strings.ReplaceAll(event.Stream, "@depth20@1000ms", "")))
				depth.UpdatedAt = time.Now().Unix()
				s.ExchangeRepository.SetDepth(depth, 20, 25)
				swapSymbol = depth.Symbol
			}

			if swapSymbol == "" {
				continue
			}

			swapPair, err := s.ExchangeRepository.GetSwapPair(swapSymbol)
			if err == nil {
				s.SwapUpdater.UpdateSwapPair(swapPair)

				possibleSwap := s.SwapRepository.GetSwapChainCache(swapPair.BaseAsset)
				if possibleSwap != nil {
					go func(asset string) {
						s.SwapManager.CalculateSwapOptions(asset)
					}(swapPair.BaseAsset)
				}
			}
		}
	}()

	swapWebsockets := make([]*websocket.Conn, 0)

	swapPairCollection := make([]model.SymbolInterface, 0)
	for _, swapPair := range s.ExchangeRepository.GetSwapPairs() {
		swapPairCollection = append(swapPairCollection, swapPair)
	}

	lock := sync.Mutex{}
	sWg := sync.WaitGroup{}

	for index, streamBatchItem := range client.GetStreamBatch(swapPairCollection, []string{"@kline_1m", "@depth20@1000ms"}) {
		sWg.Add(1)
		go func(sbi []string, i int) {
			defer sWg.Done()
			lock.Lock()
			swapWebsockets = append(swapWebsockets, client.Listen(fmt.Sprintf(
				"%s/stream?streams=%s",
				"wss://stream.binance.com:9443",
				strings.Join(sbi, "/"),
			), swapKlineChannel, []string{}, 10000+int64(i)))
			lock.Unlock()
			log.Printf("Swap batch %d websocket: %d connected", i, len(sbi))
		}(streamBatchItem, index)
	}

	sWg.Wait()
}
