package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/config"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/http"
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

	router := gin.Default()
	router.Use(http.CORSMiddleware())
	container := config.InitServiceContainer()

	router.GET("/stats/pivot/:symbol", container.StatsController.GetStats)
	router.GET("/stats/pivot/grid", container.StatsController.GetStatsGrid)

	go func() {
		err := router.Run(":8080")
		if err != nil {
			log.Panicln(err)
		}
	}()
	container.Start()
}
