package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"log"
	"slices"
	"strconv"
	"time"
)

type ByBit struct {
	CurrentBot *model.Bot
	HttpClient HttpClientInterface
	DSN        string
	ApiKey     string
	ApiSecret  string
	Formatter  *utils.Formatter

	RDB *redis.Client
	Ctx *context.Context

	APIKeyCheckCompleted bool
}

func (b *ByBit) IsConnected() bool {
	return true
}

func (b *ByBit) IsWaitMode() bool {
	return false
}

func (b *ByBit) IsAPIKeyCheckCompleted() bool {
	return b.APIKeyCheckCompleted == true
}

func (b *ByBit) QueryOrder(symbol string, orderId string) (model.BinanceOrder, error) {
	var order model.BinanceOrder
	queryString := fmt.Sprintf("category=spot&limit=1&orderId=%s&symbol=%s&openOnly=0", orderId, symbol)
	url := fmt.Sprintf("%s/v5/order/realtime?%s", b.DSN, queryString)
	result, err := b.HttpClient.Get(url, b.GetHeaders(queryString))

	if err != nil {
		return order, err
	}

	var orderHistoryResponse model.ByBitOrderListResponse
	err = json.Unmarshal(result, &orderHistoryResponse)
	if err != nil {
		log.Printf("[%s] QueryOrder: %s", symbol, err.Error())
		return order, err
	}

	if orderHistoryResponse.Message != "OK" {
		log.Printf("[%s] QueryOrder: %s", symbol, orderHistoryResponse.Message)
		return order, errors.New(orderHistoryResponse.Message)
	}

	for _, byBitOrder := range orderHistoryResponse.Result.List {
		if byBitOrder.OrderId == orderId {
			return b.Formatter.ByBitOrderToBinanceOrder(orderHistoryResponse.Result.List[0]), nil
		}
	}

	return order, errors.New(fmt.Sprintf("[%s] order %s is not found", symbol, orderId))
}

func (b *ByBit) CancelOrder(symbol string, orderId string) (model.BinanceOrder, error) {
	requestBody := map[string]string{
		"category": "spot",
		"symbol":   symbol,
		"orderId":  orderId,
	}
	encoded, err := json.Marshal(requestBody)

	var order model.BinanceOrder

	if err != nil {
		return order, err
	}

	result, err := b.HttpClient.Post(fmt.Sprintf("%s/v5/order/cancel", b.DSN), encoded, b.GetHeaders(string(encoded)))
	if err != nil {
		return order, err
	}

	var byBitResult model.ByBitKeyValueResult
	err = json.Unmarshal(result, &byBitResult)
	if err != nil {
		log.Printf("[%s] CancelOrder: %s", symbol, err.Error())
		return order, err
	}

	if byBitResult.Message != "OK" {
		log.Printf("[%s] CancelOrder: %s", symbol, byBitResult.Message)
		return order, errors.New(byBitResult.Message)
	}

	return b.QueryOrder(symbol, orderId)
}

func (b *ByBit) GetDepth(symbol string, limit int64) *model.OrderBook {
	if limit > 200 {
		limit = 200
	}
	queryString := fmt.Sprintf("category=spot&symbol=%s&limit=%d", symbol, limit)
	result, err := b.HttpClient.Get(fmt.Sprintf(
		"%s/v5/market/orderbook?%s",
		b.DSN,
		queryString,
	), b.GetHeaders(queryString))

	if err != nil {
		return nil
	}
	var orderBookResult model.ByBitOrderBookResponse
	err = json.Unmarshal(result, &orderBookResult)
	if err != nil {
		log.Printf("[%s] GetDepth: %s", symbol, err.Error())
		return nil
	}

	if orderBookResult.Message != "OK" {
		log.Printf("[%s] GetDepth: %s", symbol, orderBookResult.Message)
		return nil
	}

	return &model.OrderBook{
		Bids: orderBookResult.Result.Bids,
		Asks: orderBookResult.Result.Asks,
	}
}

func (b *ByBit) GetOpenedOrders() ([]model.BinanceOrder, error) {
	orders := make([]model.BinanceOrder, 0)
	queryString := "category=spot&openOnly=1"
	result, err := b.HttpClient.Get(fmt.Sprintf("%s/v5/order/realtime?%s", b.DSN, queryString), b.GetHeaders(queryString))
	if err != nil {
		return orders, err
	}

	var openedOrdersResponse model.ByBitOrderListResponse
	err = json.Unmarshal(result, &openedOrdersResponse)
	if err != nil {
		log.Printf("GetOpenedOrders: %s", err.Error())
		return orders, err
	}

	if openedOrdersResponse.Message != "OK" {
		log.Printf("GetOpenedOrders: %s", openedOrdersResponse.Message)
		return orders, errors.New(openedOrdersResponse.Message)
	}

	for _, byBitOrder := range openedOrdersResponse.Result.List {
		order := b.Formatter.ByBitOrderToBinanceOrder(byBitOrder)
		if order.IsNew() || order.IsPartiallyFilled() {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

func (b *ByBit) GetKLines(symbol string, interval string, limit int64) []model.KLineHistory {
	kLines := make([]model.KLineHistory, 0)
	queryString := fmt.Sprintf(
		"category=spot&symbol=%s&interval=%s&limit=%d",
		symbol,
		b.Formatter.BinanceIntervalToByBitInterval(interval),
		limit,
	)
	result, err := b.HttpClient.Get(fmt.Sprintf(
		"%s/v5/market/kline?%s",
		b.DSN,
		queryString,
	), b.GetHeaders(queryString))

	if err != nil {
		return kLines
	}
	var klineHistoryResponse model.ByBitKLineHistoryResponse
	err = json.Unmarshal(result, &klineHistoryResponse)
	if err != nil {
		log.Printf("[%s] GetKLines: %s", symbol, err.Error())
		return kLines
	}

	if klineHistoryResponse.Message != "OK" {
		log.Printf("[%s] GetKLines: %s", symbol, klineHistoryResponse.Message)
		return kLines
	}

	for _, byBitKLine := range klineHistoryResponse.Result.List {
		kLines = append(kLines, b.Formatter.ByBitHistoryKlineToBinanceHistoryKline(byBitKLine))
	}

	// Reverse list (Doc: Sort in reverse by startTime)
	slices.Reverse(kLines)

	return kLines
}

func (b *ByBit) TradesAggregate(symbol string, limit int64, startTime int64, endTime int64) []model.Trade {
	queryString := fmt.Sprintf("category=spot&symbol=%s&limit=%d", symbol, limit)
	trades := make([]model.Trade, 0)
	result, err := b.HttpClient.Get(fmt.Sprintf(
		"%s/v5/market/recent-trade?%s",
		b.DSN,
		queryString,
	), b.GetHeaders(queryString))
	if err != nil {
		return trades
	}
	var tradesHistory model.ByBitTradeHistoryResponse
	err = json.Unmarshal(result, &tradesHistory)
	if err != nil {
		log.Printf("[%s] TradesAggregate: %s", symbol, err.Error())
		return trades
	}

	if tradesHistory.Message != "OK" {
		log.Printf("[%s] TradesAggregate: %s", symbol, tradesHistory.Message)
		return trades
	}

	for _, byBitTrade := range tradesHistory.Result.List {
		if byBitTrade.Time.Gt(model.TimestampMilli(startTime)) && byBitTrade.Time.Lte(model.TimestampMilli(endTime)) {
			trades = append(trades, b.Formatter.ByBitTradeToBinanceTrade(byBitTrade))
		}
	}

	return trades
}

func (b *ByBit) GetKLinesCached(symbol string, interval string, limit int64) []model.KLine {
	cacheKey := fmt.Sprintf("interval-klines-history-%s-%s-%d-%d", symbol, interval, limit, b.CurrentBot.Id)

	res := b.RDB.Get(*b.Ctx, cacheKey).Val()
	if len(res) > 0 {
		var batch model.KlineBatch

		err := json.Unmarshal([]byte(res), &batch)
		if err == nil {
			return batch.Items
		}
		log.Printf("[%s] kline[%s] history cache invalid", symbol, interval)
	}

	historyKLines := b.GetKLines(symbol, interval, limit)
	kLines := make([]model.KLine, 0)
	for _, historyKLine := range historyKLines {
		kLines = append(kLines, historyKLine.ToKLine(symbol))
	}

	batch := model.KlineBatch{
		Items: kLines,
	}
	encoded, err := json.Marshal(batch)
	if err == nil {
		b.RDB.Set(*b.Ctx, cacheKey, string(encoded), time.Second*15)
	}

	return batch.Items
}

func (b *ByBit) GetExchangeData(symbols []string) (*model.ExchangeInfo, error) {
	queryString := "category=spot"
	result, err := b.HttpClient.Get(fmt.Sprintf(
		"%s/v5/market/instruments-info?%s",
		b.DSN,
		queryString,
	), b.GetHeaders(queryString))
	if err != nil {
		return nil, err
	}
	var exchangeInfoResponse model.ByBitExchangeInfoResponse
	err = json.Unmarshal(result, &exchangeInfoResponse)
	if err != nil {
		log.Printf("GetExchangeData: %s", err.Error())
		return nil, err
	}

	if exchangeInfoResponse.Message != "OK" {
		log.Printf("GetExchangeData: %s", exchangeInfoResponse.Message)
		return nil, errors.New(exchangeInfoResponse.Message)
	}

	exchangeSymbols := make([]model.ExchangeSymbol, 0)
	for _, byBitSymbol := range exchangeInfoResponse.Result.List {
		if len(symbols) == 0 || slices.Contains(symbols, byBitSymbol.Symbol) {
			exchangeSymbols = append(exchangeSymbols, b.Formatter.ByBitExchangeSymbolToBinanceExchangeSymbol(byBitSymbol))
		}
	}

	return &model.ExchangeInfo{
		Symbols:    exchangeSymbols,
		Timezone:   "UTC",
		RateLimits: make([]model.RateLimit, 0), // unused
		ServerTime: time.Now().UnixMilli(),
	}, nil
}

func (b *ByBit) GetAccountStatus() (*model.AccountStatus, error) {
	accountType := model.ByBitAccountTypeUnified
	queryString := fmt.Sprintf("accountType=%s", accountType)
	result, err := b.HttpClient.Get(fmt.Sprintf(
		"%s/v5/account/wallet-balance?%s",
		b.DSN,
		queryString,
	), b.GetHeaders(queryString))
	if err != nil {
		return nil, err
	}
	var balanceResponse model.ByBitBalanceResponse
	err = json.Unmarshal(result, &balanceResponse)
	if err != nil {
		log.Printf("GetAccountStatus: %s", err.Error())
		return nil, err
	}
	if balanceResponse.Message != "OK" {
		log.Printf("GetAccountStatus: %s", balanceResponse.Message)
		return nil, errors.New(balanceResponse.Message)
	}

	balances := make([]model.Balance, 0)

	for _, byBitBalance := range balanceResponse.Result.List {
		if byBitBalance.AccountType == accountType {
			for _, coin := range byBitBalance.Coin {
				balances = append(balances, model.Balance{
					Asset:  coin.Coin,
					Free:   coin.AvailableToWithdraw,
					Locked: coin.Locked,
				})
			}
		}
	}

	return &model.AccountStatus{
		Balances: balances,
	}, nil
}
func (b *ByBit) GetTickers(symbols []string) []model.WSTickerPrice {
	tickers := make([]model.WSTickerPrice, 0)
	queryString := "category=spot"

	result, err := b.HttpClient.Get(fmt.Sprintf(
		"%s/v5/market/tickers?%s",
		b.DSN,
		queryString,
	), b.GetHeaders(queryString))
	if err != nil {
		return tickers
	}
	var tickerResponse model.ByBitTickerResponse
	err = json.Unmarshal(result, &tickerResponse)
	if err != nil {
		log.Printf("GetTickers: %s", err.Error())
		return tickers
	}
	if tickerResponse.Message != "OK" {
		log.Printf("GetTickers: %s", tickerResponse.Message)
		return tickers
	}

	for _, byBitTicker := range tickerResponse.Result.List {
		if len(symbols) == 0 || slices.Contains(symbols, byBitTicker.Symbol) {
			tickers = append(tickers, b.Formatter.ByBitTickerToBinanceTicker(byBitTicker))
		}
	}

	return tickers
}
func (b *ByBit) LimitOrder(symbol string, quantity float64, price float64, operation string, timeInForce string) (model.BinanceOrder, error) {
	requestBody := map[string]any{
		"category":    "spot",
		"symbol":      symbol,
		"side":        b.Formatter.BinanceSideToByBitSide(operation),
		"orderType":   "Limit",
		"qty":         strconv.FormatFloat(quantity, 'f', -1, 64),
		"price":       strconv.FormatFloat(price, 'f', -1, 64),
		"timeInForce": timeInForce,
		"isLeverage":  0,
		"orderFilter": "Order",
	}
	encoded, err := json.Marshal(requestBody)
	if err != nil {
		return model.BinanceOrder{}, err
	}
	result, err := b.HttpClient.Post(fmt.Sprintf("%s/v5/order/create", b.DSN), encoded, b.GetHeaders(string(encoded)))
	if err != nil {
		return model.BinanceOrder{}, err
	}
	var byBitResult model.ByBitKeyValueResult
	err = json.Unmarshal(result, &byBitResult)
	if err != nil {
		log.Printf("[%s] LimitOrder: %s", symbol, err.Error())
		return model.BinanceOrder{}, err
	}

	if byBitResult.Message != "OK" {
		log.Printf("[%s] LimitOrder: %s", symbol, byBitResult.Message)
		return model.BinanceOrder{}, errors.New(byBitResult.Message)
	}

	orderIdRaw, ok := byBitResult.Result["orderId"]
	if !ok {
		return model.BinanceOrder{}, errors.New("can't get orderId")
	}

	if orderId, ok := orderIdRaw.(string); ok {
		exchangeOrder, err := b.QueryOrder(symbol, orderId)
		if err == nil {
			return exchangeOrder, nil
		}

		return model.BinanceOrder{
			OrderId:             orderId,
			Symbol:              symbol,
			TransactTime:        0,
			Price:               price,
			OrigQty:             quantity,
			ExecutedQty:         0.00,
			CummulativeQuoteQty: price * quantity,
			Status:              model.ExchangeOrderStatusNew,
			Type:                "LIMIT",
			Side:                operation,
			WorkingTime:         0,
			Timestamp:           time.Now().UnixMilli(),
		}, nil
	}

	return model.BinanceOrder{}, errors.New("orderId is not string")
}

func (b *ByBit) GetHeaders(payload string) map[string]string {
	timestamp := time.Now().UnixMilli()
	val := strconv.FormatInt(timestamp, 10) + b.ApiKey
	val = val + payload
	h := hmac.New(sha256.New, []byte(b.ApiSecret))
	h.Write([]byte(val))

	return map[string]string{
		"X-BAPI-API-KEY":   b.ApiKey,
		"X-BAPI-TIMESTAMP": strconv.FormatInt(timestamp, 10),
		"X-BAPI-SIGN":      hex.EncodeToString(h.Sum(nil)),
	}
}
