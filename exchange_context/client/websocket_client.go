package client

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
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

				_ = connection.Close()
				log.Printf("Binance WS Events, wait and reconnect...")
				time.Sleep(time.Second * 20)
				Listen(address, tradeChannel)
				return
			}

			tradeChannel <- message
		}
	}()

	return connection
}
