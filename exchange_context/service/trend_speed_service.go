package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"slices"
	"time"
)

type TrendSpeedService struct {
	RDB        *redis.Client
	Ctx        *context.Context
	CurrentBot *model.Bot
}

func (t *TrendSpeedService) ProcessKline(kLine model.KLine) {
	klineSpeed := model.KlineSpeed{
		Time:  time.Now(),
		KLine: kLine,
	}

	encoded, _ := json.Marshal(klineSpeed)
	t.RDB.LPush(*t.Ctx, fmt.Sprintf("k-line-speed-%s-%d", kLine.Symbol, t.CurrentBot.Id), string(encoded))
	t.RDB.LTrim(*t.Ctx, fmt.Sprintf("k-line-speed-%s-%d", kLine.Symbol, t.CurrentBot.Id), 0, 5)
}

func (t *TrendSpeedService) GetPriceSpeedPoints(limit model.TradeLimit) float64 {
	res := t.RDB.LRange(*t.Ctx, fmt.Sprintf("k-line-speed-%s-%d", limit.Symbol, t.CurrentBot.Id), 0, 5).Val()
	list := make([]model.KlineSpeed, 0)

	for _, str := range res {
		var dto model.KlineSpeed
		_ = json.Unmarshal([]byte(str), &dto)
		list = append(list, dto)
	}

	slices.Reverse(list)

	var previousItem *model.KlineSpeed

	speedList := make([]float64, 0)

	for _, item := range list {
		if previousItem == nil {
			previousItem = &item
			continue
		}

		secondsDiff := float64(previousItem.Time.Unix() - item.Time.Unix())

		if secondsDiff <= 0 {
			continue
		}

		priceDiff := item.KLine.Close - previousItem.KLine.Close
		priceStepDiff := priceDiff / limit.MinPrice

		speedList = append(speedList, priceStepDiff/secondsDiff)
	}

	averageSpeed := 0.00

	if len(speedList) > 0 {
		speedSum := 0.00

		for _, speed := range speedList {
			speedSum += speed
		}

		averageSpeed = speedSum / float64(len(speedList))
	}

	return averageSpeed
}
