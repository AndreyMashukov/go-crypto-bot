package controller

import (
	"encoding/json"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/validator"
	"log"
	"net/http"
	"strings"
)

type TradeController struct {
	CurrentBot          *model.Bot
	ExchangeRepository  *repository.ExchangeRepository
	TradeStack          *exchange.TradeStack
	TradeLimitValidator *validator.TradeLimitValidator
	SignalRepository    *repository.SignalRepository
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

	violation := t.TradeLimitValidator.Validate(tradeLimit)

	if violation != nil {
		http.Error(w, violation.Error(), http.StatusBadRequest)

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
	t.TradeStack.InvalidateBuyPriceCache(tradeLimit.Symbol)

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

	violation := t.TradeLimitValidator.Validate(tradeLimit)

	if violation != nil {
		http.Error(w, violation.Error(), http.StatusBadRequest)

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

func (t *TradeController) PostSignalAction(w http.ResponseWriter, req *http.Request) {
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
		http.Error(w, "Only POST method are allowed", http.StatusMethodNotAllowed)

		return
	}

	var signal model.Signal

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&signal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if t.CurrentBot.Exchange != signal.Exchange {
		http.Error(w, fmt.Sprintf("Wrong exchange '%s', expected: %s", signal.Exchange, t.CurrentBot.Exchange), http.StatusBadRequest)

		return
	}

	t.SignalRepository.SaveSignal(signal)
	t.TradeStack.InvalidateBuyPriceCache(signal.Symbol)
	fmt.Fprintf(w, "OK")
}

func (t *TradeController) GetTradeStackAction(w http.ResponseWriter, req *http.Request) {
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
		http.Error(w, "Only GET method are allowed", http.StatusMethodNotAllowed)

		return
	}

	stack := t.TradeStack.GetTradeStack(exchange.TradeStackParams{
		SkipFiltered:    false,
		SkipLocked:      false,
		SkipDisabled:    false,
		BalanceFilter:   false,
		SkipPending:     false,
		WithValidPrice:  false,
		AttachDecisions: true,
	})

	encodedRes, err := json.Marshal(stack)
	if err != nil {
		log.Printf("Trade stack marshal error: %s", err.Error())
		http.Error(w, "Something went wrong", http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintf(w, string(encodedRes))
}

func (t *TradeController) SwitchTradeLimitAction(w http.ResponseWriter, req *http.Request) {
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

	symbol := strings.TrimPrefix(req.URL.Path, "/trade/limit/switch/")

	entity, err := t.ExchangeRepository.GetTradeLimit(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	entity.IsEnabled = !entity.IsEnabled
	err = t.ExchangeRepository.UpdateTradeLimit(entity)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	entity, err = t.ExchangeRepository.GetTradeLimit(entity.Symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)

		return
	}

	encodedRes, _ := json.Marshal(entity)
	fmt.Fprintf(w, string(encodedRes))
}
