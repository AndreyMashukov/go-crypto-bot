package service

import (
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/repository"
)

type BotServiceInterface interface {
	IsSwapEnabled() bool
	IsMasterBot() bool
}

type BotService struct {
	CurrentBot    *model.Bot
	BotRepository *repository.BotRepository
}

func (b *BotService) IsSwapEnabled() bool {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id).IsSwapEnabled
}

func (b *BotService) IsMasterBot() bool {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id).IsMasterBot
}
