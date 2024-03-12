package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/config"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"math"
	"os"
	"slices"
	"strings"
	"time"
)

func getStreamBatch(tradeLimits []model.SymbolInterface, events []string) [][]string {
	streamBatch := make([][]string, 0)

	streams := make([]string, 0)

	for _, tradeLimit := range tradeLimits {
		for i := 0; i < len(events); i++ {
			event := events[i]
			streams = append(streams, fmt.Sprintf("%s%s", strings.ToLower(tradeLimit.GetSymbol()), event))
		}

		if len(streams) >= 24 {
			streamBatch = append(streamBatch, streams)
			streams = make([]string, 0)
		}
	}

	if len(streams) > 0 {
		streamBatch = append(streamBatch, streams)
	}

	return streamBatch
}

func main() {
	pwd, _ := os.Getwd()
	if _, err := os.Stat(fmt.Sprintf("%s/.env", pwd)); err == nil {
		log.Println(".env is found, loading variables...")
		err = godotenv.Load()
		if err != nil {
			log.Println(err)
		}
	}

	container := config.InitServiceContainer()
	defer container.Db.Close()
	defer container.DbSwap.Close()
	container.PythonMLBridge.Initialize()
	defer container.PythonMLBridge.Finalize()
	container.StartHttpServer()
	log.Printf("Bot [%s] is initialized successfully", container.CurrentBot.BotUuid)

	container.Binance.Connect(os.Getenv("BINANCE_WS_DSN")) // "wss://testnet.binance.vision/ws-api/v3"

	usdtBalance, err := container.BalanceService.GetAssetBalance("USDT", false)
	if err != nil {
		log.Printf("Balance check error: %s", err.Error())

		if err.Error() == model.BinanceErrorInvalidAPIKeyOrPermissions {
			log.Println("Notify SaaS system about error")
			container.CallbackManager.Error(
				*container.CurrentBot,
				model.BinanceErrorInvalidAPIKeyOrPermissions,
				"Please check API Key permissions or IP address binding",
				true,
			)
			os.Exit(0)
		}
	}
	log.Printf("API Key permission check passed, balance is: %.2f", usdtBalance)
	container.Binance.APIKeyCheckCompleted = true

	tradeLimits := container.ExchangeRepository.GetTradeLimits()
	symbols := make([]string, 0)
	for _, limit := range tradeLimits {
		symbols = append(symbols, limit.Symbol)
	}

	binanceOrders, err := container.Binance.GetOpenedOrders()
	if err == nil {
		for _, binanceOrder := range binanceOrders {
			if !slices.Contains(symbols, binanceOrder.Symbol) {
				log.Printf("[%s] binance order %d skipped", binanceOrder.Symbol, binanceOrder.OrderId)

				continue
			}

			log.Printf("[%s] loaded binance order %d", binanceOrder.Symbol, binanceOrder.OrderId)
			container.OrderRepository.SetBinanceOrder(binanceOrder)
		}
	}

	// Wait 5 seconds, here API can update some settings...
	time.Sleep(time.Second * 5)
	eventChannel := make(chan []byte)
	depthChannel := make(chan model.Depth)
	runChannel := make(chan string)

	go func(container *config.Container) {
		for {
			container.MakerService.UpdateLimits()
			time.Sleep(time.Minute * 5)
		}
	}(&container)

	if container.IsMasterBot {
		container.MakerService.UpdateSwapPairs()
	}

	if container.IsMasterBot {
		swapKlineChannel := make(chan []byte)
		defer close(swapKlineChannel)

		go func(container *config.Container) {
			for {
				baseAssets := make([]string, 0)
				for _, pair := range container.ExchangeRepository.GetSwapPairs() {
					if !slices.Contains(baseAssets, pair.BaseAsset) {
						baseAssets = append(baseAssets, pair.BaseAsset)
					}
				}

				for _, baseAsset := range baseAssets {
					container.SwapManager.CalculateSwapOptions(baseAsset)
				}

				time.Sleep(time.Millisecond * 250)
			}
		}(&container)

		// existing swaps real time monitoring
		go func(container *config.Container) {
			for {
				swapMsg := <-swapKlineChannel
				swapSymbol := ""

				if strings.Contains(string(swapMsg), "kline") {
					var event model.KlineEvent
					json.Unmarshal(swapMsg, &event)
					kLine := event.KlineData.Kline
					container.ExchangeRepository.AddKLine(kLine)
					swapSymbol = kLine.Symbol
				}

				if strings.Contains(string(swapMsg), "@depth20") {
					var event model.OrderBookEvent
					json.Unmarshal(swapMsg, &event)

					depth := event.Depth.ToDepth(strings.ToUpper(strings.ReplaceAll(event.Stream, "@depth20@1000ms", "")))
					container.ExchangeRepository.SetDepth(depth)
					swapSymbol = depth.Symbol
				}

				if swapSymbol == "" {
					continue
				}

				swapPair, err := container.ExchangeRepository.GetSwapPair(swapSymbol)
				if err == nil {
					container.SwapUpdater.UpdateSwapPair(swapPair)

					possibleSwap := container.SwapRepository.GetSwapChainCache(swapPair.BaseAsset)
					if possibleSwap != nil {
						go func(asset string) {
							container.SwapManager.CalculateSwapOptions(asset)
						}(swapPair.BaseAsset)
					}
				}
			}
		}(&container)

		swapWebsockets := make([]*websocket.Conn, 0)

		swapPairCollection := make([]model.SymbolInterface, 0)
		for _, swapPair := range container.ExchangeRepository.GetSwapPairs() {
			swapPairCollection = append(swapPairCollection, swapPair)
		}

		for index, streamBatchItem := range getStreamBatch(swapPairCollection, []string{"@kline_1m", "@depth20@1000ms"}) {
			swapWebsockets = append(swapWebsockets, client.Listen(fmt.Sprintf(
				"%s/stream?streams=%s",
				"wss://stream.binance.com:9443",
				strings.Join(streamBatchItem, "/"),
			), swapKlineChannel, []string{}, 10000+int64(index)))

			log.Printf("Swap batch %d websocket: %s", index, strings.Join(streamBatchItem, ", "))

			defer swapWebsockets[index].Close()
		}
	}

	for _, limit := range tradeLimits {
		go func(limit model.TradeLimit, container *config.Container) {
			for {
				// todo: write to database and read from database
				err := container.PythonMLBridge.LearnModel(limit.Symbol)
				if err != nil {
					log.Printf("[%s] %s", limit.Symbol, err.Error())
					container.TimeService.WaitSeconds(60)
					continue
				}

				container.TimeService.WaitSeconds(3600 * 6)
			}
		}(limit, &container)
		// learn every 1000 minutes

		go func(symbol string, container *config.Container) {
			for {
				container.MakerService.Make(symbol)

				time.Sleep(time.Millisecond * 500)
			}
		}(limit.Symbol, &container)
	}

	predictChannel := make(chan string)
	go func(channel chan string, container *config.Container) {
		for {
			symbol := <-channel
			predicted, err := container.PythonMLBridge.Predict(symbol)
			if err == nil && predicted > 0.00 {
				kLine := container.ExchangeRepository.GetLastKLine(symbol)
				if kLine != nil {
					container.ExchangeRepository.SaveKLinePredict(predicted, *kLine)
					// todo: write only master bot???
					interpolation := container.PriceCalculator.InterpolatePrice(kLine.Symbol)
					container.ExchangeRepository.SaveInterpolation(interpolation, *kLine)
				}
				container.ExchangeRepository.SavePredict(predicted, symbol)
			}
		}
	}(predictChannel, &container)

	klineChannel := make(chan model.KLine)

	go func(tradeLimits []model.TradeLimit, container *config.Container, klineChannel chan model.KLine) {
		for {
			invalidPriceSymbols := make([]string, 0)
			for _, limit := range tradeLimits {
				kline := container.ExchangeRepository.GetLastKLine(limit.Symbol)
				if kline != nil && kline.IsPriceNotActual() {
					invalidPriceSymbols = append(invalidPriceSymbols, limit.Symbol)
				}
			}

			if len(invalidPriceSymbols) > 0 {
				log.Printf("Price is invalid for: %s", strings.Join(invalidPriceSymbols, ", "))
				tickers := container.Binance.GetTickers(invalidPriceSymbols)
				updated := make([]string, 0)
				for _, ticker := range tickers {
					kline := container.ExchangeRepository.GetLastKLine(ticker.Symbol)
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

			container.TimeService.WaitSeconds(2)
		}
	}(tradeLimits, &container, klineChannel)

	go func(klineChannel chan model.KLine) {
		for {
			kLine := <-klineChannel
			container.ExchangeRepository.AddKLine(kLine)

			go func(channel chan string, symbol string) {
				predictChannel <- symbol
			}(predictChannel, kLine.Symbol)

			container.ExchangeRepository.SetDecision(container.BaseKLineStrategy.Decide(kLine), kLine.Symbol)
			container.ExchangeRepository.SetDecision(container.OrderBasedStrategy.Decide(kLine), kLine.Symbol)
		}
	}(klineChannel)

	go func(container *config.Container) {
		for {
			message := <-eventChannel

			switch true {
			case strings.Contains(string(message), "miniTicker"):
				var tickerEvent model.MiniTickerEvent
				json.Unmarshal(message, &tickerEvent)
				ticker := tickerEvent.MiniTicker
				kLine := container.ExchangeRepository.GetLastKLine(ticker.Symbol)

				if kLine != nil && kLine.Includes(ticker) {
					kLine.Update(ticker)
					klineChannel <- *kLine
				}

				break
			case strings.Contains(string(message), "aggTrade"):
				var tradeEvent model.TradeEvent
				json.Unmarshal(message, &tradeEvent)
				smaDecision := container.SmaTradeStrategy.Decide(tradeEvent.Trade)
				container.ExchangeRepository.SetDecision(smaDecision, tradeEvent.Trade.Symbol)

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
				depthDecision := container.MarketDepthStrategy.Decide(depth)
				container.ExchangeRepository.SetDecision(depthDecision, depth.Symbol)
				go func() {
					depthChannel <- depth
				}()
				break
			}
		}
	}(&container)

	go func(container *config.Container) {
		for {
			depth := <-depthChannel
			container.ExchangeRepository.SetDepth(depth)
		}
	}(&container)

	websockets := make([]*websocket.Conn, 0)

	tradeLimitCollection := make([]model.SymbolInterface, 0)
	hasBtcUsdt := false
	hasEthUsdt := false
	for _, limit := range tradeLimits {
		tradeLimitCollection = append(tradeLimitCollection, limit)

		go func(tradeLimit model.TradeLimit) {
			klineAmount := 0

			history := container.Binance.GetKLines(tradeLimit.GetSymbol(), "1m", 200)

			for _, kline := range history {
				klineAmount++
				container.ExchangeRepository.AddKLine(kline.ToKLine(tradeLimit.GetSymbol()))
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

	for index, streamBatchItem := range getStreamBatch(tradeLimitCollection, []string{"@aggTrade", "@kline_1m@2000ms", "@depth20@100ms", "@miniTicker"}) {
		websockets = append(websockets, client.Listen(fmt.Sprintf(
			"%s/stream?streams=%s",
			os.Getenv("BINANCE_STREAM_DSN"),
			strings.Join(streamBatchItem, "/"),
		), eventChannel, []string{}, int64(index)))

		log.Printf("Batch %d websocket: %s", index, strings.Join(streamBatchItem, ", "))

		defer websockets[index].Close()
	}

	// just to keep running
	runChannel <- "run"
	log.Panic("Stopped")
}
