package repository

import (
	"database/sql"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
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
