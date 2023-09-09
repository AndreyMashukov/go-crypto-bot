package exchange_context

import model "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"

func GetSubscribedSymbols() []model.Symbol {
	symbolSlice := make([]model.Symbol, 0)
	symbolSlice = append(symbolSlice, model.Symbol{Value: "BTCUSDT"})

	return symbolSlice
}
