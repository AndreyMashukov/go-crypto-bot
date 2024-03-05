package exchange

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"strconv"
	"time"
)

type BalanceServiceInterface interface {
	GetAssetBalance(asset string, cache bool) (float64, error)
	InvalidateBalanceCache(asset string)
}

type BalanceService struct {
	RDB        *redis.Client
	Ctx        *context.Context
	CurrentBot *model.Bot
	Binance    *client.Binance
}

func (b *BalanceService) InvalidateBalanceCache(asset string) {
	b.RDB.Del(*b.Ctx, b.getBalanceCacheKey(asset))
}

func (b *BalanceService) GetAssetBalance(asset string, cache bool) (float64, error) {
	cached := b.RDB.Get(*b.Ctx, b.getBalanceCacheKey(asset)).Val()

	if len(cached) > 0 && cache {
		balanceCached, err := strconv.ParseFloat(cached, 64)

		if err == nil {
			return balanceCached, nil
		}
	}

	accountInfo, err := b.Binance.GetAccountStatus()

	if err != nil {
		return 0.00, err
	}

	for _, assetBalance := range accountInfo.Balances {
		if assetBalance.Asset == asset {
			log.Printf("[%s] Free balance is: %f", asset, assetBalance.Free)
			log.Printf("[%s] Locked balance is: %f", asset, assetBalance.Locked)

			b.RDB.Set(*b.Ctx, b.getBalanceCacheKey(asset), assetBalance.Free, time.Minute)
			return assetBalance.Free, nil
		}
	}

	return 0.00, nil
}

func (b *BalanceService) getBalanceCacheKey(asset string) string {
	return fmt.Sprintf("balance-%s-account-%d", asset, b.CurrentBot.Id)
}
