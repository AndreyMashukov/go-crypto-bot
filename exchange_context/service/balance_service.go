package service

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	ExchangeClient "gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
	"strconv"
	"time"
)

type BalanceService struct {
	RDB        *redis.Client
	Ctx        *context.Context
	CurrentBot *ExchangeModel.Bot
	Binance    *ExchangeClient.Binance
}

func (b *BalanceService) InvalidateBalanceCache(asset string) {
	b.RDB.Del(*b.Ctx, b.getBalanceCacheKey(asset))
}

func (b *BalanceService) GetAssetBalance(asset string) (float64, error) {
	cached := b.RDB.Get(*b.Ctx, b.getBalanceCacheKey(asset)).Val()

	if len(cached) > 0 {
		balanceCached, err := strconv.ParseFloat(cached, 64)

		if err == nil {
			log.Printf("[%s] Free balance is: %f (cached)", asset, balanceCached)
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

			b.RDB.Set(*b.Ctx, b.getBalanceCacheKey(asset), assetBalance.Free, time.Minute*5)
			return assetBalance.Free, nil
		}
	}

	return 0.00, nil
}

func (b *BalanceService) getBalanceCacheKey(asset string) string {
	return fmt.Sprintf("balance-%s-account-%d", asset, b.CurrentBot.Id)
}
