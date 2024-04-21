package tests

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"testing"
)

func TestOrderExtraBudget(t *testing.T) {
	assertion := assert.New(t)

	kline := model.KLine{
		Close: 85.00,
	}

	order := model.Order{
		UsedExtraBudget:  49.95,
		ExecutedQuantity: 1.00,
		Price:            100.00,
		ExtraChargeOptions: model.ExtraChargeOptions{
			{
				Index:      0,
				Percent:    model.Percent(-10.00),
				AmountUsdt: 50.00,
			},
			{
				Index:      1,
				Percent:    model.Percent(-15.00),
				AmountUsdt: 50.00,
			},
			{
				Index:      2,
				Percent:    model.Percent(-20.00),
				AmountUsdt: 50.00,
			},
		},
	}

	extraBudget := order.GetAvailableExtraBudget(kline, false)
	assertion.Equal(50.05, extraBudget)

	order.UsedExtraBudget = 99.95

	extraBudget = order.GetAvailableExtraBudget(kline, false)
	assertion.Equal(0.04999999999999716, extraBudget)
}
