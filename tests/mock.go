package tests

import (
	"github.com/stretchr/testify/mock"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"sync"
)

type ExchangeRepositoryMock struct {
	mock.Mock
}

func (m *ExchangeRepositoryMock) CreateSwapPair(swapPair model.SwapPair) (*int64, error) {
	args := m.Called(swapPair)
	id := int64(args.Int(1))
	return &id, args.Error(0)
}
func (m *ExchangeRepositoryMock) UpdateSwapPair(swapPair model.SwapPair) error {
	args := m.Called(swapPair)
	return args.Error(0)
}
func (m *ExchangeRepositoryMock) GetSwapPairs() []model.SwapPair {
	args := m.Called()
	return args.Get(0).([]model.SwapPair)
}
func (m *ExchangeRepositoryMock) GetSwapPairsByBaseAsset(baseAsset string) []model.SwapPair {
	args := m.Called(baseAsset)
	return args.Get(0).([]model.SwapPair)
}
func (m *ExchangeRepositoryMock) GetSwapPairsByQuoteAsset(quoteAsset string) []model.SwapPair {
	args := m.Called(quoteAsset)
	return args.Get(0).([]model.SwapPair)
}
func (m *ExchangeRepositoryMock) GetSwapPair(symbol string) (model.SwapPair, error) {
	args := m.Called(symbol)
	return args.Get(0).(model.SwapPair), args.Error(1)
}

type SwapRepositoryMock struct {
	mock.Mock
	savedChain model.SwapChainEntity
	swapAction model.SwapAction
}

func (m *SwapRepositoryMock) GetSwapChain(hash string) (model.SwapChainEntity, error) {
	args := m.Called(hash)
	return args.Get(0).(model.SwapChainEntity), args.Error(1)
}
func (m *SwapRepositoryMock) CreateSwapChain(swapChain model.SwapChainEntity) (*int64, error) {
	m.savedChain = swapChain
	args := m.Called(swapChain)
	id := int64(args.Int(0))
	return &id, args.Error(1)
}
func (m *SwapRepositoryMock) UpdateSwapChain(swapChain model.SwapChainEntity) error {
	args := m.Called(swapChain)
	return args.Error(0)
}
func (m *SwapRepositoryMock) SaveSwapChainCache(asset string, entity model.SwapChainEntity) {
	m.Called(asset, entity)
}
func (m *SwapRepositoryMock) GetSwapPairBySymbol(symbol string) (model.SwapPair, error) {
	args := m.Called(symbol)
	return args.Get(0).(model.SwapPair), args.Error(1)
}
func (s *SwapRepositoryMock) GetActiveSwapAction(order model.Order) (model.SwapAction, error) {
	args := s.Called(order)
	return args.Get(0).(model.SwapAction), args.Error(1)
}
func (s *SwapRepositoryMock) UpdateSwapAction(action model.SwapAction) error {
	s.swapAction = action
	args := s.Called(action)
	return args.Error(0)
}
func (s *SwapRepositoryMock) GetSwapChainById(id int64) (model.SwapChainEntity, error) {
	args := s.Called(id)
	return args.Get(0).(model.SwapChainEntity), args.Error(1)
}
func (s *SwapRepositoryMock) InvalidateSwapChainCache(asset string) {
	_ = s.Called(asset)
}
func (s *SwapRepositoryMock) GetSwapChainCache(asset string) *model.SwapChainEntity {
	args := s.Called(asset)
	entity := args.Get(0)
	if nil == entity {
		return nil
	}
	return entity.(*model.SwapChainEntity)
}
func (s *SwapRepositoryMock) GetSwapChains(baseAsset string) []model.SwapChainEntity {
	args := s.Called(baseAsset)
	return args.Get(0).([]model.SwapChainEntity)
}
func (s *SwapRepositoryMock) CreateSwapAction(action model.SwapAction) (*int64, error) {
	s.swapAction = action
	args := s.Called(action)
	return args.Get(0).(*int64), args.Error(1)
}

type OrderUpdaterMock struct {
	mock.Mock
	order model.Order
}

func (s *OrderUpdaterMock) Update(order model.Order) error {
	s.order = order
	args := s.Called(order)
	return args.Error(0)
}

type BalanceServiceMock struct {
	mock.Mock
}

func (b *BalanceServiceMock) GetAssetBalance(asset string, cache bool) (float64, error) {
	args := b.Called(asset, cache)
	return args.Get(0).(float64), args.Error(1)
}
func (b *BalanceServiceMock) InvalidateBalanceCache(asset string) {
	_ = b.Called(asset)
}

type ExchangeOrderAPIMock struct {
	mock.Mock
}

func (b *ExchangeOrderAPIMock) GetOpenedOrders() ([]model.BinanceOrder, error) {
	args := b.Called()
	return args.Get(0).([]model.BinanceOrder), args.Error(1)
}
func (b *ExchangeOrderAPIMock) LimitOrder(symbol string, quantity float64, price float64, operation string, timeInForce string) (model.BinanceOrder, error) {
	args := b.Called(symbol, quantity, price, operation, timeInForce)
	return args.Get(0).(model.BinanceOrder), args.Error(1)
}
func (b *ExchangeOrderAPIMock) QueryOrder(symbol string, orderId int64) (model.BinanceOrder, error) {
	args := b.Called(symbol, orderId)
	return args.Get(0).(model.BinanceOrder), args.Error(1)
}
func (b *ExchangeOrderAPIMock) CancelOrder(symbol string, orderId int64) (model.BinanceOrder, error) {
	args := b.Called(symbol, orderId)
	return args.Get(0).(model.BinanceOrder), args.Error(1)
}

type TimeServiceMock struct {
	mock.Mock
}

func (t *TimeServiceMock) WaitSeconds(seconds int64) {
	_ = t.Called(seconds)
}
func (t *TimeServiceMock) WaitMilliseconds(milliseconds int64) {
	_ = t.Called(milliseconds)
}
func (t *TimeServiceMock) GetNowDiffMinutes(unixTime int64) float64 {
	args := t.Called(unixTime)
	return args.Get(0).(float64)
}
func (t *TimeServiceMock) GetNowDateTimeString() string {
	args := t.Called()
	return args.String(0)
}
func (t *TimeServiceMock) GetNowUnix() int64 {
	args := t.Called()
	return int64(args.Int(0))
}

type ExchangePriceStorageMock struct {
	mock.Mock
}

func (e *ExchangePriceStorageMock) GetCurrentKline(symbol string) *model.KLine {
	args := e.Called(symbol)
	kLine := args.Get(0)

	if kLine != nil {
		return kLine.(*model.KLine)
	}

	return nil
}
func (e *ExchangePriceStorageMock) GetSwapPairsByBaseAsset(baseAsset string) []model.SwapPair {
	args := e.Called(baseAsset)
	return args.Get(0).([]model.SwapPair)
}
func (e *ExchangePriceStorageMock) GetSwapPairsByQuoteAsset(quoteAsset string) []model.SwapPair {
	args := e.Called(quoteAsset)
	return args.Get(0).([]model.SwapPair)
}
func (e *ExchangePriceStorageMock) GetSwapPairsByAssets(quoteAsset string, baseAsset string) (model.SwapPair, error) {
	args := e.Called(quoteAsset, baseAsset)
	return args.Get(0).(model.SwapPair), args.Error(1)
}
func (e *ExchangePriceStorageMock) GetPeriodMinPrice(symbol string, period int64) float64 {
	args := e.Called(symbol, period)
	return args.Get(0).(float64)
}
func (e *ExchangePriceStorageMock) GetDepth(symbol string, limit int64) model.OrderBookModel {
	args := e.Called(symbol, limit)
	return args.Get(0).(model.OrderBookModel)
}
func (e *ExchangePriceStorageMock) SetDepth(depth model.OrderBookModel, limit int64, expires int64) {
	_ = e.Called(depth, limit, expires)
}
func (e *ExchangePriceStorageMock) GetPredict(symbol string) (float64, error) {
	args := e.Called(symbol)
	return args.Get(0).(float64), args.Error(1)
}

type OrderCachedReaderMock struct {
	mock.Mock
}

func (o *OrderCachedReaderMock) GetOpenedOrderCached(symbol string, operation string) *model.Order {
	args := o.Called(symbol, operation)
	order := args.Get(0)

	if order == nil {
		return nil
	}

	return order.(*model.Order)
}

type FrameServiceMock struct {
	mock.Mock
}

func (f *FrameServiceMock) GetFrame(symbol string, interval string, limit int64) model.Frame {
	args := f.Called(symbol, interval, limit)
	return args.Get(0).(model.Frame)
}

type ExchangePriceAPIMock struct {
	mock.Mock
}

func (e *ExchangePriceAPIMock) GetOpenedOrders() ([]model.BinanceOrder, error) {
	args := e.Called()
	return args.Get(0).([]model.BinanceOrder), args.Error(1)
}
func (e *ExchangePriceAPIMock) GetDepth(symbol string, limit int64) *model.OrderBook {
	args := e.Called(symbol, limit)

	val := args.Get(0)
	if val == nil {
		return nil
	}

	return val.(*model.OrderBook)
}
func (e *ExchangePriceAPIMock) GetKLines(symbol string, interval string, limit int64) []model.KLineHistory {
	args := e.Called(symbol, interval, limit)
	return args.Get(0).([]model.KLineHistory)
}
func (e *ExchangePriceAPIMock) GetKLinesCached(symbol string, interval string, limit int64) []model.KLine {
	args := e.Called(symbol, interval, limit)
	return args.Get(0).([]model.KLine)
}
func (e *ExchangePriceAPIMock) GetExchangeData(symbols []string) (*model.ExchangeInfo, error) {
	args := e.Called(symbols)
	return args.Get(0).(*model.ExchangeInfo), args.Error(1)
}

type OrderStorageMock struct {
	mock.Mock
	Created model.Order
	Updated model.Order
}

func (e *OrderStorageMock) Create(order model.Order) (*int64, error) {
	e.Created = order
	args := e.Called(order)
	return args.Get(0).(*int64), args.Error(1)
}
func (e *OrderStorageMock) Update(order model.Order) error {
	e.Updated = order
	args := e.Called(order)
	return args.Error(0)
}
func (e *OrderStorageMock) DeleteManualOrder(symbol string) {
	_ = e.Called(symbol)
}
func (e *OrderStorageMock) Find(id int64) (model.Order, error) {
	args := e.Called(id)
	return args.Get(0).(model.Order), args.Error(1)
}
func (e *OrderStorageMock) GetClosesOrderList(buyOrder model.Order) []model.Order {
	args := e.Called(buyOrder)
	return args.Get(0).([]model.Order)
}
func (e *OrderStorageMock) DeleteBinanceOrder(order model.BinanceOrder) {
	_ = e.Called(order)
}
func (e *OrderStorageMock) GetTodayExtraOrderMap() *sync.Map {
	args := e.Called()
	return args.Get(0).(*sync.Map)
}
func (e *OrderStorageMock) GetOpenedOrderCached(symbol string, operation string) *model.Order {
	args := e.Called(symbol, operation)
	order := args.Get(0)

	if order == nil {
		return nil
	}

	return order.(*model.Order)
}
func (e *OrderStorageMock) GetManualOrder(symbol string) *model.ManualOrder {
	args := e.Called(symbol)
	manual := args.Get(0)

	if nil == manual {
		return nil
	}

	return manual.(*model.ManualOrder)
}
func (e *OrderStorageMock) SetBinanceOrder(order model.BinanceOrder) {
	_ = e.Called(order)
}
func (e *OrderStorageMock) GetBinanceOrder(symbol string, operation string) *model.BinanceOrder {
	args := e.Called(symbol, operation)

	order := args.Get(0)

	if nil == order {
		return nil
	}

	return order.(*model.BinanceOrder)
}
func (e *OrderStorageMock) LockBuy(symbol string, seconds int64) {
	_ = e.Called(symbol, seconds)
}
func (e *OrderStorageMock) HasBuyLock(symbol string) bool {
	args := e.Called(symbol)
	return args.Get(0).(bool)
}

type ExchangeTradeInfoMock struct {
	mock.Mock
}

func (e *ExchangeTradeInfoMock) GetCurrentKline(symbol string) *model.KLine {
	args := e.Called(symbol)
	kLine := args.Get(0)
	if kLine != nil {
		return kLine.(*model.KLine)
	}

	return nil
}
func (e *ExchangeTradeInfoMock) GetTradeLimit(symbol string) (model.TradeLimit, error) {
	args := e.Called(symbol)
	return args.Get(0).(model.TradeLimit), args.Error(1)
}
func (e *ExchangeTradeInfoMock) GetPeriodMinPrice(symbol string, period int64) float64 {
	args := e.Called(symbol, period)
	return args.Get(0).(float64)
}
func (e *ExchangeTradeInfoMock) GetPredict(symbol string) (float64, error) {
	args := e.Called(symbol)
	return args.Get(0).(float64), args.Error(1)
}
func (e *ExchangeTradeInfoMock) GetInterpolation(kLine model.KLine) (model.Interpolation, error) {
	args := e.Called(kLine)
	return args.Get(0).(model.Interpolation), args.Error(1)
}

type PriceCalculatorMock struct {
	mock.Mock
}

func (p *PriceCalculatorMock) CalculateBuy(tradeLimit model.TradeLimit) (float64, error) {
	args := p.Called(tradeLimit)
	return args.Get(0).(float64), args.Error(1)
}
func (p *PriceCalculatorMock) CalculateSell(tradeLimit model.TradeLimit, order model.Order) (float64, error) {
	args := p.Called(tradeLimit, order)
	return args.Get(0).(float64), args.Error(1)
}
func (p *PriceCalculatorMock) GetDepth(symbol string, limit int64) model.OrderBookModel {
	args := p.Called(symbol, limit)
	return args.Get(0).(model.OrderBookModel)
}

type SwapExecutorMock struct {
	mock.Mock
}

func (p *SwapExecutorMock) Execute(order model.Order) {
	_ = p.Called(order)
}

type SwapValidatorMock struct {
	mock.Mock
}

func (s *SwapValidatorMock) Validate(entity model.SwapChainEntity, order model.Order) error {
	args := s.Called(entity, order)
	return args.Error(0)
}
func (s *SwapValidatorMock) CalculatePercent(entity model.SwapChainEntity) model.Percent {
	args := s.Called(entity)
	return args.Get(0).(model.Percent)
}

type TelegramNotificatorMock struct {
	mock.Mock
}

func (s *TelegramNotificatorMock) Error(bot model.Bot, code string, message string, stop bool) {
	_ = s.Called(bot, code, message, stop)
}
func (s *TelegramNotificatorMock) SellOrder(order model.Order, bot model.Bot, details string) {
	_ = s.Called(order, bot, details)
}
func (s *TelegramNotificatorMock) BuyOrder(order model.Order, bot model.Bot, details string) {
	_ = s.Called(order, bot, details)
}

type LossSecurityMock struct {
	mock.Mock
}

func (l *LossSecurityMock) IsRiskyBuy(binanceOrder model.BinanceOrder, limit model.TradeLimit) bool {
	args := l.Called(binanceOrder, limit)
	return args.Get(0).(bool)
}
func (l *LossSecurityMock) BuyPriceCorrection(price float64, limit model.TradeLimit) float64 {
	args := l.Called(price, limit)
	return args.Get(0).(float64)
}
func (l *LossSecurityMock) CheckBuyPriceOnHistory(limit model.TradeLimit, buyPrice float64) float64 {
	args := l.Called(limit, buyPrice)
	return args.Get(0).(float64)
}

type ProfitServiceMock struct {
	mock.Mock
}

func (p *ProfitServiceMock) CheckBuyPriceOnHistory(limit model.TradeLimit, buyPrice float64) float64 {
	args := p.Called(limit, buyPrice)
	return args.Get(0).(float64)
}
func (p *ProfitServiceMock) GetMinClosePrice(order model.ProfitPositionInterface, currentPrice float64) float64 {
	args := p.Called(order, currentPrice)
	return args.Get(0).(float64)
}
func (p *ProfitServiceMock) GetMinProfitPercent(order model.ProfitPositionInterface) model.Percent {
	args := p.Called(order)
	return args.Get(0).(model.Percent)
}

type BotServiceMock struct {
	mock.Mock
}

func (b *BotServiceMock) GetBot() model.Bot {
	args := b.Called()
	return args.Get(0).(model.Bot)
}
func (b *BotServiceMock) IsSwapEnabled() bool {
	args := b.Called()
	return args.Get(0).(bool)
}
func (b *BotServiceMock) IsMasterBot() bool {
	args := b.Called()
	return args.Get(0).(bool)
}
func (b *BotServiceMock) GetTradeStackSorting() string {
	args := b.Called()
	return args.Get(0).(string)
}
func (b *BotServiceMock) UseSwapCapital() bool {
	args := b.Called()
	return args.Get(0).(bool)
}
func (b *BotServiceMock) GetSwapConfig() model.SwapConfig {
	args := b.Called()
	return args.Get(0).(model.SwapConfig)
}

type BuyOrderStackMock struct {
	mock.Mock
}

func (b *BuyOrderStackMock) CanBuy(limit model.TradeLimit) bool {
	args := b.Called(limit)
	return args.Get(0).(bool)
}

type DecisionReadStorageMock struct {
	mock.Mock
}

func (d *DecisionReadStorageMock) GetDecisions(symbol string) []model.Decision {
	args := d.Called(symbol)
	return args.Get(0).([]model.Decision)
}

type BaseTradeStorageMock struct {
	mock.Mock
}

func (e *BaseTradeStorageMock) GetCurrentKline(symbol string) *model.KLine {
	args := e.Called(symbol)
	kLine := args.Get(0)
	if kLine != nil {
		return kLine.(*model.KLine)
	}

	return nil
}
func (e *BaseTradeStorageMock) GetTradeLimits() []model.TradeLimit {
	args := e.Called()
	return args.Get(0).([]model.TradeLimit)
}
func (e *BaseTradeStorageMock) CreateSwapPair(swapPair model.SwapPair) (*int64, error) {
	args := e.Called(swapPair)
	id := int64(args.Int(1))
	return &id, args.Error(0)
}
func (e *BaseTradeStorageMock) GetSwapPair(symbol string) (model.SwapPair, error) {
	args := e.Called(symbol)
	return args.Get(0).(model.SwapPair), args.Error(1)
}
func (e *BaseTradeStorageMock) GetTradeLimit(symbol string) (model.TradeLimit, error) {
	args := e.Called(symbol)
	return args.Get(0).(model.TradeLimit), args.Error(1)
}
func (e *BaseTradeStorageMock) UpdateSwapPair(swapPair model.SwapPair) error {
	args := e.Called(swapPair)
	return args.Error(0)
}
func (e *BaseTradeStorageMock) UpdateTradeLimit(limit model.TradeLimit) error {
	args := e.Called(limit)
	return args.Error(0)
}

type StrategyFacadeMock struct {
	mock.Mock
}

func (s *StrategyFacadeMock) Decide(symbol string) (model.FacadeResponse, error) {
	args := s.Called(symbol)
	return args.Get(0).(model.FacadeResponse), args.Error(1)
}

type OrderExecutorMock struct {
	mock.Mock
}

func (o *OrderExecutorMock) BuyExtra(tradeLimit model.TradeLimit, order model.Order, price float64) error {
	args := o.Called(tradeLimit, order, price)
	return args.Error(0)
}
func (o *OrderExecutorMock) Buy(tradeLimit model.TradeLimit, price float64, quantity float64) error {
	args := o.Called(tradeLimit, price, quantity)
	return args.Error(0)
}
func (o *OrderExecutorMock) Sell(tradeLimit model.TradeLimit, opened model.Order, price float64, quantity float64, isManual bool) error {
	args := o.Called(tradeLimit, opened, price, quantity, isManual)
	return args.Error(0)
}
func (o *OrderExecutorMock) ProcessSwap(order model.Order) bool {
	args := o.Called(order)
	return args.Get(0).(bool)
}
func (o *OrderExecutorMock) TrySwap(order model.Order) {
	_ = o.Called(order)
}
func (o *OrderExecutorMock) CheckMinBalance(limit model.TradeLimit, kLine model.KLine) error {
	args := o.Called(limit, kLine)
	return args.Error(0)
}
func (o *OrderExecutorMock) CalculateSellQuantity(order model.Order) float64 {
	args := o.Called(order)
	return args.Get(0).(float64)
}

type TradeFilterServiceMock struct {
	mock.Mock
}

func (o *TradeFilterServiceMock) CanBuy(limit model.TradeLimit) bool {
	args := o.Called(limit)
	return args.Get(0).(bool)
}
func (o *TradeFilterServiceMock) CanExtraBuy(limit model.TradeLimit) bool {
	args := o.Called(limit)
	return args.Get(0).(bool)
}
func (o *TradeFilterServiceMock) CanSell(limit model.TradeLimit) bool {
	args := o.Called(limit)
	return args.Get(0).(bool)
}
