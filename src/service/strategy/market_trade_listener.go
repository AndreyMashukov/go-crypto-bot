package strategy

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/service/ml"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"log"
	"math"
	"os"
	"strings"
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
}

func (m *MarketTradeListener) ListenAll() {
	klineChannel := make(chan model.KLine)
	predictChannel := make(chan string)
	depthChannel := make(chan model.Depth)

	go func(channel chan string) {
		for {
			symbol := <-channel
			predicted, err := m.PythonMLBridge.Predict(symbol)
			if err == nil && predicted > 0.00 {
				kLine := m.ExchangeRepository.GetLastKLine(symbol)
				if kLine != nil {
					m.ExchangeRepository.SaveKLinePredict(predicted, *kLine)
					// todo: write only master bot???
					interpolation := m.PriceCalculator.InterpolatePrice(kLine.Symbol)
					m.ExchangeRepository.SaveInterpolation(interpolation, *kLine)
				}
				m.ExchangeRepository.SavePredict(predicted, symbol)
			}
		}
	}(predictChannel)

	go func(klineChannel chan model.KLine) {
		for {
			invalidPriceSymbols := make([]string, 0)
			for _, limit := range m.ExchangeRepository.GetTradeLimits() {
				kline := m.ExchangeRepository.GetLastKLine(limit.Symbol)
				if kline != nil && kline.IsPriceNotActual() {
					invalidPriceSymbols = append(invalidPriceSymbols, limit.Symbol)
				}
			}

			if len(invalidPriceSymbols) > 0 {
				log.Printf("Price is invalid for: %s", strings.Join(invalidPriceSymbols, ", "))
				tickers := m.Binance.GetTickers(invalidPriceSymbols)
				updated := make([]string, 0)
				for _, ticker := range tickers {
					kline := m.ExchangeRepository.GetLastKLine(ticker.Symbol)
					if kline != nil && kline.IsPriceNotActual() {
						kline.UpdatedAt = time.Now().Unix()
						kline.Close = ticker.Price
						kline.High = math.Max(ticker.Price, kline.High)
						kline.Low = math.Min(ticker.Price, kline.Low)
						klineChannel <- *kline

						updated = append(updated, kline.Symbol)
					}
				}
				if len(updated) > 0 {
					log.Printf("Price updated for: %s", strings.Join(updated, ", "))
				}
			}

			m.TimeService.WaitSeconds(2)
		}
	}(klineChannel)

	go func(klineChannel chan model.KLine) {
		for {
			kLine := <-klineChannel
			m.ExchangeRepository.AddKLine(kLine)

			go func(channel chan string, symbol string) {
				predictChannel <- symbol
			}(predictChannel, kLine.Symbol)

			m.ExchangeRepository.SetDecision(m.BaseKLineStrategy.Decide(kLine), kLine.Symbol)
			m.ExchangeRepository.SetDecision(m.OrderBasedStrategy.Decide(kLine), kLine.Symbol)
		}
	}(klineChannel)

	eventChannel := make(chan []byte)

	go func() {
		for {
			message := <-eventChannel

			switch true {
			case strings.Contains(string(message), "miniTicker"):
				var tickerEvent model.MiniTickerEvent
				json.Unmarshal(message, &tickerEvent)
				ticker := tickerEvent.MiniTicker
				kLine := m.ExchangeRepository.GetLastKLine(ticker.Symbol)

				if kLine != nil && kLine.Includes(ticker) {
					kLine.Update(ticker)
					klineChannel <- *kLine
				}

				break
			case strings.Contains(string(message), "aggTrade"):
				var tradeEvent model.TradeEvent
				json.Unmarshal(message, &tradeEvent)
				smaDecision := m.SmaTradeStrategy.Decide(tradeEvent.Trade)
				m.ExchangeRepository.SetDecision(smaDecision, tradeEvent.Trade.Symbol)

				go func(channel chan string, symbol string) {
					predictChannel <- symbol
				}(predictChannel, tradeEvent.Trade.Symbol)
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

	go func() {
		for {
			depth := <-depthChannel
			m.ExchangeRepository.SetDepth(depth)
		}
	}()

	websockets := make([]*websocket.Conn, 0)

	tradeLimitCollection := make([]model.SymbolInterface, 0)
	hasBtcUsdt := false
	hasEthUsdt := false
	for _, limit := range m.ExchangeRepository.GetTradeLimits() {
		tradeLimitCollection = append(tradeLimitCollection, limit)

		go func(tradeLimit model.TradeLimit) {
			klineAmount := 0

			history := m.Binance.GetKLines(tradeLimit.GetSymbol(), "1m", 200)

			for _, kline := range history {
				klineAmount++
				m.ExchangeRepository.AddKLine(kline.ToKLine(tradeLimit.GetSymbol()))
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
}
