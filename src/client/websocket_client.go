package client

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"strings"
	"time"
)

func GetStreamBatch(tradeLimits []model.SymbolInterface, events []string) [][]string {
	streamBatch := make([][]string, 0)

	streams := make([]string, 0)

	for _, tradeLimit := range tradeLimits {
		for i := 0; i < len(events); i++ {
			event := events[i]
			streams = append(streams, fmt.Sprintf("%s%s", strings.ToLower(tradeLimit.GetSymbol()), event))
		}

		if len(streams) >= 24 {
			streamBatch = append(streamBatch, streams)
			streams = make([]string, 0)
		}
	}

	if len(streams) > 0 {
		streamBatch = append(streamBatch, streams)
	}

	return streamBatch
}

func GetStreamBatchByBit(tradeLimits []model.SymbolInterface, events []string) [][]string {
	streamBatch := make([][]string, 0)

	streams := make([]string, 0)

	for _, tradeLimit := range tradeLimits {
		for i := 0; i < len(events); i++ {
			event := events[i]
			streams = append(streams, fmt.Sprintf("%s%s", event, strings.ToUpper(tradeLimit.GetSymbol())))
		}

		if len(streams) >= 8 {
			streamBatch = append(streamBatch, streams)
			streams = make([]string, 0)
		}
	}

	if len(streams) > 0 {
		streamBatch = append(streamBatch, streams)
	}

	return streamBatch
}

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

func ListenByBit(address string, tradeChannel chan<- []byte, streams []string, connectionId int64) *websocket.Conn {
	connection, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Printf("ByBit [err_1] WS Events [%s]: %s, wait and reconnect...", address, err.Error())
		time.Sleep(time.Second * 3)
		connectionId++

		return ListenByBit(address, tradeChannel, streams, connectionId)
	}

	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Printf("ByBit [err_2] WS Events, read [%s]: %s", address, err.Error())

				_ = connection.Close()
				log.Printf("ByBit [err_2] WS Events, wait and reconnect...")
				time.Sleep(time.Second * 3)
				connectionId++
				ListenByBit(address, tradeChannel, streams, connectionId)
				return
			}

			tradeChannel <- message
		}
	}()

	if len(streams) > 0 {
		socketRequest := model.ByBitSocketStreamsRequest{
			Operation: "subscribe",
			Arguments: streams,
		}
		serialized, _ := json.Marshal(socketRequest)
		_ = connection.WriteMessage(websocket.TextMessage, serialized)
	}

	return connection
}
