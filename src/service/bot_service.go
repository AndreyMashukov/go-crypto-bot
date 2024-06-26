package service

import (
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
)

type BotServiceInterface interface {
	GetBot() model.Bot
	IsSwapEnabled() bool
	IsMasterBot() bool
	GetTradeStackSorting() string
	UseSwapCapital() bool
	GetSwapConfig() model.SwapConfig
}

type BotService struct {
	CurrentBot    *model.Bot
	BotRepository *repository.BotRepository
}

func (b *BotService) GetBot() model.Bot {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id)
}
func (b *BotService) IsSwapEnabled() bool {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id).IsSwapEnabled
}
func (b *BotService) IsMasterBot() bool {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id).IsMasterBot
}
func (b *BotService) UseSwapCapital() bool {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id).SwapConfig.UseSwapCapital
}
func (b *BotService) GetSwapConfig() model.SwapConfig {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id).SwapConfig
}
func (b *BotService) GetTradeStackSorting() string {
	return b.BotRepository.GetCurrentBotCached(b.CurrentBot.Id).TradeStackSorting
}
