package strategy

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

type ByBitWsStreamer struct {
	ExchangeRepository  *repository.ExchangeRepository
	SmaTradeStrategy    *SmaTradeStrategy
	MarketDepthStrategy *MarketDepthStrategy
	Formatter           *utils.Formatter
}

func (b *ByBitWsStreamer) StartStream(
	tradeLimitCollection []model.SymbolInterface,
	klineChannel chan model.KLine,
	depthChannel chan model.OrderBookModel,
) {
	eventChannel := make(chan []byte, 1000)

	go func() {
		for {
			message := <-eventChannel

			switch true {
			case strings.Contains(string(message), "tickers."):
				var tickerEvent model.ByBitWsTickerEvent
				err := json.Unmarshal(message, &tickerEvent)
				if err == nil {
					tickerData := tickerEvent.Data
					ticker := tickerData.ToBinanceMiniTicker(tickerEvent.Ts)
					kLine := b.ExchangeRepository.GetCurrentKline(ticker.Symbol)
					periodTo := model.TimestampMilli(ticker.EventTime.GetPeriodToMinute())

					if kLine != nil && kLine.Timestamp.Lte(periodTo) {
						klineChannel <- kLine.Update(ticker, model.KLineSourceTickerStream)
					}
				} else {
					log.Printf("Ticker error bybit: %s", err.Error())
				}

				break
			case strings.Contains(string(message), "publicTrade."):
				var tradeEvent model.ByBitWsPublicTradeEvent
				err := json.Unmarshal(message, &tradeEvent)
				if err == nil {
					for _, byBitTrade := range tradeEvent.Data {
						trade := byBitTrade.ToBinanceTrade()

						b.ExchangeRepository.AddTrade(trade)
						smaDecision := b.SmaTradeStrategy.Decide(trade)
						b.ExchangeRepository.SetDecision(smaDecision, trade.Symbol)
					}
				} else {
					log.Printf("Public trade error bybit: %s", err.Error())
				}

				break
			case strings.Contains(string(message), "kline.1."):
				var event model.ByBitWsKLineEvent
				err := json.Unmarshal(message, &event)
				if err == nil {
					symbol := strings.ReplaceAll(event.Topic, "kline.1.", "")

					if len(event.Data) > 0 {
						byBitKline := event.Data[0]
						kLine := byBitKline.ToBinanceKline(symbol, b.Formatter.ByBitIntervalToBinanceInterval(byBitKline.Interval))
						kLine.UpdatedAt = time.Now().Unix()

						lastKline := b.ExchangeRepository.GetCurrentKline(kLine.Symbol)
						if lastKline == nil || lastKline.Timestamp.Lte(kLine.Timestamp) {
							kLine.Source = model.KLineSourceKLineStream
							klineChannel <- kLine
						}
					}
				} else {
					log.Printf("Kline error bybit: %s", err.Error())
				}

				break
			case strings.Contains(string(message), "orderbook.50."):
				var event model.ByBitWsOrderBookEvent
				err := json.Unmarshal(message, &event)
				if err == nil {
					depth := event.Data.ToOrderBookModel()
					depthDecision := b.MarketDepthStrategy.Decide(depth)
					b.ExchangeRepository.SetDecision(depthDecision, depth.Symbol)
					depthChannel <- depth
				} else {
					log.Printf("Order book error bybit: %s", err.Error())
				}
				break
			}
		}
	}()
	websockets := make([]*websocket.Conn, 0)

	lock := sync.Mutex{}
	sWg := sync.WaitGroup{}

	for index, streamBatchItem := range client.GetStreamBatchByBit(tradeLimitCollection, []string{"publicTrade.", "kline.1.", "orderbook.50.", "tickers."}) {
		sWg.Add(1)
		go func(sbi []string, i int) {
			defer sWg.Done()
			lock.Lock()
			websockets = append(websockets, client.ListenByBit(os.Getenv("BYBIT_STREAM_DSN"), eventChannel, sbi, int64(i)))
			lock.Unlock()
			log.Printf("Batch %d websocket: %d connected", i, len(sbi))
		}(streamBatchItem, index)
	}

	sWg.Wait()
}
