package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/config"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
	"os"
	"slices"
	"time"
)

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

	// ToDo: Get bot exchange from the bot entity!
	// ToDo: Exchange Interface!
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

	// todo: Exchange Interface
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

	swapKlineChannel := make(chan []byte)
	swapUpdateChannel := make(chan string)
	defer close(swapKlineChannel)
	defer close(swapUpdateChannel)

	if container.IsMasterBot {
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
		go func(swapUpdateChannel chan string) {
			for {
				swapUpdateSymbol := <-swapUpdateChannel
				swapPair, err := container.ExchangeRepository.GetSwapPair(swapUpdateSymbol)
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
		}(swapUpdateChannel)
	}

	for _, limit := range tradeLimits {
		// ML learn every 1000 minutes
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

		// Trade engine (Maker)
		go func(symbol string, container *config.Container) {
			for {
				currentDecisions := container.ExchangeRepository.GetDecisions(symbol)

				if len(currentDecisions) > 0 {
					container.MakerService.Make(symbol, currentDecisions)
				}
				time.Sleep(time.Millisecond * 500)
			}
		}(limit.Symbol, &container)

		// History update
		go func(tradeLimit model.TradeLimit) {
			klineAmount := 0

			history := container.Binance.GetKLines(tradeLimit.GetSymbol(), "1m", 200)

			for _, kline := range history {
				klineAmount++
				container.ExchangeRepository.AddKLine(kline.ToKLine(tradeLimit.GetSymbol()))
			}
			log.Printf("Loaded history %s -> %d klines", tradeLimit.Symbol, klineAmount)
		}(limit)
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

	go func(container *config.Container) {
		for {
			depth := <-depthChannel
			container.ExchangeRepository.SetDepth(depth)
		}
	}(&container)

	// todo: Exchange Interface
	container.Binance.ListenAll(
		tradeLimits,
		container,
		swapKlineChannel,
		swapUpdateChannel,
		predictChannel,
		depthChannel,
	)

	// just to keep running
	runChannel <- "run"
	log.Panic("Stopped")
}
