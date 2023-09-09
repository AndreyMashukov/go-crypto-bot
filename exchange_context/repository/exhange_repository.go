package exchange_context

import model "github.com/AndreyMashukov/go-crypto-bot/server/exchange_context/model"

func GetSubscribedSymbols() []model.Symbol {
	symbolSlice := make([]model.Symbol, 0)
	symbolSlice = append(symbolSlice, model.Symbol{Value: "BTCUSDT"})

	return symbolSlice
}
