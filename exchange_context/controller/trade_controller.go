package controller

import (
	"encoding/json"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"net/http"
)

type TradeController struct {
	CurrentBot         *model.Bot
	ExchangeRepository *ExchangeRepository.ExchangeRepository
}

func (t *TradeController) UpdateTradeLimitAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if req.Method == "OPTIONS" {
		fmt.Fprintf(w, "OK")
		return
	}

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != t.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	if req.Method != "PUT" {
		http.Error(w, "Разрешены только PUT методы", http.StatusMethodNotAllowed)

		return
	}

	var tradeLimit model.TradeLimit

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&tradeLimit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	entity, err := t.ExchangeRepository.GetTradeLimit(tradeLimit.Symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	tradeLimit.Id = entity.Id
	err = t.ExchangeRepository.UpdateTradeLimit(tradeLimit)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	entity, err = t.ExchangeRepository.GetTradeLimit(tradeLimit.Symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)

		return
	}

	encodedRes, _ := json.Marshal(entity)
	fmt.Fprintf(w, string(encodedRes))
}

func (t *TradeController) CreateTradeLimitAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if req.Method == "OPTIONS" {
		fmt.Fprintf(w, "OK")
		return
	}

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != t.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	if req.Method != "POST" {
		http.Error(w, "Разрешены только POST методы", http.StatusMethodNotAllowed)

		return
	}

	var tradeLimit model.TradeLimit

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&tradeLimit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	_, err = t.ExchangeRepository.GetTradeLimit(tradeLimit.Symbol)
	if err == nil {
		http.Error(w, "Trade limit has already existed", http.StatusBadRequest)

		return
	}

	_, err = t.ExchangeRepository.CreateTradeLimit(tradeLimit)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	entity, err := t.ExchangeRepository.GetTradeLimit(tradeLimit.Symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)

		return
	}

	encodedRes, _ := json.Marshal(entity)
	fmt.Fprintf(w, string(encodedRes))
}

func (t *TradeController) GetTradeLimitsAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if req.Method == "OPTIONS" {
		fmt.Fprintf(w, "OK")
		return
	}

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != t.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	if req.Method != "GET" {
		http.Error(w, "Разрешены только GET методы", http.StatusMethodNotAllowed)

		return
	}

	limits := t.ExchangeRepository.GetTradeLimits()

	encodedRes, _ := json.Marshal(limits)
	fmt.Fprintf(w, string(encodedRes))
}
