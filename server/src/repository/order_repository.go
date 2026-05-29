package repository

// All time.Time values written to the DB are UTC-normalised at the
// repository edge; all reads return UTC time.Time. The DB columns use
// timestamptz so the conversion is enforced at the storage layer.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
)

type OrderUpdaterInterface interface {
	Update(order model.Order) error
}

type OrderCachedReaderInterface interface {
	GetOpenedOrderCached(symbol string, operation string) *model.Order
}

type OrderStorageInterface interface {
	Create(order model.Order) (*int64, error)
	Update(order model.Order) error
	DeleteManualOrder(symbol string)
	Find(id int64) (model.Order, error)
	GetClosesOrderList(buyOrder model.Order) []model.Order
	DeleteBinanceOrder(order model.BinanceOrder)
	GetOpenedOrderCached(symbol string, operation string) *model.Order
	GetManualOrder(symbol string) *model.ManualOrder
	SetBinanceOrder(order model.BinanceOrder)
	GetBinanceOrder(symbol string, operation string) *model.BinanceOrder
	LockBuy(symbol string, seconds int64)
	HasBuyLock(symbol string) bool
	GetTodayExtraOrderMap() *sync.Map
}

type OrderRepository struct {
	DB               *sql.DB
	CurrentBot       *model.Bot
	RDB              *redis.Client
	Ctx              *context.Context
	ObjectRepository *ObjectRepository
}

const orderSelectColumns = `
	o.id                   as Id,
	o.symbol               as Symbol,
	o.quantity             as Quantity,
	o.executed_quantity    as ExecutedQuantity,
	o.price                as Price,
	o.created_at           as CreatedAt,
	o.operation            as Operation,
	o.status               as Status,
	o.sell_volume          as SellVolume,
	o.buy_volume           as BuyVolume,
	o.sma_value            as SmaValue,
	o.external_id          as ExternalId,
	o.closes_order         as ClosesOrder,
	o.used_extra_budget    as UsedExtraBudget,
	o.commission           as Commission,
	o.commission_asset     as CommissionAsset,
	COALESCE(SUM(sell.executed_quantity), 0) as SoldQuantity,
	o.swap                 as Swap,
	o.extra_charge_options as ExtraChargeOptions,
	o.profit_options       as ProfitOptions,
	0::double precision    as SwapQuantity,
	COUNT(extra.id)        as ExtraOrdersCount,
	o.exchange             as Exchange
`

const orderFromJoins = `
	FROM orders o
	LEFT JOIN orders sell  ON o.id = sell.closes_order  AND sell.operation = 'SELL'
	LEFT JOIN orders extra ON o.id = extra.closes_order AND extra.operation = 'BUY'
`

func (repo *OrderRepository) getOpenedOrderCacheKey(symbol string, operation string) string {
	return fmt.Sprintf(
		"opened-order-%s-%s-bot-%d",
		symbol,
		strings.ToLower(operation),
		repo.CurrentBot.Id,
	)
}

func (repo *OrderRepository) GetOpenedOrderCached(symbol string, operation string) *model.Order {
	res := repo.RDB.Get(*repo.Ctx, repo.getOpenedOrderCacheKey(symbol, operation)).Val()
	if len(res) > 0 {
		var dto model.Order
		err := json.Unmarshal([]byte(res), &dto)

		if err == nil && dto.GetPositionQuantityWithSwap() > 0 && dto.IsOpened() {
			return &dto
		}
	}

	order, err := repo.GetOpenedOrder(symbol, operation)

	if err != nil {
		return nil
	}

	repo.SaveOrderCache(order)

	return &order
}

func (repo *OrderRepository) SaveOrderCache(order model.Order) {
	encoded, err := json.Marshal(order)

	if err == nil {
		repo.RDB.Set(*repo.Ctx, repo.getOpenedOrderCacheKey(order.Symbol, order.Operation), string(encoded), time.Second*30)
	} else {
		repo.DeleteOpenedOrderCache(order)
	}
}

func (repo *OrderRepository) DeleteOpenedOrderCache(order model.Order) {
	repo.RDB.Del(*repo.Ctx, repo.getOpenedOrderCacheKey(order.Symbol, order.Operation)).Val()
}

func scanOrder(row interface{ Scan(...interface{}) error }, order *model.Order) error {
	return row.Scan(
		&order.Id,
		&order.Symbol,
		&order.Quantity,
		&order.ExecutedQuantity,
		&order.Price,
		&order.CreatedAt,
		&order.Operation,
		&order.Status,
		&order.SellVolume,
		&order.BuyVolume,
		&order.SmaValue,
		&order.ExternalId,
		&order.ClosesOrder,
		&order.UsedExtraBudget,
		&order.Commission,
		&order.CommissionAsset,
		&order.SoldQuantity,
		&order.Swap,
		&order.ExtraChargeOptions,
		&order.ProfitOptions,
		&order.SwapQuantity,
		&order.ExtraOrdersCount,
		&order.Exchange,
	)
}

// GetOpenedOrder loads the single opened order for (symbol, operation) on
// the current bot from Postgres, aggregating SoldQuantity from related
// SELL orders and ExtraOrdersCount from related extra BUY orders.
func (repo *OrderRepository) GetOpenedOrder(symbol, operation string) (model.Order, error) {
	var order model.Order

	err := scanOrder(repo.DB.QueryRow(`
		SELECT `+orderSelectColumns+orderFromJoins+`
		WHERE o.status = $1 AND o.symbol = $2 AND o.operation = $3 AND o.bot_id = $4 AND o.exchange = $5
		GROUP BY o.id`,
		"opened",
		symbol,
		operation,
		repo.CurrentBot.Id,
		repo.CurrentBot.Exchange,
	), &order)

	if err != nil {
		return order, err
	}

	return order, nil
}

func (repo *OrderRepository) Create(order model.Order) (*int64, error) {
	var lastID int64
	err := repo.DB.QueryRow(`
		INSERT INTO orders (
			symbol, quantity, executed_quantity, price, created_at,
			sell_volume, buy_volume, sma_value, operation, status,
			external_id, closes_order, used_extra_budget,
			commission, commission_asset,
			extra_charge_options, profit_options,
			bot_id, exchange
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13,
			$14, $15,
			$16, $17,
			$18, $19
		)
		RETURNING id
	`,
		order.Symbol,
		order.Quantity,
		order.ExecutedQuantity,
		order.Price,
		order.CreatedAt,
		order.SellVolume,
		order.BuyVolume,
		order.SmaValue,
		order.Operation,
		order.Status,
		order.ExternalId,
		order.ClosesOrder,
		order.UsedExtraBudget,
		order.Commission,
		order.CommissionAsset,
		order.ExtraChargeOptions,
		order.ProfitOptions,
		repo.CurrentBot.Id,
		repo.CurrentBot.Exchange,
	).Scan(&lastID)

	if err != nil {
		log.Println(err)

		return nil, err
	}

	return &lastID, nil
}

func (repo *OrderRepository) Update(order model.Order) error {
	repo.DeleteOpenedOrderCache(order)

	_, err := repo.DB.Exec(`
		UPDATE orders SET
			symbol               = $1,
			quantity             = $2,
			executed_quantity    = $3,
			price                = $4,
			created_at           = $5,
			sell_volume          = $6,
			buy_volume           = $7,
			sma_value            = $8,
			operation            = $9,
			status               = $10,
			external_id          = $11,
			closes_order         = $12,
			used_extra_budget    = $13,
			commission           = $14,
			commission_asset     = $15,
			swap                 = $16,
			extra_charge_options = $17,
			profit_options       = $18
		WHERE id = $19 AND bot_id = $20 AND exchange = $21
	`,
		order.Symbol,
		order.Quantity,
		order.ExecutedQuantity,
		order.Price,
		order.CreatedAt,
		order.SellVolume,
		order.BuyVolume,
		order.SmaValue,
		order.Operation,
		order.Status,
		order.ExternalId,
		order.ClosesOrder,
		order.UsedExtraBudget,
		order.Commission,
		order.CommissionAsset,
		order.Swap,
		order.ExtraChargeOptions,
		order.ProfitOptions,
		order.Id,
		repo.CurrentBot.Id,
		repo.CurrentBot.Exchange,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (repo *OrderRepository) Find(id int64) (model.Order, error) {
	var order model.Order

	err := scanOrder(repo.DB.QueryRow(`
		SELECT `+orderSelectColumns+orderFromJoins+`
		WHERE o.id = $1 AND o.bot_id = $2 AND o.exchange = $3
		GROUP BY o.id`,
		id, repo.CurrentBot.Id, repo.CurrentBot.Exchange,
	), &order)

	if err != nil {
		return order, err
	}

	return order, nil
}

func (repo *OrderRepository) GetTrades() []model.OrderTrade {
	res, err := repo.DB.Query(`
		SELECT
			trade.id                                                                 as OrderId,
			initial.created_at                                                       as Open,
			trade.created_at                                                         as Close,
			initial.price                                                            as Buy,
			trade.price                                                              as Sell,
			trade.executed_quantity                                                  as BuyQuantity,
			trade.executed_quantity                                                  as SellQuantity,
			(trade.price * trade.executed_quantity) - (initial.price * trade.executed_quantity) as Profit,
			trade.symbol                                                             as Symbol,
			EXTRACT(EPOCH FROM (trade.created_at - initial.created_at)) / 3600       as HoursOpened,
			(initial.price * initial.executed_quantity)                              as Budget,
			((trade.price * trade.executed_quantity) - (initial.price * trade.executed_quantity)) * 100 / (initial.price * trade.quantity) as Percent
		FROM orders trade
		INNER JOIN orders initial ON initial.id = trade.closes_order AND initial.operation = 'buy' AND initial.bot_id = $1
		WHERE trade.operation = 'sell' AND trade.status = 'closed' AND trade.bot_id = $2 AND trade.exchange = $3
		ORDER BY Close DESC
	`, repo.CurrentBot.Id, repo.CurrentBot.Id, repo.CurrentBot.Exchange)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	list := make([]model.OrderTrade, 0)

	for res.Next() {
		var orderTrade model.OrderTrade
		err := res.Scan(
			&orderTrade.OrderId,
			&orderTrade.Open,
			&orderTrade.Close,
			&orderTrade.Buy,
			&orderTrade.Sell,
			&orderTrade.BuyQuantity,
			&orderTrade.SellQuantity,
			&orderTrade.Profit,
			&orderTrade.Symbol,
			&orderTrade.HoursOpened,
			&orderTrade.Budget,
			&orderTrade.Percent,
		)

		if err != nil {
			log.Fatal(err)
		}

		list = append(list, orderTrade)
	}

	return list
}

func (repo *OrderRepository) GetHistoryList(symbol string, from time.Time, to time.Time) []model.Order {
	res, err := repo.DB.Query(`
		SELECT `+orderSelectColumns+orderFromJoins+`
		WHERE o.bot_id = $1 AND o.exchange = $2 AND o.symbol = $3 AND o.created_at >= $4 AND o.created_at <= $5
		GROUP BY o.id
	`,
		repo.CurrentBot.Id,
		repo.CurrentBot.Exchange,
		symbol,
		from.UTC(),
		to.UTC(),
	)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	list := make([]model.Order, 0)

	for res.Next() {
		var order model.Order
		if err := scanOrder(res, &order); err != nil {
			log.Fatal(err)
		}
		list = append(list, order)
	}

	return list
}

func (repo *OrderRepository) GetList() []model.Order {
	res, err := repo.DB.Query(`
		SELECT `+orderSelectColumns+orderFromJoins+`
		WHERE o.bot_id = $1 AND o.exchange = $2
		GROUP BY o.id
	`, repo.CurrentBot.Id, repo.CurrentBot.Exchange)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	list := make([]model.Order, 0)

	for res.Next() {
		var order model.Order
		if err := scanOrder(res, &order); err != nil {
			log.Fatal(err)
		}
		list = append(list, order)
	}

	return list
}

func (repo *OrderRepository) GetClosesOrderList(buyOrder model.Order) []model.Order {
	res, err := repo.DB.Query(`
		SELECT `+orderSelectColumns+orderFromJoins+`
		WHERE o.bot_id = $1 AND o.closes_order = $2 AND o.operation = $3 AND o.exchange = $4
		GROUP BY o.id
	`, repo.CurrentBot.Id, buyOrder.Id, "SELL", repo.CurrentBot.Exchange)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	list := make([]model.Order, 0)

	for res.Next() {
		var order model.Order
		if err := scanOrder(res, &order); err != nil {
			log.Fatal(err)
		}
		list = append(list, order)
	}

	return list
}

func (repo *OrderRepository) SetBinanceOrder(order model.BinanceOrder) {
	storageKey := fmt.Sprintf(
		"binance-order-%s-%s-bot-%d",
		order.Symbol,
		strings.ToLower(order.Side),
		repo.CurrentBot.Id,
	)
	err := repo.ObjectRepository.SaveObject(storageKey, order)

	if err == nil {
		repo.RDB.Del(*repo.Ctx, storageKey).Val()

		return
	} else {
		log.Printf("[%s] storage order save error: %s", order.Symbol, err.Error())
	}
}

func (repo *OrderRepository) GetBinanceOrder(symbol string, operation string) *model.BinanceOrder {
	var dto model.BinanceOrder
	storageKey := fmt.Sprintf(
		"binance-order-%s-%s-bot-%d",
		symbol,
		strings.ToLower(operation),
		repo.CurrentBot.Id,
	)

	err := repo.ObjectRepository.LoadObject(storageKey, &dto)
	if err == nil {
		repo.RDB.Del(*repo.Ctx, storageKey).Val()

		return &dto
	} else {
		if !strings.Contains(err.Error(), "no rows in result set") {
			log.Printf("[%s] storage order load error: %s", symbol, err.Error())
		}
	}

	res := repo.RDB.Get(*repo.Ctx, storageKey).Val()
	if len(res) == 0 {
		return nil
	}

	err = json.Unmarshal([]byte(res), &dto)

	if err != nil {
		return nil
	}

	if len(dto.OrderId) == 0 {
		return nil
	}

	_ = repo.ObjectRepository.SaveObject(storageKey, dto)

	return &dto
}

func (repo *OrderRepository) DeleteBinanceOrder(order model.BinanceOrder) {
	storageKey := fmt.Sprintf(
		"binance-order-%s-%s-bot-%d",
		order.Symbol,
		strings.ToLower(order.Side),
		repo.CurrentBot.Id,
	)

	err := repo.ObjectRepository.DeleteObject(storageKey)
	if err != nil {
		if !strings.Contains(err.Error(), "no rows in result set") {
			log.Printf("[%s] storage order delete error: %s", order.Symbol, err.Error())
		}
	}

	repo.RDB.Del(*repo.Ctx, storageKey).Val()
}

func (repo *OrderRepository) GetManualOrder(symbol string) *model.ManualOrder {
	res := repo.RDB.Get(*repo.Ctx, fmt.Sprintf(
		"manual-order-%s-bot-%d",
		strings.ToLower(symbol),
		repo.CurrentBot.Id,
	)).Val()
	if len(res) == 0 {
		return nil
	}

	var dto model.ManualOrder
	err := json.Unmarshal([]byte(res), &dto)

	if err != nil {
		return nil
	}

	return &dto
}

func (repo *OrderRepository) SetManualOrder(order model.ManualOrder) {
	encoded, _ := json.Marshal(order)
	repo.RDB.Set(*repo.Ctx, fmt.Sprintf(
		"manual-order-%s-bot-%d",
		strings.ToLower(order.Symbol),
		repo.CurrentBot.Id,
	), string(encoded), time.Second*time.Duration(order.Ttl))
}

func (repo *OrderRepository) DeleteManualOrder(symbol string) {
	repo.RDB.Del(*repo.Ctx, fmt.Sprintf(
		"manual-order-%s-bot-%d",
		strings.ToLower(symbol),
		repo.CurrentBot.Id,
	)).Val()
}

func (repo *OrderRepository) HasBuyLock(symbol string) bool {
	value := repo.RDB.Get(*repo.Ctx, fmt.Sprintf(
		"buy-lock-%s-bot-%d",
		strings.ToLower(symbol),
		repo.CurrentBot.Id,
	)).Val()

	return len(value) > 0
}

func (repo *OrderRepository) LockBuy(symbol string, seconds int64) {
	repo.RDB.Set(*repo.Ctx, fmt.Sprintf(
		"buy-lock-%s-bot-%d",
		strings.ToLower(symbol),
		repo.CurrentBot.Id,
	), "lock", time.Second*time.Duration(seconds))
}

func (repo *OrderRepository) GetTodayExtraOrderMap() *sync.Map {
	res, err := repo.DB.Query(`
		SELECT
			origin.symbol           as OriginSymbol,
			COUNT(DISTINCT extra.id) as Extras
		FROM orders extra
		INNER JOIN orders origin ON origin.id = extra.closes_order AND origin.status = 'opened' AND origin.operation = 'BUY'
		WHERE extra.operation = 'BUY' AND extra.bot_id = $1 AND extra.created_at >= CURRENT_DATE AND extra.exchange = $2
		GROUP BY origin.symbol
		ORDER BY Extras DESC
	`, repo.CurrentBot.Id, repo.CurrentBot.Exchange)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	extraOrderMap := sync.Map{}

	for res.Next() {
		var symbol string
		var count float64
		err := res.Scan(
			&symbol,
			&count,
		)

		if err != nil {
			log.Fatal(err)
		}
		extraOrderMap.Store(symbol, count)
	}

	return &extraOrderMap
}
