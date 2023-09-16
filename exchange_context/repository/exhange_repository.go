package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	model "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
	"slices"
	"time"
)

type ExchangeRepository struct {
	DB  *sql.DB
	RDB *redis.Client
	Ctx *context.Context
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

func (e *ExchangeRepository) AddKLine(kLine model.KLine) {
	list := e.KLineList(kLine.Symbol, false, 1)

	if len(list) > 0 {
		firstKLine := list[0]
		if firstKLine.Timestamp == kLine.Timestamp {
			e.RDB.LPop(*e.Ctx, fmt.Sprintf("k-lines-%s", kLine.Symbol)).Val()
		}
	}

	encoded, _ := json.Marshal(kLine)
	e.RDB.LPush(*e.Ctx, fmt.Sprintf("k-lines-%s", kLine.Symbol), string(encoded))
	e.RDB.LTrim(*e.Ctx, fmt.Sprintf("k-lines-%s", kLine.Symbol), 0, 2880)
}

func (e *ExchangeRepository) KLineList(symbol string, reverse bool, size int64) []model.KLine {
	res := e.RDB.LRange(*e.Ctx, fmt.Sprintf("k-lines-%s", symbol), 0, size).Val()
	list := make([]model.KLine, 0)

	for _, str := range res {
		var dto model.KLine
		json.Unmarshal([]byte(str), &dto)
		list = append(list, dto)
	}

	if reverse {
		slices.Reverse(list)
	}

	return list
}

func (e *ExchangeRepository) SetDepth(depth model.Depth) {
	encoded, _ := json.Marshal(depth)
	e.RDB.Set(*e.Ctx, fmt.Sprintf("depth-%s", depth.Symbol), string(encoded), time.Second*5)
}

func (e *ExchangeRepository) GetDepth(symbol string) model.Depth {
	res := e.RDB.Get(*e.Ctx, fmt.Sprintf("depth-%s", symbol)).Val()
	if len(res) == 0 {
		return model.Depth{
			Asks:      make([][2]model.Number, 0),
			Bids:      make([][2]model.Number, 0),
			Symbol:    symbol,
			Timestamp: time.Now().UnixMilli(),
		}
	}

	var dto model.Depth
	json.Unmarshal([]byte(res), &dto)

	return dto
}

func (e *ExchangeRepository) AddTrade(trade model.Trade) {
	encoded, _ := json.Marshal(trade)
	e.RDB.LPush(*e.Ctx, fmt.Sprintf("trades-%s", trade.Symbol), string(encoded))
	e.RDB.LTrim(*e.Ctx, fmt.Sprintf("trades-%s", trade.Symbol), 0, 100)
}

func (e *ExchangeRepository) TradeList(symbol string) []model.Trade {
	res := e.RDB.LRange(*e.Ctx, fmt.Sprintf("trades-%s", symbol), 0, 100).Val()
	list := make([]model.Trade, 0)

	for _, str := range res {
		var dto model.Trade
		json.Unmarshal([]byte(str), &dto)
		list = append(list, dto)
	}

	slices.Reverse(list)
	return list
}

func (e *ExchangeRepository) SetDecision(decision model.Decision) {
	encoded, _ := json.Marshal(decision)
	e.RDB.Set(*e.Ctx, fmt.Sprintf("decision-%s", decision.StrategyName), string(encoded), time.Second*5)
}

func (e *ExchangeRepository) GetDecision(strategy string) *model.Decision {
	res := e.RDB.Get(*e.Ctx, fmt.Sprintf("decision-%s", strategy)).Val()
	if len(res) == 0 {
		return nil
	}

	var dto model.Decision
	json.Unmarshal([]byte(res), &dto)

	return &dto
}
