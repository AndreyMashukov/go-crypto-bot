package strategy

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/event"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/service/ml"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

type MarketTradeListener struct {
	BaseKLineStrategy   *BaseKLineStrategy
	OrderBasedStrategy  *OrderBasedStrategy
	SmaTradeStrategy    *SmaTradeStrategy
	MarketDepthStrategy *MarketDepthStrategy
	ExchangeRepository  *repository.ExchangeRepository
	TimeService         *utils.TimeHelper
	Binance             *client.Binance
	PythonMLBridge      *ml.PythonMLBridge
	PriceCalculator     *exchange.PriceCalculator
	EventDispatcher     *service.EventDispatcher
}

func (m *MarketTradeListener) ListenAll() {
	klineChannel := make(chan model.KLine)
	predictChannel := make(chan string)
	depthChannel := make(chan model.Depth)

	go func() {
		for {
			symbol := <-predictChannel
			predicted, err := m.PythonMLBridge.Predict(symbol)
			if err == nil && predicted > 0.00 {
				kLine := m.ExchangeRepository.GetCurrentKline(symbol)
				if kLine != nil {
					m.ExchangeRepository.SaveKLinePredict(predicted, *kLine)
					// todo: write only master bot???
					interpolation := m.PriceCalculator.InterpolatePrice(kLine.Symbol)
					m.ExchangeRepository.SaveInterpolation(interpolation, *kLine)
				}
				m.ExchangeRepository.SavePredict(predicted, symbol)
			}
		}
	}()

	go func() {
		for {
			kLine := <-klineChannel

			lastKline := m.ExchangeRepository.GetCurrentKline(kLine.Symbol)
			m.ExchangeRepository.SetCurrentKline(kLine)

			if lastKline != nil && lastKline.Timestamp.Gt(kLine.Timestamp) {
				log.Printf(
					"[%s] Exchange sent expired stream price. T= %d < %d",
					kLine.Symbol,
					kLine.Timestamp.Value(),
					lastKline.Timestamp.Value(),
				)
				continue
			}

			if lastKline != nil && lastKline.Timestamp.GetPeriodToMinute() != kLine.Timestamp.GetPeriodToMinute() {
				m.EventDispatcher.Dispatch(event.NewKlineReceived{
					Previous: lastKline,
					Current:  &kLine,
				}, event.EventNewKLineReceived)
			}

			go func(symbol string) {
				predictChannel <- symbol
			}(kLine.Symbol)

			m.ExchangeRepository.SetDecision(m.BaseKLineStrategy.Decide(kLine), kLine.Symbol)
			m.ExchangeRepository.SetDecision(m.OrderBasedStrategy.Decide(kLine), kLine.Symbol)
		}
	}()

	eventChannel := make(chan []byte)

	go func() {
		for {
			depth := <-depthChannel
			m.ExchangeRepository.SetDepth(depth)
		}
	}()

	go func() {
		for {
			message := <-eventChannel

			switch true {
			case strings.Contains(string(message), "miniTicker"):
				var tickerEvent model.MiniTickerEvent
				json.Unmarshal(message, &tickerEvent)
				ticker := tickerEvent.MiniTicker
				kLine := m.ExchangeRepository.GetCurrentKline(ticker.Symbol)

				if kLine != nil && kLine.Includes(ticker) {
					kLine.Update(ticker)
					klineChannel <- *kLine
				}

				break
			case strings.Contains(string(message), "aggTrade"):
				var tradeEvent model.TradeEvent
				json.Unmarshal(message, &tradeEvent)

				m.ExchangeRepository.AddTrade(tradeEvent.Trade)
				smaDecision := m.SmaTradeStrategy.Decide(tradeEvent.Trade)
				m.ExchangeRepository.SetDecision(smaDecision, tradeEvent.Trade.Symbol)

				break
			case strings.Contains(string(message), "kline"):
				var event model.KlineEvent
				json.Unmarshal(message, &event)
				kLine := event.KlineData.Kline
				kLine.UpdatedAt = time.Now().Unix()

				klineChannel <- kLine

				break
			case strings.Contains(string(message), "depth20"):
				var event model.OrderBookEvent
				json.Unmarshal(message, &event)

				depth := event.Depth.ToDepth(strings.ToUpper(strings.ReplaceAll(event.Stream, "@depth20@100ms", "")))
				depthDecision := m.MarketDepthStrategy.Decide(depth)
				m.ExchangeRepository.SetDecision(depthDecision, depth.Symbol)
				go func() {
					depthChannel <- depth
				}()
				break
			}
		}
	}()

	websockets := make([]*websocket.Conn, 0)

	tradeLimitCollection := make([]model.SymbolInterface, 0)
	hasBtcUsdt := false
	hasEthUsdt := false

	waitGroup := sync.WaitGroup{}
	log.Printf("Price history recovery started")

	for _, limit := range m.ExchangeRepository.GetTradeLimits() {
		waitGroup.Add(1)
		tradeLimitCollection = append(tradeLimitCollection, limit)

		go func(tradeLimit model.TradeLimit) {
			defer waitGroup.Done()
			klineAmount := 0

			lastKline := m.ExchangeRepository.GetCurrentKline(tradeLimit.Symbol)
			if lastKline != nil && !lastKline.IsPriceExpired() {
				log.Printf("Price is not expired for %s history recovery skipped", tradeLimit.Symbol)
				return
			}

			history := m.Binance.GetKLines(tradeLimit.GetSymbol(), "1m", 200)

			if len(history) > 0 {
				m.ExchangeRepository.ClearKlineHistory(tradeLimit.GetSymbol())
			}

			for _, kline := range history {
				klineAmount++
				m.ExchangeRepository.SaveKlineHistory(kline.ToKLine(tradeLimit.GetSymbol()))
			}
			log.Printf("Loaded history %s -> %d klines", tradeLimit.Symbol, klineAmount)
		}(limit)
		if "BTCUSDT" == limit.GetSymbol() {
			hasBtcUsdt = true
		}
		if "ETHUSDT" == limit.GetSymbol() {
			hasEthUsdt = true
		}
	}

	waitGroup.Wait()
	log.Printf("Price history recovery finished")
	m.EventDispatcher.Enabled = true
	log.Printf("Event subscribers are enabled")

	if !hasBtcUsdt {
		tradeLimitCollection = append(tradeLimitCollection, model.DummySymbol{Symbol: "BTCUSDT"})
	}
	if !hasEthUsdt {
		tradeLimitCollection = append(tradeLimitCollection, model.DummySymbol{Symbol: "ETHUSDT"})
	}

	for index, streamBatchItem := range client.GetStreamBatch(tradeLimitCollection, []string{"@aggTrade", "@kline_1m@2000ms", "@depth20@100ms", "@miniTicker"}) {
		websockets = append(websockets, client.Listen(fmt.Sprintf(
			"%s/stream?streams=%s",
			os.Getenv("BINANCE_STREAM_DSN"),
			strings.Join(streamBatchItem, "/"),
		), eventChannel, []string{}, int64(index)))

		log.Printf("Batch %d websocket: %s", index, strings.Join(streamBatchItem, ", "))
	}

	// Price recovery watcher
	go func() {
		for {
			lock := sync.Mutex{}
			wg := sync.WaitGroup{}

			invalidPriceSymbols := make([]string, 0)
			for _, limit := range m.ExchangeRepository.GetTradeLimits() {
				wg.Add(1)
				go func(l model.TradeLimit) {
					defer wg.Done()

					k := m.ExchangeRepository.GetCurrentKline(l.Symbol)
					if k != nil && k.IsPriceNotActual() {
						lock.Lock()
						invalidPriceSymbols = append(invalidPriceSymbols, l.Symbol)
						lock.Unlock()
					}
				}(limit)
			}
			wg.Wait()

			if len(invalidPriceSymbols) > 0 {
				log.Printf("Price is invalid for: %s", strings.Join(invalidPriceSymbols, ", "))
				tickers := m.Binance.GetTickers(invalidPriceSymbols)
				updated := make([]string, 0)
				wg = sync.WaitGroup{}

				for _, ticker := range tickers {
					wg.Add(1)
					go func(t model.WSTickerPrice) {
						defer wg.Done()

						k := m.ExchangeRepository.GetCurrentKline(t.Symbol)
						if k != nil && k.IsPriceNotActual() {
							currentInterval := model.TimestampMilli(time.Now().UnixMilli()).GetPeriodToMinute()
							if k.Timestamp.GetPeriodToMinute() < currentInterval {
								log.Printf(
									"[%s] New time interval reached %d -> %d, price is unknown",
									k.Symbol,
									k.Timestamp.GetPeriodToMinute(),
									currentInterval,
								)
								k.Timestamp = model.TimestampMilli(currentInterval)
								k.Open = t.Price
								k.Close = t.Price
								k.High = t.Price
								k.Low = t.Price
							}

							// todo: update timestamp and recover max, min prices...
							k.UpdatedAt = time.Now().Unix()
							k.Close = t.Price
							k.High = math.Max(t.Price, k.High)
							k.Low = math.Min(t.Price, k.Low)
							klineChannel <- *k
							lock.Lock()
							updated = append(updated, k.Symbol)
							lock.Unlock()
						}
					}(ticker)
				}
				wg.Wait()

				if len(updated) > 0 {
					log.Printf("Price updated for: %s", strings.Join(updated, ", "))
				}
			}
			m.TimeService.WaitSeconds(2)
		}
	}()
	log.Printf("Price recovery watcher started")

	runChannel := make(chan string)
	// just to keep running
	runChannel <- "run"
	log.Panic("Trade Listener Stopped")
}
