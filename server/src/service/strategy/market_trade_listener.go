package strategy

import (
	"context"
	"log"
	"math"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/AndreyMashukov/go-crypto-bot/server/market-watcher/enrichment"
	"github.com/AndreyMashukov/go-crypto-bot/server/market-watcher/publisher"
	tickevent "github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/tickstore"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/event"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/repository"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/service"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/service/exchange"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/service/ml"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/utils"
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

	// CurrentBot exists so the MarketTick stamp can name its exchange
	// without reaching back into a repo. Nil-tolerant: the legacy path
	// runs unchanged when CurrentBot or Publisher is unset (tests).
	CurrentBot *model.Bot
	// Publisher fans every observed kline out to the new shared/transport
	// pub/sub channel + ClickHouse market_tick table. The legacy decision
	// path still runs inline; Publisher is a parallel side-effect only.
	Publisher *publisher.Publisher
	// Enrichment maintains the rolling kline window per symbol so the
	// emitted MarketTick carries real Candles + Indicators. Phase D adds
	// this — earlier phases shipped a skeleton tick with empty Candles.
	Enrichment *enrichment.Store
	// TickStore is the in-process latest-tick map the StrategyFacade
	// reads on the decision path. Watcher writes here right after Emit.
	TickStore tickstore.Store
}

// emitMarketTick records the kline in the enrichment window, builds a
// MarketTick carrying Candles + Indicators computed from that window,
// and fans the tick out to both the network publisher (Redis pub/sub +
// ClickHouse) and the in-process TickStore that the strategy facade
// reads on the decision path. OpenPosition + RiskEnvelope stay zero
// for Phase D; a later phase wires their real sources.
func (m *MarketTradeListener) emitMarketTick(kLine model.KLine) {
	if m.CurrentBot == nil {
		return
	}
	if m.Enrichment != nil {
		m.Enrichment.OnKLine(enrichment.KLine{
			Symbol:    kLine.Symbol,
			OpenTime:  time.UnixMilli(kLine.OpenTime.Value()).UTC(),
			Open:      kLine.Open.Value(),
			High:      kLine.High.Value(),
			Low:       kLine.Low.Value(),
			Close:     kLine.Close.Value(),
			Volume:    kLine.Volume.Value(),
			Timestamp: time.UnixMilli(kLine.Timestamp.Value()).UTC(),
		})
	}
	var (
		candles    tickevent.Candles
		indicators tickevent.Indicators
	)
	if m.Enrichment != nil {
		candles, indicators = m.Enrichment.Snapshot(kLine.Symbol)
	}
	tick := tickevent.MarketTick{
		Exchange:   m.CurrentBot.Exchange,
		Symbol:     kLine.Symbol,
		Source:     tickevent.SourceTrade,
		EventTime:  time.UnixMilli(kLine.Timestamp.Value()).UTC(),
		IngestedAt: time.Now().UTC(),
		Price:      decimal.NewFromFloat(kLine.Close.Value()),
		BestBid:    decimal.Zero,
		BestAsk:    decimal.Zero,
		Volume24h:  decimal.Zero,
		Candles:    candles,
		Indicators: indicators,
	}
	if m.Publisher != nil {
		m.Publisher.Emit(context.Background(), tick)
	}
	if m.TickStore != nil {
		m.TickStore.Set(tick)
	}
}

func (m *MarketTradeListener) ListenAll() {
	klineChannel := make(chan model.KLine, 1000)
	predictChannel := make(chan string, 1000)
	depthChannel := make(chan model.OrderBookModel, 1000)

	predictMap := sync.Map{}

	go func(pMap *sync.Map) {
		for {
			symbol := <-predictChannel

			if status, ok := pMap.Load(symbol); ok {
				log.Printf("[%s] Prediction status: %s, skip", symbol, status)
				continue
			}
			pMap.Store(symbol, "processing")

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
			pMap.Delete(symbol)
		}
	}(&predictMap)

	klineConsumerCount := 8

	for i := 0; i < klineConsumerCount; i++ {
		go func() {
			for {
				afterEach := func() {
					runtime.GC()
					runtime.Gosched()
				}

				kLine := <-klineChannel
				lastKline := m.ExchangeRepository.GetCurrentKline(kLine.Symbol)

				if lastKline != nil && (lastKline.Timestamp.Gt(kLine.Timestamp) || kLine.IsPriceNotActual() || kLine.IsPriceWrongTimestamp()) {
					log.Printf(
						"[%s] (%s) Exchange sent expired stream price. T = %d < %d, UpdAt: %d, Now: %d [%d]",
						kLine.Symbol,
						kLine.Source,
						kLine.Timestamp.Value(),
						lastKline.Timestamp.Value(),
						kLine.UpdatedAt,
						time.Now().Unix(),
						model.TimestampMilli(time.Now().UnixMilli()).GetPeriodToMinute(),
					)
					afterEach()
					continue
				}

				m.ExchangeRepository.SetCurrentKline(kLine)
				m.emitMarketTick(kLine)
				if lastKline != nil && lastKline.Timestamp.GetPeriodToMinute() != kLine.Timestamp.GetPeriodToMinute() {
					m.EventDispatcher.Dispatch(event.NewKlineReceived{
						Previous: lastKline,
						Current:  &kLine,
					}, event.EventNewKLineReceived)
				}

				predictChannel <- kLine.Symbol
				m.ExchangeRepository.SetDecision(m.BaseKLineStrategy.Decide(kLine), kLine.Symbol)
				m.ExchangeRepository.SetDecision(m.OrderBasedStrategy.Decide(kLine), kLine.Symbol)
				afterEach()
			}
		}()
	}

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

	go func() {
		for {
			invalidPriceSymbols := make([]string, 0)
			for _, limit := range m.ExchangeRepository.GetTradeLimits() {
				k := m.ExchangeRepository.GetCurrentKline(limit.Symbol)
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
					if k == nil {
						k = &model.KLine{
							Symbol:    t.Symbol,
							Interval:  "1m",
							Low:       model.Price(t.Price),
							Open:      model.Price(t.Price),
							Close:     model.Price(t.Price),
							High:      model.Price(t.Price),
							Timestamp: model.TimestampMilli(0),
							UpdatedAt: 0,
							OpenTime:  model.TimestampMilli(model.TimestampMilli(currentInterval).GetPeriodFromMinute()),
						}
					}

					if k.IsPriceNotActual() {
						k.High = model.Price(math.Max(t.Price, k.High.Value()))
						k.Low = model.Price(math.Min(t.Price, k.Low.Value()))
						k.Close = model.Price(t.Price)

						if k.Timestamp.GetPeriodToMinute() < currentInterval {
							k.Timestamp = model.TimestampMilli(currentInterval)
							k.Open = model.Price(t.Price)
							k.Close = model.Price(t.Price)
							k.High = model.Price(t.Price)
							k.Low = model.Price(t.Price)
						}

						// todo: update timestamp and recover max, min prices...
						k.UpdatedAt = time.Now().Unix()
						k.Source = model.KLineSourceKLineFetch
						klineChannel <- *k
						updated = append(updated, k.Symbol)
					}
				}

				if len(updated) > 0 {
					log.Printf("Price updated for: %s", strings.Join(updated, ", "))
				}
			}
			m.TimeService.WaitSeconds(4)
		}
	}()
	log.Printf("Price recovery watcher started")

	// todo: order book recovery watcher is needed!

	runChannel := make(chan string)
	runChannel <- "run"
	log.Panic("Trade Listener Stopped")
}
