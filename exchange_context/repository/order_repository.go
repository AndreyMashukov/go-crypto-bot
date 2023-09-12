package exchange_context

import (
	"database/sql"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"log"
)

type OrderRepository struct {
	DB *sql.DB
}

func (repo *OrderRepository) GetOpenedOrder(symbol string, operation string) (ExchangeModel.Order, error) {
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
			o.used_extra_budget as UsedExtraBudget
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
			used_extra_budget = ?
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
	)

	if err != nil {
		log.Println(err)

		return nil, err
	}

	lastId, err := res.LastInsertId()

	return &lastId, err
}

func (repo *OrderRepository) Update(order ExchangeModel.Order) error {
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
			o.used_extra_budget = ?
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
			o.used_extra_budget as UsedExtraBudget
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
	)

	if err != nil {
		return order, err
	}

	return order, nil
}
