package controller

import (
	"encoding/json"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"net/http"
	"slices"
)

type OrderController struct {
	OrderRepository    *ExchangeRepository.OrderRepository
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Formatter          *service.Formatter
	MakerService       *service.MakerService
}

func (o *OrderController) GetOrderListAction(w http.ResponseWriter, req *http.Request) {
	list := o.OrderRepository.GetList()
	encoded, _ := json.Marshal(list)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(encoded))
}

func (o *OrderController) PostManualOrderAction(w http.ResponseWriter, req *http.Request) {
	if req.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "OK")
		return
	}

	if req.Method != "POST" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Разрешены только POST методы", http.StatusMethodNotAllowed)

		return
	}

	var manual model.ManualOrder

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&manual)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	allowedOperations := []string{"BUY", "SELL"}
	if !slices.Contains(allowedOperations, manual.Operation) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Поддерживаются только операции BUY/SELL", http.StatusBadRequest)

		return
	}

	tradeLimit, err := o.ExchangeRepository.GetTradeLimit(manual.Symbol)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf("%s не поддерживается", manual.Symbol), http.StatusBadRequest)

		return
	}

	opened, err := o.OrderRepository.GetOpenedOrderCached(manual.Symbol, "BUY")
	if err == nil && manual.Operation == "SELL" {
		minPrice := o.Formatter.FormatPrice(tradeLimit, opened.Price*(100+tradeLimit.MinProfitPercent)/100)
		if minPrice > manual.Price {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, fmt.Sprintf("Цена не может быть ниже %.6f", minPrice), http.StatusBadRequest)

			return
		}
	}

	if err != nil && manual.Operation == "SELL" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Нет открытых ордеров", http.StatusBadRequest)

		return
	}

	if err == nil && manual.Operation == "BUY" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Докупать вручную временно запрещено", http.StatusBadRequest)

		return
	}

	minPrice, err := o.MakerService.CalculateBuyPrice(tradeLimit)

	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf("Ошибка: %s", err.Error()), http.StatusBadRequest)

		return
	}

	if err != nil && manual.Operation == "BUY" && minPrice < manual.Price {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf("Покупать выше цены %f запрещено", minPrice), http.StatusBadRequest)

		return
	}

	binanceOrder := o.OrderRepository.GetBinanceOrder(manual.Symbol, manual.Operation)

	if binanceOrder != nil && binanceOrder.Status == "PARTIALLY_FILLED" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Ордер исполняется, дождитесь завершения операции", http.StatusBadRequest)

		return
	}

	manual.Price = o.Formatter.FormatPrice(tradeLimit, manual.Price)
	o.OrderRepository.SetManualOrder(manual)

	encoded, _ := json.Marshal(manual)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(encoded))
}
