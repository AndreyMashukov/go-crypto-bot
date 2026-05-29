package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/tickstore"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/utils"
)

type DecisionReadStorageInterface interface {
	GetDecisions(symbol string) []model.Decision
}

type ExchangeTradeInfoInterface interface {
	GetCurrentKline(symbol string) *model.KLine
	GetTradeLimit(symbol string) (model.TradeLimit, error)
	GetPeriodMinPrice(symbol string, period int64) float64
	GetPredict(symbol string) (float64, error)
	GetInterpolation(kLine model.KLine) (model.Interpolation, error)
	GetTradeLimitCached(symbol string) *model.TradeLimit
}

type BaseTradeStorageInterface interface {
	GetCurrentKline(symbol string) *model.KLine
	GetTradeLimits() []model.TradeLimit
	GetTradeLimit(symbol string) (model.TradeLimit, error)
	UpdateTradeLimit(limit model.TradeLimit) error
}

type ExchangeRepositoryInterface interface {
	GetSubscribedSymbols() []model.Symbol
	GetTradeLimits() []model.TradeLimit
	GetTradeLimit(symbol string) (model.TradeLimit, error)
	CreateTradeLimit(limit model.TradeLimit) (*int64, error)
	UpdateTradeLimit(limit model.TradeLimit) error
	GetCurrentKline(symbol string) *model.KLine
	GetPeriodMinPrice(symbol string, period int64) float64
	GetDepth(symbol string, limit int64) model.OrderBookModel
	SetDepth(depth model.OrderBookModel, limit int64, expires int64)
	AddTrade(trade model.Trade)
	TradeList(symbol string) []model.Trade
	SetDecision(decision model.Decision, symbol string)
	GetDecision(strategy string, symbol string) *model.Decision
	GetDecisions(symbol string) []model.Decision
	GetInterpolation(kLine model.KLine) (model.Interpolation, error)
	GetCapitalization(symbol string, timestamp model.TimestampMilli) *model.MCObject
	GetPredict(symbol string) (float64, error)
}

type ExchangePriceStorageInterface interface {
	GetCurrentKline(symbol string) *model.KLine
	GetPeriodMinPrice(symbol string, period int64) float64
	GetDepth(symbol string, limit int64) model.OrderBookModel
	SetDepth(depth model.OrderBookModel, limit int64, expires int64)
	GetPredict(symbol string) (float64, error)
}

type ExchangeRepository struct {
	DB               *sql.DB
	RDB              *redis.Client
	Ctx              *context.Context
	CurrentBot       *model.Bot
	Formatter        *utils.Formatter
	Binance          client.ExchangePriceAPIInterface
	ObjectRepository *ObjectRepository
	// TickStore is the source of truth for "current price view" after
	// Phase E. GetCurrentKline derives a model.KLine on demand from the
	// latest tick; there is no Redis kline cache any more.
	TickStore tickstore.Store
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
		    tl.min_notional as MinNotional,
		    tl.is_enabled as IsEnabled,
		    tl.min_price_minutes_period as MinPriceMinutesPeriod,
		    tl.frame_interval as FrameInterval,
		    tl.frame_period as FramePeriod,
		    tl.buy_price_history_check_interval as BuyPriceHistoryCheckInterval,
		    tl.buy_price_history_check_period as BuyPriceHistoryCheckPeriod,
		    tl.extra_charge_options as ExtraChargeOptions,
		    tl.profit_options as ProfitOptions,
		    tl.trade_filters_buy as TradeFiltersBuy,
		    tl.trade_filters_sell as TradeFiltersSell,
		    tl.trade_filters_extra_charge as TradeFiltersExtraCharge,
		    tl.sentiment_label as SentimentLabel,
		    tl.sentiment_score as SentimentScore
		FROM trade_limit tl WHERE tl.bot_id = $1
	`, e.CurrentBot.Id)
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
			&tradeLimit.MinNotional,
			&tradeLimit.IsEnabled,
			&tradeLimit.MinPriceMinutesPeriod,
			&tradeLimit.FrameInterval,
			&tradeLimit.FramePeriod,
			&tradeLimit.BuyPriceHistoryCheckInterval,
			&tradeLimit.BuyPriceHistoryCheckPeriod,
			&tradeLimit.ExtraChargeOptions,
			&tradeLimit.ProfitOptions,
			&tradeLimit.TradeFiltersBuy,
			&tradeLimit.TradeFiltersSell,
			&tradeLimit.TradeFiltersExtraCharge,
			&tradeLimit.SentimentLabel,
			&tradeLimit.SentimentScore,
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
		    tl.min_notional as MinNotional,
		    tl.is_enabled as IsEnabled,
		    tl.min_price_minutes_period as MinPriceMinutesPeriod,
		    tl.frame_interval as FrameInterval,
		    tl.frame_period as FramePeriod,
		    tl.buy_price_history_check_interval as BuyPriceHistoryCheckInterval,
		    tl.buy_price_history_check_period as BuyPriceHistoryCheckPeriod,
		    tl.extra_charge_options as ExtraChargeOptions,
		    tl.profit_options as ProfitOptions,
		    tl.trade_filters_buy as TradeFiltersBuy,
		    tl.trade_filters_sell as TradeFiltersSell,
		    tl.trade_filters_extra_charge as TradeFiltersExtraCharge,
		    tl.sentiment_label as SentimentLabel,
		    tl.sentiment_score as SentimentScore
		FROM trade_limit tl
		WHERE tl.symbol = $1 AND tl.bot_id = $2
	`,
		symbol,
		e.CurrentBot.Id,
	).Scan(
		&tradeLimit.Id,
		&tradeLimit.Symbol,
		&tradeLimit.USDTLimit,
		&tradeLimit.MinPrice,
		&tradeLimit.MinQuantity,
		&tradeLimit.MinNotional,
		&tradeLimit.IsEnabled,
		&tradeLimit.MinPriceMinutesPeriod,
		&tradeLimit.FrameInterval,
		&tradeLimit.FramePeriod,
		&tradeLimit.BuyPriceHistoryCheckInterval,
		&tradeLimit.BuyPriceHistoryCheckPeriod,
		&tradeLimit.ExtraChargeOptions,
		&tradeLimit.ProfitOptions,
		&tradeLimit.TradeFiltersBuy,
		&tradeLimit.TradeFiltersSell,
		&tradeLimit.TradeFiltersExtraCharge,
		&tradeLimit.SentimentLabel,
		&tradeLimit.SentimentScore,
	)
	if err != nil {
		return tradeLimit, err
	}

	return tradeLimit, nil
}

func (e *ExchangeRepository) CreateTradeLimit(limit model.TradeLimit) (*int64, error) {
	var lastID int64
	err := e.DB.QueryRow(`
		INSERT INTO trade_limit (
			symbol, usdt_limit, min_price, min_quantity, min_notional,
			is_enabled, min_price_minutes_period, frame_interval, frame_period,
			buy_price_history_check_interval, buy_price_history_check_period,
			extra_charge_options, profit_options,
			trade_filters_buy, trade_filters_sell, trade_filters_extra_charge,
			sentiment_label, sentiment_score,
			bot_id
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11,
			$12, $13,
			$14, $15, $16,
			$17, $18,
			$19
		)
		RETURNING id
	`,
		limit.Symbol,
		limit.USDTLimit,
		limit.MinPrice,
		limit.MinQuantity,
		limit.MinNotional,
		limit.IsEnabled,
		limit.MinPriceMinutesPeriod,
		limit.FrameInterval,
		limit.FramePeriod,
		limit.BuyPriceHistoryCheckInterval,
		limit.BuyPriceHistoryCheckPeriod,
		limit.ExtraChargeOptions,
		limit.ProfitOptions,
		limit.TradeFiltersBuy,
		limit.TradeFiltersSell,
		limit.TradeFiltersExtraCharge,
		limit.SentimentLabel,
		limit.SentimentScore,
		e.CurrentBot.Id,
	).Scan(&lastID)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &lastID, nil
}

func (e *ExchangeRepository) UpdateTradeLimit(limit model.TradeLimit) error {
	_, err := e.DB.Exec(`
		UPDATE trade_limit SET
			symbol                            = $1,
			usdt_limit                        = $2,
			min_price                         = $3,
			min_quantity                      = $4,
			min_notional                      = $5,
			is_enabled                        = $6,
			min_price_minutes_period          = $7,
			frame_interval                    = $8,
			frame_period                      = $9,
			buy_price_history_check_interval  = $10,
			buy_price_history_check_period    = $11,
			extra_charge_options              = $12,
			profit_options                    = $13,
			trade_filters_buy                 = $14,
			trade_filters_sell                = $15,
			trade_filters_extra_charge        = $16,
			sentiment_label                   = $17,
			sentiment_score                   = $18
		WHERE id = $19
	`,
		limit.Symbol,
		limit.USDTLimit,
		limit.MinPrice,
		limit.MinQuantity,
		limit.MinNotional,
		limit.IsEnabled,
		limit.MinPriceMinutesPeriod,
		limit.FrameInterval,
		limit.FramePeriod,
		limit.BuyPriceHistoryCheckInterval,
		limit.BuyPriceHistoryCheckPeriod,
		limit.ExtraChargeOptions,
		limit.ProfitOptions,
		limit.TradeFiltersBuy,
		limit.TradeFiltersSell,
		limit.TradeFiltersExtraCharge,
		limit.SentimentLabel,
		limit.SentimentScore,
		limit.Id,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// GetCurrentKline is the Phase E shim. The legacy Redis kline cache is
// gone; we synthesise a model.KLine from the latest tick the watcher
// produced. Strategy callers that still take a *model.KLine on their
// hot path continue to compile and read "current price view" without
// touching Redis; Phase H will move them to MarketTick.Candles
// directly when the legacy src/ tree is cannibalised.
func (e *ExchangeRepository) GetCurrentKline(symbol string) *model.KLine {
	if e.TickStore == nil {
		return nil
	}
	tick, ok := e.TickStore.Latest(symbol)
	if !ok {
		return nil
	}
	closePrice, _ := tick.Price.Float64()
	open := closePrice
	high := closePrice
	low := closePrice
	volume := 0.0
	if n := len(tick.Candles.Series); n > 0 {
		c := tick.Candles.Series[n-1]
		open, _ = c.Open.Float64()
		high, _ = c.High.Float64()
		low, _ = c.Low.Float64()
		volume, _ = c.Volume.Float64()
	}
	return &model.KLine{
		Symbol:    symbol,
		Open:      model.Price(open),
		High:      model.Price(high),
		Low:       model.Price(low),
		Close:     model.Price(closePrice),
		Volume:    model.Volume(volume),
		Timestamp: model.TimestampMilli(tick.EventTime.UnixMilli()),
		OpenTime:  model.TimestampMilli(tick.EventTime.Truncate(time.Minute).UnixMilli()),
		UpdatedAt: tick.EventTime.Unix(),
		Source:    "tick",
		Interval:  "1m",
	}
}

// GetPeriodMinPrice walks the tick's Candles snapshot and returns the
// minimum Low across them. Period names how many candles back to look;
// the function caps at the available window size and returns 0 when
// no candles are buffered.
func (e *ExchangeRepository) GetPeriodMinPrice(symbol string, period int64) float64 {
	if e.TickStore == nil {
		return 0.00
	}
	tick, ok := e.TickStore.Latest(symbol)
	if !ok {
		return 0.00
	}
	series := tick.Candles.Series
	if int64(len(series)) > period {
		series = series[int64(len(series))-period:]
	}
	minPrice := 0.00
	for i := range series {
		low, _ := series[i].Low.Float64()
		if minPrice == 0.00 || low < minPrice {
			minPrice = low
		}
	}
	return minPrice
}

func (e *ExchangeRepository) SetDepth(depth model.OrderBookModel, limit int64, expires int64) {
	if len(depth.Asks) == 0 || len(depth.Bids) == 0 {
		res := e.RDB.Get(*e.Ctx, fmt.Sprintf("depth-%s-%d", depth.Symbol, limit)).Val()

		if len(res) > 0 {
			var prevDepth model.OrderBookModel
			err := json.Unmarshal([]byte(res), &prevDepth)
			if err == nil {
				if len(depth.Asks) == 0 && len(prevDepth.Asks) > 0 {
					depth.Asks = prevDepth.Asks
				}
				if len(depth.Bids) == 0 && len(prevDepth.Bids) > 0 {
					depth.Bids = prevDepth.Bids
				}
			} else {
				log.Printf("[%s] SetDepth recover error: %s", depth.Symbol, err.Error())
			}
		}
	}

	encoded, err := json.Marshal(depth)
	if err == nil {
		e.RDB.Set(*e.Ctx, fmt.Sprintf("depth-%s-%d", depth.Symbol, limit), string(encoded), time.Second*time.Duration(expires))
	} else {
		log.Printf("[%s] SetDepth save error: %s", depth.Symbol, err.Error())
	}
}

func (e *ExchangeRepository) GetDepth(symbol string, limit int64) model.OrderBookModel {
	expiresSec := int64(25)

	if limit >= 500 {
		expiresSec = 15
	}

	res := e.RDB.Get(*e.Ctx, fmt.Sprintf("depth-%s-%d", symbol, limit)).Val()
	if len(res) == 0 {
		book := e.Binance.GetDepth(symbol, limit)
		if book != nil {
			depth := book.ToOrderBookModel(symbol)
			depth.UpdatedAt = time.Now().Unix()
			e.SetDepth(depth, limit, expiresSec)

			return depth
		}

		return model.OrderBookModel{
			Asks:      make([][2]model.Number, 0),
			Bids:      make([][2]model.Number, 0),
			Symbol:    symbol,
			Timestamp: time.Now().UnixMilli(),
		}
	}

	var dto model.OrderBookModel
	err := json.Unmarshal([]byte(res), &dto)
	if err != nil {
		log.Printf("[%s] GetDepth error: %s", symbol, err.Error())
		book := e.Binance.GetDepth(symbol, limit)
		if book != nil {
			depth := book.ToOrderBookModel(symbol)
			depth.UpdatedAt = time.Now().Unix()
			e.SetDepth(depth, limit, expiresSec)

			return depth
		}

		return model.OrderBookModel{
			Asks:      make([][2]model.Number, 0),
			Bids:      make([][2]model.Number, 0),
			Symbol:    symbol,
			Timestamp: time.Now().UnixMilli(),
		}
	}

	return dto
}

func (e *ExchangeRepository) GetTradeVolumes(kLine model.KLine) (float64, float64) {
	buyVolume := 0.00
	sellVolume := 0.00

	for _, trade := range e.TradeList(kLine.Symbol) {
		if trade.Timestamp.Value() >= (time.Now().UnixMilli() - 60000) {
			if trade.GetOperation() == "BUY" {
				buyVolume += trade.Price * trade.Quantity
			} else {
				sellVolume += trade.Price * trade.Quantity
			}
			continue
		}

		break
	}

	return buyVolume, sellVolume
}

func (e *ExchangeRepository) UpdateTradeVolume(trade model.Trade) {
	tradeVolume := e.GetTradeVolume(trade.Symbol, trade.Timestamp)

	if tradeVolume == nil {
		tradeVolume = &model.TradeVolume{
			Symbol:     trade.Symbol,
			Timestamp:  trade.Timestamp,
			PeriodFrom: model.TimestampMilli(trade.Timestamp.GetPeriodFromMinute()),
			PeriodTo:   model.TimestampMilli(trade.Timestamp.GetPeriodToMinute()),
			SellQty:    0.00,
			BuyQty:     0.00,
		}
	}

	if trade.IsSell() {
		tradeVolume.SellQty += trade.Quantity
	}
	if trade.IsBuy() {
		tradeVolume.BuyQty += trade.Quantity
	}
	tradeVolume.PeriodFrom = model.TimestampMilli(trade.Timestamp.GetPeriodFromMinute())
	tradeVolume.PeriodTo = model.TimestampMilli(trade.Timestamp.GetPeriodToMinute())
	e.SetTradeVolume(*tradeVolume)
}

func (e *ExchangeRepository) AddTrade(trade model.Trade) {
	tradeCacheKey := fmt.Sprintf("trades-%s-%d", trade.Symbol, e.CurrentBot.Id)

	e.UpdateTradeVolume(trade)

	lastTrades := e.TradeList(trade.Symbol)
	encoded, _ := json.Marshal(trade)

	for _, lastTrade := range lastTrades {
		if lastTrade.AggregateTradeId == trade.AggregateTradeId {
			e.RDB.LPop(*e.Ctx, tradeCacheKey).Val()
		}
	}

	e.RDB.LPush(*e.Ctx, tradeCacheKey, string(encoded))
	e.RDB.LTrim(*e.Ctx, tradeCacheKey, 0, 2000)
}

func (e *ExchangeRepository) TradeList(symbol string) []model.Trade {
	tradeCacheKey := fmt.Sprintf("trades-%s-%d", symbol, e.CurrentBot.Id)
	res := e.RDB.LRange(*e.Ctx, tradeCacheKey, 0, 2000).Val()
	list := make([]model.Trade, 0)

	for _, str := range res {
		var dto model.Trade
		json.Unmarshal([]byte(str), &dto)
		list = append(list, dto)
	}

	return list
}

func (e *ExchangeRepository) SetTradeLimit(limit model.TradeLimit) {
	encoded, _ := json.Marshal(limit)
	e.RDB.Set(*e.Ctx, fmt.Sprintf("trade-limit-%s-bot-%d", limit.Symbol, e.CurrentBot.Id), string(encoded), time.Second*60)
}

func (e *ExchangeRepository) GetTradeLimitCached(symbol string) *model.TradeLimit {
	res := e.RDB.Get(*e.Ctx, fmt.Sprintf("trade-limit-%s-bot-%d", symbol, e.CurrentBot.Id)).Val()
	if len(res) == 0 {
		tradeLimit, err := e.GetTradeLimit(symbol)

		if err == nil {
			e.SetTradeLimit(tradeLimit)

			return &tradeLimit
		}

		return nil
	}

	var dto model.TradeLimit
	json.Unmarshal([]byte(res), &dto)

	return &dto
}

func (e *ExchangeRepository) SetDecision(decision model.Decision, symbol string) {
	encoded, _ := json.Marshal(decision)
	e.RDB.Set(*e.Ctx, fmt.Sprintf("decision-%s-%s-bot-%d", decision.StrategyName, symbol, e.CurrentBot.Id), string(encoded), time.Second*model.PriceValidSeconds*2)
}

func (e *ExchangeRepository) DeleteDecision(strategy string, symbol string) {
	e.RDB.Del(*e.Ctx, fmt.Sprintf("decision-%s-%s-bot-%d", strategy, symbol, e.CurrentBot.Id))
}

func (e *ExchangeRepository) GetDecision(strategy string, symbol string) *model.Decision {
	res := e.RDB.Get(*e.Ctx, fmt.Sprintf("decision-%s-%s-bot-%d", strategy, symbol, e.CurrentBot.Id)).Val()
	if len(res) == 0 {
		return nil
	}

	var dto model.Decision
	err := json.Unmarshal([]byte(res), &dto)

	if err == nil {
		return &dto
	}

	return nil
}

func (e *ExchangeRepository) getPredictedCacheKey(symbol string) string {
	return fmt.Sprintf("predicted-price-%s-%d", symbol, e.CurrentBot.Id)
}

func (e *ExchangeRepository) GetPredict(symbol string) (float64, error) {
	var predictedPrice float64

	predictedPriceCacheKey := e.getPredictedCacheKey(symbol)
	predictedPriceCached := e.RDB.Get(*e.Ctx, predictedPriceCacheKey).Val()

	if len(predictedPriceCached) > 0 {
		err := json.Unmarshal([]byte(predictedPriceCached), &predictedPrice)
		if err == nil {
			return predictedPrice, nil
		}
	}

	return 0.00, errors.New("predict is not found")
}

func (e *ExchangeRepository) SavePredict(predicted float64, symbol string) {
	predictedPriceCacheKey := e.getPredictedCacheKey(symbol)

	encoded, err := json.Marshal(predicted)
	if err == nil {
		e.RDB.Set(*e.Ctx, predictedPriceCacheKey, string(encoded), time.Minute)
	}
}

func (e *ExchangeRepository) GetKLinePredict(kLine model.KLine) (float64, error) {
	var predictedPrice float64

	predictedPriceCacheKey := fmt.Sprintf("%s-%d", e.getPredictedCacheKey(kLine.Symbol), kLine.Timestamp.GetPeriodToMinute())
	predictedPriceCached := e.RDB.Get(*e.Ctx, predictedPriceCacheKey).Val()

	if len(predictedPriceCached) > 0 {
		err := json.Unmarshal([]byte(predictedPriceCached), &predictedPrice)
		if err == nil {
			return predictedPrice, nil
		}
	}

	return 0.00, errors.New("predict is not found")
}

func (e *ExchangeRepository) SaveKLinePredict(predicted float64, kLine model.KLine) {
	predictedPriceCacheKey := fmt.Sprintf("%s-%d", e.getPredictedCacheKey(kLine.Symbol), kLine.Timestamp.GetPeriodToMinute())

	encoded, err := json.Marshal(predicted)
	if err == nil {
		e.RDB.Set(*e.Ctx, predictedPriceCacheKey, string(encoded), time.Minute*600)
	}
}

func (e *ExchangeRepository) getInterpolationCacheKey(symbol string) string {
	return fmt.Sprintf("interpolation-price-%s-%d", symbol, e.CurrentBot.Id)
}
func (e *ExchangeRepository) GetInterpolation(kLine model.KLine) (model.Interpolation, error) {
	var interpolation model.Interpolation

	cacheKey := fmt.Sprintf("%s-%d", e.getInterpolationCacheKey(kLine.Symbol), kLine.Timestamp.GetPeriodToMinute())
	interpolationCached := e.RDB.Get(*e.Ctx, cacheKey).Val()

	if len(interpolationCached) > 0 {
		err := json.Unmarshal([]byte(interpolationCached), &interpolation)
		if err == nil {
			return interpolation, nil
		}
	}

	return model.Interpolation{
		Asset:                strings.ReplaceAll(kLine.Symbol, "USDT", ""),
		EthInterpolationUsdt: 0.00,
		BtcInterpolationUsdt: 0.00,
	}, errors.New("interpolation is not found")
}

func (e *ExchangeRepository) SaveInterpolation(interpolation model.Interpolation, kLine model.KLine) {
	cacheKey := fmt.Sprintf("%s-%d", e.getInterpolationCacheKey(kLine.Symbol), kLine.Timestamp.GetPeriodToMinute())

	encoded, err := json.Marshal(interpolation)
	if err == nil {
		e.RDB.Set(*e.Ctx, cacheKey, string(encoded), time.Minute*600)
	}
}

func (e *ExchangeRepository) GetDecisions(symbol string) []model.Decision {
	currentDecisions := make([]model.Decision, 0)
	smaDecision := e.GetDecision(model.SmaTradeStrategyName, symbol)
	kLineDecision := e.GetDecision(model.BaseKlineStrategyName, symbol)
	marketDepthDecision := e.GetDecision(model.MarketDepthStrategyName, symbol)
	orderBasedDecision := e.GetDecision(model.OrderBasedStrategyName, symbol)

	if smaDecision != nil {
		currentDecisions = append(currentDecisions, *smaDecision)
	}
	if kLineDecision != nil {
		currentDecisions = append(currentDecisions, *kLineDecision)
	}
	if marketDepthDecision != nil {
		currentDecisions = append(currentDecisions, *marketDepthDecision)
	}
	if orderBasedDecision != nil {
		currentDecisions = append(currentDecisions, *orderBasedDecision)
	}

	return currentDecisions
}

func (e *ExchangeRepository) SetTradeVolume(volume model.TradeVolume) {
	encoded, _ := json.Marshal(volume)
	e.RDB.Set(*e.Ctx, fmt.Sprintf("trade-volume-%s-%d-bot-%d", strings.ToUpper(volume.Symbol), volume.Timestamp.GetPeriodToMinute(), e.CurrentBot.Id), string(encoded), time.Minute*400)
}

func (e *ExchangeRepository) GetTradeVolume(symbol string, timestamp model.TimestampMilli) *model.TradeVolume {
	res := e.RDB.Get(*e.Ctx, fmt.Sprintf("trade-volume-%s-%d-bot-%d", strings.ToUpper(symbol), timestamp.GetPeriodToMinute(), e.CurrentBot.Id)).Val()
	if len(res) == 0 {
		return nil
	}

	var dto model.TradeVolume
	err := json.Unmarshal([]byte(res), &dto)

	if err != nil {
		log.Printf("[%s] error during trade volume reading: %s", symbol, err.Error())
		return nil
	}

	return &dto
}

func (e *ExchangeRepository) SetPriceChangeSpeed(speed model.PriceChangeSpeed) {
	encoded, _ := json.Marshal(speed)
	e.RDB.Set(*e.Ctx, fmt.Sprintf("price-change-speed-%s-%d-bot-%d", strings.ToUpper(speed.Symbol), speed.Timestamp.GetPeriodToMinute(), e.CurrentBot.Id), string(encoded), time.Minute*400)
}

func (e *ExchangeRepository) GetPriceChangeSpeed(symbol string, timestamp model.TimestampMilli) *model.PriceChangeSpeed {
	res := e.RDB.Get(*e.Ctx, fmt.Sprintf("price-change-speed-%s-%d-bot-%d", strings.ToUpper(symbol), timestamp.GetPeriodToMinute(), e.CurrentBot.Id)).Val()
	if len(res) == 0 {
		return nil
	}

	var dto model.PriceChangeSpeed
	err := json.Unmarshal([]byte(res), &dto)

	if err != nil {
		log.Printf("[%s] error during PCS reading: %s", symbol, err.Error())
		return nil
	}

	return &dto
}

func (e *ExchangeRepository) SetCapitalization(event model.MCEvent) {
	encoded, _ := json.Marshal(event.Data)
	e.RDB.Set(*e.Ctx, fmt.Sprintf("capitalization-%s-%d", strings.ToUpper(event.Data.Symbol()), event.Timestamp.GetPeriodToMinute()), string(encoded), time.Minute*400)
}

func (e *ExchangeRepository) GetCapitalization(symbol string, timestamp model.TimestampMilli) *model.MCObject {
	res := e.RDB.Get(*e.Ctx, fmt.Sprintf("capitalization-%s-%d", strings.ToUpper(symbol), timestamp.GetPeriodToMinute())).Val()
	if len(res) == 0 {
		return nil
	}

	var dto model.MCObject
	err := json.Unmarshal([]byte(res), &dto)

	if err != nil {
		return nil
	}

	return &dto
}
