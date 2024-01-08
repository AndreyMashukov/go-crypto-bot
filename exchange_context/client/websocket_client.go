package client

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
	"time"
)

func Listen(address string, tradeChannel chan<- []byte, streams []string, connectionId int64) *websocket.Conn {
	connection, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Printf("Binance [err_1] WS Events [%s]: %s, wait and reconnect...", address, err.Error())
		time.Sleep(time.Second * 3)
		connectionId++

		return Listen(address, tradeChannel, streams, connectionId)
	}

	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Printf("Binance [err_2] WS Events, read [%s]: %s", address, err.Error())

				_ = connection.Close()
				log.Printf("Binance [err_2] WS Events, wait and reconnect...")
				time.Sleep(time.Second * 3)
				connectionId++
				Listen(address, tradeChannel, streams, connectionId)
				return
			}

			tradeChannel <- message
		}
	}()

	if len(streams) > 0 {
		socketRequest := model.SocketStreamsRequest{
			Id:     connectionId,
			Method: "SUBSCRIBE",
			Params: streams,
		}
		serialized, _ := json.Marshal(socketRequest)
		_ = connection.WriteMessage(websocket.TextMessage, serialized)
	}

	return connection
}
