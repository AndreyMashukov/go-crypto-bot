package controller

import (
	"encoding/json"
	"fmt"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"net/http"
	"strings"
)

type ExchangeController struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	ChartService       *service.ChartService
}

func (e *ExchangeController) GetKlineListAction(w http.ResponseWriter, req *http.Request) {
	symbol := strings.TrimPrefix(req.URL.Path, "/kline/list/")

	list := e.ExchangeRepository.KLineList(symbol, true)
	encoded, _ := json.Marshal(list)
	w.Header().Set("content-type", "application/json")
	fmt.Fprintf(w, string(encoded))
}

func (e *ExchangeController) GetDepthAction(w http.ResponseWriter, req *http.Request) {
	symbol := strings.TrimPrefix(req.URL.Path, "/depth/")

	list := e.ExchangeRepository.GetDepth(symbol)
	encoded, _ := json.Marshal(list)
	w.Header().Set("content-type", "application/json")
	fmt.Fprintf(w, string(encoded))
}

func (e *ExchangeController) GetTradeListAction(w http.ResponseWriter, req *http.Request) {
	symbol := strings.TrimPrefix(req.URL.Path, "/trade/list/")

	list := e.ExchangeRepository.TradeList(symbol)
	encoded, _ := json.Marshal(list)
	w.Header().Set("content-type", "application/json")
	fmt.Fprintf(w, string(encoded))
}

func (e *ExchangeController) GetChartListAction(w http.ResponseWriter, req *http.Request) {
	chart := e.ChartService.GetCharts()
	encoded, _ := json.Marshal(chart)
	w.Header().Set("content-type", "application/json")
	fmt.Fprintf(w, string(encoded))
}
