package service

import (
	"fmt"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"math"
	"strconv"
	"strings"
)

type Formatter struct {
}

func (m *Formatter) FormatPrice(limit ExchangeModel.TradeLimit, price float64) float64 {
	if price < limit.MinPrice {
		return limit.MinPrice
	}

	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.MinPrice, 'f', -1, 64)), ".")
	precision := len(split[1])
	ratio := math.Pow(10, float64(precision))
	return math.Round(price*ratio) / ratio
}

func (m *Formatter) FormatQuantity(limit ExchangeModel.TradeLimit, quantity float64) float64 {
	if quantity < limit.MinQuantity {
		return limit.MinQuantity
	}

	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.MinQuantity, 'f', -1, 64)), ".")
	precision := len(split[1])
	ratio := math.Pow(10, float64(precision))
	return math.Round(quantity*ratio) / ratio
}
