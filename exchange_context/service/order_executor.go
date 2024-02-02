package service

import (
	"errors"
	"fmt"
	ExchangeClient "gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"strings"
	"sync"
)

type OrderExecutor struct {
	CurrentBot             *ExchangeModel.Bot
	TimeService            TimeServiceInterface
	BalanceService         BalanceServiceInterface
	Binance                ExchangeClient.ExchangeOrderAPIInterface
	OrderRepository        ExchangeRepository.OrderStorageInterface
	ExchangeRepository     ExchangeRepository.ExchangeTradeInfoInterface
	LossSecurity           LossSecurityInterface
	PriceCalculator        PriceCalculatorInterface
	SwapRepository         ExchangeRepository.SwapBasicRepositoryInterface
	SwapExecutor           SwapExecutorInterface
	SwapValidator          SwapValidatorInterface
	CallbackManager        CallbackManagerInterface
	Formatter              *Formatter
	SwapSellOrderDays      int64
	SwapEnabled            bool
	SwapProfitPercent      float64
	TurboSwapProfitPercent float64
	Lock                   map[string]bool
	TradeLockMutex         sync.RWMutex
	LockChannel            *chan ExchangeModel.Lock
	CancelRequestMap       map[string]bool
}

func (m *OrderExecutor) BuyExtra(tradeLimit ExchangeModel.TradeLimit, order ExchangeModel.Order, price float64) error {
	lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

	if lastKline == nil {
		return errors.New(fmt.Sprintf("[%s] Price is unknown", tradeLimit.Symbol))
	}

	if tradeLimit.GetBuyOnFallPercent(order, *lastKline).Gte(0.00) {
		return errors.New(fmt.Sprintf("[%s] Extra buy is disabled", tradeLimit.Symbol))
	}

	if !order.CanExtraBuy(tradeLimit, *lastKline) {
		return errors.New(fmt.Sprintf("[%s] Not enough budget to buy more", tradeLimit.Symbol))
	}

	profit := order.GetProfitPercent(lastKline.Close)

	if profit.Gt(tradeLimit.GetBuyOnFallPercent(order, *lastKline)) {
		return errors.New(fmt.Sprintf(
			"[%s] Extra buy percent is not reached %.2f of %.2f",
			tradeLimit.Symbol,
			profit,
			tradeLimit.GetBuyOnFallPercent(order, *lastKline).Value()),
		)
	}

	m.acquireLock(order.Symbol)
	defer m.releaseLock(order.Symbol)
	// todo: get buy quantity, buy to all cutlet! check available balance!
	quantity := m.Formatter.FormatQuantity(tradeLimit, order.GetAvailableExtraBudget(tradeLimit, *lastKline)/price)

	if (quantity * price) < tradeLimit.MinNotional {
		return errors.New(fmt.Sprintf("[%s] Extra BUY Notional: %.8f < %.8f", order.Symbol, quantity*price, tradeLimit.MinNotional))
	}

	balanceErr := m.CheckBalance(order.Symbol, price, quantity)

	if balanceErr != nil {
		m.CallbackManager.Error(
			*m.CurrentBot,
			"balance_error",
			balanceErr.Error(),
			false,
		)

		return balanceErr
	}

	var extraOrder = ExchangeModel.Order{
		Symbol:      order.Symbol,
		Quantity:    quantity,
		Price:       price,
		CreatedAt:   m.TimeService.GetNowDateTimeString(),
		SellVolume:  0.00,
		BuyVolume:   0.00,
		SmaValue:    0.00,
		Status:      "closed",
		Operation:   "buy",
		ExternalId:  nil,
		ClosesOrder: &order.Id,
		// todo: add commission???
	}

	balanceBefore, balanceErr := m.BalanceService.GetAssetBalance(order.GetBaseAsset(), true)

	binanceOrder, err := m.tryLimitOrder(extraOrder, "BUY", 120)

	if err != nil {
		m.BalanceService.InvalidateBalanceCache("USDT")
		return err
	}

	executedQty := binanceOrder.GetExecutedQuantity()

	// fill from API
	extraOrder.ExternalId = &binanceOrder.OrderId
	extraOrder.ExecutedQuantity = executedQty
	extraOrder.Price = binanceOrder.Price
	extraOrder.CreatedAt = m.TimeService.GetNowDateTimeString()

	avgPrice := m.getAvgPrice(order, extraOrder)

	_, err = m.OrderRepository.Create(extraOrder)
	if err != nil {
		// remove binance order from cache if we have already had saved in database
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "order_external_id_symbol") {
			m.OrderRepository.DeleteBinanceOrder(binanceOrder)
		}

		return err
	}

	if balanceErr == nil {
		m.UpdateCommission(balanceBefore, extraOrder)
	}

	order.ExecutedQuantity = executedQty + order.ExecutedQuantity
	order.Price = avgPrice
	order.UsedExtraBudget = order.UsedExtraBudget + (extraOrder.Price * executedQty)
	commission := 0.00
	if order.Commission != nil {
		commission = *order.Commission
	}
	extraCommission := 0.00
	if extraOrder.Commission != nil {
		extraCommission = *extraOrder.Commission
	}
	commissionSum := commission + extraCommission
	if commissionSum < 0 {
		commissionSum = 0
	}

	order.Commission = &commissionSum

	err = m.OrderRepository.Update(order)
	_, err = m.OrderRepository.Create(order)
	m.BalanceService.InvalidateBalanceCache("USDT")
	m.BalanceService.InvalidateBalanceCache(order.GetBaseAsset())

	if err != nil {
		return err
	}

	m.OrderRepository.DeleteManualOrder(order.Symbol)

	go func(extraOrder ExchangeModel.Order, tradeLimit ExchangeModel.TradeLimit) {
		m.CallbackManager.BuyOrder(
			extraOrder,
			*m.CurrentBot,
			fmt.Sprintf("Extra Charge! Sell when price will be around: %f USDT", m.PriceCalculator.CalculateSell(tradeLimit, extraOrder)),
		)
	}(extraOrder, tradeLimit)

	return nil
}

func (m *OrderExecutor) Buy(tradeLimit ExchangeModel.TradeLimit, symbol string, price float64, quantity float64) error {
	if m.isTradeLocked(symbol) {
		return errors.New(fmt.Sprintf("Operation Buy is Locked %s", symbol))
	}

	if quantity <= 0.00 {
		return errors.New(fmt.Sprintf("Available quantity is %f", quantity))
	}

	balanceErr := m.CheckBalance(symbol, price, quantity)

	if balanceErr != nil {
		m.CallbackManager.Error(
			*m.CurrentBot,
			"balance_error",
			balanceErr.Error(),
			false,
		)

		return balanceErr
	}

	// to avoid concurrent map writes
	m.acquireLock(symbol)
	defer m.releaseLock(symbol)

	// todo: commission
	// You place an order to buy 10 ETH for 3,452.55 USDT each:
	// Trading fee = 10 ETH * 0.1% = 0.01 ETH

	// todo: check min quantity

	var order = ExchangeModel.Order{
		Symbol:             symbol,
		Quantity:           quantity,
		Price:              price,
		CreatedAt:          m.TimeService.GetNowDateTimeString(),
		SellVolume:         0.00,
		BuyVolume:          0.00,
		SmaValue:           0.00,
		Status:             "opened",
		Operation:          "buy",
		ExternalId:         nil,
		ClosesOrder:        nil,
		ExtraChargeOptions: tradeLimit.ExtraChargeOptions,
		// todo: add commission???
	}

	balanceBefore, balanceErr := m.BalanceService.GetAssetBalance(order.GetBaseAsset(), true)

	binanceOrder, err := m.tryLimitOrder(order, "BUY", 480)

	if err != nil {
		m.BalanceService.InvalidateBalanceCache("USDT")
		return err
	}

	// fill from API
	order.ExternalId = &binanceOrder.OrderId
	order.ExecutedQuantity = binanceOrder.GetExecutedQuantity()
	order.Price = binanceOrder.Price
	order.CreatedAt = m.TimeService.GetNowDateTimeString()

	_, err = m.OrderRepository.Create(order)
	m.BalanceService.InvalidateBalanceCache("USDT")
	m.BalanceService.InvalidateBalanceCache(order.GetBaseAsset())

	if err != nil {
		// remove binance order from cache if we have already had saved in database
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "order_external_id_symbol") {
			m.OrderRepository.DeleteBinanceOrder(binanceOrder)
		}

		log.Printf("Can't create order: %s", order.Symbol)

		return err
	}

	m.OrderRepository.DeleteManualOrder(order.Symbol)

	if balanceErr == nil {
		m.UpdateCommission(balanceBefore, order)
	}

	go func(order ExchangeModel.Order, tradeLimit ExchangeModel.TradeLimit) {
		m.CallbackManager.BuyOrder(
			order,
			*m.CurrentBot,
			fmt.Sprintf("Sell when price will be around: %f USDT", m.PriceCalculator.CalculateSell(tradeLimit, order)),
		)
	}(order, tradeLimit)

	return nil
}

func (m *OrderExecutor) Sell(tradeLimit ExchangeModel.TradeLimit, opened ExchangeModel.Order, symbol string, price float64, quantity float64, isManual bool) error {
	if m.isTradeLocked(symbol) {
		return errors.New(fmt.Sprintf("Operation Sell is Locked %s", symbol))
	}

	m.acquireLock(symbol)
	defer m.releaseLock(symbol)

	// todo: commission
	// Or you place an order to sell 10 ETH for 3,452.55 USDT each:
	// Trading fee = (10 ETH * 3,452.55 USDT) * 0.1% = 34.5255 USDT

	profit := (price - opened.Price) * quantity

	// loose money control
	if opened.Price >= price {
		return errors.New(fmt.Sprintf(
			"[%s] Bad deal, wait for positive profit: %.6f [o:%.6f, c:%.6f]",
			symbol,
			profit,
			opened.Price,
			price,
		))
	}

	minPrice := m.Formatter.FormatPrice(tradeLimit, opened.GetMinClosePrice(tradeLimit))

	if isManual {
		minPrice = m.Formatter.FormatPrice(tradeLimit, opened.GetManualMinClosePrice())
	}

	if price < minPrice {
		return errors.New(fmt.Sprintf(
			"[%s] Minimum profit is not reached, Price %.6f < %.6f",
			symbol,
			price,
			minPrice,
		))
	}

	var order = ExchangeModel.Order{
		Symbol:      symbol,
		Quantity:    quantity,
		Price:       price,
		CreatedAt:   m.TimeService.GetNowDateTimeString(),
		SellVolume:  0.00,
		BuyVolume:   0.00,
		SmaValue:    0.00,
		Status:      "closed",
		Operation:   "sell",
		ExternalId:  nil,
		ClosesOrder: &opened.Id,
		// todo: add commission???
	}

	binanceOrder, err := m.tryLimitOrder(order, "SELL", 480)

	if err != nil {
		m.BalanceService.InvalidateBalanceCache("USDT")
		m.BalanceService.InvalidateBalanceCache(order.GetBaseAsset())
		return err
	}

	// fill from API
	order.ExternalId = &binanceOrder.OrderId
	order.ExecutedQuantity = binanceOrder.GetExecutedQuantity()
	order.Price = binanceOrder.Price
	order.CreatedAt = m.TimeService.GetNowDateTimeString()

	lastId, err := m.OrderRepository.Create(order)

	if err != nil {
		// todo: test 2024/02/02 08:24:29 [XLMUSDT] Error 1062 (23000): Duplicate entry '207993-XLMUSDT' for key 'order_external_id_symbol'
		// remove binance order from cache if we have already had saved in database
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "order_external_id_symbol") {
			m.OrderRepository.DeleteBinanceOrder(binanceOrder)
		}

		log.Printf("Can't create order: %s", order.Symbol)

		return err
	}

	m.OrderRepository.DeleteManualOrder(order.Symbol)
	_, err = m.OrderRepository.Find(*lastId)

	if err != nil {
		log.Printf("Can't get created order [%d]: %s", lastId, order.Symbol)

		return err
	}

	closings := m.OrderRepository.GetClosesOrderList(opened)
	totalExecuted := 0.00
	commission := 0.00
	// @see https://www.binance.com/en/fee/trading
	commission += opened.ExecutedQuantity * 0.0015
	for _, closeOrder := range closings {
		if closeOrder.IsClosed() {
			totalExecuted += closeOrder.ExecutedQuantity
			commission += closeOrder.ExecutedQuantity * 0.0015
		}
	}

	if (opened.ExecutedQuantity - (totalExecuted + commission)) <= tradeLimit.MinQuantity {
		opened.Status = "closed"
	}

	err = m.OrderRepository.Update(opened)
	m.BalanceService.InvalidateBalanceCache("USDT")
	m.BalanceService.InvalidateBalanceCache(opened.GetBaseAsset())

	if err != nil {
		log.Printf("Can't udpdate order [%d]: %s", order.Id, order.Symbol)

		return err
	}

	go func(order ExchangeModel.Order, profit float64) {
		m.CallbackManager.SellOrder(
			order,
			*m.CurrentBot,
			fmt.Sprintf("Profit is: %f USDT", m.Formatter.ToFixed(profit, 2)),
		)
	}(order, profit)

	return nil
}

func (m *OrderExecutor) ProcessSwap(order ExchangeModel.Order) bool {
	if m.SwapEnabled && order.IsSwap() {
		log.Printf("[%s] Swap Order [%d] Mode: processing...", order.Symbol, order.Id)
		m.SwapExecutor.Execute(order)
		return true
	} else if m.SwapEnabled {
		swapAction, err := m.SwapRepository.GetActiveSwapAction(order)
		if err == nil && swapAction.OrderId == order.Id {
			log.Printf("[%s] Swap Recovered for Order [%d] Mode: processing...", order.Symbol, order.Id)
			m.SwapExecutor.Execute(order)
			return true
		}
	}

	return false
}

func (m *OrderExecutor) TrySwap(order ExchangeModel.Order) {
	swapChain := m.SwapRepository.GetSwapChainCache(order.GetBaseAsset())
	if swapChain != nil && m.SwapEnabled {
		possibleSwaps := m.SwapRepository.GetSwapChains(order.GetBaseAsset())

		if len(possibleSwaps) == 0 {
			m.SwapRepository.InvalidateSwapChainCache(order.GetBaseAsset())
		}

		for _, possibleSwap := range possibleSwaps {
			violation := m.SwapValidator.Validate(possibleSwap, order)

			if violation == nil {
				chainCurrentPercent := m.SwapValidator.CalculatePercent(possibleSwap)
				log.Printf(
					"[%s] TRY SWAP -> Swap chain [%s] is found for order #%d, initial percent: %.2f, current = %.2f",
					order.Symbol,
					swapChain.Title,
					order.Id,
					swapChain.Percent,
					chainCurrentPercent,
				)
				m.MakeSwap(order, possibleSwap)
			}
		}
	}
}

// todo: order has to be Interface
func (m *OrderExecutor) tryLimitOrder(order ExchangeModel.Order, operation string, ttl int64) (ExchangeModel.BinanceOrder, error) {
	// todo: extra order flag...
	binanceOrder, err := m.findOrCreateOrder(order, operation)

	if err != nil {
		return binanceOrder, err
	}

	if (binanceOrder.IsCanceled() || binanceOrder.IsExpired()) && binanceOrder.ExecutedQty == 0 {
		m.OrderRepository.DeleteBinanceOrder(binanceOrder)

		return binanceOrder, errors.New(fmt.Sprintf("binance order [%d] is cancelled or expired", binanceOrder.OrderId))
	}

	if binanceOrder.IsFilled() {
		return binanceOrder, nil
	}

	if binanceOrder.IsCanceled() && binanceOrder.ExecutedQty > 0.00 {
		return binanceOrder, nil
	}

	// todo: save sell order in buy order to make sure it is saved after processing...
	binanceOrder, err = m.waitExecution(binanceOrder, ttl)

	if err != nil {
		return binanceOrder, err
	}

	return binanceOrder, nil
}

func (m *OrderExecutor) waitExecution(binanceOrder ExchangeModel.BinanceOrder, seconds int64) (ExchangeModel.BinanceOrder, error) {
	defer m.OrderRepository.DeleteBinanceOrder(binanceOrder)

	if binanceOrder.IsFilled() {
		return binanceOrder, nil
	}

	depth := m.PriceCalculator.GetDepth(binanceOrder.Symbol)

	var currentPosition int
	var book [2]ExchangeModel.Number
	if "BUY" == binanceOrder.Side {
		currentPosition, book = depth.GetBidPosition(binanceOrder.Price)
	} else {
		currentPosition, book = depth.GetAskPosition(binanceOrder.Price)
	}
	log.Printf(
		"[%s] Order Book start position is [%d] %.6f\n",
		binanceOrder.Symbol,
		currentPosition,
		book[0],
	)

	executedQty := 0.00

	orderManageChannel := make(chan string)
	control := make(chan string)
	defer close(orderManageChannel)
	defer close(control)

	tradeLimit, err := m.ExchangeRepository.GetTradeLimit(binanceOrder.Symbol)

	if err != nil {
		return binanceOrder, err
	}

	go func(
		tradeLimit ExchangeModel.TradeLimit,
		binanceOrder *ExchangeModel.BinanceOrder,
		ttl *int64,
		control chan string,
		orderManageChannel chan string,
	) {
		timer := 0
		start := m.TimeService.GetNowUnix()

		for {
			if binanceOrder.IsCanceled() || binanceOrder.IsExpired() || binanceOrder.IsFilled() {
				orderManageChannel <- "status"
				action := <-control
				if action == "stop" {
					return
				}

				m.TimeService.WaitSeconds(40)
				continue
			}

			end := m.TimeService.GetNowUnix()
			kline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

			if kline != nil && binanceOrder.IsSell() && binanceOrder.IsNew() && m.SwapEnabled {
				openedBuyPosition, err := m.OrderRepository.GetOpenedOrderCached(binanceOrder.Symbol, "BUY")
				// Try arbitrage for long orders >= 4 hours and with profit < -1.00%
				if err == nil {
					swapChain := m.SwapRepository.GetSwapChainCache(openedBuyPosition.GetBaseAsset())
					if swapChain != nil {
						possibleSwaps := m.SwapRepository.GetSwapChains(openedBuyPosition.GetBaseAsset())

						if len(possibleSwaps) == 0 {
							m.SwapRepository.InvalidateSwapChainCache(openedBuyPosition.GetBaseAsset())
						}

						for _, possibleSwap := range possibleSwaps {
							turboSwap := possibleSwap.Percent.Gte(ExchangeModel.Percent(m.TurboSwapProfitPercent))
							isTimeToSwap := openedBuyPosition.GetHoursOpened() >= m.SwapSellOrderDays && openedBuyPosition.GetProfitPercent(kline.Close).Lte(ExchangeModel.Percent(m.SwapProfitPercent)) && !openedBuyPosition.IsSwap()

							if !turboSwap && !isTimeToSwap {
								break
							}

							violation := m.SwapValidator.Validate(possibleSwap, openedBuyPosition)

							if violation == nil {
								chainCurrentPercent := m.SwapValidator.CalculatePercent(possibleSwap)
								log.Printf(
									"[%s] Swap chain [%s] is found for order #%d, initial percent: %.2f, current = %.2f",
									binanceOrder.Symbol,
									swapChain.Title,
									openedBuyPosition.Id,
									swapChain.Percent,
									chainCurrentPercent,
								)
								orderManageChannel <- "status"
								action := <-control
								if action == "stop" {
									return
								}

								if binanceOrder.IsNew() {
									log.Printf("[%s] Cancel signal sent!", binanceOrder.Symbol)
									orderManageChannel <- "cancel"
									action := <-control
									if action == "stop" {
										m.MakeSwap(openedBuyPosition, possibleSwap)
										return
									}
								}
							}
						}
					}
				}
			}

			if m.LossSecurity.IsRiskyBuy(*binanceOrder, tradeLimit) {
				log.Printf("[%s] (LossSecurity) Check status signal sent!", binanceOrder.Symbol)
				orderManageChannel <- "status"
				action := <-control
				if action == "stop" {
					return
				}
				log.Printf("[%s] (LossSecurity) Order status is [%s]", binanceOrder.Symbol, binanceOrder.Status)

				if binanceOrder.IsNew() {
					log.Printf("[%s] (LossSecurity) Cancel signal sent!", binanceOrder.Symbol)
					orderManageChannel <- "cancel"
					action := <-control
					if action == "stop" {
						m.OrderRepository.LockBuy(binanceOrder.Symbol, 10)

						return
					}
				}
			}

			if m.HasCancelRequest(binanceOrder.Symbol) && binanceOrder.IsNew() {
				log.Printf(
					"[%s] Cancel request received from user",
					binanceOrder.Symbol,
				)
				orderManageChannel <- "cancel"
				action := <-control
				if action == "stop" {
					return
				}
			}

			// [BUY] Check is it time to sell (maybe we have already partially filled)
			if kline != nil && binanceOrder.IsBuy() && binanceOrder.IsPartiallyFilled() && binanceOrder.GetProfitPercent(kline.Close).Gte(tradeLimit.GetMinProfitPercent()) {
				log.Printf(
					"[%s] Max profit percent reached, current profit is: %.2f, %s [%d] order is cancelled",
					binanceOrder.Symbol,
					binanceOrder.GetProfitPercent(kline.Close).Value(),
					binanceOrder.Side,
					binanceOrder.OrderId,
				)
				orderManageChannel <- "cancel"
				action := <-control
				if action == "stop" {
					return
				}
			}

			// Check is time to extra buy, but we have sell partial...
			if kline != nil && binanceOrder.IsSell() && binanceOrder.IsPartiallyFilled() {
				openedBuyPosition, err := m.OrderRepository.GetOpenedOrderCached(binanceOrder.Symbol, "BUY")
				if err == nil && openedBuyPosition.GetProfitPercent(kline.Close).Lte(tradeLimit.GetBuyOnFallPercent(openedBuyPosition, *kline)) {
					log.Printf(
						"[%s] Extra Charge percent reached, current profit is: %.2f, SELL order is cancelled",
						binanceOrder.Symbol,
						openedBuyPosition.GetProfitPercent(kline.Close).Value(),
					)
					orderManageChannel <- "cancel"
					action := <-control
					if action == "stop" {
						return
					}
				}
			}

			if timer >= 30000 {
				orderManageChannel <- "status"
				action := <-control
				log.Printf(
					"[%s] %s Order [%d] status [%s] wait handler (%s), current price is [%.8f], order price [%.8f], ExecutedQty: %.6f of %.6f\"",
					binanceOrder.Symbol,
					binanceOrder.Side,
					binanceOrder.OrderId,
					binanceOrder.Status,
					action,
					kline.Close,
					binanceOrder.Price,
					binanceOrder.ExecutedQty,
					binanceOrder.OrigQty,
				)
				if action == "stop" {
					return
				}

				timer = 0
				m.TimeService.WaitSeconds(1)

				// check only new timeout
				if end >= (start+*ttl) && binanceOrder.IsNew() {
					if kline != nil && binanceOrder.IsSell() {
						openedBuyPosition, err := m.OrderRepository.GetOpenedOrderCached(binanceOrder.Symbol, "BUY")
						if err == nil {
							profitPercent := openedBuyPosition.GetProfitPercent(kline.Close)
							if profitPercent.Lte(0.00) {
								log.Printf(
									"[%s] %s Order [%d] status [%s] ttl reached, current price is [%.8f], order price [%.8f], open [%.8f], profit: %.2f",
									binanceOrder.Symbol,
									binanceOrder.Side,
									binanceOrder.OrderId,
									binanceOrder.Status,
									kline.Close,
									binanceOrder.Price,
									openedBuyPosition.Price,
									profitPercent.Value(),
								)
								orderManageChannel <- "cancel"
								action := <-control
								if action == "stop" {
									return
								}
							} else {
								log.Printf(
									"[%s] %s Order [%d] status [%s] ttl ignored, current price is [%.8f], order price [%.8f], open [%.8f], profit: %.2f",
									binanceOrder.Symbol,
									binanceOrder.Side,
									binanceOrder.OrderId,
									binanceOrder.Status,
									kline.Close,
									binanceOrder.Price,
									openedBuyPosition.Price,
									profitPercent.Value(),
								)
							}
						} else {
							log.Printf(
								"[%s] %s Order [%d] %s",
								binanceOrder.Symbol,
								binanceOrder.Side,
								binanceOrder.OrderId,
								err.Error(),
							)
							orderManageChannel <- "cancel"
							action := <-control
							if action == "stop" {
								return
							}
						}
					}

					if kline != nil && binanceOrder.IsBuy() {
						positionPercentage := m.Formatter.ComparePercentage(binanceOrder.Price, kline.Close)
						if positionPercentage.Gte(101.00) {
							log.Printf(
								"[%s] %s Order [%d] status [%s] ttl reached, current price is [%.8f], order price [%.8f], diff percent: %.2f",
								binanceOrder.Symbol,
								binanceOrder.Side,
								binanceOrder.OrderId,
								binanceOrder.Status,
								kline.Close,
								binanceOrder.Price,
								positionPercentage.Value(),
							)
							orderManageChannel <- "cancel"
							action := <-control
							if action == "stop" {
								return
							}
						} else {
							log.Printf(
								"[%s] %s Order [%d] status [%s] ttl ignored, current price is [%.8f], order price [%.8f], diff percent: %.2f",
								binanceOrder.Symbol,
								binanceOrder.Side,
								binanceOrder.OrderId,
								binanceOrder.Status,
								kline.Close,
								binanceOrder.Price,
								positionPercentage.Value(),
							)
						}
					}
				}
			} else {
				manualOrder := m.OrderRepository.GetManualOrder(binanceOrder.Symbol)
				// cancel current immediately on new manual order
				if manualOrder != nil && manualOrder.Price != binanceOrder.Price {
					orderManageChannel <- "cancel"
					action := <-control
					if action == "stop" {
						return
					}
				} else {
					orderManageChannel <- "continue"
					action := <-control
					if action == "stop" {
						return
					}
				}
			}

			timer = timer + 30
			m.TimeService.WaitMilliseconds(20)
		}
	}(tradeLimit, &binanceOrder, &seconds, control, orderManageChannel)

	for {
		action := <-orderManageChannel
		if action == "continue" {
			control <- "continue"
			continue
		}

		if action == "cancel" {
			log.Printf(
				"[%s] %s Order %d, cancel signal has received",
				binanceOrder.Symbol,
				binanceOrder.Side,
				binanceOrder.OrderId,
			)
			break
		}

		queryOrder, err := m.Binance.QueryOrder(binanceOrder.Symbol, binanceOrder.OrderId)

		if err != nil {
			log.Printf("[%s] QueryOrder: %s", binanceOrder.Symbol, err.Error())

			if strings.Contains(err.Error(), "Order was canceled or expired") {
				control <- "stop"
				return binanceOrder, err
			}

			if strings.Contains(err.Error(), "Order does not exist") {
				control <- "stop"
				return binanceOrder, err
			}

			log.Printf("[%s] Retry query order...", binanceOrder.Symbol)
			m.TimeService.WaitSeconds(120)

			control <- "continue"
			continue
		}

		binanceOrder = queryOrder
		m.OrderRepository.SetBinanceOrder(binanceOrder)

		if binanceOrder.IsPartiallyFilled() {
			// Add 5 minutes more if ExecutedQty moves up!
			if binanceOrder.GetExecutedQuantity() > executedQty {
				seconds = seconds + (60 * 5)
			}

			executedQty = binanceOrder.GetExecutedQuantity()
			m.OrderRepository.SetBinanceOrder(binanceOrder)
			control <- "continue"
			continue
		}

		if binanceOrder.IsExpired() {
			if binanceOrder.HasExecutedQuantity() {
				control <- "stop"
				return binanceOrder, nil
			}

			return binanceOrder, errors.New("Order is expired")
		}

		if binanceOrder.IsCanceled() {
			if binanceOrder.HasExecutedQuantity() {
				control <- "stop"
				return binanceOrder, nil
			}

			control <- "stop"
			return binanceOrder, errors.New("Order is cancelled")
		}

		if binanceOrder.IsFilled() {
			log.Printf("[%s] Order [%d] is executed [%s]", binanceOrder.Symbol, binanceOrder.OrderId, binanceOrder.Status)

			control <- "stop"
			return binanceOrder, nil
		}

		control <- "continue"
	}

	// If you cancel an order that has already been partially filled,
	// the cryptocurrency or fiat currency that was used to fill the order
	// will be returned to your account. The remaining cryptocurrency or fiat
	// currency will be used to fill other orders that are waiting to be executed.
	// {
	//    "symbol": "ETHUSDT",
	//    "origClientOrderId": "aSUn6e7pktn5fVFuNEb0TK",
	//    "orderId": 31314,
	//    "orderListId": -1,
	//    "clientOrderId": "wnpmGUt6RgoyuZB48NXbFG",
	//    "transactTime": 1701886419972,
	//    "price": "2100.93000000",
	//    "origQty": "0.04750000",
	//    "executedQty": "0.04670000",
	//    "cummulativeQuoteQty": "98.11343100",
	//    "status": "CANCELED",
	//    "timeInForce": "GTC",
	//    "type": "LIMIT",
	//    "side": "BUY",
	//    "selfTradePreventionMode": "EXPIRE_MAKER"
	//}
	cancelOrder, err := m.Binance.CancelOrder(binanceOrder.Symbol, binanceOrder.OrderId)

	if err != nil {
		// Possible case: {"code": -2011,"msg": "Order was not canceled due to cancel restrictions."}
		log.Printf("[%s] Cancel failed: %s", binanceOrder.Symbol, err.Error())
		queryOrder, retryErr := m.Binance.QueryOrder(binanceOrder.Symbol, binanceOrder.OrderId)

		if retryErr == nil {
			binanceOrder = queryOrder
			control <- "stop"
			log.Printf("[%s] Order [%d] is recovered [%s]", binanceOrder.Symbol, binanceOrder.OrderId, binanceOrder.Status)

			if binanceOrder.IsFilled() {
				return binanceOrder, nil
			}

			// Just in case of bug...
			if binanceOrder.IsPartiallyFilled() {
				log.Printf(
					"[%s] Order [%d] status is [%s], try again waitExecution...",
					binanceOrder.Symbol,
					binanceOrder.OrderId,
					binanceOrder.Status,
				)

				return m.waitExecution(binanceOrder, 120)
			}

			// Just in case of bug...
			if binanceOrder.IsNew() {
				log.Printf(
					"[%s] Order [%d] status is [%s], try again waitExecution...",
					binanceOrder.Symbol,
					binanceOrder.OrderId,
					binanceOrder.Status,
				)

				return m.waitExecution(binanceOrder, 120)
			}

			if binanceOrder.HasExecutedQuantity() {
				log.Printf(
					"Order [%d] is [%s], ExecutedQty = %.8f",
					binanceOrder.OrderId,
					binanceOrder.Status,
					binanceOrder.GetExecutedQuantity(),
				)

				return binanceOrder, nil
			} else {
				return binanceOrder, errors.New(fmt.Sprintf("Order %d was CANCELED", binanceOrder.OrderId))
			}
		} else {
			// todo: loop??? timeout + loop???
			control <- "stop"
			return binanceOrder, err
		}
	}

	binanceOrder = cancelOrder
	control <- "stop"

	// handle cancel error and get again

	if binanceOrder.HasExecutedQuantity() {
		log.Printf(
			"Order [%d] is [%s], ExecutedQty = %.8f",
			binanceOrder.OrderId,
			binanceOrder.Status,
			binanceOrder.GetExecutedQuantity(),
		)

		return binanceOrder, nil
	}

	log.Printf("Order [%d] is [%s]", binanceOrder.OrderId, binanceOrder.Status)

	return binanceOrder, errors.New(fmt.Sprintf("Order %d was CANCELED", binanceOrder.OrderId))
}

func (m *OrderExecutor) CalculateSellQuantity(order ExchangeModel.Order) float64 {
	binanceOrder := m.OrderRepository.GetBinanceOrder(order.Symbol, "SELL")

	if binanceOrder != nil {
		return binanceOrder.OrigQty
	}

	m.recoverCommission(order)
	sellQuantity := order.GetRemainingToSellQuantity()
	balance, err := m.BalanceService.GetAssetBalance(order.GetBaseAsset(), true)

	if err != nil {
		return sellQuantity
	}

	if balance > sellQuantity {
		// User can have own asset which bot is not allowed to sell!
		return sellQuantity
	}

	return balance
}

func (m *OrderExecutor) MakeSwap(order ExchangeModel.Order, swapChain ExchangeModel.SwapChainEntity) {
	assetBalance, err := m.BalanceService.GetAssetBalance(swapChain.SwapOne.BaseAsset, false)

	if err != nil {
		return
	}

	swapAction, err := m.SwapRepository.GetActiveSwapAction(order)

	if err == nil {
		log.Printf("[%s] Swap has already exists: %s", swapChain.SwapOne.BaseAsset, swapAction.Status)

		return
	}

	// todo: transaction
	// create swap
	_, err = m.SwapRepository.CreateSwapAction(ExchangeModel.SwapAction{
		Id:              0,
		OrderId:         order.Id,
		BotId:           m.CurrentBot.Id,
		SwapChainId:     swapChain.Id,
		Asset:           swapChain.SwapOne.BaseAsset,
		Status:          ExchangeModel.SwapActionStatusPending,
		StartTimestamp:  m.TimeService.GetNowUnix(),
		StartQuantity:   assetBalance,
		SwapOneSymbol:   swapChain.SwapOne.GetSymbol(),
		SwapOnePrice:    swapChain.SwapOne.Price,
		SwapTwoSymbol:   swapChain.SwapTwo.GetSymbol(),
		SwapTwoPrice:    swapChain.SwapTwo.Price,
		SwapThreeSymbol: swapChain.SwapThree.GetSymbol(),
		SwapThreePrice:  swapChain.SwapThree.Price,
	})

	if err != nil {
		log.Printf(
			"[%s] Swap couldn't be created: %s",
			swapChain.SwapOne.BaseAsset,
			err.Error(),
		)

		return
	}

	// enable order swap mode
	order.Swap = true
	err = m.OrderRepository.Update(order)
	if err == nil {
		log.Printf("[%s] Swap order mode enabled [%s]", order.Symbol, swapChain.Title)
	}
}

func (m *OrderExecutor) UpdateCommission(balanceBefore float64, order ExchangeModel.Order) {
	assetSymbol := order.GetBaseAsset()
	balanceAfter, err := m.BalanceService.GetAssetBalance(assetSymbol, true)

	if err != nil {
		log.Printf("[%s] Can't update commission: %s", order.Status, err.Error())
		return
	}

	arrived := balanceAfter - balanceBefore

	commission := order.ExecutedQuantity - arrived

	if commission < 0 {
		commission = 0.00
	}

	order.Commission = &commission
	order.CommissionAsset = &assetSymbol

	err = m.OrderRepository.Update(order)
	if err != nil {
		log.Printf("[%s] Order Commission Update: %s", order.Symbol, err.Error())
	}
}

func (m *OrderExecutor) isTradeLocked(symbol string) bool {
	m.TradeLockMutex.Lock()
	isLocked, _ := m.Lock[symbol]
	m.TradeLockMutex.Unlock()

	return isLocked
}

func (m *OrderExecutor) acquireLock(symbol string) {
	*m.LockChannel <- ExchangeModel.Lock{IsLocked: true, Symbol: symbol}
}

func (m *OrderExecutor) releaseLock(symbol string) {
	*m.LockChannel <- ExchangeModel.Lock{IsLocked: false, Symbol: symbol}
}

func (m *OrderExecutor) findBinanceOrder(symbol string, operation string, cachedOnly bool) (*ExchangeModel.BinanceOrder, error) {
	cached := m.OrderRepository.GetBinanceOrder(symbol, operation)

	if cached != nil {
		log.Printf("[%s] Found cached %s order %d in binance", symbol, operation, cached.OrderId)

		return cached, nil
	}

	if cachedOnly {
		return nil, errors.New(fmt.Sprintf("[%s] Cached binance order is not found", symbol))
	}

	openedOrders, err := m.Binance.GetOpenedOrders()

	if err != nil {
		log.Printf("[%s] Opened: %s", symbol, err.Error())
		return nil, err
	}

	for _, opened := range openedOrders {
		if opened.Side == operation && opened.Symbol == symbol {
			log.Printf("[%s] Found opened %s order %d in binance", symbol, operation, opened.OrderId)
			m.OrderRepository.SetBinanceOrder(opened)

			return &opened, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("[%s] Binance order is not found", symbol))
}

func (m *OrderExecutor) findOrCreateOrder(order ExchangeModel.Order, operation string) (ExchangeModel.BinanceOrder, error) {
	// todo: extra order flag...
	cached, err := m.findBinanceOrder(order.Symbol, operation, false)

	if cached != nil {
		log.Printf("[%s] Found cached %s order %d in binance", order.Symbol, operation, cached.OrderId)

		return *cached, nil
	}

	binanceOrder, err := m.Binance.LimitOrder(order.Symbol, order.Quantity, order.Price, operation, "GTC")

	if err != nil {
		log.Printf("[%s] Limit: %s", order.Symbol, err.Error())
		return binanceOrder, err
	}

	log.Printf("[%s] %s Order created %d, Price: %.6f", order.Symbol, operation, binanceOrder.OrderId, binanceOrder.Price)
	m.OrderRepository.SetBinanceOrder(binanceOrder)
	if order.IsBuy() {
		m.BalanceService.InvalidateBalanceCache("USDT")
	} else {
		m.BalanceService.InvalidateBalanceCache(order.GetBaseAsset())
	}

	return binanceOrder, nil
}

func (m *OrderExecutor) getAvgPrice(opened ExchangeModel.Order, extra ExchangeModel.Order) float64 {
	return ((opened.ExecutedQuantity * opened.Price) + (extra.ExecutedQuantity * extra.Price)) / (opened.ExecutedQuantity + extra.ExecutedQuantity)
}

func (m *OrderExecutor) recoverCommission(order ExchangeModel.Order) {
	if order.Commission != nil {
		return
	}
	assetSymbol := order.GetBaseAsset()

	balanceAfter, err := m.BalanceService.GetAssetBalance(assetSymbol, true)

	if err != nil {
		log.Printf("[%s] Can't recover commission: %s", order.Status, err.Error())
		return
	}

	commission := order.ExecutedQuantity - balanceAfter
	if commission < 0 {
		commission = 0
	}
	order.Commission = &commission
	order.CommissionAsset = &assetSymbol

	err = m.OrderRepository.Update(order)
	if err != nil {
		log.Printf("[%s] Order Commission Recover: %s", order.Symbol, err.Error())
	}
}

func (m *OrderExecutor) CheckBalance(symbol string, priceUsdt float64, quantity float64) error {
	cached, _ := m.findBinanceOrder(symbol, "BUY", true)

	// Check balance for new order
	if cached == nil {
		usdtAvailableBalance, err := m.BalanceService.GetAssetBalance("USDT", true)

		if err != nil {
			return errors.New(fmt.Sprintf("[%s] BUY balance error: %s", symbol, err.Error()))
		}

		requiredUsdtAmount := priceUsdt * quantity

		if requiredUsdtAmount > usdtAvailableBalance {
			return errors.New(fmt.Sprintf("[%s] BUY not enough balance: %f/%f", symbol, usdtAvailableBalance, requiredUsdtAmount))
		}
	}

	return nil
}

func (m *OrderExecutor) CheckMinBalance(limit ExchangeModel.TradeLimit, kLine ExchangeModel.KLine) error {
	opened, err := m.OrderRepository.GetOpenedOrderCached(limit.Symbol, "BUY")
	limitUsdt := limit.USDTLimit

	if err == nil {
		limitUsdt = opened.GetAvailableExtraBudget(limit, kLine)
	}

	cached, _ := m.findBinanceOrder(limit.Symbol, "BUY", true)

	// Check balance for new order
	if cached == nil {
		usdtAvailableBalance, err := m.BalanceService.GetAssetBalance("USDT", true)

		if err != nil {
			return errors.New(fmt.Sprintf("[%s] BUY balance error: %s", limit.Symbol, err.Error()))
		}

		if limitUsdt > usdtAvailableBalance {
			return errors.New(fmt.Sprintf("[%s] BUY not enough balance: %f/%f", limit.Symbol, usdtAvailableBalance, limitUsdt))
		}
	}

	return nil
}

func (o *OrderExecutor) HasCancelRequest(symbol string) bool {
	value, ok := o.CancelRequestMap[symbol]

	if !ok {
		return false
	}

	defer func(symbol string) {
		delete(o.CancelRequestMap, symbol)
	}(symbol)

	if value {
		return true
	}

	return false
}

func (o *OrderExecutor) SetCancelRequest(symbol string) {
	o.TradeLockMutex.Lock()
	o.CancelRequestMap[symbol] = true
	o.TradeLockMutex.Unlock()
}
