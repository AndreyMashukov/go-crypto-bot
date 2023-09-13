package controller

import (
	"encoding/json"
	"fmt"
	ExchangeRepository "github.com/AndreyMashukov/go-crypto-bot/server/exchange_context/repository"
	"net/http"
	"strings"
)

type ExchangeController struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
}

func (e *ExchangeController) GetKlineListAction(w http.ResponseWriter, req *http.Request) {
	symbol := strings.TrimPrefix(req.URL.Path, "/kline/list/")

	list := e.ExchangeRepository.KLineList(symbol)
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
