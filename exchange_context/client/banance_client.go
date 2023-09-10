package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	Model "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"io/ioutil"
	"net/http"
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

	var depth Model.MarketDepth
	json.Unmarshal(body, &depth)

	return &depth, nil
}

func (b *Binance) QueryOrder(symbol string, orderId int64) (*Model.BinanceOrder, error) {
	queryString := fmt.Sprintf(
		"symbol=%s&orderId=%d&timestamp=%d",
		symbol,
		orderId,
		time.Now().UTC().Unix()*1000,
	)
	request, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/order?%s&signature=%s", b.DestinationURI, queryString, b._Sign(queryString)), nil)
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

	var binanceOrder Model.BinanceOrder
	json.Unmarshal(body, &binanceOrder)

	return &binanceOrder, nil
}

func (b *Binance) CancelOrder(symbol string, orderId int64) (*Model.BinanceOrder, error) {
	queryString := fmt.Sprintf(
		"symbol=%s&orderId=%d&timestamp=%d",
		symbol,
		orderId,
		time.Now().UTC().Unix()*1000,
	)
	request, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v3/order?%s&signature=%s", b.DestinationURI, queryString, b._Sign(queryString)), nil)
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

	var binanceOrder Model.BinanceOrder
	json.Unmarshal(body, &binanceOrder)

	return &binanceOrder, nil
}

func (b *Binance) GetOpenedOrders() (*[]Model.BinanceOrder, error) {
	queryString := fmt.Sprintf(
		"timestamp=%d",
		time.Now().UTC().Unix()*1000,
	)
	request, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/openOrders?%s&signature=%s", b.DestinationURI, queryString, b._Sign(queryString)), nil)
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

	var binanceOrders []Model.BinanceOrder
	json.Unmarshal(body, &binanceOrders)

	return &binanceOrders, nil
}

func (b *Binance) LimitOrder(order Model.Order, operation string) (*Model.BinanceOrder, error) {
	queryString := fmt.Sprintf(
		"symbol=%s&side=%s&type=LIMIT&timeInForce=GTC&quantity=%f&price=%f&timestamp=%d",
		order.Symbol,
		operation,
		order.Quantity,
		order.Price,
		time.Now().UTC().Unix()*1000,
	)
	request, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v3/order?%s&signature=%s", b.DestinationURI, queryString, b._Sign(queryString)), nil)
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

	var binanceOrder Model.BinanceOrder
	json.Unmarshal(body, &binanceOrder)

	return &binanceOrder, nil
}

func (b *Binance) _Sign(url string) string {
	mac := hmac.New(sha256.New, []byte(b.ApiSecret))
	mac.Write([]byte(url))
	signingKey := fmt.Sprintf("%x", mac.Sum(nil))

	return signingKey
}
