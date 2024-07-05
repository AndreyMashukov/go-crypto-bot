package strategy

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type ExchangeWSStreamer interface {
	StartStream(
		tradeLimitCollection []model.SymbolInterface,
		klineChannel chan model.KLine,
		depthChannel chan model.OrderBookModel,
	)
}

type BinanceWSStreamer struct {
	ExchangeRepository  *repository.ExchangeRepository
	SmaTradeStrategy    *SmaTradeStrategy
	MarketDepthStrategy *MarketDepthStrategy
}

func (b *BinanceWSStreamer) StartStream(
	tradeLimitCollection []model.SymbolInterface,
	klineChannel chan model.KLine,
	depthChannel chan model.OrderBookModel,
) {
	eventChannel := make(chan []byte, 1000)

	go func() {
		for {
			message := <-eventChannel

			switch true {
			case strings.Contains(string(message), "miniTicker"):
				var tickerEvent model.MiniTickerEvent
				err := json.Unmarshal(message, &tickerEvent)
				if err == nil {
					ticker := tickerEvent.MiniTicker
					kLine := b.ExchangeRepository.GetCurrentKline(ticker.Symbol)

					if kLine != nil && kLine.Includes(ticker) {
						klineChannel <- kLine.Update(ticker)
					}
				}

				break
			case strings.Contains(string(message), "aggTrade"):
				var tradeEvent model.TradeEvent
				err := json.Unmarshal(message, &tradeEvent)

				if err == nil {
					b.ExchangeRepository.AddTrade(tradeEvent.Trade)
					smaDecision := b.SmaTradeStrategy.Decide(tradeEvent.Trade)
					b.ExchangeRepository.SetDecision(smaDecision, tradeEvent.Trade.Symbol)
				}

				break
			case strings.Contains(string(message), "kline"):
				var event model.KlineEvent
				err := json.Unmarshal(message, &event)
				if err == nil {
					kLine := event.KlineData.Kline
					kLine.UpdatedAt = time.Now().Unix()

					klineChannel <- kLine
				}

				break
			case strings.Contains(string(message), "depth20"):
				var event model.OrderBookEvent
				err := json.Unmarshal(message, &event)

				if err == nil {
					depth := event.Depth.ToOrderBookModel(strings.ToUpper(strings.ReplaceAll(event.Stream, "@depth20@100ms", "")))
					depthDecision := b.MarketDepthStrategy.Decide(depth)
					b.ExchangeRepository.SetDecision(depthDecision, depth.Symbol)
					depthChannel <- depth
				}
				break
			}
		}
	}()
	websockets := make([]*websocket.Conn, 0)

	lock := sync.Mutex{}
	sWg := sync.WaitGroup{}

	for index, streamBatchItem := range client.GetStreamBatch(tradeLimitCollection, []string{"@aggTrade", "@kline_1m@2000ms", "@depth20@100ms", "@miniTicker"}) {
		sWg.Add(1)
		go func(sbi []string, i int) {
			defer sWg.Done()
			lock.Lock()
			websockets = append(websockets, client.Listen(fmt.Sprintf(
				"%s/stream?streams=%s",
				os.Getenv("BINANCE_STREAM_DSN"),
				strings.Join(sbi, "/"),
			), eventChannel, []string{}, int64(i)))
			lock.Unlock()
			log.Printf("Batch %d websocket: %d connected", i, len(sbi))
		}(streamBatchItem, index)
	}

	sWg.Wait()
}
