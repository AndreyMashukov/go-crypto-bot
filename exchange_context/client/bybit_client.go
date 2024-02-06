package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	bybit "github.com/wuhewuhe/bybit.go.api"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/config"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log"
	"net/http"
	"strings"
	"time"
)

type ByBitClient struct {
	ApiKey         string
	ApiSecret      string
	DestinationURI string

	HttpClient   *http.Client
	connection   *websocket.Conn
	Channel      chan []byte
	SocketWriter chan []byte
	RDB          *redis.Client
	Ctx          *context.Context

	WaitMode             bool
	Connected            bool
	APIKeyCheckCompleted bool
	IsMasterBot          bool
}

func (b *ByBitClient) CheckWait() {
	for {
		if !b.WaitMode {
			break
		}
	}
}

func (b *ByBitClient) Connect(address string) {
	connection, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		b.Connected = false
		log.Printf("Binance WS [%s]: %s, wait and reconnect...", address, err.Error())
		time.Sleep(time.Second * 10)
		b.Connect(address)
		return
	}

	// reader channel
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

	// writer channel
	go func() {
		for {
			serialized := <-b.SocketWriter
			_ = b.connection.WriteMessage(websocket.TextMessage, serialized)
		}
	}()

	b.connection = connection
	b.Connected = true
}

func (b *ByBitClient) socketRequest(req model.SocketRequest, channel chan []byte) {
	b.CheckWait()

	go func(req model.SocketRequest) {
		for {
			msg := <-b.Channel

			// todo: stable check???
			if strings.Contains(string(msg), "Too much request weight used; current limit is 6000 request weight per 1 MINUTE") {
				b.WaitMode = true

				log.Printf(
					"[%s] Socket error [%s]: %s, wait 1 min and retry...",
					req.Method,
					req.Id,
					string(msg),
				)

				time.Sleep(time.Minute)
				serialized, _ := json.Marshal(req)
				b.WaitMode = false
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
func (b *ByBitClient) catchEvent(entityId string, channel chan []byte) {
	b.CheckWait()

	go func(entityId string) {
		for {
			msg := <-b.Channel

			if strings.Contains(string(msg), entityId) {
				channel <- msg
				return
			}

			b.Channel <- msg
		}
	}(entityId)
}
func (b *ByBitClient) QueryOrder(symbol string, orderId int64) (model.BinanceOrder, error) {
	client := bybit.NewBybitHttpClient(b.ApiKey, b.ApiSecret, bybit.WithBaseURL(b.DestinationURI))
	params := map[string]interface{}{
		"symbol":  symbol,
		"orderId": fmt.Sprintf("%d", orderId),
	}
	var domainOrder model.BinanceOrder
	orderResult, err := client.NewTradeService(params).GetOrderHistory(context.Background())
	if err != nil {
		fmt.Println(err)
		return domainOrder, err
	}
	result := orderResult.Result.(map[string]interface{})

	// todo: implement...
	log.Println(result)

	return domainOrder, nil
}
func (b *ByBitClient) CancelOrder(symbol string, orderId int64) (model.BinanceOrder, error) {
	// todo: implement
	return model.BinanceOrder{}, nil
}
func (b *ByBitClient) GetDepth(symbol string) (model.OrderBook, error) {
	// todo: implement
	return model.OrderBook{}, nil
}
func (b *ByBitClient) GetOpenedOrders() ([]model.BinanceOrder, error) {
	// todo: implement
	return make([]model.BinanceOrder, 0), nil
}
func (b *ByBitClient) GetKLines(symbol string, interval string, limit int64) []model.KLineHistory {
	// todo: implement
	return make([]model.KLineHistory, 0)
}
func (b *ByBitClient) GetKLinesCached(symbol string, interval string, limit int64) []model.KLine {
	// todo: implement
	return make([]model.KLine, 0)
}
func (b *ByBitClient) GetExchangeData(symbols []string) (*model.ExchangeInfo, error) {
	// todo: implement
	return nil, nil
}
func (b *ByBitClient) GetAccountStatus() (*model.AccountStatus, error) {
	client := bybit.NewBybitHttpClient(b.ApiKey, b.ApiSecret, bybit.WithBaseURL(b.DestinationURI))
	params := map[string]interface{}{
		"accountType": "SPOT",
	}
	wallet, err := client.NewAccountService(params).GetAccountWallet(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	result := wallet.Result.(map[string]interface{})
	var domainAccount model.AccountStatus

	// todo: implement
	log.Println(result)

	return &domainAccount, nil
}
func (b *ByBitClient) LimitOrder(symbol string, quantity float64, price float64, operation string, timeInForce string) (model.BinanceOrder, error) {
	client := bybit.NewBybitHttpClient(b.ApiKey, b.ApiSecret, bybit.WithBaseURL(b.DestinationURI))
	params := map[string]interface{}{
		"category":    "spot",
		"symbol":      symbol,
		"side":        cases.Title(language.English, cases.Compact).String(operation),
		"positionIdx": 0,
		"orderType":   "Limit",
		"qty":         fmt.Sprintf("%f", quantity),
		"price":       fmt.Sprintf("%f", price),
		"timeInForce": timeInForce,
	}
	var domainOrder model.BinanceOrder
	orderResult, err := client.NewTradeService(params).PlaceOrder(context.Background())
	if err != nil {
		fmt.Println(err)
		return domainOrder, err
	}

	result := orderResult.Result.(map[string]interface{})
	channel := make(chan []byte)
	b.catchEvent(result["orderId"].(string), channel)
	order := <-channel

	// todo: implement...
	log.Println(order)

	return domainOrder, nil
}

func (b *ByBitClient) ListenAll(
	tradeLimits []model.TradeLimit,
	container config.Container,
	swapKlineChannel chan []byte,
	swapUpdateChannel chan string,
	predictChannel chan string,
	depthChannel chan model.Depth,
) {

}
