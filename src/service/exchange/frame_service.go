package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"time"
)

type FrameServiceInterface interface {
	GetFrame(symbol string, interval string, limit int64) model.Frame
}

type FrameService struct {
	CurrentBot *model.Bot
	Binance    *client.Binance
	RDB        *redis.Client
	Ctx        *context.Context
}

func (f *FrameService) GetFrame(symbol string, interval string, limit int64) model.Frame {
	key := fmt.Sprintf("kline-frame-results-%s-%s-%d-%d", symbol, interval, limit, f.CurrentBot.Id)
	cached := f.RDB.Get(*f.Ctx, key).String()

	if len(cached) > 0 {
		var frameResult model.Frame
		err := json.Unmarshal([]byte(cached), &frameResult)

		if err == nil {
			return frameResult
		}
	}

	kLines := f.Binance.GetKLinesCached(symbol, interval, limit)

	highSum := 0.00
	lowSum := 0.00
	amountHigh := 0.00
	amountLow := 0.00
	highestPrice := 0.00
	lowestPrice := 0.00

	for _, kLine := range kLines {
		if lowestPrice == 0.00 || lowestPrice > kLine.Low {
			lowestPrice = kLine.Low
		}

		if highestPrice < kLine.High {
			highestPrice = kLine.High
		}

		if kLine.IsPositive() {
			highSum += kLine.High
			amountHigh++
		}

		if kLine.IsNegative() {
			lowSum += kLine.Low
			amountLow++
		}
	}

	avgLow := 0.00

	if amountLow > 0 {
		avgLow = lowSum / amountLow
	}

	avgHigh := 0.00

	if amountHigh == 0.00 {
		avgHigh = avgLow
	} else {
		avgHigh = highSum / amountHigh
	}

	frame := model.Frame{
		High:    highestPrice,
		Low:     lowestPrice,
		AvgHigh: avgHigh,
		AvgLow:  avgLow,
	}

	result, _ := json.Marshal(frame)
	f.RDB.Set(*f.Ctx, key, string(result), time.Second*15)

	return frame
}
