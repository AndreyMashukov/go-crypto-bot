package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"io"
	"log"
	"net/http"
	"strings"
)

type TelegramNotificatorInterface interface {
	SellOrder(order model.Order, bot model.Bot, details string)
	BuyOrder(order model.Order, bot model.Bot, details string)
}

type TelegramNotificator struct {
	AutoTradeHost string
}

func (t *TelegramNotificator) SellOrder(order model.Order, bot model.Bot, details string) {
	encoded, _ := json.Marshal(model.TgOrderNotification{
		BotId:     bot.Id,
		Price:     order.Price,
		Quantity:  order.Quantity,
		Symbol:    order.Symbol,
		Operation: strings.ToUpper(order.Operation),
		DateTime:  order.CreatedAt,
		Details:   details,
	})
	err := t.SendMessage(encoded)
	if err == nil {
		log.Printf("[%s] Telegram SELL notification sent", order.Symbol)
	}
}

func (t *TelegramNotificator) BuyOrder(order model.Order, bot model.Bot, details string) {
	encoded, _ := json.Marshal(model.TgOrderNotification{
		BotId:     bot.Id,
		Price:     order.Price,
		Quantity:  order.Quantity,
		Symbol:    order.Symbol,
		Operation: strings.ToUpper(order.Operation),
		DateTime:  order.CreatedAt,
		Details:   details,
	})
	err := t.SendMessage(encoded)
	if err == nil {
		log.Printf("[%s] Telegram BUY notification sent", order.Symbol)
	}
}

func (t *TelegramNotificator) SendMessage(message []byte) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/public/callback/telegram", t.AutoTradeHost), bytes.NewReader(message))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	res, err := client.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("Request failed with error code: %d", res.StatusCode))
	}

	_, err = io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return err
	}

	return nil
}
