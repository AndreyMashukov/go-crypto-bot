package controller

import (
	"encoding/json"
	"fmt"
	ExchangeRepository "github.com/AndreyMashukov/go-crypto-bot/server/exchange_context/repository"
	"net/http"
)

type OrderController struct {
	OrderRepository *ExchangeRepository.OrderRepository
}

func (o *OrderController) GetOrderListAction(w http.ResponseWriter, req *http.Request) {
	list := o.OrderRepository.GetList()
	encoded, _ := json.Marshal(list)
	w.Header().Set("content-type", "application/json")
	fmt.Fprintf(w, string(encoded))
}
