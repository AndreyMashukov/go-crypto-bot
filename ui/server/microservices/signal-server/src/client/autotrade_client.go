package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/model"
	"io"
	"log"
	"net/http"
)

type AutoTradeClient struct {
	InternalAPIToken string
}

func (a *AutoTradeClient) SendSignal(signal model.Signal) {
	if len(a.InternalAPIToken) == 0 {
		return
	}

	encoded, _ := json.Marshal(signal)
	err := a.Post("/public/internal/signal", encoded)
	if err != nil {
		log.Printf("Signal send error, code: %s", err.Error())
	}
}

func (a *AutoTradeClient) Post(path string, message []byte) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost%s", path), bytes.NewReader(message))

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Crypto-Internal-Token", a.InternalAPIToken)

	client := &http.Client{}

	res, err := client.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		body, bodyErr := io.ReadAll(res.Body)
		defer func() {
			_ = res.Body.Close()
		}()

		if bodyErr == nil {
			return errors.New(fmt.Sprintf("Request failed with error code: %d -> %s", res.StatusCode, string(body)))
		}

		return errors.New(fmt.Sprintf("Request failed with error code: %d", res.StatusCode))
	}

	_, err = io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return err
	}

	return nil
}
