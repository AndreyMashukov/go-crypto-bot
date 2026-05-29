package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	uuid2 "github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
)

type ExchangeOrderAPIInterface interface {
	LimitOrder(symbol string, quantity float64, price float64, operation string, timeInForce string) (model.BinanceOrder, error)
	QueryOrder(symbol string, orderId string) (model.BinanceOrder, error)
	CancelOrder(symbol string, orderId string) (model.BinanceOrder, error)
	GetOpenedOrders() ([]model.BinanceOrder, error)
}

type ExchangePriceAPIInterface interface {
	GetOpenedOrders() ([]model.BinanceOrder, error)
	GetDepth(symbol string, limit int64) *model.OrderBook
	GetKLines(symbol string, interval string, limit int64) []model.KLineHistory
	GetKLinesCached(symbol string, interval string, limit int64) []model.KLine
	GetExchangeData(symbols []string) (*model.ExchangeInfo, error)
}

type ExchangeAPIInterface interface {
	QueryOrder(symbol string, orderId string) (model.BinanceOrder, error)
	CancelOrder(symbol string, orderId string) (model.BinanceOrder, error)
	GetDepth(symbol string, limit int64) *model.OrderBook
	GetOpenedOrders() ([]model.BinanceOrder, error)
	GetKLines(symbol string, interval string, limit int64) []model.KLineHistory
	TradesAggregate(symbol string, limit int64, startTime int64, endTime int64) []model.Trade
	GetKLinesCached(symbol string, interval string, limit int64) []model.KLine
	GetExchangeData(symbols []string) (*model.ExchangeInfo, error)
	GetAccountStatus() (*model.AccountStatus, error)
	GetTickers(symbols []string) []model.WSTickerPrice
	LimitOrder(symbol string, quantity float64, price float64, operation string, timeInForce string) (model.BinanceOrder, error)
	IsConnected() bool
	IsWaitMode() bool
	IsAPIKeyCheckCompleted() bool
}

type Binance struct {
	CurrentBot *model.Bot
	ApiKey     string
	ApiSecret  string

	HttpClient   *http.Client
	connection   *websocket.Conn
	Channel      chan []byte
	SocketWriter chan []byte
	RDB          *redis.Client
	Ctx          *context.Context

	WaitMode             bool
	Connected            bool
	APIKeyCheckCompleted bool
	Lock                 *sync.Mutex
}

func (b *Binance) IsConnected() bool {
	return b.Connected
}

func (b *Binance) IsWaitMode() bool {
	return b.IsWaitingMode()
}

func (b *Binance) IsAPIKeyCheckCompleted() bool {
	return b.APIKeyCheckCompleted
}

func (b *Binance) IsWaitingMode() bool {
	b.Lock.Lock()
	isWaiting := b.WaitMode
	b.Lock.Unlock()

	return isWaiting
}

func (b *Binance) SetWaitingMode(isEnabled bool) {
	b.Lock.Lock()
	b.WaitMode = isEnabled
	b.Lock.Unlock()
}

func (b *Binance) CheckWait() {
	for {
		if !b.IsWaitingMode() {
			break
		}
	}
}

func (b *Binance) Connect(address string) {
	connection, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		b.Connected = false
		log.Printf("Binance WS [%s]: %s, wait and reconnect...", address, err.Error())
		time.Sleep(time.Second * 10)
		b.Connect(address)
		return
	}

	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Println("read: ", err)

				_ = connection.Close()
				b.Connected = false
				log.Printf("Binance WS, wait and reconnect...")
				time.Sleep(time.Second * 10)
				b.Connect(address)
				return
			}

			b.Channel <- message
		}
	}()

	go func() {
		for {
			serialized := <-b.SocketWriter
			_ = b.connection.WriteMessage(websocket.TextMessage, serialized)
		}
	}()

	b.connection = connection
	b.Connected = true
	b.connection.SetPingHandler(nil)
}

func (b *Binance) socketRequest(req model.SocketRequest, channel chan []byte) {
	b.CheckWait()

	go func(req model.SocketRequest) {
		for {
			msg := <-b.Channel

			if strings.Contains(string(msg), "Too much request weight used; current limit is 6000 request weight per 1 MINUTE") {
				b.SetWaitingMode(true)

				log.Printf(
					"[%s] Socket error [%s]: %s, wait 1 min and retry...",
					req.Method,
					req.Id,
					string(msg),
				)

				time.Sleep(time.Minute)
				serialized, _ := json.Marshal(req)
				b.SetWaitingMode(false)

				b.SocketWriter <- serialized
				log.Printf("[%s] retried...", req.Id)

				continue
			}

			if strings.Contains(string(msg), req.Id) {
				channel <- msg
				return
			}

			b.Channel <- msg
			time.Sleep(time.Millisecond)
		}
	}(req)

	serialized, _ := json.Marshal(req)
	b.SocketWriter <- serialized
}

func (b *Binance) QueryOrder(symbol string, orderId string) (model.BinanceOrder, error) {
	b.CheckWait()

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
		return model.BinanceOrder{}, errors.New(response.Error.GetMessage())
	}

	return response.Result.ToModern(), nil
}

func (b *Binance) CancelOrder(symbol string, orderId string) (model.BinanceOrder, error) {
	b.CheckWait()

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
		return model.BinanceOrder{}, errors.New(response.Error.GetMessage())
	}

	return response.Result.ToModern(), nil
}

func (b *Binance) UserDataStreamStart() (model.UserDataStreamStart, error) {
	b.CheckWait()

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
		return model.UserDataStreamStart{}, errors.New(response.Error.GetMessage())
	}

	return response.Result, nil
}

func (b *Binance) GetDepth(symbol string, limit int64) *model.OrderBook {
	b.CheckWait()

	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "depth",
		Params: make(map[string]any),
	}
	socketRequest.Params["limit"] = limit
	socketRequest.Params["symbol"] = symbol
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.OrderBookResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		return nil
	}

	return &response.Result
}

func (b *Binance) GetOpenedOrders() ([]model.BinanceOrder, error) {
	b.CheckWait()

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
		return list, errors.New(response.Error.GetMessage())
	}

	list := make([]model.BinanceOrder, 0)
	for _, orderLegacy := range response.Result {
		list = append(list, orderLegacy.ToModern())
	}

	return list, nil
}

func (b *Binance) GetKLines(symbol string, interval string, limit int64) []model.KLineHistory {
	b.CheckWait()

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

func (b *Binance) TradesAggregate(symbol string, limit int64, startTime int64, endTime int64) []model.Trade {
	b.CheckWait()

	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "trades.aggregate",
		Params: make(map[string]any),
	}

	socketRequest.Params["symbol"] = symbol
	socketRequest.Params["limit"] = limit
	if startTime > 0 {
		socketRequest.Params["startTime"] = startTime
	}
	if endTime > 0 {
		socketRequest.Params["endTime"] = endTime
	}
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.BinanceAggTradesResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		log.Println(socketRequest)
		list := make([]model.Trade, 0)
		return list
	}

	return response.Result
}

func (b *Binance) GetKLinesCached(symbol string, interval string, limit int64) []model.KLine {
	cacheKey := fmt.Sprintf("interval-klines-history-%s-%s-%d-%d", symbol, interval, limit, b.CurrentBot.Id)

	res := b.RDB.Get(*b.Ctx, cacheKey).Val()
	if len(res) > 0 {
		var batch model.KlineBatch

		err := json.Unmarshal([]byte(res), &batch)
		if err == nil {
			return batch.Items
		}
		log.Printf("[%s] kline[%s] history cache invalid", symbol, interval)
	}

	historyKLines := b.GetKLines(symbol, interval, limit)
	kLines := make([]model.KLine, 0)
	for _, historyKLine := range historyKLines {
		kLines = append(kLines, historyKLine.ToKLine(symbol))
	}

	batch := model.KlineBatch{
		Items: kLines,
	}
	encoded, err := json.Marshal(batch)
	if err == nil {
		b.RDB.Set(*b.Ctx, cacheKey, string(encoded), time.Second*15)
	}

	return batch.Items
}

func (b *Binance) GetExchangeData(symbols []string) (*model.ExchangeInfo, error) {
	b.CheckWait()

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
		return &model.ExchangeInfo{}, errors.New(response.Error.GetMessage())
	}

	return &response.Result, nil
}

func (b *Binance) GetAccountStatus() (*model.AccountStatus, error) {
	b.CheckWait()

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

		return nil, errors.New(response.Error.GetMessage())
	}

	return &response.Result, nil
}

func (b *Binance) GetTrades(order model.Order) ([]model.MyTrade, error) {
	b.CheckWait()

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
		return list, errors.New(response.Error.GetMessage())
	}

	return response.Result, nil
}

func (b *Binance) GetTickers(symbols []string) []model.WSTickerPrice {
	b.CheckWait()

	channel := make(chan []byte)
	defer close(channel)

	socketRequest := model.SocketRequest{
		Id:     uuid2.New().String(),
		Method: "ticker.price",
		Params: make(map[string]any),
	}

	socketRequest.Params["symbols"] = symbols
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.BinanceTickersPriceResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		log.Println(response.Error)
		list := make([]model.WSTickerPrice, 0)
		return list
	}

	return response.Result
}

func (b *Binance) LimitOrder(symbol string, quantity float64, price float64, operation string, timeInForce string) (model.BinanceOrder, error) {
	b.CheckWait()

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
	socketRequest.Params["timeInForce"] = timeInForce
	socketRequest.Params["price"] = strconv.FormatFloat(price, 'f', -1, 64)
	socketRequest.Params["apiKey"] = b.ApiKey
	socketRequest.Params["timestamp"] = time.Now().Unix() * 1000
	socketRequest.Params["signature"] = b.signature(socketRequest.Params)
	b.socketRequest(socketRequest, channel)
	message := <-channel

	var response model.BinanceOrderResponse
	json.Unmarshal(message, &response)

	if response.Error != nil {
		log.Printf("[%s] Limit Order: %s -> %s", symbol, response.Error.GetMessage(), socketRequest)

		if response.Error.IsNotional() {
			log.Printf("[%s] Sleep 1 minute", symbol)
			time.Sleep(time.Minute)
		}

		return model.BinanceOrder{}, errors.New(response.Error.GetMessage())
	}

	return response.Result.ToModern(), nil
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
