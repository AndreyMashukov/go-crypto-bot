package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	Model "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Binance struct {
	ApiKey         string
	ApiSecret      string
	DestinationURI string

	HttpClient *http.Client
}

func (b *Binance) GetDepth(symbol string) (*Model.MarketDepth, error) {
	url := fmt.Sprintf("%s/api/v3/depth?symbol=%s", b.DestinationURI, symbol)
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	response, err := b.HttpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(string(body))
	}

	var depth Model.MarketDepth
	json.Unmarshal(body, &depth)

	return &depth, nil
}

func (b *Binance) QueryOrder(symbol string, orderId int64) (Model.BinanceOrder, error) {
	queryString := fmt.Sprintf(
		"symbol=%s&orderId=%d&timestamp=%d",
		symbol,
		orderId,
		time.Now().UTC().Unix()*1000,
	)
	request, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/order?%s&signature=%s", b.DestinationURI, queryString, b.sign(queryString)), nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("X-MBX-APIKEY", b.ApiKey)

	response, err := b.HttpClient.Do(request)
	var binanceOrder Model.BinanceOrder
	if err != nil {
		return binanceOrder, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return binanceOrder, err
	}

	if response.StatusCode != 200 {
		return binanceOrder, errors.New(string(body))
	}

	json.Unmarshal(body, &binanceOrder)

	return binanceOrder, nil
}

func (b *Binance) CancelOrder(symbol string, orderId int64) (Model.BinanceOrder, error) {
	queryString := fmt.Sprintf(
		"symbol=%s&orderId=%d&timestamp=%d",
		symbol,
		orderId,
		time.Now().UTC().Unix()*1000,
	)
	request, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v3/order?%s&signature=%s", b.DestinationURI, queryString, b.sign(queryString)), nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("X-MBX-APIKEY", b.ApiKey)

	response, err := b.HttpClient.Do(request)
	var binanceOrder Model.BinanceOrder
	if err != nil {
		return binanceOrder, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return binanceOrder, err
	}

	if response.StatusCode != 200 {
		return binanceOrder, errors.New(string(body))
	}

	json.Unmarshal(body, &binanceOrder)

	return binanceOrder, nil
}

func (b *Binance) GetOpenedOrders() (*[]Model.BinanceOrder, error) {
	queryString := fmt.Sprintf(
		"timestamp=%d",
		time.Now().UTC().Unix()*1000,
	)
	request, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/openOrders?%s&signature=%s", b.DestinationURI, queryString, b.sign(queryString)), nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("X-MBX-APIKEY", b.ApiKey)

	response, err := b.HttpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(string(body))
	}

	var binanceOrders []Model.BinanceOrder
	json.Unmarshal(body, &binanceOrders)

	return &binanceOrders, nil
}

func (b *Binance) LimitOrder(order Model.Order, operation string) (Model.BinanceOrder, error) {
	queryString := fmt.Sprintf(
		"symbol=%s&side=%s&type=LIMIT&timeInForce=GTC&quantity=%s&price=%s&timestamp=%d",
		order.Symbol,
		operation,
		strconv.FormatFloat(order.Quantity, 'f', -1, 64),
		strconv.FormatFloat(order.Price, 'f', -1, 64),
		time.Now().UTC().Unix()*1000,
	)
	request, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v3/order?%s&signature=%s", b.DestinationURI, queryString, b.sign(queryString)), nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("X-MBX-APIKEY", b.ApiKey)

	response, err := b.HttpClient.Do(request)
	var binanceOrder Model.BinanceOrder
	if err != nil {
		return binanceOrder, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if strings.Contains(string(body), "Account has insufficient balance for requested action") {
		return binanceOrder, errors.New(fmt.Sprintf("[%s] Account has insufficient balance for requested action", order.Symbol))
	}

	if err != nil {
		return binanceOrder, err
	}

	if response.StatusCode != 200 {
		return binanceOrder, errors.New(string(body))
	}

	json.Unmarshal(body, &binanceOrder)

	return binanceOrder, nil
}

func (b *Binance) sign(url string) string {
	mac := hmac.New(sha256.New, []byte(b.ApiSecret))
	mac.Write([]byte(url))
	signingKey := fmt.Sprintf("%x", mac.Sum(nil))

	return signingKey
}
