package client

import (
	"github.com/gorilla/websocket"
	"log"
	"os"
)

func Listen(address string, tradeChannel chan<- []byte) *websocket.Conn {
	connection, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Printf("Binance WS Events [%s]: %s", address, err.Error())
		log.Fatal("Quit!")
	}

	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Printf("Binance WS Events, read [%s]: %s", address, err.Error())

				os.Exit(1)
			}

			tradeChannel <- message
		}
	}()

	return connection
}
