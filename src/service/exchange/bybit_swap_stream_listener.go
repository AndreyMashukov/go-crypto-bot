package exchange

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type BybitSwapStreamListener struct {
	ExchangeRepository *repository.ExchangeRepository
	SwapUpdater        *SwapUpdater
	SwapRepository     *repository.SwapRepository
	SwapManager        *SwapManager
	Formatter          *utils.Formatter
}

func (s *BybitSwapStreamListener) StartListening() {
	swapKlineChannel := make(chan []byte, 1000)
	// existing swaps real time monitoring
	go func() {
		for {
			swapMsg := <-swapKlineChannel
			swapSymbol := ""

			if strings.Contains(string(swapMsg), "kline.1.") {
				var event model.ByBitWsKLineEvent
				err := json.Unmarshal(swapMsg, &event)
				if err == nil {
					symbol := strings.ReplaceAll(event.Topic, "kline.1.", "")

					if len(event.Data) > 0 {
						byBitKline := event.Data[0]
						kLine := byBitKline.ToBinanceKline(symbol, s.Formatter.ByBitIntervalToBinanceInterval(byBitKline.Interval))
						kLine.UpdatedAt = time.Now().Unix()

						s.ExchangeRepository.SetCurrentKline(kLine)
						swapSymbol = kLine.Symbol
					}
				} else {
					log.Printf("Swap Kline error bybit: %s", err.Error())
				}
			}

			if strings.Contains(string(swapMsg), "orderbook.50.") {
				var event model.ByBitWsOrderBookEvent
				err := json.Unmarshal(swapMsg, &event)
				if err == nil {
					depth := event.Data.ToOrderBookModel()
					depth.UpdatedAt = time.Now().Unix()
					s.ExchangeRepository.SetDepth(depth, 20, 25)
					swapSymbol = depth.Symbol
				} else {
					log.Printf("Swap Order book error bybit: %s", err.Error())
				}
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

	for index, streamBatchItem := range client.GetStreamBatchByBit(swapPairCollection, []string{"kline.1.", "orderbook.50."}) {
		sWg.Add(1)
		go func(sbi []string, i int) {
			defer sWg.Done()
			lock.Lock()
			swapWebsockets = append(swapWebsockets, client.ListenByBit(os.Getenv("BYBIT_STREAM_DSN"), swapKlineChannel, sbi, int64(i)))
			lock.Unlock()
			log.Printf("Batch %d websocket: %d connected", i, len(sbi))
		}(streamBatchItem, index)
	}

	sWg.Wait()
}
