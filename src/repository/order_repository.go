package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"strings"
	"sync"
	"time"
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
	DB         *sql.DB
	RDB        *redis.Client
	Ctx        *context.Context
	CurrentBot *model.Bot
}

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

func (repo *OrderRepository) GetOpenedOrder(symbol string, operation string) (model.Order, error) {
	var order model.Order

	err := repo.DB.QueryRow(`
		SELECT 
			o.id as Id, 
			o.symbol as Symbol, 
			o.quantity as Quantity,
			o.executed_quantity as ExecutedQuantity,
			o.price as Price,
			o.created_at as CreatedAt,
			o.operation as Operation,
			o.status as Status,
			o.sell_volume as SellVolume,
			o.buy_volume as BuyVolume,
			o.sma_value as SmaValue,
			o.external_id as ExternalId,
			o.closes_order as ClosesOrder,
			o.used_extra_budget as UsedExtraBudget,
			o.commission as Commission,
			o.commission_asset as CommissionAsset,
			SUM(IFNULL(sell.executed_quantity, 0)) as SoldQuantity,
			o.swap as Swap,
			o.extra_charge_options as ExtraChargeOptions,
			o.profit_options as ProfitOptions,
    		IFNULL(SUM(sa.end_quantity - sa.start_quantity), 0) as SwapQuantity,
    		COUNT(extra.id) as ExtraOrdersCount,
    		o.exchange as Exchange
		FROM orders o
		LEFT JOIN orders sell ON o.id = sell.closes_order AND sell.operation = 'SELL'
     	LEFT JOIN orders extra ON o.id = extra.closes_order AND extra.operation = 'BUY'
		LEFT JOIN swap_action sa on o.id = sa.order_id AND sa.status = ?
		WHERE o.status = ? AND o.symbol = ? AND o.operation = ? AND o.bot_id = ? AND o.exchange = ?
		GROUP BY o.id`,
		"success",
		"opened",
		symbol,
		operation,
		repo.CurrentBot.Id,
		repo.CurrentBot.Exchange,
	).Scan(
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

	if err != nil {
		return order, err
	}

	return order, nil
}

func (repo *OrderRepository) Create(order model.Order) (*int64, error) {
	res, err := repo.DB.Exec(`
		INSERT INTO orders SET
	  		symbol = ?,
		    quantity = ?,
		    executed_quantity = ?,
	        price = ?,
		    created_at = ?,
		    sell_volume = ?,
	        buy_volume = ?,
		    sma_value = ?,
		    operation = ?,
		    status = ?,
		    external_id = ?,
		    closes_order = ?,
			used_extra_budget = ?,
			commission = ?,
			commission_asset = ?,
			extra_charge_options = ?,
			profit_options = ?,
			bot_id = ?,
			exchange = ?
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
	)

	if err != nil {
		log.Println(err)

		return nil, err
	}

	lastId, err := res.LastInsertId()

	return &lastId, err
}

func (repo *OrderRepository) Update(order model.Order) error {
	repo.DeleteOpenedOrderCache(order)

	_, err := repo.DB.Exec(`
		UPDATE orders o SET
	  		o.symbol = ?,
		    o.quantity = ?,
		    o.executed_quantity = ?,
	        o.price = ?,
		    o.created_at = ?,
		    o.sell_volume = ?,
	        o.buy_volume = ?,
		    o.sma_value = ?,
		    o.operation = ?,
		    o.status = ?,
		    o.external_id = ?,
			o.closes_order = ?,
			o.used_extra_budget = ?,
			o.commission = ?,
			o.commission_asset = ?,
			o.swap = ?,
			o.extra_charge_options = ?,
			o.profit_options = ?
		WHERE o.id = ? AND o.bot_id = ? AND exchange = ?
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

	err := repo.DB.QueryRow(`
		SELECT 
			o.id as Id, 
			o.symbol as Symbol, 
			o.quantity as Quantity,
			o.executed_quantity as ExecutedQuantity,
			o.price as Price,
			o.created_at as CreatedAt,
			o.operation as Operation,
			o.status as Status,
			o.sell_volume as SellVolume,
			o.buy_volume as BuyVolume,
			o.sma_value as SmaValue,
			o.external_id as ExternalId,
			o.closes_order as ClosesOrder,
			o.used_extra_budget as UsedExtraBudget,
			o.commission as Commission,
			o.commission_asset as CommissionAsset,
			SUM(IFNULL(sell.executed_quantity, 0)) as SoldQuantity,
			o.swap as Swap,
			o.extra_charge_options as ExtraChargeOptions,
			o.profit_options as ProfitOptions,
    		IFNULL(SUM(sa.end_quantity - sa.start_quantity), 0) as SwapQuantity,
    		COUNT(extra.id) as ExtraOrdersCount,
    		o.exchange as Exchange
		FROM orders o
		LEFT JOIN orders sell ON o.id = sell.closes_order AND sell.operation = 'SELL'
     	LEFT JOIN orders extra ON o.id = extra.closes_order AND extra.operation = 'BUY'
		LEFT JOIN swap_action sa on o.id = sa.order_id AND sa.status = ?
		WHERE o.id = ? AND o.bot_id = ? AND o.exchange = ?
		GROUP BY o.id`, "success", id, repo.CurrentBot.Id, repo.CurrentBot.Exchange,
	).Scan(
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

	if err != nil {
		return order, err
	}

	return order, nil
}

func (repo *OrderRepository) GetTrades() []model.OrderTrade {
	res, err := repo.DB.Query(`
		SELECT
			trade.id as OrderId,
			initial.created_at as Open,
			trade.created_at as Close,
			initial.price as Buy,
			trade.price as Sell,
			trade.executed_quantity as BuyQuantity,
			trade.executed_quantity as SellQuantity,
			(trade.price * trade.executed_quantity) - (initial.price * trade.executed_quantity) as Profit,
			trade.symbol as Symbol,
			TIMESTAMPDIFF(HOUR, initial.created_at, trade.created_at) as HoursOpened,
			(initial.price * initial.executed_quantity) as Budget,
			((trade.price * trade.executed_quantity) - (initial.price * trade.executed_quantity)) * 100 / (initial.price * trade.quantity) as Percent
		FROM orders trade
		INNER JOIN orders initial ON initial.id = trade.closes_order AND initial.operation = 'buy' AND initial.bot_id = ?
		WHERE trade.operation = 'sell' and trade.status = 'closed' AND trade.bot_id = ? AND trade.exchange = ?
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

func (repo *OrderRepository) GetList() []model.Order {
	res, err := repo.DB.Query(`
		SELECT
		    o.id as Id, 
			o.symbol as Symbol, 
			o.quantity as Quantity,
			o.executed_quantity as ExecutedQuantity,
			o.price as Price,
			o.created_at as CreatedAt,
			o.operation as Operation,
			o.status as Status,
			o.sell_volume as SellVolume,
			o.buy_volume as BuyVolume,
			o.sma_value as SmaValue,
			o.external_id as ExternalId,
			o.closes_order as ClosesOrder,
			o.used_extra_budget as UsedExtraBudget,
			o.commission as Commission,
			o.commission_asset as CommissionAsset,
			SUM(IFNULL(sell.executed_quantity, 0)) as SoldQuantity,
			o.swap as Swap,
			o.extra_charge_options as ExtraChargeOptions,
			o.profit_options as ProfitOptions,
    		IFNULL(SUM(sa.end_quantity - sa.start_quantity), 0) as SwapQuantity,
    		COUNT(extra.id) as ExtraOrdersCount,
    		o.exchange as Exchange
		FROM orders o 
		LEFT JOIN orders sell ON o.id = sell.closes_order AND sell.operation = 'SELL'
     	LEFT JOIN orders extra ON o.id = extra.closes_order AND extra.operation = 'BUY'
		LEFT JOIN swap_action sa on o.id = sa.order_id AND sa.status = ?
		WHERE o.bot_id = ? AND o.exchange = ?
		GROUP BY o.id
	`, "success", repo.CurrentBot.Id, repo.CurrentBot.Exchange)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	list := make([]model.Order, 0)

	for res.Next() {
		var order model.Order
		err := res.Scan(
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

		if err != nil {
			log.Fatal(err)
		}

		list = append(list, order)
	}

	return list
}

func (repo *OrderRepository) GetClosesOrderList(buyOrder model.Order) []model.Order {
	res, err := repo.DB.Query(`
		SELECT
		    o.id as Id, 
			o.symbol as Symbol, 
			o.quantity as Quantity,
			o.executed_quantity as ExecutedQuantity,
			o.price as Price,
			o.created_at as CreatedAt,
			o.operation as Operation,
			o.status as Status,
			o.sell_volume as SellVolume,
			o.buy_volume as BuyVolume,
			o.sma_value as SmaValue,
			o.external_id as ExternalId,
			o.closes_order as ClosesOrder,
			o.used_extra_budget as UsedExtraBudget,
			o.commission as Commission,
			o.commission_asset as CommissionAsset,
			SUM(IFNULL(sell.executed_quantity, 0)) as SoldQuantity,
			o.swap as Swap,
			o.extra_charge_options as ExtraChargeOptions,
			o.profit_options as ProfitOptions,
    		IFNULL(SUM(sa.end_quantity - sa.start_quantity), 0) as SwapQuantity,
    		COUNT(extra.id) as ExtraOrdersCount,
    		o.exchange as Exchange
		FROM orders o 
		LEFT JOIN orders sell ON o.id = sell.closes_order AND sell.operation = 'SELL'
     	LEFT JOIN orders extra ON o.id = extra.closes_order AND extra.operation = 'BUY'
		LEFT JOIN swap_action sa on o.id = sa.order_id AND sa.status = ?
		WHERE o.bot_id = ? AND o.closes_order = ? AND o.operation = ? AND o.exchange = ?
		GROUP BY o.id
	`, "success", repo.CurrentBot.Id, buyOrder.Id, "SELL", repo.CurrentBot.Exchange)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	list := make([]model.Order, 0)

	for res.Next() {
		var order model.Order
		err := res.Scan(
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

		if err != nil {
			log.Fatal(err)
		}

		list = append(list, order)
	}

	return list
}

func (repo *OrderRepository) SetBinanceOrder(order model.BinanceOrder) {
	encoded, _ := json.Marshal(order)
	repo.RDB.Set(*repo.Ctx, fmt.Sprintf(
		"binance-order-%s-%s-bot-%d",
		order.Symbol,
		strings.ToLower(order.Side),
		repo.CurrentBot.Id,
	), string(encoded), time.Hour*24*90)
}

func (repo *OrderRepository) GetBinanceOrder(symbol string, operation string) *model.BinanceOrder {
	res := repo.RDB.Get(*repo.Ctx, fmt.Sprintf(
		"binance-order-%s-%s-bot-%d",
		symbol,
		strings.ToLower(operation),
		repo.CurrentBot.Id,
	)).Val()
	if len(res) == 0 {
		return nil
	}

	var dto model.BinanceOrder
	err := json.Unmarshal([]byte(res), &dto)

	if err != nil {
		return nil
	}

	return &dto
}

func (repo *OrderRepository) DeleteBinanceOrder(order model.BinanceOrder) {
	repo.RDB.Del(*repo.Ctx, fmt.Sprintf(
		"binance-order-%s-%s-bot-%d",
		order.Symbol,
		strings.ToLower(order.Side),
		repo.CurrentBot.Id,
	)).Val()
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
			origin.symbol as OriginSymbol,
			COUNT(DISTINCT extra.id) as Extras
		FROM orders extra
		INNER JOIN orders origin ON origin.id = extra.closes_order and origin.status = 'opened' AND origin.operation = 'BUY'
		WHERE extra.operation = 'BUY' AND extra.bot_id = ? AND extra.created_at >= CURDATE() AND extra.exchange = ?
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
