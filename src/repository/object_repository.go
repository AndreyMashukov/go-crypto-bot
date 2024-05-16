package repository

import (
	"database/sql"
	"encoding/json"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"time"
)

type ObjectRepository struct {
	DB         *sql.DB
	CurrentBot *model.Bot
}

func (o *ObjectRepository) LoadObject(key string, object interface{}) error {
	var jsonString string
	err := o.DB.QueryRow(`
		SELECT 
			os.object as ObjectJSON
		FROM object_storage os WHERE os.storage_key = ? AND os.bot_id = ?
	`, key, o.CurrentBot.Id).Scan(
		&jsonString,
	)

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(jsonString), object)
	if err != nil {
		return err
	}

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

	return nil
}

func (o *ObjectRepository) DeleteObject(key string) error {
	_, err := o.DB.Exec(`
		DELETE FROM object_storage WHERE storage_key = ? AND bot_id = ?
	`, key, o.CurrentBot.Id)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
