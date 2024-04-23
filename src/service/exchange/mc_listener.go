package exchange

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"log"
	"strings"
	"time"
)

type MCListener struct {
	MSGatewayAddress   string
	ExchangeRepository *repository.ExchangeRepository
}

func (m *MCListener) ListenAll() {
	if len(m.MSGatewayAddress) == 0 {
		log.Printf("MC address is not defined.")
		return
	}

	mcChannel := make(chan []byte)

	// existing swaps real time monitoring
	go func() {
		for {
			msg := <-mcChannel

			if strings.Contains(string(msg), "@crypto_price_15s@") {
				var mcEvent model.MCEvent
				err := json.Unmarshal(msg, &mcEvent)
				if err == nil {
					m.ExchangeRepository.SetCapitalization(mcEvent)
				} else {
					log.Printf("Error during parsing crypto price event: %s", err.Error())
				}
			}
		}
	}()

	m.Listen(m.MSGatewayAddress, mcChannel)

	runChannel := make(chan string)
	// just to keep running
	runChannel <- "run"
	log.Panic("Swap Listener Stopped")
}

func (m *MCListener) Listen(address string, tradeChannel chan<- []byte) *websocket.Conn {
	connection, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Printf("MC [err_1] WS Events [%s]: %s, wait and reconnect...", address, err.Error())
		time.Sleep(time.Second * 3)

		return m.Listen(address, tradeChannel)
	}

	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Printf("MC [err_2] WS Events, read [%s]: %s", address, err.Error())

				_ = connection.Close()
				log.Printf("MC [err_2] WS Events, wait and reconnect...")
				time.Sleep(time.Second * 3)
				m.Listen(address, tradeChannel)
				return
			}

			tradeChannel <- message
		}
	}()

	serialized := `
		{
			"method":"RSUBSCRIPTION",
			"params":[
				"main-site@crypto_price_15s@{}@detail",
				"1,1027,1839,5426,52,74,2010,5994,5805,6636,1831,1958,1975,6535,3890,2,7083,1321,3794,512,1376,1437"
			]
		}
	`
	_ = connection.WriteMessage(websocket.TextMessage, []byte(serialized))

	return connection
}
