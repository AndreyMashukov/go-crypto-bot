package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"sentiment/src/http"
	"sentiment/src/service"
	"sentiment/src/service/ml"
)

func main() {
	router := gin.Default()
	router.Use(http.CORSMiddleware())

	coinDetector := service.CoinDetector{}
	mlBridge := ml.PythonMLBridge{
		CoinDetector: &coinDetector,
	}
	mlBridge.Initialize()
	defer mlBridge.Finalize()

	controller := http.SentimentController{
		PythonMLBridge: &mlBridge,
	}

	router.POST("/sentiment/predict", controller.PostPredict)

	err := router.Run(":8080")
	if err != nil {
		log.Panicln(err)
	}
}
