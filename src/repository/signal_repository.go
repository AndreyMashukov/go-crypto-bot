package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"strings"
	"time"
)

type SignalStorageInterface interface {
	SaveSignal(signal model.Signal)
	GetSignal(symbol string) *model.Signal
}

type SignalRepository struct {
	RDB        *redis.Client
	Ctx        *context.Context
	CurrentBot *model.Bot
}

func (s *SignalRepository) SaveSignal(signal model.Signal) {
	encoded, _ := json.Marshal(signal)
	s.RDB.Set(*s.Ctx, fmt.Sprintf("signal-%s-%d", strings.ToUpper(signal.Symbol), s.CurrentBot.Id), string(encoded), time.Millisecond*signal.GetTTLMilli())
}

func (s *SignalRepository) GetSignal(symbol string) *model.Signal {
	res := s.RDB.Get(*s.Ctx, fmt.Sprintf("signal-%s-%d", strings.ToUpper(symbol), s.CurrentBot.Id)).Val()
	if len(res) == 0 {
		return nil
	}

	var dto model.Signal
	err := json.Unmarshal([]byte(res), &dto)
	if err != nil {
		log.Printf("[%s] signal storage error: %s", symbol, err.Error())
		return nil
	}

	return &dto
}
