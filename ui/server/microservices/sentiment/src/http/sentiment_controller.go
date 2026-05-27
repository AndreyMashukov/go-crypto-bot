package http

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"sentiment/src/form"
	"sentiment/src/service/ml"
)

type SentimentController struct {
	PythonMLBridge *ml.PythonMLBridge
}

func (s *SentimentController) PostPredict(context *gin.Context) {
	jsonData, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.JSON(400, err.Error())
		return
	}
	var sentiment form.Sentiment
	err = json.Unmarshal(jsonData, &sentiment)
	if err != nil {
		log.Println(err)
		context.JSON(400, err.Error())
		return
	}

	result, err := s.PythonMLBridge.Predict(sentiment.Text)

	if err != nil {
		context.JSON(503, err.Error())
		return
	}

	context.JSON(200, result)
}
