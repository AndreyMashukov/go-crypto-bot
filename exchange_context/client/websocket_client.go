package client

import (
	"github.com/gorilla/websocket"
	"log"
	"os"
)

func Listen(address string, tradeChannel chan<- []byte) *websocket.Conn {
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

			tradeChannel <- message
		}
	}()

	return connection
}
