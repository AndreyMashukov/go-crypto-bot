package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gitlab.com/open-soft/go-crypto-bot/src/config"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"os"
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
	container.PingDB()

	defer container.Db.Close()
	defer container.DbSwap.Close()
	container.PythonMLBridge.Initialize()
	defer container.PythonMLBridge.Finalize()
	container.StartHttpServer()
	log.Printf("Bot [%s] is initialized successfully", container.CurrentBot.BotUuid)

	container.Binance.Connect(os.Getenv("BINANCE_WS_DSN"))

	usdtBalance, err := container.BalanceService.GetAssetBalance("USDT", false)
	if err != nil {
		log.Printf("Balance check error: %s", err.Error())

		// todo: `Invalid account.`
		if err.Error() == model.BinanceErrorInvalidAPIKeyOrPermissions {
			log.Println("Notify SaaS system about error")
			container.CallbackManager.Error(
				*container.CurrentBot,
				model.BinanceErrorInvalidAPIKeyOrPermissions,
				"Please check API Key permissions or IP address binding",
				true,
			)
		}

		os.Exit(0)
	}
	log.Printf("API Key permission check passed, balance is: %.2f", usdtBalance)
	container.PythonMLBridge.StartAutoLearn()
	container.Binance.APIKeyCheckCompleted = true

	container.MakerService.RecoverOrders()

	if container.IsMasterBot {
		container.MakerService.UpdateSwapPairs()
		go func() {
			container.MarketSwapListener.ListenAll()
		}()
		go func() {
			container.MCListener.ListenAll()
		}()
	}

	container.TimeService.WaitSeconds(10)
	container.MakerService.StartTrade()

	container.MarketTradeListener.ListenAll()
}
