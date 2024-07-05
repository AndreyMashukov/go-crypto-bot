package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"time"
)

type ObjectRepository struct {
	DB         *sql.DB
	CurrentBot *model.Bot
	RDB        *redis.Client
	Ctx        *context.Context
}

func (o *ObjectRepository) LoadObject(key string, object interface{}) error {
	res := o.RDB.Get(*o.Ctx, key).Val()
	if len(res) > 0 {
		err := json.Unmarshal([]byte(res), object)
		if err == nil {
			return nil
		}
	}

	var jsonString string
	err := o.DB.QueryRow(`
		SELECT 
			os.object as ObjectJSON
		FROM object_storage os WHERE os.storage_key = ?
	`, key).Scan(
		&jsonString,
	)

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(jsonString), object)
	if err != nil {
		return err
	}

	o.RDB.Set(*o.Ctx, key, jsonString, time.Hour*2)

	return nil
}

func (o *ObjectRepository) SaveObject(key string, object interface{}) error {
	jsonString, err := json.Marshal(object)
	if err != nil {
		return err
	}

	_, err = o.DB.Exec(`
		INSERT INTO object_storage SET
		    storage_key = ?,
			object = ?,
			created_at = ?,
			updated_at = ?,
			bot_id = ?
		ON DUPLICATE KEY UPDATE 
			object = ?,
			updated_at = ?
	`,
		key,
		jsonString,
		time.Now(),
		time.Now(),
		o.CurrentBot.Id,
		jsonString,
		time.Now(),
	)

	if err != nil {
		log.Println(err)

		return err
	}

	o.RDB.Set(*o.Ctx, key, string(jsonString), time.Hour*2)

	return nil
}

func (o *ObjectRepository) DeleteObject(key string) error {
	_, err := o.DB.Exec(`
		DELETE FROM object_storage WHERE storage_key = ? AND bot_id = ?
	`, key, o.CurrentBot.Id)

	o.RDB.Del(*o.Ctx, key)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
