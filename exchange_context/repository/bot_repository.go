package repository

import (
	"context"
	"database/sql"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
	"os"
)

type BotRepository struct {
	DB  *sql.DB
	RDB *redis.Client
	Ctx *context.Context
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
			b.uuid as Uuid
		FROM bots b
		WHERE b.uuid = ?`, botUuid,
	).Scan(
		&bot.Id,
		&bot.BotUuid,
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
