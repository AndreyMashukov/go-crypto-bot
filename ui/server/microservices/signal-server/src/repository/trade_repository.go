package repository

import (
	"database/sql"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/model"
	"log"
	"sync"
)

type TradeRepository struct {
	DB *sql.DB
}

func (t *TradeRepository) GetTradeStat(tradingInterval int64, periodDays int64, minPercent model.Percent, exchange string) *sync.Map {
	result := sync.Map{}

	res, err := t.DB.Query(
		`
		SELECT
			ROUND(last_value(t0.Close), 10) as Close,
			ROUND(min(t0.Low), 10) as Low,
			ROUND(max(t0.High), 10) as High,
			((Close + High + Low) / 3) as Pivot,
			(2 * Pivot - High) as S1,
			(Pivot - (High - Low)) as S2,
			(Low - 2*(High - Pivot)) as S3,
			(2 * Pivot - Low) as R1,
			(Pivot + (High - Low)) as R2,
			(High + 2*(Pivot - Low)) as R3,
			t0.Symbol as Symbol,
			S2 as BuyPrice,
			R3 as SellPrice,
			R2 as HalfSellPrice,
			R1 as SellPriceNegative,
			BuyPrice + (BuyPrice * ExtraChargePercent / 100) as ExtraChargePrice,
			((BuyPrice + ExtraChargePrice) / 2) as AvgPrice,
			Round((SellPrice * 100 / BuyPrice) - 100, 2) as Percent,
			Round((HalfSellPrice * 100 / BuyPrice) - 100, 2) as PercentHalf,
			Round((SellPriceNegative * 100 / BuyPrice) - 100, 2) as PercentNegative,
			Percent * -1 as ExtraChargePercent
		FROM (
			SELECT
				t4.symbol as Symbol,
				ROUND(last_value(t4.close), 10) as Close,
				ROUND(min(t4.low), 10) as Low,
				ROUND(max(t4.high), 10) as High,
				toStartOfInterval(t4.timestamp, INTERVAL ? MINUTE) AS Period,
				t4.symbol as Symbol
			FROM (
				SELECT * FROM default.trades t
				WHERE t.timestamp >= now() - toIntervalDay(?) AND t.exchange = ?
				ORDER BY t.timestamp
			) t4
		GROUP BY t4.symbol, Period
		ORDER BY Period
		) t0 GROUP BY t0.Symbol HAVING PercentNegative >= ? ORDER BY Percent DESC
	`, tradingInterval, periodDays, exchange, minPercent.Value())
	defer res.Close()

	if err != nil {
		log.Fatalf("TradeRepository: %s", err.Error())
	}

	for res.Next() {
		var tradeStatItem model.TradeStat
		err := res.Scan(
			&tradeStatItem.Close,
			&tradeStatItem.Low,
			&tradeStatItem.High,
			&tradeStatItem.Pivot,
			&tradeStatItem.S1,
			&tradeStatItem.S2,
			&tradeStatItem.S3,
			&tradeStatItem.R1,
			&tradeStatItem.R2,
			&tradeStatItem.R3,
			&tradeStatItem.Symbol,
			&tradeStatItem.BuyPrice,
			&tradeStatItem.SellPrice,
			&tradeStatItem.HalfSellPrice,
			&tradeStatItem.SellPriceNegative,
			&tradeStatItem.ExtraChargePrice,
			&tradeStatItem.AvgPrice,
			&tradeStatItem.Percent,
			&tradeStatItem.PercentHalf,
			&tradeStatItem.PercentNegative,
			&tradeStatItem.ExtraChargePercent,
		)
		tradeStatItem.PeriodDays = periodDays
		tradeStatItem.IntervalMinutes = tradingInterval

		if err != nil {
			log.Fatal(err)
		}

		result.Store(tradeStatItem.Symbol, tradeStatItem)
	}

	return &result
}
