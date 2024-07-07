package exchange

import (
	"context"
	"encoding/json"
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
	Binance    client.ExchangeAPIInterface
}

func (b *BalanceService) InvalidateBalanceCache(asset string) {
	b.RDB.Del(*b.Ctx, b.getBalanceCacheKey(asset))
	b.RDB.Del(*b.Ctx, b.getAccountCacheKey())
}

func (b *BalanceService) GetBalance(hideZero bool) map[string]model.Balance {
	cached := b.RDB.Get(*b.Ctx, b.getAccountCacheKey()).Val()

	balanceMap := make(map[string]model.Balance)

	if len(cached) > 0 {
		var account model.AccountStatus
		err := json.Unmarshal([]byte(cached), &account)

		if err == nil {
			for _, balance := range account.Balances {
				if hideZero && (balance.Locked+balance.Free) == 0.00 {
					continue
				}

				balanceMap[balance.Asset] = balance
			}

			return balanceMap
		}
	}

	if account, err := b.Binance.GetAccountStatus(); err == nil {
		if encoded, err := json.Marshal(account); err == nil {
			b.RDB.Set(*b.Ctx, b.getAccountCacheKey(), encoded, time.Minute)
			for _, balance := range account.Balances {
				if hideZero && (balance.Locked+balance.Free) == 0.00 {
					continue
				}

				balanceMap[balance.Asset] = balance
			}

			return balanceMap
		}
	}

	return balanceMap
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

func (b *BalanceService) getAccountCacheKey() string {
	return fmt.Sprintf("account-status-%d", b.CurrentBot.Id)
}
