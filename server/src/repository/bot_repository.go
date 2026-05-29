package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
)

type BotRepository struct {
	DB  *sql.DB
	RDB *redis.Client
	Ctx *context.Context
}

func (b *BotRepository) GetCurrentBotCached(botId int64) model.Bot {
	botUuid := os.Getenv("BOT_UUID")

	if len(botUuid) == 0 {
		panic("'BOT_UUID' variable must be set!")
	}

	cacheKey := b.GetCacheKey(botUuid)
	cachedBot := b.RDB.Get(*b.Ctx, cacheKey).Val()

	if len(cachedBot) > 0 {
		var bot model.Bot
		err := json.Unmarshal([]byte(cachedBot), &bot)
		if err == nil {
			if bot.Id != botId {
				panic(fmt.Sprintf("Bot ID is different! %d != %d", bot.Id, botId))
			}

			return bot
		}
	}

	bot := b.GetCurrentBot()

	if bot == nil {
		panic("Current bot is not found!")
	}

	if bot.Id != botId {
		panic(fmt.Sprintf("Bot ID is different! %d != %d", bot.Id, botId))
	}

	botEncoded, err := json.Marshal(bot)
	if err == nil {
		b.RDB.Set(*b.Ctx, cacheKey, string(botEncoded), time.Minute)
	}

	return *bot
}

func (b *BotRepository) GetCurrentBot() *model.Bot {
	botUuid := os.Getenv("BOT_UUID")
	botExchange := os.Getenv("BOT_EXCHANGE")

	if len(botUuid) == 0 {
		panic("'BOT_UUID' variable must be set!")
	}

	var bot model.Bot

	err := b.DB.QueryRow(`
		SELECT
			b.id                  as Id,
			b.exchange            as Exchange,
			b.uuid                as Uuid,
			b.is_master_bot       as IsMasterBot,
			b.is_swap_enabled     as IsSwapEnabled,
			b.swap_config         as SwapConfig,
			b.trade_stack_sorting as TradeStackSorting
		FROM bots b
		WHERE b.uuid = $1 AND b.exchange = $2`, botUuid, botExchange,
	).Scan(
		&bot.Id,
		&bot.Exchange,
		&bot.BotUuid,
		&bot.IsMasterBot,
		&bot.IsSwapEnabled,
		&bot.SwapConfig,
		&bot.TradeStackSorting,
	)

	if err != nil {
		log.Println(err)
		return nil
	}

	cacheKey := b.GetCacheKey(botUuid)
	botEncoded, err := json.Marshal(bot)
	if err == nil {
		b.RDB.Set(*b.Ctx, cacheKey, string(botEncoded), time.Minute)
	}

	return &bot
}

func (b *BotRepository) Create(bot model.Bot) error {
	_, err := b.DB.Exec(`
		INSERT INTO bots (
			uuid, exchange, is_swap_enabled, is_master_bot, swap_config, trade_stack_sorting
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`,
		bot.BotUuid,
		bot.Exchange,
		bot.IsSwapEnabled,
		bot.IsMasterBot,
		bot.SwapConfig,
		bot.TradeStackSorting,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (b *BotRepository) Update(bot model.Bot) error {
	_, err := b.DB.Exec(`
		UPDATE bots SET
			is_swap_enabled     = $1,
			is_master_bot       = $2,
			swap_config         = $3,
			trade_stack_sorting = $4
		WHERE uuid = $5 AND id = $6
	`,
		bot.IsSwapEnabled,
		bot.IsMasterBot,
		bot.SwapConfig,
		bot.TradeStackSorting,
		bot.BotUuid,
		bot.Id,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	b.RDB.Del(*b.Ctx, b.GetCacheKey(bot.BotUuid))

	return nil
}

func (b *BotRepository) GetCacheKey(botUuid string) string {
	return fmt.Sprintf("bot-cached-%s", botUuid)
}
