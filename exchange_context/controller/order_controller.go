package controller

import (
	"encoding/json"
	"fmt"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"net/http"
)

type OrderController struct {
	OrderRepository *ExchangeRepository.OrderRepository
}

func (o *OrderController) GetOrderListAction(w http.ResponseWriter, req *http.Request) {
	list := o.OrderRepository.GetList()
	encoded, _ := json.Marshal(list)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(encoded))
}
