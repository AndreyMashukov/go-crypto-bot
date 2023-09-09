package client

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	exchange_context "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
	"os"
)

func Listen(address string, callback func(candle exchange_context.Candle)) *websocket.Conn {
	connection, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Println("read: ", err)

				os.Exit(1)
			}

			var decodedModel exchange_context.Event
			json.Unmarshal(message, &decodedModel)
			callback(decodedModel.Candle)
		}
	}()

	return connection
}
