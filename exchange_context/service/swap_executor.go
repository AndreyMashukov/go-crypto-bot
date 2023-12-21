package service

import (
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"strings"
	"time"
)

type SwapExecutor struct {
	SwapRepository  repository.SwapBasicRepositoryInterface
	OrderRepository repository.OrderUpdaterInterface
	BalanceService  BalanceServiceInterface
	Binance         client.ExchangeOrderAPIInterface
	TimeoutService  TimeoutServiceInterface
	Formatter       *Formatter
}

func (s *SwapExecutor) Execute(order ExchangeModel.Order) {
	swapAction, err := s.SwapRepository.GetActiveSwapAction(order)

	if err != nil {
		log.Printf("[%s] Swap processing error: %s", order.Symbol, err.Error())

		if strings.Contains(err.Error(), "no rows in result set") {
			order.Swap = false
			_ = s.OrderRepository.Update(order)
		}

		return
	}

	balanceBefore, _ := s.BalanceService.GetAssetBalance(swapAction.Asset, false)

	if swapAction.IsPending() {
		swapAction.Status = ExchangeModel.SwapActionStatusProcess
		_ = s.SwapRepository.UpdateSwapAction(swapAction)
	}

	swapChain, err := s.SwapRepository.GetSwapChainById(swapAction.SwapChainId)

	if err != nil {
		log.Printf("Swap chain %d is not found", swapAction.SwapChainId)
		return
	}

	swapOneOrder := s.ExecuteSwapOne(&swapAction, order)

	if swapOneOrder == nil {
		return
	}

	swapTwoOrder := s.ExecuteSwapTwo(&swapAction, swapChain, *swapOneOrder)

	if swapTwoOrder == nil {
		return
	}

	assetTwo := strings.ReplaceAll(swapOneOrder.Symbol, swapAction.Asset, "")

	// step 3
	swapThreeOrder := s.ExecuteSwapThree(&swapAction, swapChain, *swapTwoOrder, assetTwo)

	if swapThreeOrder == nil {
		return
	}

	endQuantity := swapThreeOrder.ExecutedQty
	if swapChain.IsSBS() {
		endQuantity = swapThreeOrder.ExecutedQty * swapThreeOrder.Price
	}

	order.Swap = false
	swapAction.Status = ExchangeModel.SwapActionStatusSuccess
	nowTimestamp := time.Now().Unix()
	swapAction.EndTimestamp = &nowTimestamp
	swapAction.SwapThreeTimestamp = &nowTimestamp
	swapAction.EndQuantity = &endQuantity
	_ = s.SwapRepository.UpdateSwapAction(swapAction)
	_ = s.OrderRepository.Update(order)

	s.BalanceService.InvalidateBalanceCache(swapAction.Asset)
	balanceAfter, _ := s.BalanceService.GetAssetBalance(swapAction.Asset, false)

	log.Printf(
		"[%s] Swap funished, balance %s before = %f after = %f",
		order.Symbol,
		swapAction.Asset,
		balanceBefore,
		balanceAfter,
	)
}

func (s *SwapExecutor) ExecuteSwapOne(swapAction *ExchangeModel.SwapAction, order ExchangeModel.Order) *ExchangeModel.BinanceOrder {
	var swapOneOrder *ExchangeModel.BinanceOrder = nil

	if swapAction.SwapOneExternalId == nil {
		swapPair, err := s.SwapRepository.GetSwapPairBySymbol(swapAction.SwapOneSymbol)
		binanceOrder, err := s.Binance.LimitOrder(
			swapAction.SwapOneSymbol,
			s.Formatter.FormatQuantity(swapPair, swapAction.StartQuantity),
			s.Formatter.FormatPrice(swapPair, swapAction.SwapOnePrice),
			"SELL",
			"GTC",
		)

		if err != nil {
			log.Printf(
				"[%s] Swap [%d] error: %s",
				order.Symbol,
				swapAction.Id,
				err.Error(),
			)

			orderStatus := "ERROR"
			swapAction.SwapOneExternalStatus = &orderStatus
			swapAction.Status = ExchangeModel.SwapActionStatusCanceled
			nowTimestamp := time.Now().Unix()
			swapAction.EndTimestamp = &nowTimestamp
			swapAction.EndQuantity = &swapAction.StartQuantity
			_ = s.SwapRepository.UpdateSwapAction(*swapAction)
			order.Swap = false
			_ = s.OrderRepository.Update(order)
			// invalidate balance cache
			s.BalanceService.InvalidateBalanceCache(swapAction.Asset)
			return nil
		}

		swapOneOrder = &binanceOrder
		swapAction.SwapOneExternalId = &binanceOrder.OrderId
		nowTimestamp := time.Now().Unix()
		swapAction.SwapOneTimestamp = &nowTimestamp
		swapAction.SwapOneExternalStatus = &binanceOrder.Status
		_ = s.SwapRepository.UpdateSwapAction(*swapAction)
	} else {
		binanceOrder, err := s.Binance.QueryOrder(swapAction.SwapOneSymbol, *swapAction.SwapOneExternalId)
		if err != nil {
			log.Printf("[%s] Swap error: %s", order.Symbol, err.Error())
			return nil
		}

		if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
			swapAction.SwapOneExternalId = nil
			swapAction.SwapOneTimestamp = nil
			swapAction.SwapOneExternalStatus = nil
			_ = s.SwapRepository.UpdateSwapAction(*swapAction)
			s.BalanceService.InvalidateBalanceCache(swapAction.Asset)

			return nil
		}

		swapOneOrder = &binanceOrder
		swapAction.SwapOneExternalStatus = &binanceOrder.Status
		_ = s.SwapRepository.UpdateSwapAction(*swapAction)
	}

	// step 1

	// todo: if expired, clear and call recursively
	if !swapOneOrder.IsFilled() {
		s.TimeoutService.WaitSeconds(5)
		for {
			binanceOrder, err := s.Binance.QueryOrder(swapOneOrder.Symbol, swapOneOrder.OrderId)
			if err != nil {
				log.Printf("[%s] Swap Binance error: %s", order.Symbol, err.Error())

				continue
			}

			swapPair, err := s.SwapRepository.GetSwapPairBySymbol(binanceOrder.Symbol)

			log.Printf(
				"[%s] Swap [%d] one [%s] processing, status %s [%d], price %f, current = %f, Executed %f of %f",
				swapAction.SwapOneSymbol,
				swapAction.Id,
				binanceOrder.Side,
				binanceOrder.Status,
				binanceOrder.OrderId,
				binanceOrder.Price,
				swapPair.SellPrice,
				binanceOrder.ExecutedQty,
				binanceOrder.OrigQty,
			)

			// update value, set new memory address
			swapOneOrder = &binanceOrder

			nowTimestamp := time.Now().Unix()
			swapAction.SwapOneTimestamp = &nowTimestamp
			swapAction.SwapOneExternalStatus = &binanceOrder.Status
			_ = s.SwapRepository.UpdateSwapAction(*swapAction)

			if binanceOrder.IsFilled() {
				break
			}

			// todo: timeout... cancel and remove swap action...

			if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
				swapAction.SwapOneExternalStatus = &binanceOrder.Status
				swapAction.Status = ExchangeModel.SwapActionStatusCanceled
				nowTimestamp := time.Now().Unix()
				swapAction.EndTimestamp = &nowTimestamp
				swapAction.EndQuantity = &swapAction.StartQuantity
				_ = s.SwapRepository.UpdateSwapAction(*swapAction)
				order.Swap = false
				_ = s.OrderRepository.Update(order)
				// invalidate balance cache
				s.BalanceService.InvalidateBalanceCache(swapAction.Asset)
				log.Printf("[%s] Swap one process cancelled, cancel all the operation!", order.Symbol)

				return nil
			}

			if binanceOrder.IsNew() && (time.Now().Unix()-swapAction.StartTimestamp) >= 60 {
				cancelOrder, err := s.Binance.CancelOrder(binanceOrder.Symbol, binanceOrder.OrderId)
				if err == nil {
					swapAction.SwapOneExternalStatus = &cancelOrder.Status
					swapAction.Status = ExchangeModel.SwapActionStatusCanceled
					nowTimestamp := time.Now().Unix()
					swapAction.EndTimestamp = &nowTimestamp
					swapAction.EndQuantity = &swapAction.StartQuantity
					_ = s.SwapRepository.UpdateSwapAction(*swapAction)
					order.Swap = false
					_ = s.OrderRepository.Update(order)
					// invalidate balance cache
					s.BalanceService.InvalidateBalanceCache(swapAction.Asset)
					log.Printf("[%s] Swap process cancelled, couldn't be processed more than 60 seconds", order.Symbol)

					return nil
				}
			}

			if binanceOrder.IsPartiallyFilled() {
				s.TimeoutService.WaitSeconds(7)
			} else {
				s.TimeoutService.WaitSeconds(15)
			}
		}
	}

	return swapOneOrder
}

func (s *SwapExecutor) ExecuteSwapTwo(
	swapAction *ExchangeModel.SwapAction,
	swapChain ExchangeModel.SwapChainEntity,
	swapOneOrder ExchangeModel.BinanceOrder,
) *ExchangeModel.BinanceOrder {
	assetTwo := strings.ReplaceAll(swapOneOrder.Symbol, swapAction.Asset, "")

	var swapTwoOrder *ExchangeModel.BinanceOrder = nil

	if swapAction.SwapTwoExternalId == nil {
		balance, _ := s.BalanceService.GetAssetBalance(assetTwo, false)
		// Calculate how much we earn, and sell it!
		quantity := swapOneOrder.ExecutedQty * swapOneOrder.Price

		if quantity > balance {
			quantity = balance
		}

		log.Printf(
			"[%s] Swap [%d] two balance %s is %f, operation SELL %s",
			swapChain.SwapTwo.Symbol,
			swapAction.Id,
			assetTwo,
			balance,
			swapAction.SwapTwoSymbol,
		)

		swapPair, err := s.SwapRepository.GetSwapPairBySymbol(swapAction.SwapTwoSymbol)
		var binanceOrder ExchangeModel.BinanceOrder

		if swapChain.IsSSB() {
			binanceOrder, err = s.Binance.LimitOrder(
				swapAction.SwapTwoSymbol,
				s.Formatter.FormatQuantity(swapPair, quantity),
				s.Formatter.FormatPrice(swapPair, swapAction.SwapTwoPrice),
				"SELL",
				"GTC",
			)
		}

		if swapChain.IsSBB() || swapChain.IsSBS() {
			binanceOrder, err = s.Binance.LimitOrder(
				swapAction.SwapTwoSymbol,
				s.Formatter.FormatQuantity(swapPair, quantity/swapAction.SwapTwoPrice),
				s.Formatter.FormatPrice(swapPair, swapAction.SwapTwoPrice),
				"BUY",
				"GTC",
			)
		}

		if err != nil {
			log.Printf(
				"[%s] Swap [%d] error: %s",
				swapAction.SwapTwoSymbol,
				swapAction.Id,
				err.Error(),
			)
			return nil
		}

		swapTwoOrder = &binanceOrder
		swapAction.SwapTwoExternalId = &binanceOrder.OrderId
		nowTimestamp := time.Now().Unix()
		swapAction.SwapTwoTimestamp = &nowTimestamp
		swapAction.SwapTwoExternalStatus = &binanceOrder.Status
		_ = s.SwapRepository.UpdateSwapAction(*swapAction)
	} else {
		binanceOrder, err := s.Binance.QueryOrder(swapAction.SwapTwoSymbol, *swapAction.SwapTwoExternalId)
		if err != nil {
			log.Printf("[%s] Swap error: %s", swapAction.SwapTwoSymbol, err.Error())
			return nil
		}

		if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
			swapAction.SwapTwoExternalId = nil
			swapAction.SwapTwoTimestamp = nil
			swapAction.SwapTwoExternalStatus = nil
			_ = s.SwapRepository.UpdateSwapAction(*swapAction)
			s.BalanceService.InvalidateBalanceCache(assetTwo)

			return nil
		}

		swapTwoOrder = &binanceOrder
		swapAction.SwapTwoExternalStatus = &binanceOrder.Status
		_ = s.SwapRepository.UpdateSwapAction(*swapAction)
	}

	// todo: if expired, clear and call recursively
	if !swapTwoOrder.IsFilled() {
		s.TimeoutService.WaitSeconds(5)
		for {
			binanceOrder, err := s.Binance.QueryOrder(swapTwoOrder.Symbol, swapTwoOrder.OrderId)
			if err != nil {
				log.Printf("[%s] Swap Binance error: %s", swapAction.SwapTwoSymbol, err.Error())

				continue
			}

			swapPair, err := s.SwapRepository.GetSwapPairBySymbol(binanceOrder.Symbol)

			log.Printf(
				"[%s] Swap [%d] two [%s] processing, status %s [%d], price %f, current = %f, Executed %f of %f",
				swapAction.SwapTwoSymbol,
				swapAction.Id,
				binanceOrder.Side,
				binanceOrder.Status,
				binanceOrder.OrderId,
				binanceOrder.Price,
				swapPair.SellPrice,
				binanceOrder.ExecutedQty,
				binanceOrder.OrigQty,
			)

			// update value, set new memory address
			swapTwoOrder = &binanceOrder

			nowTimestamp := time.Now().Unix()
			swapAction.SwapTwoTimestamp = &nowTimestamp
			swapAction.SwapTwoExternalStatus = &binanceOrder.Status
			_ = s.SwapRepository.UpdateSwapAction(*swapAction)

			if binanceOrder.IsFilled() {
				break
			}

			if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
				swapAction.SwapTwoExternalId = nil
				swapAction.SwapTwoTimestamp = nil
				swapAction.SwapTwoExternalStatus = nil
				_ = s.SwapRepository.UpdateSwapAction(*swapAction)
				s.BalanceService.InvalidateBalanceCache(assetTwo)

				return nil
			}

			if binanceOrder.IsPartiallyFilled() {
				s.TimeoutService.WaitSeconds(7)
			} else {
				s.TimeoutService.WaitSeconds(15)
			}
		}
	}

	return swapTwoOrder
}

func (s *SwapExecutor) ExecuteSwapThree(
	swapAction *ExchangeModel.SwapAction,
	swapChain ExchangeModel.SwapChainEntity,
	swapTwoOrder ExchangeModel.BinanceOrder,
	assetTwo string,
) *ExchangeModel.BinanceOrder {
	assetThree := strings.ReplaceAll(swapTwoOrder.Symbol, assetTwo, "")
	var swapThreeOrder *ExchangeModel.BinanceOrder = nil

	if swapAction.SwapThreeExternalId == nil {
		balance, _ := s.BalanceService.GetAssetBalance(assetThree, false)
		quantity := swapTwoOrder.ExecutedQty * swapTwoOrder.Price

		if swapChain.IsSBS() {
			quantity = swapTwoOrder.ExecutedQty
		}

		if quantity > balance {
			quantity = balance
		}

		log.Printf(
			"[%s] Swap [%d] three balance %s is %f, operation BUY %s",
			swapChain.SwapThree.Symbol,
			swapAction.Id,
			assetThree,
			balance,
			swapAction.SwapThreeSymbol,
		)

		swapPair, err := s.SwapRepository.GetSwapPairBySymbol(swapAction.SwapThreeSymbol)
		var binanceOrder ExchangeModel.BinanceOrder

		if swapChain.IsSSB() || swapChain.IsSBB() {
			binanceOrder, err = s.Binance.LimitOrder(
				swapAction.SwapThreeSymbol,
				s.Formatter.FormatQuantity(swapPair, quantity/swapAction.SwapThreePrice),
				s.Formatter.FormatPrice(swapPair, swapAction.SwapThreePrice),
				"BUY",
				"GTC",
			)
		}

		if swapChain.IsSBS() {
			binanceOrder, err = s.Binance.LimitOrder(
				swapAction.SwapThreeSymbol,
				s.Formatter.FormatQuantity(swapPair, quantity),
				s.Formatter.FormatPrice(swapPair, swapAction.SwapThreePrice),
				"SELL",
				"GTC",
			)
		}

		if err != nil {
			log.Printf(
				"[%s] Swap [%d] three error: %s",
				swapChain.SwapThree.Symbol,
				swapAction.Id,
				err.Error(),
			)
			return nil
		}

		swapThreeOrder = &binanceOrder
		swapAction.SwapThreeExternalId = &binanceOrder.OrderId
		nowTimestamp := time.Now().Unix()
		swapAction.SwapThreeTimestamp = &nowTimestamp
		swapAction.SwapThreeExternalStatus = &binanceOrder.Status
		_ = s.SwapRepository.UpdateSwapAction(*swapAction)
	} else {
		binanceOrder, err := s.Binance.QueryOrder(swapAction.SwapThreeSymbol, *swapAction.SwapThreeExternalId)
		if err != nil {
			log.Printf("[%s] Swap error: %s", swapChain.SwapThree.Symbol, err.Error())
			return nil
		}

		if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
			swapAction.SwapThreeExternalId = nil
			swapAction.SwapThreeTimestamp = nil
			swapAction.SwapThreeExternalStatus = nil
			_ = s.SwapRepository.UpdateSwapAction(*swapAction)
			s.BalanceService.InvalidateBalanceCache(assetThree)

			return nil
		}

		swapThreeOrder = &binanceOrder
		swapAction.SwapThreeExternalStatus = &binanceOrder.Status
		_ = s.SwapRepository.UpdateSwapAction(*swapAction)
	}

	// todo: if expired, clear and call recursively
	if !swapThreeOrder.IsFilled() {
		s.TimeoutService.WaitSeconds(5)
		for {
			binanceOrder, err := s.Binance.QueryOrder(swapThreeOrder.Symbol, swapThreeOrder.OrderId)
			if err != nil {
				log.Printf("[%s] Swap Binance error: %s", swapChain.SwapThree.Symbol, err.Error())

				continue
			}

			swapPair, err := s.SwapRepository.GetSwapPairBySymbol(binanceOrder.Symbol)

			log.Printf(
				"[%s] Swap [%d] three [%s] processing, status %s [%d], price %f, current = %f, Executed %f of %f",
				swapAction.SwapThreeSymbol,
				swapAction.Id,
				binanceOrder.Side,
				binanceOrder.Status,
				binanceOrder.OrderId,
				binanceOrder.Price,
				swapPair.BuyPrice,
				binanceOrder.ExecutedQty,
				binanceOrder.OrigQty,
			)

			// update value, set new memory address
			swapThreeOrder = &binanceOrder

			nowTimestamp := time.Now().Unix()
			swapAction.SwapThreeTimestamp = &nowTimestamp
			swapAction.SwapThreeExternalStatus = &binanceOrder.Status
			_ = s.SwapRepository.UpdateSwapAction(*swapAction)

			if binanceOrder.IsFilled() {
				break
			}

			if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
				swapAction.SwapThreeExternalId = nil
				swapAction.SwapThreeTimestamp = nil
				swapAction.SwapThreeExternalStatus = nil
				_ = s.SwapRepository.UpdateSwapAction(*swapAction)
				s.BalanceService.InvalidateBalanceCache(assetThree)

				return nil
			}

			if binanceOrder.IsPartiallyFilled() {
				swapAction.EndQuantity = &binanceOrder.ExecutedQty
				_ = s.SwapRepository.UpdateSwapAction(*swapAction)

				if (nowTimestamp-swapAction.StartTimestamp) > (3600*4) && binanceOrder.IsNearlyFilled() {
					break // Do not cancel order, but check it later...
				}
				s.TimeoutService.WaitSeconds(7)
			} else {
				s.TimeoutService.WaitSeconds(15)
			}
		}
	}

	return swapThreeOrder
}
