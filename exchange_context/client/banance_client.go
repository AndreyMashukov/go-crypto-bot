package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	uuid2 "github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Binance struct {
	ApiKey         string
	ApiSecret      string
	DestinationURI string

	HttpClient   *http.Client
	connection   *websocket.Conn
	channel      chan []byte
	socketWriter chan []byte
}

func (b *Binance) Connect(address string) {
	connection, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	b.channel = make(chan []byte)
	b.socketWriter = make(chan []byte)

	// reader channel
	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Println("read: ", err)

				os.Exit(1)
			}

			b.channel <- message
		}
	}()

	// writer channel
	go func() {
		for {
			serialized := <-b.socketWriter
			_ = b.connection.WriteMessage(websocket.TextMessage, serialized)
		}
	}()

	b.connection = connection
}

func (b *Binance) socketRequest(req model.SocketRequest, channel chan []byte) {
	go func() {
		for {
			msg := <-b.channel

			if strings.Contains(string(msg), req.Id) {
				//log.Printf("[%s], %s", req.Method, string(msg))
				channel <- msg
				return
			}

			b.channel <- msg
		}
	}()

	serialized, _ := json.Marshal(req)
	b.socketWriter <- serialized
}

func (b *Binance) QueryOrder(symbol string, orderId int64) (model.BinanceOrder, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "order.status",
		Params: make(map[string]any),
	}
	socketRequest.Params["apiKey"] = b.ApiKey
	socketRequest.Params["orderId"] = orderId
	socketRequest.Params["symbol"] = symbol
	socketRequest.Params["timestamp"] = time.Now().Unix() * 1000
	socketRequest.Params["signature"] = b.signature(socketRequest.Params)
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.BinanceOrderResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		return model.BinanceOrder{}, errors.New(response.Error.Message)
	}

	return response.Result, nil
}

func (b *Binance) CancelOrder(symbol string, orderId int64) (model.BinanceOrder, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "order.cancel",
		Params: make(map[string]any),
	}
	socketRequest.Params["apiKey"] = b.ApiKey
	socketRequest.Params["orderId"] = orderId
	socketRequest.Params["symbol"] = symbol
	socketRequest.Params["timestamp"] = time.Now().Unix() * 1000
	socketRequest.Params["signature"] = b.signature(socketRequest.Params)
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.BinanceOrderResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		return model.BinanceOrder{}, errors.New(response.Error.Message)
	}

	return response.Result, nil
}

func (b *Binance) GetOpenedOrders() (*[]model.BinanceOrder, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "openOrders.status",
		Params: make(map[string]any),
	}
	socketRequest.Params["apiKey"] = b.ApiKey
	socketRequest.Params["timestamp"] = time.Now().Unix() * 1000
	socketRequest.Params["signature"] = b.signature(socketRequest.Params)
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.BinanceOrderListResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		log.Println(socketRequest)
		list := make([]model.BinanceOrder, 0)
		return &list, errors.New(response.Error.Message)
	}

	return &response.Result, nil
}

func (b *Binance) GetExchangeData(symbols []string) (*model.ExchangeInfo, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "exchangeInfo",
		Params: make(map[string]any),
	}
	socketRequest.Params["symbols"] = symbols
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.BinanceExchangeInfoResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		log.Println(socketRequest)
		return &model.ExchangeInfo{}, errors.New(response.Error.Message)
	}

	return &response.Result, nil
}

func (b *Binance) GetAccountStatus() (*model.AccountStatus, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "account.status",
		Params: make(map[string]any),
	}

	socketRequest.Params["apiKey"] = b.ApiKey
	socketRequest.Params["timestamp"] = time.Now().Unix() * 1000
	socketRequest.Params["signature"] = b.signature(socketRequest.Params)
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.AccountStatusResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		log.Println(socketRequest)

		return nil, errors.New(response.Error.Message)
	}

	return &response.Result, nil
}

func (b *Binance) GetTrades(order model.Order) ([]model.MyTrade, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "myTrades",
		Params: make(map[string]any),
	}

	socketRequest.Params["apiKey"] = b.ApiKey
	socketRequest.Params["timestamp"] = time.Now().Unix() * 1000
	socketRequest.Params["symbol"] = order.Symbol
	socketRequest.Params["signature"] = b.signature(socketRequest.Params)
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.TradesResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		log.Println(socketRequest)
		list := make([]model.MyTrade, 0)
		return list, errors.New(response.Error.Message)
	}

	return response.Result, nil
}

func (b *Binance) LimitOrder(order model.Order, operation string) (model.BinanceOrder, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "order.place",
		Params: make(map[string]any),
	}
	socketRequest.Params["symbol"] = order.Symbol
	socketRequest.Params["side"] = operation
	socketRequest.Params["type"] = "LIMIT"
	socketRequest.Params["quantity"] = strconv.FormatFloat(order.Quantity, 'f', -1, 64)
	socketRequest.Params["timeInForce"] = "GTC"
	socketRequest.Params["price"] = strconv.FormatFloat(order.Price, 'f', -1, 64)
	socketRequest.Params["apiKey"] = b.ApiKey
	socketRequest.Params["timestamp"] = time.Now().Unix() * 1000
	socketRequest.Params["signature"] = b.signature(socketRequest.Params)
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.BinanceOrderResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		log.Printf("[%s] Limit Order: %s -> %s", order.Symbol, response.Error.Message, socketRequest)
		return model.BinanceOrder{}, errors.New(response.Error.Message)
	}

	return response.Result, nil
}

func (b *Binance) signature(params map[string]any) string {
	parts := make([]string, 0)

	keys := make([]string, 0, len(params))

	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, params[key]))
	}

	mac := hmac.New(sha256.New, []byte(b.ApiSecret))
	mac.Write([]byte(strings.Join(parts, "&")))
	signingKey := fmt.Sprintf("%x", mac.Sum(nil))

	return signingKey
}

func (b *Binance) sign(url string) string {
	mac := hmac.New(sha256.New, []byte(b.ApiSecret))
	mac.Write([]byte(url))
	signingKey := fmt.Sprintf("%x", mac.Sum(nil))

	return signingKey
}
