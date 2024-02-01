package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/config"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
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

	tradeLimits := container.ExchangeRepository.GetTradeLimits()

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
				currentDecisions := container.ExchangeRepository.GetDecisions()

				if len(currentDecisions) > 0 {
					container.MakerService.Make(symbol, currentDecisions)
				}
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

					//if interpolation.HasBoth() {
					//	log.Printf(
					//		"[%s] price interpolation btc -> %f, eth -> %f, current: %f",
					//		symbol,
					//		interpolation.BtcInterpolationUsdt,
					//		interpolation.EthInterpolationUsdt,
					//		kLine.Close,
					//	)
					//} else {
					//	if interpolation.HasBtc() {
					//		log.Printf(
					//			"[%s] price interpolation btc -> %f, current: %f",
					//			symbol,
					//			interpolation.BtcInterpolationUsdt,
					//			kLine.Close,
					//		)
					//	}
					//
					//	if interpolation.HasEth() {
					//		log.Printf(
					//			"[%s] price interpolation eth -> %f, current: %f",
					//			symbol,
					//			interpolation.EthInterpolationUsdt,
					//			kLine.Close,
					//		)
					//	}
					//}
				}
				container.ExchangeRepository.SavePredict(predicted, symbol)
			}
		}
	}(predictChannel, &container)

	go func(container *config.Container) {
		for {
			message := <-eventChannel

			switch true {
			case strings.Contains(string(message), "aggTrade"):
				var tradeEvent model.TradeEvent
				json.Unmarshal(message, &tradeEvent)
				smaDecision := container.SmaTradeStrategy.Decide(tradeEvent.Trade)
				container.ExchangeRepository.SetDecision(smaDecision)

				go func(channel chan string, symbol string) {
					predictChannel <- symbol
				}(predictChannel, tradeEvent.Trade.Symbol)
				break
			case strings.Contains(string(message), "kline"):
				var event model.KlineEvent
				json.Unmarshal(message, &event)
				kLine := event.KlineData.Kline
				kLine.UpdatedAt = time.Now().Unix()
				container.ExchangeRepository.AddKLine(kLine)

				go func(channel chan string, symbol string) {
					predictChannel <- symbol
				}(predictChannel, kLine.Symbol)

				baseKLineDecision := container.BaseKLineStrategy.Decide(kLine)
				container.ExchangeRepository.SetDecision(baseKLineDecision)
				orderBasedDecision := container.OrderBasedStrategy.Decide(kLine)
				container.ExchangeRepository.SetDecision(orderBasedDecision)

				break
			case strings.Contains(string(message), "depth20"):
				var event model.OrderBookEvent
				json.Unmarshal(message, &event)

				depth := event.Depth.ToDepth(strings.ToUpper(strings.ReplaceAll(event.Stream, "@depth20@100ms", "")))
				depthDecision := container.MarketDepthStrategy.Decide(depth)
				container.ExchangeRepository.SetDecision(depthDecision)
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

	for index, streamBatchItem := range getStreamBatch(tradeLimitCollection, []string{"@aggTrade", "@kline_1m@2000ms", "@depth20@100ms"}) {
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
