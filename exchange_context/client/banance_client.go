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
	Channel      chan []byte
	SocketWriter chan []byte
}

func (b *Binance) Connect(address string) {
	connection, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Printf("Binance WS [%s]: %s", address, err.Error())
		log.Fatal("Quit!")
	}

	// 2023/12/11 05:56:32 [SOLUSDT] QueryOrder: Too much request weight used; current limit is 6000 request weight per 1 MINUTE. Please use WebSocket Streams for live updates to avoid polling the API.
	// 2023/12/11 05:56:32 [SOLUSDT] Retry query order...
	// 2023/12/11 05:56:38 [AVAXUSDT] Opened: Way too much request weight used; IP banned until 1702275878212. Please use WebSocket Streams for live updates to avoid bans.
	// 2023/12/11 05:56:38 read:  websocket: close 1008 (policy violation): disconnected

	// reader channel
	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Println("read: ", err)

				_ = connection.Close()
				log.Printf("Binance WS, wait and reconnect...")
				time.Sleep(time.Second * 20)
				b.Connect(address)
				return
			}

			b.Channel <- message
		}
	}()

	// writer channel
	go func() {
		for {
			serialized := <-b.SocketWriter
			_ = b.connection.WriteMessage(websocket.TextMessage, serialized)
		}
	}()

	b.connection = connection
}

func (b *Binance) socketRequest(req model.SocketRequest, channel chan []byte) {

	go func(req model.SocketRequest) {
		for {
			msg := <-b.Channel

			if strings.Contains(string(msg), "Too much request weight used; current limit is 6000 request weight per 1 MINUTE") {
				log.Printf("Socket error [%s]: %s, wait 1 min and retry...", req.Id, string(msg))
				time.Sleep(time.Minute)
				serialized, _ := json.Marshal(req)
				b.SocketWriter <- serialized
				log.Printf("[%s] retried...", req.Id)

				continue
			}

			if strings.Contains(string(msg), req.Id) {
				//log.Printf("[%s], %s", req.Method, string(msg))
				channel <- msg
				return
			}

			b.Channel <- msg
		}
	}(req)

	serialized, _ := json.Marshal(req)
	b.SocketWriter <- serialized
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

func (b *Binance) UserDataStreamStart() (model.UserDataStreamStart, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "userDataStream.start",
		Params: make(map[string]any),
	}
	socketRequest.Params["apiKey"] = b.ApiKey
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.UserDataStreamStartResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		return model.UserDataStreamStart{}, errors.New(response.Error.Message)
	}

	return response.Result, nil
}

func (b *Binance) GetDepth(symbol string) (model.OrderBook, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "depth",
		Params: make(map[string]any),
	}
	socketRequest.Params["limit"] = 20
	socketRequest.Params["symbol"] = symbol
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.OrderBookResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		return model.OrderBook{}, errors.New(response.Error.Message)
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

func (b *Binance) GetKLines(symbol string, interval string, limit int64) []model.KLineHistory {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "klines",
		Params: make(map[string]any),
	}

	socketRequest.Params["symbol"] = symbol
	socketRequest.Params["interval"] = interval
	socketRequest.Params["limit"] = limit
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.BinanceKLineResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		log.Println(socketRequest)
		list := make([]model.KLineHistory, 0)
		return list
	}

	return response.Result
}

func (b *Binance) GetExchangeData(symbols []string) (*model.ExchangeInfo, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "exchangeInfo",
		Params: make(map[string]any),
	}
	if len(symbols) > 0 {
		socketRequest.Params["symbols"] = symbols
	}
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

func (b *Binance) LimitOrder(symbol string, quantity float64, price float64, operation string) (model.BinanceOrder, error) {
	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "order.place",
		Params: make(map[string]any),
	}
	socketRequest.Params["symbol"] = symbol
	socketRequest.Params["side"] = operation
	socketRequest.Params["type"] = "LIMIT"
	socketRequest.Params["quantity"] = strconv.FormatFloat(quantity, 'f', -1, 64)
	// [FOK] - Fill or kill (FOK) is a conditional type of time-in-force order used in
	// securities trading that instructs a brokerage to execute a
	// transaction immediately and completely or not at all.
	// This type of order is most often used by active traders and is usually for a
	// large quantity of stock. The order must be filled in its entirety or else canceled (killed)
	// --------
	// [IOC] - An immediate or cancel order (IOC) is an order to buy or sell a security that attempts
	// to execute all or part immediately and then cancels any unfilled portion of the order.
	// An IOC order is one of several "duration," or time in force orders, that investors can use
	// to specify how long the order remains active in the market and under what conditions the order is canceled.
	// -------
	// [GTC] - Good ’til canceled (GTC) describes a type of order that an investor may place to buy or sell
	// a security that remains active until either the order is filled or the investor cancels it.
	// Brokerages will typically limit the maximum time you can keep a GTC order open (active) to 90 days.
	socketRequest.Params["timeInForce"] = "GTC"
	socketRequest.Params["price"] = strconv.FormatFloat(price, 'f', -1, 64)
	socketRequest.Params["apiKey"] = b.ApiKey
	socketRequest.Params["timestamp"] = time.Now().Unix() * 1000
	socketRequest.Params["signature"] = b.signature(socketRequest.Params)
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.BinanceOrderResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		log.Printf("[%s] Limit Order: %s -> %s", symbol, response.Error.Message, socketRequest)

		if strings.Contains(response.Error.Message, "Filter failure: NOTIONAL") {
			log.Printf("[%s] Sleep 1 minute", symbol)
			time.Sleep(time.Minute) // wait one minute
		}

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
