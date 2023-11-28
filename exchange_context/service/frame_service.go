package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"time"
)

type FrameService struct {
	Binance *client.Binance
	RDB     *redis.Client
	Ctx     *context.Context
}

func (f *FrameService) GetFrame(symbol string, interval string, limit int64) model.Frame {
	key := fmt.Sprintf("kline-frame-results-%s-%s-%d", symbol, interval, limit)
	cached := f.RDB.Get(*f.Ctx, key).String()

	if len(cached) > 0 {
		var frameResult model.Frame
		err := json.Unmarshal([]byte(cached), &frameResult)

		if err == nil {
			return frameResult
		}
	}

	kLines := f.Binance.GetKLines(symbol, interval, limit)

	highSum := 0.00
	lowSum := 0.00
	amount := 0.00
	highestPrice := 0.00
	lowestPrice := 0.00

	for _, kLine := range kLines {
		highSum += kLine.GetHighPrice()
		lowSum += kLine.GetLowPrice()

		if lowestPrice == 0.00 || lowestPrice > kLine.GetLowPrice() {
			lowestPrice = kLine.GetLowPrice()
		}

		if highestPrice < kLine.GetHighPrice() {
			highestPrice = kLine.GetHighPrice()
		}

		amount++
	}

	frame := model.Frame{
		High:    highestPrice,
		Low:     lowestPrice,
		AvgHigh: highSum / amount,
		AvgLow:  lowSum / amount,
	}

	result, _ := json.Marshal(frame)
	f.RDB.Set(*f.Ctx, key, string(result), time.Second*15)

	return frame
}
