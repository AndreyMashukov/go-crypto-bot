package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"os"
	"time"
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

	if len(botUuid) == 0 {
		panic("'BOT_UUID' variable must be set!")
	}

	var bot model.Bot

	err := b.DB.QueryRow(`
		SELECT 
			b.id as Id, 
			b.uuid as Uuid,
			b.is_master_bot as IsMasterBot,
			b.is_swap_enabled as IsSwapEnabled,
			b.swap_config as SwapConfig,
			b.trade_stack_sorting as TradeStackSorting
		FROM bots b
		WHERE b.uuid = ?`, botUuid,
	).Scan(
		&bot.Id,
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

	return &bot
}

func (b *BotRepository) Create(bot model.Bot) error {
	_, err := b.DB.Exec(`INSERT INTO bots SET	uuid = ?`, bot.BotUuid)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (b *BotRepository) Update(bot model.Bot) error {
	_, err := b.DB.Exec(`
		UPDATE bots b SET
			b.is_swap_enabled = ?,
			b.is_master_bot = ?,
			b.swap_config = ?,
			b.trade_stack_sorting = ?
	    WHERE b.uuid = ? AND b.id = ?
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

	// Invalidate cache
	b.RDB.Del(*b.Ctx, b.GetCacheKey(bot.BotUuid))

	return nil
}

func (b *BotRepository) GetCacheKey(botUuid string) string {
	return fmt.Sprintf("bot-cached-%s", botUuid)
}
