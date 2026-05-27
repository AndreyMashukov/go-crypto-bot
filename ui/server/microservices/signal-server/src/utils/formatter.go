package utils

import (
	"fmt"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/model"
	"math"
	"strconv"
	"strings"
)

type Formatter struct {
}

func (m *Formatter) DetectPrecision(value float64) int {
	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(value, 'f', -1, 64)), ".")
	precision := 0
	if len(split) > 1 {
		precision = len(split[1])
	}

	return precision
}

func (m *Formatter) ComparePercentage(first float64, second float64) model.Percent {
	return model.Percent(second * 100.00 / first)
}

func (m *Formatter) Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func (m *Formatter) ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(m.Round(num*output)) / output
}

func (m *Formatter) Floor(num float64) int64 {
	return int64(math.Floor(num))
}
