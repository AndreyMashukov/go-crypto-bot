package tests

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"os"
	"testing"
	"time"
)

func TestSignalTTLCalculation(t *testing.T) {
	assertion := assert.New(t)

	signal := model.Signal{
		ExpireTimestamp: time.Now().Add(time.Minute).UnixMilli(),
	}
	assertion.Equal(time.Duration(60000), signal.GetTTLMilli())
	assertion.False(signal.IsExpired())
}

func TestSignalProfitAndExtraChargeOptions(t *testing.T) {
	assertion := assert.New(t)

	content, _ := os.ReadFile("example/signal.json")
	var signal model.Signal
	err := json.Unmarshal(content, &signal)
	assertion.Nil(err)

	assertion.Equal("PERPUSDT", signal.Symbol)
	assertion.Equal(int64(1714754067823), signal.ExpireTimestamp)
	assertion.Equal(2, len(signal.ProfitOptions))
	assertion.Equal(1, len(signal.ExtraChargeOptions))
	profitOptions := signal.GetProfitOptions()
	assertion.Equal(2, len(profitOptions))
	assertion.True(profitOptions[0].IsTriggerOption)
	assertion.False(profitOptions[1].IsTriggerOption)
	assertion.Equal(12.00, profitOptions[0].OptionValue)
	assertion.Equal(model.Percent(6.35), profitOptions[0].OptionPercent)
	assertion.Equal("h", profitOptions[0].OptionUnit)

	limit := model.TradeLimit{
		USDTLimit: 1000.00,
	}
	extraChargeOptions := signal.GetProfitExtraChargeOptions(limit)
	assertion.Equal(1, len(extraChargeOptions))
	assertion.Equal(int64(0), extraChargeOptions[0].Index)
	assertion.Equal(model.Percent(-9.41), extraChargeOptions[0].Percent)
	assertion.Equal(800.00, extraChargeOptions[0].AmountUsdt)
}
