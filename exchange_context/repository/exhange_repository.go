package repository

import (
	"database/sql"
	model "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
)

type ExchangeRepository struct {
	DB *sql.DB
}

func (e *ExchangeRepository) GetSubscribedSymbols() []model.Symbol {
	symbolSlice := make([]model.Symbol, 0)
	symbolSlice = append(symbolSlice, model.Symbol{Value: "BTCUSDT"})

	return symbolSlice
}

func (e *ExchangeRepository) GetTradeLimits() []model.TradeLimit {
	res, err := e.DB.Query(`
		SELECT
		    tl.id as Id,
		    tl.symbol as Symbol,
		    tl.usdt_limit as USDTLimit,
		    tl.min_price as MinPrice,
		    tl.min_quantity as MinQuantity,
		    tl.min_profit_percent as MinProfitPercent,
		    tl.is_enabled as IsEnabled,
		    tl.usdt_extra_budget as USDTExtraBudget,
		    tl.buy_on_fall_percent as BuyOnFallPercent
		FROM trade_limit tl
	`)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	list := make([]model.TradeLimit, 0)

	for res.Next() {
		var tradeLimit model.TradeLimit
		err := res.Scan(
			&tradeLimit.Id,
			&tradeLimit.Symbol,
			&tradeLimit.USDTLimit,
			&tradeLimit.MinPrice,
			&tradeLimit.MinQuantity,
			&tradeLimit.MinProfitPercent,
			&tradeLimit.IsEnabled,
			&tradeLimit.USDTExtraBudget,
			&tradeLimit.BuyOnFallPercent,
		)

		if err != nil {
			log.Fatal(err)
		}

		list = append(list, tradeLimit)
	}

	return list
}

func (e *ExchangeRepository) GetTradeLimit(symbol string) (model.TradeLimit, error) {
	var tradeLimit model.TradeLimit
	err := e.DB.QueryRow(`
		SELECT
		    tl.id as Id,
		    tl.symbol as Symbol,
		    tl.usdt_limit as USDTLimit,
		    tl.min_price as MinPrice,
		    tl.min_quantity as MinQuantity,
		    tl.min_profit_percent as MinProfitPercent,
		    tl.is_enabled as IsEnabled,
		    tl.usdt_extra_budget as USDTExtraBudget,
		    tl.buy_on_fall_percent as BuyOnFallPercent
		FROM trade_limit tl
		WHERE tl.symbol = ?
	`,
		symbol,
	).Scan(
		&tradeLimit.Id,
		&tradeLimit.Symbol,
		&tradeLimit.USDTLimit,
		&tradeLimit.MinPrice,
		&tradeLimit.MinQuantity,
		&tradeLimit.MinProfitPercent,
		&tradeLimit.IsEnabled,
		&tradeLimit.USDTExtraBudget,
		&tradeLimit.BuyOnFallPercent,
	)
	if err != nil {
		return tradeLimit, err
	}

	return tradeLimit, nil
}
