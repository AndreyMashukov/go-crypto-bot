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

func (m *Formatter) FormatPrice(limit ExchangeModel.TradeLimitInterface, price float64) float64 {
	if price < limit.GetMinPrice() {
		return limit.GetMinPrice()
	}

	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.GetMinPrice(), 'f', -1, 64)), ".")
	precision := 0
	if len(split) > 1 {
		precision = len(split[1])
	}
	ratio := math.Pow(10, float64(precision))
	return math.Round(price*ratio) / ratio
}

func (m *Formatter) FormatQuantity(limit ExchangeModel.TradeLimitInterface, quantity float64) float64 {
	if quantity < limit.GetMinQuantity() {
		return limit.GetMinQuantity()
	}

	splitQty := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(quantity, 'f', -1, 64)), ".")
	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.GetMinQuantity(), 'f', -1, 64)), ".")
	precision := 0
	if len(split) > 1 {
		precision = len(split[1])
	}

	second := "00"
	if precision > 0 && len(splitQty) > 1 {
		substr := precision
		if len(splitQty[1]) < substr {
			substr = len(splitQty[1])
		}

		second = splitQty[1][0:substr]
	}
	quantity, _ = strconv.ParseFloat(fmt.Sprintf("%s.%s", splitQty[0], second), 64)

	return quantity
}

func (m *Formatter) ComparePercentage(first float64, second float64) ExchangeModel.Percent {
	return ExchangeModel.Percent(second * 100 / first)
}

func (m *Formatter) Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func (m *Formatter) ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(m.Round(num*output)) / output
}
