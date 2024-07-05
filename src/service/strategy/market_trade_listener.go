package strategy

import (
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
	Binance             client.ExchangeAPIInterface
	PythonMLBridge      *ml.PythonMLBridge
	PriceCalculator     *exchange.PriceCalculator
	EventDispatcher     *service.EventDispatcher

	ExchangeWSStreamer ExchangeWSStreamer
}

func (m *MarketTradeListener) ListenAll() {
	klineChannel := make(chan model.KLine, 1000)
	predictChannel := make(chan string, 1000)
	depthChannel := make(chan model.OrderBookModel, 1000)

	go func() {
		for {
			symbol := <-predictChannel
			predicted, _ := m.PythonMLBridge.Predict(symbol)

			kLine := m.ExchangeRepository.GetCurrentKline(symbol)
			if predicted > 0.00 {
				if kLine != nil {
					m.ExchangeRepository.SaveKLinePredict(predicted, *kLine)
				}
				m.ExchangeRepository.SavePredict(predicted, symbol)
			}

			if kLine != nil {
				limit := m.ExchangeRepository.GetTradeLimitCached(kLine.Symbol)
				if limit != nil {
					interpolation := m.PriceCalculator.InterpolatePrice(*limit)
					m.ExchangeRepository.SaveInterpolation(interpolation, *kLine)
				}
			}
		}
	}()

	go func() {
		for {
			kLine := <-klineChannel
			predictChannel <- kLine.Symbol

			lastKline := m.ExchangeRepository.GetCurrentKline(kLine.Symbol)

			if lastKline != nil && lastKline.Timestamp.Gt(kLine.Timestamp) {
				log.Printf(
					"[%s] Exchange sent expired stream price. T = %d < %d",
					kLine.Symbol,
					kLine.Timestamp.Value(),
					lastKline.Timestamp.Value(),
				)
				continue
			}

			m.ExchangeRepository.SetCurrentKline(kLine)
			if lastKline != nil && lastKline.Timestamp.GetPeriodToMinute() != kLine.Timestamp.GetPeriodToMinute() {
				m.EventDispatcher.Dispatch(event.NewKlineReceived{
					Previous: lastKline,
					Current:  &kLine,
				}, event.EventNewKLineReceived)
			}

			m.ExchangeRepository.SetDecision(m.BaseKLineStrategy.Decide(kLine), kLine.Symbol)
			m.ExchangeRepository.SetDecision(m.OrderBasedStrategy.Decide(kLine), kLine.Symbol)
		}
	}()

	go func() {
		for {
			depth := <-depthChannel
			depth.UpdatedAt = time.Now().Unix()
			m.ExchangeRepository.SetDepth(depth, 20, 25)
		}
	}()

	tradeLimitCollection := make([]model.SymbolInterface, 0)
	hasBtcUsdt := false
	hasEthUsdt := false

	waitGroup := sync.WaitGroup{}
	log.Printf("Price history recovery started")
	for _, limit := range m.ExchangeRepository.GetTradeLimits() {
		waitGroup.Add(1)
		tradeLimitCollection = append(tradeLimitCollection, limit)

		go func(l model.TradeLimit) {
			defer waitGroup.Done()
			klineAmount := 0
			history := m.Binance.GetKLines(l.GetSymbol(), "1m", 200)

			if len(history) > 0 {
				m.ExchangeRepository.ClearKlineHistory(l.GetSymbol())
			}

			for _, kline := range history {
				klineAmount++
				m.ExchangeRepository.SaveKlineHistory(kline.ToKLine(l.GetSymbol()))
			}
			log.Printf("Loaded history %s -> %d klines", l.Symbol, klineAmount)
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

	m.ExchangeWSStreamer.StartStream(tradeLimitCollection, klineChannel, depthChannel)
	log.Printf("WS Price stream started.")

	// Price recovery watcher
	go func() {
		for {
			invalidPriceSymbols := make([]string, 0)
			for _, limit := range m.ExchangeRepository.GetTradeLimits() {
				k := m.ExchangeRepository.GetCurrentKline(limit.Symbol)
				// If update is not received from WS or price is not actual
				if k == nil || k.IsPriceNotActual() {
					invalidPriceSymbols = append(invalidPriceSymbols, limit.Symbol)
				}
			}

			if len(invalidPriceSymbols) > 0 {
				log.Printf("Price is invalid for: %s", strings.Join(invalidPriceSymbols, ", "))
				tickers := m.Binance.GetTickers(invalidPriceSymbols)
				updated := make([]string, 0)

				for _, t := range tickers {
					k := m.ExchangeRepository.GetCurrentKline(t.Symbol)
					currentInterval := model.TimestampMilli(time.Now().UnixMilli()).GetPeriodToMinute()
					// Recover Kline
					if k == nil {
						k = &model.KLine{
							Symbol:    t.Symbol,
							Interval:  "1m",
							Low:       t.Price,
							Open:      t.Price,
							Close:     t.Price,
							High:      t.Price,
							Timestamp: model.TimestampMilli(0),
							UpdatedAt: 0,
							OpenTime:  model.TimestampMilli(model.TimestampMilli(currentInterval).GetPeriodFromMinute()),
						}
					}

					if k.IsPriceNotActual() {
						k.High = math.Max(t.Price, k.High)
						k.Low = math.Min(t.Price, k.Low)
						k.Close = t.Price

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
						klineChannel <- *k
						updated = append(updated, k.Symbol)
					}
				}

				if len(updated) > 0 {
					log.Printf("Price updated for: %s", strings.Join(updated, ", "))
				}
			}
			m.TimeService.WaitSeconds(2)
		}
	}()
	log.Printf("Price recovery watcher started")

	// todo: order book recovery watcher is needed!

	runChannel := make(chan string)
	// just to keep running
	runChannel <- "run"
	log.Panic("Trade Listener Stopped")
}
