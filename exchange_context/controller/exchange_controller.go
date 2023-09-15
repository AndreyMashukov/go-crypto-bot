package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"net/http"
	"strings"
	"time"
)

type ExchangeController struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	ChartService       *service.ChartService
	RDB                *redis.Client
	Ctx                *context.Context
}

func (e *ExchangeController) GetKlineListAction(w http.ResponseWriter, req *http.Request) {
	symbol := strings.TrimPrefix(req.URL.Path, "/kline/list/")

	list := e.ExchangeRepository.KLineList(symbol, true, 200)
	encoded, _ := json.Marshal(list)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(encoded))
}

func (e *ExchangeController) GetDepthAction(w http.ResponseWriter, req *http.Request) {
	symbol := strings.TrimPrefix(req.URL.Path, "/depth/")

	list := e.ExchangeRepository.GetDepth(symbol)
	encoded, _ := json.Marshal(list)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(encoded))
}

func (e *ExchangeController) GetTradeListAction(w http.ResponseWriter, req *http.Request) {
	symbol := strings.TrimPrefix(req.URL.Path, "/trade/list/")

	list := e.ExchangeRepository.TradeList(symbol)
	encoded, _ := json.Marshal(list)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(encoded))
}

func (e *ExchangeController) GetChartListAction(w http.ResponseWriter, req *http.Request) {
	encoded := e.RDB.Get(*e.Ctx, "chart-cache").Val()

	if len(encoded) == 0 {
		chart := e.ChartService.GetCharts()
		encodedRes, _ := json.Marshal(chart)
		encoded = string(encodedRes)
		e.RDB.Set(*e.Ctx, "chart-cache", encoded, time.Second*5)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, encoded)
}
