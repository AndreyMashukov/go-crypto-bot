package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
	"strings"
	"time"
)

type OrderRepository struct {
	DB  *sql.DB
	RDB *redis.Client
	Ctx *context.Context
}

func (repo *OrderRepository) GetOpenedOrderCached(symbol string, operation string) (ExchangeModel.Order, error) {
	res := repo.RDB.Get(*repo.Ctx, fmt.Sprintf("opened-order-%s-%s", symbol, strings.ToLower(operation))).Val()
	if len(res) > 0 {
		var dto ExchangeModel.Order
		json.Unmarshal([]byte(res), &dto)

		cached := repo.GetBinanceOrder(symbol, operation)
		if cached != nil && cached.OrderId == *dto.ExternalId {
			repo.DeleteBinanceOrder(*cached)
		}

		return dto, nil
	}

	order, err := repo.getOpenedOrder(symbol, operation)

	if err != nil {
		return order, err
	}

	encoded, _ := json.Marshal(order)
	repo.RDB.Set(*repo.Ctx, fmt.Sprintf("opened-order-%s-%s", symbol, operation), string(encoded), time.Minute*60)

	cached := repo.GetBinanceOrder(symbol, operation)
	if cached != nil && cached.OrderId == *order.ExternalId {
		repo.DeleteBinanceOrder(*cached)
	}

	return order, nil
}

func (repo *OrderRepository) DeleteOpenedOrderCache(order ExchangeModel.Order) {
	repo.RDB.Del(*repo.Ctx, fmt.Sprintf("opened-order-%s-%s", order.Symbol, strings.ToLower(order.Operation))).Val()
}

func (repo *OrderRepository) getOpenedOrder(symbol string, operation string) (ExchangeModel.Order, error) {
	var order ExchangeModel.Order

	err := repo.DB.QueryRow(`
		SELECT 
			o.id as Id, 
			o.symbol as Symbol, 
			o.quantity as Quantity,
			o.price as Price,
			o.created_at as CreatedAt,
			o.operation as Operation,
			o.status as Status,
			o.sell_volume as SellVolume,
			o.buy_volume as BuyVolume,
			o.sma_value as SmaValue,
			o.external_id as ExternalId,
			o.closed_by as ClosedBy,
			o.used_extra_budget as UsedExtraBudget,
			o.commission as Commission,
			o.commission_asset as CommissionAsset
		FROM orders o
		WHERE o.status = ? AND o.symbol = ? AND o.operation = ?`, "opened", symbol, operation,
	).Scan(
		&order.Id,
		&order.Symbol,
		&order.Quantity,
		&order.Price,
		&order.CreatedAt,
		&order.Operation,
		&order.Status,
		&order.SellVolume,
		&order.BuyVolume,
		&order.SmaValue,
		&order.ExternalId,
		&order.ClosedBy,
		&order.UsedExtraBudget,
		&order.Commission,
		&order.CommissionAsset,
	)

	if err != nil {
		return order, err
	}

	return order, nil
}

func (repo *OrderRepository) Create(order ExchangeModel.Order) (*int64, error) {
	res, err := repo.DB.Exec(`
		INSERT INTO orders SET
	  		symbol = ?,
		    quantity = ?,
	        price = ?,
		    created_at = ?,
		    sell_volume = ?,
	        buy_volume = ?,
		    sma_value = ?,
		    operation = ?,
		    status = ?,
		    external_id = ?,
		    closed_by = ?,
			used_extra_budget = ?,
			commission = ?,
			commission_asset = ?
	`,
		order.Symbol,
		order.Quantity,
		order.Price,
		order.CreatedAt,
		order.SellVolume,
		order.BuyVolume,
		order.SmaValue,
		order.Operation,
		order.Status,
		order.ExternalId,
		order.ClosedBy,
		order.UsedExtraBudget,
		order.Commission,
		order.CommissionAsset,
	)

	if err != nil {
		log.Println(err)

		return nil, err
	}

	lastId, err := res.LastInsertId()

	return &lastId, err
}

func (repo *OrderRepository) Update(order ExchangeModel.Order) error {
	repo.DeleteOpenedOrderCache(order)
	_, err := repo.DB.Exec(`
		UPDATE orders o SET
	  		o.symbol = ?,
		    o.quantity = ?,
	        o.price = ?,
		    o.created_at = ?,
		    o.sell_volume = ?,
	        o.buy_volume = ?,
		    o.sma_value = ?,
		    o.operation = ?,
		    o.status = ?,
		    o.external_id = ?,
		    o.closed_by = ?,
			o.used_extra_budget = ?,
			o.commission = ?,
			o.commission_asset = ?
		WHERE o.id = ?
	`,
		order.Symbol,
		order.Quantity,
		order.Price,
		order.CreatedAt,
		order.SellVolume,
		order.BuyVolume,
		order.SmaValue,
		order.Operation,
		order.Status,
		order.ExternalId,
		order.ClosedBy,
		order.UsedExtraBudget,
		order.Commission,
		order.CommissionAsset,
		order.Id,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (repo *OrderRepository) Find(id int64) (ExchangeModel.Order, error) {
	var order ExchangeModel.Order

	err := repo.DB.QueryRow(`
		SELECT 
			o.id as Id, 
			o.symbol as Symbol, 
			o.quantity as Quantity,
			o.price as Price,
			o.created_at as CreatedAt,
			o.operation as Operation,
			o.status as Status,
			o.sell_volume as SellVolume,
			o.buy_volume as BuyVolume,
			o.sma_value as SmaValue,
			o.external_id as ExternalId,
			o.closed_by as ClosedBy,
			o.used_extra_budget as UsedExtraBudget,
			o.commission as Commission,
			o.commission_asset as CommissionAsset
		FROM orders o
		WHERE o.id = ?`, id,
	).Scan(
		&order.Id,
		&order.Symbol,
		&order.Quantity,
		&order.Price,
		&order.CreatedAt,
		&order.Operation,
		&order.Status,
		&order.SellVolume,
		&order.BuyVolume,
		&order.SmaValue,
		&order.ExternalId,
		&order.ClosedBy,
		&order.UsedExtraBudget,
		&order.Commission,
		&order.CommissionAsset,
	)

	if err != nil {
		return order, err
	}

	return order, nil
}

func (repo *OrderRepository) GetList() []ExchangeModel.Order {
	res, err := repo.DB.Query(`
		SELECT
		    o.id as Id, 
			o.symbol as Symbol, 
			o.quantity as Quantity,
			o.price as Price,
			o.created_at as CreatedAt,
			o.operation as Operation,
			o.status as Status,
			o.sell_volume as SellVolume,
			o.buy_volume as BuyVolume,
			o.sma_value as SmaValue,
			o.external_id as ExternalId,
			o.closed_by as ClosedBy,
			o.used_extra_budget as UsedExtraBudget,
			o.commission as Commission,
			o.commission_asset as CommissionAsset
		FROM orders o
	`)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	list := make([]ExchangeModel.Order, 0)

	for res.Next() {
		var order ExchangeModel.Order
		err := res.Scan(
			&order.Id,
			&order.Symbol,
			&order.Quantity,
			&order.Price,
			&order.CreatedAt,
			&order.Operation,
			&order.Status,
			&order.SellVolume,
			&order.BuyVolume,
			&order.SmaValue,
			&order.ExternalId,
			&order.ClosedBy,
			&order.UsedExtraBudget,
			&order.Commission,
			&order.CommissionAsset,
		)

		if err != nil {
			log.Fatal(err)
		}

		list = append(list, order)
	}

	return list
}

func (repo *OrderRepository) SetBinanceOrder(order ExchangeModel.BinanceOrder) {
	encoded, _ := json.Marshal(order)
	repo.RDB.Set(*repo.Ctx, fmt.Sprintf("binance-order-%s-%s", order.Symbol, strings.ToLower(order.Side)), string(encoded), time.Hour*24*90)
}

func (repo *OrderRepository) GetBinanceOrder(symbol string, operation string) *ExchangeModel.BinanceOrder {
	res := repo.RDB.Get(*repo.Ctx, fmt.Sprintf("binance-order-%s-%s", symbol, strings.ToLower(operation))).Val()
	if len(res) == 0 {
		return nil
	}

	var dto ExchangeModel.BinanceOrder
	json.Unmarshal([]byte(res), &dto)

	return &dto
}

func (repo *OrderRepository) DeleteBinanceOrder(order ExchangeModel.BinanceOrder) {
	repo.RDB.Del(*repo.Ctx, fmt.Sprintf("binance-order-%s-%s", order.Symbol, strings.ToLower(order.Side))).Val()
}

func (repo *OrderRepository) GetManualOrder(symbol string) *ExchangeModel.ManualOrder {
	res := repo.RDB.Get(*repo.Ctx, fmt.Sprintf("manual-order-%s", strings.ToLower(symbol))).Val()
	if len(res) == 0 {
		return nil
	}

	var dto ExchangeModel.ManualOrder
	json.Unmarshal([]byte(res), &dto)

	return &dto
}

func (repo *OrderRepository) SetManualOrder(order ExchangeModel.ManualOrder) {
	encoded, _ := json.Marshal(order)
	repo.RDB.Set(*repo.Ctx, fmt.Sprintf("manual-order-%s", strings.ToLower(order.Symbol)), string(encoded), time.Hour*24)
}

func (repo *OrderRepository) DeleteManualOrder(symbol string) {
	repo.RDB.Del(*repo.Ctx, fmt.Sprintf("manual-order-%s", strings.ToLower(symbol))).Val()
}
