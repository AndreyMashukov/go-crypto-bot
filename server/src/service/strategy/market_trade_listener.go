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
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/tickbuffer"
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
	// MarketBuffer is the in-process tick buffer the StrategyFacade
	// reads on the decision path. The watcher Puts here right after
	// Emit; the trader's subscriber Puts what it pulls off pub/sub.
	MarketBuffer tickbuffer.MarketTickBufferInterface
}

// emitMarketTick records the kline in the enrichment window, builds a
// MarketTick carrying Candles + Indicators computed from that window,
// and fans the tick out to both the network publisher (Redis pub/sub +
// ClickHouse) and the in-process MarketBuffer that the strategy facade
// reads on the decision path. OpenPosition + RiskEnvelope stay zero
// for Phase D; a later phase wires their real sources.
func (m *MarketTradeListener) emitMarketTick(ctx context.Context, kLine model.KLine) {
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
		m.Publisher.Emit(ctx, tick)
	}
	if m.MarketBuffer != nil {
		m.MarketBuffer.Put(tick)
	}
}

// ListenAll wires the watcher's per-process goroutine fleet —
// 8 kline consumers, 1 prediction worker, 1 depth consumer, 1 price-
// recovery loop, plus the exchange WS streamer. Every loop observes
// ctx and exits cleanly on cancel.
func (m *MarketTradeListener) ListenAll(ctx context.Context) {
	klineChannel := make(chan model.KLine, 1000)
	predictChannel := make(chan string, 1000)
	depthChannel := make(chan model.OrderBookModel, 1000)

	predictMap := sync.Map{}

	go m.runPredictLoop(ctx, predictChannel, &predictMap)

	klineConsumerCount := 8
	for i := 0; i < klineConsumerCount; i++ {
		go m.runKlineConsumer(ctx, klineChannel, predictChannel)
	}

	go m.runDepthConsumer(ctx, depthChannel)

	tradeLimitCollection, hasBtcUsdt, hasEthUsdt := m.collectTradeLimits()
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

	go m.runPriceRecovery(ctx, klineChannel)
	log.Printf("Price recovery watcher started")

	<-ctx.Done()
	log.Printf("MarketTradeListener: shutdown signal received: %s", ctx.Err().Error())
}

func (m *MarketTradeListener) collectTradeLimits() (symbols []model.SymbolInterface, hasBtcUsdt, hasEthUsdt bool) {
	limits := m.ExchangeRepository.GetTradeLimits()
	symbols = make([]model.SymbolInterface, 0, len(limits))
	for i := range limits {
		symbols = append(symbols, limits[i])
		if limits[i].GetSymbol() == "BTCUSDT" {
			hasBtcUsdt = true
		}
		if limits[i].GetSymbol() == "ETHUSDT" {
			hasEthUsdt = true
		}
	}
	return symbols, hasBtcUsdt, hasEthUsdt
}

func (m *MarketTradeListener) runPredictLoop(ctx context.Context, predictChannel <-chan string, pMap *sync.Map) {
	for {
		var symbol string
		select {
		case <-ctx.Done():
			return
		case symbol = <-predictChannel:
		}

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
}

func (m *MarketTradeListener) runKlineConsumer(ctx context.Context, klineChannel <-chan model.KLine, predictChannel chan<- string) {
	afterEach := func() {
		runtime.GC()
		runtime.Gosched()
	}
	for {
		var kLine model.KLine
		select {
		case <-ctx.Done():
			return
		case kLine = <-klineChannel:
		}
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

		m.emitMarketTick(ctx, kLine)
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
}

func (m *MarketTradeListener) runDepthConsumer(ctx context.Context, depthChannel <-chan model.OrderBookModel) {
	for {
		var depth model.OrderBookModel
		select {
		case <-ctx.Done():
			return
		case depth = <-depthChannel:
		}
		depth.UpdatedAt = time.Now().Unix()
		m.ExchangeRepository.SetDepth(depth, 20, 25)
	}
}

func (m *MarketTradeListener) runPriceRecovery(ctx context.Context, klineChannel chan<- model.KLine) {
	for {
		if err := m.TimeService.WaitSecondsCtx(ctx, 4); err != nil {
			return
		}
		limits := m.ExchangeRepository.GetTradeLimits()
		invalidPriceSymbols := make([]string, 0)
		for i := range limits {
			k := m.ExchangeRepository.GetCurrentKline(limits[i].Symbol)
			if k == nil || k.IsPriceNotActual() {
				invalidPriceSymbols = append(invalidPriceSymbols, limits[i].Symbol)
			}
		}

		if len(invalidPriceSymbols) == 0 {
			continue
		}
		log.Printf("Price is invalid for: %s", strings.Join(invalidPriceSymbols, ", "))
		tickers := m.Binance.GetTickers(invalidPriceSymbols)
		updated := m.recoverFromTickers(tickers, klineChannel)
		if len(updated) > 0 {
			log.Printf("Price updated for: %s", strings.Join(updated, ", "))
		}
	}
}

func (m *MarketTradeListener) recoverFromTickers(tickers []model.WSTickerPrice, klineChannel chan<- model.KLine) []string {
	updated := make([]string, 0)
	for i := range tickers {
		t := &tickers[i]
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
		if !k.IsPriceNotActual() {
			continue
		}
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

		k.UpdatedAt = time.Now().Unix()
		k.Source = model.KLineSourceKLineFetch
		klineChannel <- *k
		updated = append(updated, k.Symbol)
	}
	return updated
}
