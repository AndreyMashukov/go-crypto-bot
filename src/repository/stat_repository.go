package repository

import (
	"database/sql"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"sync"
)

type StatRepository struct {
	DB         *sql.DB
	CurrentBot *model.Bot
}

func (s *StatRepository) WriteTradeStat(stat model.TradeStat) error {
	res, err := s.DB.Exec(`
		INSERT INTO default.trades (*) VALUES(
			?, -- Symbol
			?, -- Timestamp
			?, -- BotId
			?, -- Price
			?, -- BuyQty
			?, -- SellQty
			?, -- BuyVolume
			?, -- SellVolume
			?, -- TradeCount
			?, -- MaxPCS
			?, -- MinPCS
			?, -- Open
			?, -- Close
			?, -- High
			?, -- Low
			?, -- Volume
			?, -- Buy Order Book Length
			?, -- Sell Order Book Length
			?, -- Buy Order Book Qty Sum
			?, -- Sell Order Book Qty Sum
			?, -- Buy Order Book Volume Sum
			?, -- Sell Order Book Volume Sum
			?, -- Buy Iceberg Qty
			?, -- Buy Iceberg Price
			?, -- Sell Iceberg Qty
			?, -- Sell Iceberg Price
			?, -- First Buy Qty
			?, -- First Buy Price
			?, -- First Sell Qty
			? -- First Sell Price
		)
	`,
		stat.Symbol,
		stat.Timestamp.GetPeriodToMinute()/1000,
		s.CurrentBot.BotUuid,
		stat.Price,
		stat.BuyQty,
		stat.SellQty,
		stat.BuyVolume,
		stat.SellVolume,
		stat.TradeCount,
		stat.MaxPSC,
		stat.MinPCS,
		stat.Open,
		stat.Close,
		stat.High,
		stat.Low,
		stat.Volume,
		stat.OrderBookStat.BuyLength,
		stat.OrderBookStat.SellLength,
		stat.OrderBookStat.BuyQtySum,
		stat.OrderBookStat.SellQtySum,
		stat.OrderBookStat.BuyVolumeSum,
		stat.OrderBookStat.SellVolumeSum,
		stat.OrderBookStat.BuyIceberg.Quantity,
		stat.OrderBookStat.BuyIceberg.Price,
		stat.OrderBookStat.SellIceberg.Quantity,
		stat.OrderBookStat.SellIceberg.Price,
		stat.OrderBookStat.FirstBuyQty,
		stat.OrderBookStat.FirstBuyPrice,
		stat.OrderBookStat.FirstSellQty,
		stat.OrderBookStat.FirstSellPrice,
	)

	if err != nil {
		log.Printf("WriteTradeStat: %s", err.Error())
		return err
	}

	_, err = res.LastInsertId()

	return err
}

func (s *StatRepository) GetStatRange(symbol string, limit int64) *sync.Map {
	statMap := sync.Map{}

	res, err := s.DB.Query(`
		SELECT
		    symbol as Symbol,
			toUnixTimestamp64Milli(timestamp) as DateTime,
			bot_id as BotId,
			price as Price,
			buy_qty as BuyQty,
			sell_qty as SellQty,
			buy_volume as BuyVolume,
			sell_volume as SellVolume,
			trade_count as TradeCount,
			max_pcs as MaxPcs,
			min_pcs as MinPcs,
			open as Open,
			close as Close,
			high as High,
			low as Low,
			volume as Volume,
			order_book_buy_length as OrderBookBuyLength,
			order_book_sell_length as OrderBookSellLength,
			order_book_buy_qty_sum as OrderBookBuyQtySum,
			order_book_sell_qty_sum as OrderBookSellQtySum,
			order_book_buy_volume_sum as OrderBookBuyVolumeSum,
			order_book_sell_volume_sum as OrderBookSellVolumeSum,
			order_book_buy_iceberg_qty as OrderBookBuyIcebergQty,
			order_book_buy_iceberg_price as OrderBookBuyIcebergPrice,
			order_book_sell_iceberg_qty as OrderBookSellIcebergQty,
			order_book_sell_iceberg_price as OrderBookSellIcebergPrice,
			order_book_buy_first_qty as OrderBookBuyFirstQty,
			order_book_buy_first_price as OrderBookBuyFirstPrice,
			order_book_sell_first_qty as OrderBookSellFirstQty,
			order_book_sell_first_price as OrderBookSellFirstPrice
		FROM default.trades
		WHERE symbol = ?
		ORDER BY timestamp DESC LIMIT ?
	`, symbol, limit)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	for res.Next() {
		tradeStat := model.TradeStat{
			OrderBookStat: model.OrderBookStat{
				SellIceberg: model.Iceberg{
					Side: model.IcebergSideSell,
				},
				BuyIceberg: model.Iceberg{
					Side: model.IcebergSideBuy,
				},
			},
		}
		err := res.Scan(
			&tradeStat.Symbol,
			&tradeStat.Timestamp,
			&tradeStat.BotId,
			&tradeStat.Price,
			&tradeStat.BuyQty,
			&tradeStat.SellQty,
			&tradeStat.BuyVolume,
			&tradeStat.SellVolume,
			&tradeStat.TradeCount,
			&tradeStat.MaxPSC,
			&tradeStat.MinPCS,
			&tradeStat.Open,
			&tradeStat.Close,
			&tradeStat.High,
			&tradeStat.Low,
			&tradeStat.Volume,
			&tradeStat.OrderBookStat.BuyLength,
			&tradeStat.OrderBookStat.SellLength,
			&tradeStat.OrderBookStat.BuyQtySum,
			&tradeStat.OrderBookStat.SellQtySum,
			&tradeStat.OrderBookStat.BuyVolumeSum,
			&tradeStat.OrderBookStat.SellVolumeSum,
			&tradeStat.OrderBookStat.BuyIceberg.Quantity,
			&tradeStat.OrderBookStat.BuyIceberg.Price,
			&tradeStat.OrderBookStat.SellIceberg.Quantity,
			&tradeStat.OrderBookStat.SellIceberg.Price,
			&tradeStat.OrderBookStat.FirstBuyQty,
			&tradeStat.OrderBookStat.FirstBuyPrice,
			&tradeStat.OrderBookStat.FirstSellQty,
			&tradeStat.OrderBookStat.FirstSellPrice,
		)

		if err != nil {
			log.Fatal(err)
		}

		statMap.Store(tradeStat.Timestamp.GetPeriodToMinute(), tradeStat)
	}

	return &statMap
}

func (s *StatRepository) GetMLDataset(symbol string, secondary string) []model.TradeLearnDataset {
	list := make([]model.TradeLearnDataset, 0)

	res, err := s.DB.Query(
		`SELECT
			t1.order_book_buy_first_qty,
			t1.order_book_sell_first_qty,
			t1.order_book_buy_qty_sum,
			t1.order_book_sell_qty_sum,
			t1.order_book_buy_volume_sum,
			t1.order_book_sell_volume_sum,
			t2.close as secondary_price,
			t1.close
		FROM default.trades t1
	 	INNER JOIN default.trades t2 ON t1.timestamp = t2.timestamp AND t2.symbol = ?
		WHERE t1.symbol = ? AND t1.timestamp >= (toStartOfDay(now()) - toIntervalDay(1))
	`, secondary, symbol)
	defer res.Close()

	if err != nil {
		log.Fatal(err)
	}

	for res.Next() {
		var datasetItem model.TradeLearnDataset
		err := res.Scan(
			&datasetItem.OrderBookBuyFirstQty,
			&datasetItem.OrderBookSellFirstQty,
			&datasetItem.OrderBookBuyQtySum,
			&datasetItem.OrderBookSellQtySum,
			&datasetItem.OrderBookBuyVolumeSum,
			&datasetItem.OrderBookSellVolumeSum,
			&datasetItem.SecondaryPrice,
			&datasetItem.PrimaryPrice,
		)

		if err != nil {
			log.Fatal(err)
		}

		list = append(list, datasetItem)
	}

	return list
}
