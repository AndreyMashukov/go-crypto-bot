package service

import (
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/repository"
)

type BotServiceInterface interface {
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
