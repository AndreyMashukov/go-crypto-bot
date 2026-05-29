package service

import (
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/repository"
)

type BotServiceInterface interface {
	GetBot() model.Bot
	IsSwapEnabled() bool
	IsMasterBot() bool
	GetTradeStackSorting() string
	UseSwapCapital() bool
}

type BotService struct {
	CurrentBot    *model.Bot
	BotRepository *repository.BotRepository
}

func (b *BotService) GetBot() model.Bot {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id)
}
func (b *BotService) IsSwapEnabled() bool {
	return false
}
func (b *BotService) IsMasterBot() bool {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id).IsMasterBot
}
func (b *BotService) UseSwapCapital() bool {
	return false
}
func (b *BotService) GetTradeStackSorting() string {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id).TradeStackSorting
}
