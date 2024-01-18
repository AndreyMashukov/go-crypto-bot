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

type CallbackManagerInterface interface {
	Error(bot model.Bot, code string, message string, stop bool)
	SellOrder(order model.Order, bot model.Bot, details string)
	BuyOrder(order model.Order, bot model.Bot, details string)
}

type CallbackManager struct {
	AutoTradeHost string
}

func (t *CallbackManager) Error(bot model.Bot, code string, message string, stop bool) {
	encoded, _ := json.Marshal(model.ErrorNotification{
		BotId:        bot.Id,
		Stop:         stop,
		ErrorCode:    code,
		ErrorMessage: message,
	})
	err := t.Send("/callback/error", encoded)
	if err == nil {
		log.Printf("[%s] Error notification sent", message)
	} else {
		log.Printf("[%s] Error notification failed: %s", message, err.Error())
	}
}

func (t *CallbackManager) SellOrder(order model.Order, bot model.Bot, details string) {
	encoded, _ := json.Marshal(model.TgOrderNotification{
		BotId:     bot.Id,
		Price:     order.Price,
		Quantity:  order.Quantity,
		Symbol:    order.Symbol,
		Operation: strings.ToUpper(order.Operation),
		DateTime:  order.CreatedAt,
		Details:   details,
	})
	err := t.Send("/callback/telegram", encoded)
	if err == nil {
		log.Printf("[%s] Telegram SELL notification sent", order.Symbol)
	} else {
		log.Printf("[%s] Telegram notification failed: %s", order.Symbol, err.Error())
	}
}

func (t *CallbackManager) BuyOrder(order model.Order, bot model.Bot, details string) {
	encoded, _ := json.Marshal(model.TgOrderNotification{
		BotId:     bot.Id,
		Price:     order.Price,
		Quantity:  order.Quantity,
		Symbol:    order.Symbol,
		Operation: strings.ToUpper(order.Operation),
		DateTime:  order.CreatedAt,
		Details:   details,
	})
	err := t.Send("/callback/telegram", encoded)
	if err == nil {
		log.Printf("[%s] Telegram BUY notification sent", order.Symbol)
	} else {
		log.Printf("[%s] Telegram notification failed: %s", order.Symbol, err.Error())
	}
}

func (t *CallbackManager) Send(path string, message []byte) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/public%s", t.AutoTradeHost, path), bytes.NewReader(message))
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
