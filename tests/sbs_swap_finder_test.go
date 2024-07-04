package tests

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"gitlab.com/open-soft/go-crypto-bot/src/validator"
	"os"
	"testing"
	"time"
)

func TestSwapSellBuySell(t *testing.T) {
	exchangeRepoMock := new(ExchangeRepositoryMock)

	b, err := os.ReadFile("swap_pair_sbs.json") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	var options []model.SwapPair
	err = json.Unmarshal(b, &options)
	if err != nil {
		panic(err)
	}

	options0 := make([]model.SwapPair, 0)
	options[0].PriceTimestamp = time.Now().Unix() + 3600
	options0 = append(options0, options[0])

	options1 := make([]model.SwapPair, 0)
	options[1].PriceTimestamp = time.Now().Unix() + 3600
	options1 = append(options1, options[1])

	options2 := make([]model.SwapPair, 0)
	options[1].PriceTimestamp = time.Now().Unix() + 3600
	options[2].PriceTimestamp = time.Now().Unix() + 3600
	options2 = append(options2, options[1])
	options2 = append(options2, options[2])

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "ETH").Return(options0)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "BTC").Return(options1)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "XRP").Return(options2)

	sbsFinder := exchange.SBSSwapFinder{
		ExchangeRepository: exchangeRepoMock,
		Formatter:          &utils.Formatter{},
	}

	chain := sbsFinder.Find("ETH").BestChain
	assertion := assert.New(t)
	assertion.Equal(2.68, chain.Percent.Value())
	assertion.Equal("SBS", chain.Type)
	assertion.Equal("ETH sell-> BTC buy-> XRP sell-> ETH", chain.Title)
	assertion.Equal("ETHBTC", chain.SwapOne.Symbol)
	assertion.Equal(0.05355, chain.SwapOne.Price)
	assertion.Equal("XRPBTC", chain.SwapTwo.Symbol)
	assertion.Equal(0.00001436, chain.SwapTwo.Price)
	assertion.Equal("XRPETH", chain.SwapThree.Symbol)
	assertion.Equal(0.000277, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price*chain.SwapThree.Price, 103.29)

	// validate
	swapRepoMock := new(SwapRepositoryMock)

	swapRepoMock.On("GetSwapPairBySymbol", "ETHBTC").Return(options0[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPBTC").Return(options1[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPETH").Return(options2[1], nil)

	binance := new(ExchangePriceAPIMock)

	binance.On("GetKLinesCached", "ETHBTC", "1d", int64(14)).Return([]model.KLine{
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
	})
	binance.On("GetKLinesCached", "XRPBTC", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
	})
	binance.On("GetKLinesCached", "XRPETH", "1d", int64(14)).Return([]model.KLine{
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
	})

	botService := new(BotServiceMock)
	botService.On("GetSwapConfig").Return(model.SwapConfig{
		MinValidPercent: 0.1,
		HistoryInterval: "1d",
		HistoryPeriod:   14,
	})

	swapChainBuilder := exchange.SwapChainBuilder{}
	validator := validator.SwapValidator{
		Binance:        binance,
		SwapRepository: swapRepoMock,
		Formatter:      &utils.Formatter{},
		BotService:     botService,
	}

	order := model.Order{
		ExecutedQuantity: 100,
	}

	swapChain := swapChainBuilder.BuildEntity(*chain, chain.Percent, 0, 0, 0, 0, 0, 0)

	err = validator.Validate(swapChain, order)
	assertion.Nil(err)

	// execute
	balanceServiceMock := new(BalanceServiceMock)
	orderRepositoryMock := new(OrderUpdaterMock)
	binanceMock := new(ExchangeOrderAPIMock)

	assetBalance := order.ExecutedQuantity

	balanceServiceMock.On("GetAssetBalance", "ETH", false).Times(1).Return(order.ExecutedQuantity, nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	swapRepoMock.On("GetActiveSwapAction", order).Return(model.SwapAction{
		Id:              999,
		OrderId:         order.Id,
		BotId:           1,
		SwapChainId:     swapChain.Id,
		Asset:           swapChain.SwapOne.BaseAsset,
		Status:          model.SwapActionStatusPending,
		StartTimestamp:  time.Now().Unix(),
		StartQuantity:   assetBalance,
		SwapOneSymbol:   swapChain.SwapOne.GetSymbol(),
		SwapOnePrice:    swapChain.SwapOne.Price,
		SwapTwoSymbol:   swapChain.SwapTwo.GetSymbol(),
		SwapTwoPrice:    swapChain.SwapTwo.Price,
		SwapThreeSymbol: swapChain.SwapThree.GetSymbol(),
		SwapThreePrice:  swapChain.SwapThree.Price,
	}, nil)
	swapRepoMock.On("GetSwapChainById", swapChain.Id).Return(swapChain, nil)

	binanceMock.On("LimitOrder", "ETHBTC", 100.00, 0.05355, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             "12",
		Symbol:              "ETHBTC",
		ExecutedQty:         0.00,
		OrigQty:             100.00,
		Price:               0.05355,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "ETHBTC", "12").Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             "12",
		ExecutedQty:         80.00,
		OrigQty:             100.00,
		Symbol:              "ETHBTC",
		Price:               0.05355,
		Side:                "SELL",
		CummulativeQuoteQty: 80.00 * 0.05358,
	}, nil)
	binanceMock.On("QueryOrder", "ETHBTC", "12").Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             "12",
		ExecutedQty:         100.00,
		OrigQty:             100.00,
		Symbol:              "ETHBTC",
		Price:               0.05355,
		Side:                "SELL",
		CummulativeQuoteQty: 100.00 * 0.05355,
	}, nil)

	btcInitialBalance := 1.3455
	balanceServiceMock.On("GetAssetBalance", "BTC", false).Return(5.355+btcInitialBalance, nil)

	binanceMock.On("LimitOrder", "XRPBTC", 372910.00, 0.00001436, "BUY", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             "13",
		Symbol:              "XRPBTC",
		ExecutedQty:         0.00,
		OrigQty:             372910.00,
		Price:               0.00001436,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPBTC", "13").Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             "13",
		Symbol:              "XRPBTC",
		ExecutedQty:         125210.00,
		OrigQty:             372910.00,
		Price:               0.00001436,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00001436 * 125210.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPBTC", "13").Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             "13",
		Symbol:              "XRPBTC",
		ExecutedQty:         372910.00,
		OrigQty:             372910.00,
		Price:               0.00001436,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00001436 * 372910.00,
	}, nil)

	xrpInitialBalance := 4000.00
	balanceServiceMock.On("GetAssetBalance", "XRP", false).Return(372910.00+xrpInitialBalance, nil)

	binanceMock.On("LimitOrder", "XRPETH", 372910.00, 0.000277, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             "14",
		Symbol:              "XRPETH",
		ExecutedQty:         0.00,
		OrigQty:             372910.00,
		Price:               0.000277,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPETH", "14").Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             "14",
		Symbol:              "XRPETH",
		ExecutedQty:         125210.00,
		OrigQty:             372910.00,
		Price:               0.000277,
		Side:                "SELL",
		CummulativeQuoteQty: 0.000277 * 125210.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPETH", "14").Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             "14",
		Symbol:              "XRPETH",
		ExecutedQty:         372910.00,
		OrigQty:             372910.00,
		Price:               0.000277,
		Side:                "SELL",
		CummulativeQuoteQty: 0.000277 * 372910.00,
	}, nil)

	orderRepositoryMock.On("Update", mock.Anything).Once().Return(nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	balanceServiceMock.On("InvalidateBalanceCache", "ETH").Once()
	ethInitialBalance := 10.99
	balanceServiceMock.On("GetAssetBalance", "ETH", false).Times(2).Return(103.29607+ethInitialBalance, nil)

	timeServiceMock := new(TimeServiceMock)
	timeServiceMock.On("WaitSeconds", int64(5)).Times(3)
	timeServiceMock.On("WaitSeconds", int64(7)).Times(3)
	timeServiceMock.On("GetNowDiffMinutes", mock.Anything).Return(0.50)

	executor := exchange.SwapExecutor{
		SwapRepository:  swapRepoMock,
		OrderRepository: orderRepositoryMock,
		BalanceService:  balanceServiceMock,
		Binance:         binanceMock,
		TimeService:     timeServiceMock,
		Formatter:       &utils.Formatter{},
	}

	executor.Execute(order)

	assertion.Equal(103.29607, *swapRepoMock.swapAction.EndQuantity)
	assertion.Equal("12", *swapRepoMock.swapAction.SwapOneExternalId)
	assertion.Equal(0.05355, swapRepoMock.swapAction.SwapOnePrice)
	assertion.Equal("ETHBTC", swapRepoMock.swapAction.SwapOneSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapOneExternalStatus)
	assertion.Equal("13", *swapRepoMock.swapAction.SwapTwoExternalId)
	assertion.Equal(0.00001436, swapRepoMock.swapAction.SwapTwoPrice)
	assertion.Equal("XRPBTC", swapRepoMock.swapAction.SwapTwoSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapTwoExternalStatus)
	assertion.Equal("14", *swapRepoMock.swapAction.SwapThreeExternalId)
	assertion.Equal(0.000277, swapRepoMock.swapAction.SwapThreePrice)
	assertion.Equal("XRPETH", swapRepoMock.swapAction.SwapThreeSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapThreeExternalStatus)
}

func TestSwapSellBuySellForceSwap(t *testing.T) {
	exchangeRepoMock := new(ExchangeRepositoryMock)

	b, err := os.ReadFile("swap_pair_sbs.json") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	var options []model.SwapPair
	err = json.Unmarshal(b, &options)
	if err != nil {
		panic(err)
	}

	options0 := make([]model.SwapPair, 0)
	options[0].PriceTimestamp = time.Now().Unix() + 3600
	options0 = append(options0, options[0])

	options1 := make([]model.SwapPair, 0)
	options[1].PriceTimestamp = time.Now().Unix() + 3600
	options1 = append(options1, options[1])

	options2 := make([]model.SwapPair, 0)
	options[1].PriceTimestamp = time.Now().Unix() + 3600
	options[2].PriceTimestamp = time.Now().Unix() + 3600
	options2 = append(options2, options[1])
	options2 = append(options2, options[2])

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "ETH").Return(options0)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "BTC").Return(options1)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "XRP").Return(options2)

	sbsFinder := exchange.SBSSwapFinder{
		ExchangeRepository: exchangeRepoMock,
		Formatter:          &utils.Formatter{},
	}

	chain := sbsFinder.Find("ETH").BestChain
	assertion := assert.New(t)
	assertion.Equal(2.68, chain.Percent.Value())
	assertion.Equal("SBS", chain.Type)
	assertion.Equal("ETH sell-> BTC buy-> XRP sell-> ETH", chain.Title)
	assertion.Equal("ETHBTC", chain.SwapOne.Symbol)
	assertion.Equal(0.05355, chain.SwapOne.Price)
	assertion.Equal("XRPBTC", chain.SwapTwo.Symbol)
	assertion.Equal(0.00001436, chain.SwapTwo.Price)
	assertion.Equal("XRPETH", chain.SwapThree.Symbol)
	assertion.Equal(0.000277, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price*chain.SwapThree.Price, 103.296)

	// validate
	swapRepoMock := new(SwapRepositoryMock)

	swapRepoMock.On("GetSwapPairBySymbol", "ETHBTC").Return(options0[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPBTC").Return(options1[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPETH").Return(options2[1], nil)

	binance := new(ExchangePriceAPIMock)

	binance.On("GetKLinesCached", "ETHBTC", "1d", int64(14)).Return([]model.KLine{
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
	})
	binance.On("GetKLinesCached", "XRPBTC", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
	})
	binance.On("GetKLinesCached", "XRPETH", "1d", int64(14)).Return([]model.KLine{
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
	})

	botService := new(BotServiceMock)
	botService.On("GetSwapConfig").Return(model.SwapConfig{
		MinValidPercent: 0.1,
		HistoryInterval: "1d",
		HistoryPeriod:   14,
	})

	swapChainBuilder := exchange.SwapChainBuilder{}
	validator := validator.SwapValidator{
		Binance:        binance,
		SwapRepository: swapRepoMock,
		Formatter:      &utils.Formatter{},
		BotService:     botService,
	}

	order := model.Order{
		ExecutedQuantity: 100,
	}

	swapChain := swapChainBuilder.BuildEntity(*chain, chain.Percent, 0, 0, 0, 0, 0, 0)

	err = validator.Validate(swapChain, order)
	assertion.Nil(err)

	// execute

	balanceServiceMock := new(BalanceServiceMock)
	orderRepositoryMock := new(OrderUpdaterMock)
	binanceMock := new(ExchangeOrderAPIMock)

	assetBalance := order.ExecutedQuantity

	swapRepoMock.On("GetSwapPairBySymbol", "ETHBTC").Times(2).Return(model.SwapPair{
		Symbol:      "ETHBTC",
		SellPrice:   0.0536,
		MinNotional: 0.0001,
		MinQuantity: 0.0001,
		MinPrice:    0.00001,
	}, nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPBTC").Times(2).Return(model.SwapPair{
		Symbol:      "XRPBTC",
		BuyPrice:    0.00001426,
		MinNotional: 0.0001,
		MinQuantity: 1,
		MinPrice:    0.00000001,
	}, nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPETH").Times(2).Return(model.SwapPair{
		Symbol:      "XRPETH",
		SellPrice:   0.0002785,
		MinNotional: 0.001,
		MinQuantity: 1,
		MinPrice:    0.0000001,
	}, nil)

	balanceServiceMock.On("GetAssetBalance", "ETH", false).Times(1).Return(order.ExecutedQuantity, nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	swapRepoMock.On("GetActiveSwapAction", order).Return(model.SwapAction{
		Id:              999,
		OrderId:         order.Id,
		BotId:           1,
		SwapChainId:     swapChain.Id,
		Asset:           swapChain.SwapOne.BaseAsset,
		Status:          model.SwapActionStatusPending,
		StartTimestamp:  time.Now().Unix(),
		StartQuantity:   assetBalance,
		SwapOneSymbol:   swapChain.SwapOne.GetSymbol(),
		SwapOnePrice:    swapChain.SwapOne.Price,
		SwapTwoSymbol:   swapChain.SwapTwo.GetSymbol(),
		SwapTwoPrice:    swapChain.SwapTwo.Price,
		SwapThreeSymbol: swapChain.SwapThree.GetSymbol(),
		SwapThreePrice:  swapChain.SwapThree.Price,
	}, nil)
	swapRepoMock.On("GetSwapChainById", swapChain.Id).Return(swapChain, nil)

	binanceMock.On("LimitOrder", "ETHBTC", 100.00, 0.05355, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             "12",
		Symbol:              "ETHBTC",
		ExecutedQty:         0.00,
		OrigQty:             100.00,
		Price:               0.05355,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "ETHBTC", "12").Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             "12",
		ExecutedQty:         80.00,
		OrigQty:             100.00,
		Symbol:              "ETHBTC",
		Price:               0.05355,
		Side:                "SELL",
		CummulativeQuoteQty: 80.00 * 0.05355,
	}, nil)
	binanceMock.On("QueryOrder", "ETHBTC", "12").Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             "12",
		ExecutedQty:         100.00,
		OrigQty:             100.00,
		Symbol:              "ETHBTC",
		Price:               0.05355,
		Side:                "SELL",
		CummulativeQuoteQty: 100.00 * 0.05355,
	}, nil)

	btcInitialBalance := 1.3455
	balanceServiceMock.On("GetAssetBalance", "BTC", false).Return(5.355+btcInitialBalance, nil)

	binanceMock.On("LimitOrder", "XRPBTC", 372910.00, 0.00001436, "BUY", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             "13",
		Symbol:              "XRPBTC",
		ExecutedQty:         0.00,
		OrigQty:             372910.00,
		Price:               0.00001436,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPBTC", "13").Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             "13",
		Symbol:              "XRPBTC",
		ExecutedQty:         125210.00,
		OrigQty:             372910.00,
		Price:               0.00001436,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00001436 * 125210.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPBTC", "13").Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             "13",
		Symbol:              "XRPBTC",
		ExecutedQty:         372910.00,
		OrigQty:             372910.00,
		Price:               0.00001436,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00001436 * 372910.00,
	}, nil)

	xrpInitialBalance := 4000.00
	balanceServiceMock.On("GetAssetBalance", "XRP", false).Return(372910.00+xrpInitialBalance, nil)

	binanceMock.On("LimitOrder", "XRPETH", 372910.00, 0.000277, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             "14",
		Symbol:              "XRPETH",
		ExecutedQty:         0.00,
		OrigQty:             372910.00,
		Price:               0.000277,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPETH", "14").Times(1).Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             "14",
		Symbol:              "XRPETH",
		ExecutedQty:         0.00,
		OrigQty:             372910.00,
		Price:               0.000277,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("CancelOrder", "XRPETH", "14").Return(model.BinanceOrder{
		Status:              "CANCELED",
		OrderId:             "14",
		Symbol:              "XRPETH",
		ExecutedQty:         0.00,
		OrigQty:             372910.00,
		Price:               0.000277,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("LimitOrder", "XRPETH", 372910.00, 0.0002784, "SELL", "IOC").Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             "14",
		Symbol:              "XRPETH",
		ExecutedQty:         372910.00,
		OrigQty:             372910.00,
		Price:               0.000277,
		Side:                "SELL",
		CummulativeQuoteQty: 372910.00 * 0.0002784,
	}, nil)

	orderRepositoryMock.On("Update", mock.Anything).Once().Return(nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	balanceServiceMock.On("InvalidateBalanceCache", "ETH").Once()
	ethInitialBalance := 10.99
	balanceServiceMock.On("GetAssetBalance", "ETH", false).Times(2).Return(103.818144+ethInitialBalance, nil)

	timeServiceMock := new(TimeServiceMock)
	timeServiceMock.On("WaitSeconds", int64(5)).Times(3)
	timeServiceMock.On("WaitSeconds", int64(7)).Times(3)
	timeServiceMock.On("GetNowDiffMinutes", mock.Anything).Return(50.00)

	executor := exchange.SwapExecutor{
		SwapRepository:  swapRepoMock,
		OrderRepository: orderRepositoryMock,
		BalanceService:  balanceServiceMock,
		Binance:         binanceMock,
		TimeService:     timeServiceMock,
		Formatter:       &utils.Formatter{},
	}

	executor.Execute(order)

	assertion.Equal(103.818144, *swapRepoMock.swapAction.EndQuantity)
	assertion.Equal("12", *swapRepoMock.swapAction.SwapOneExternalId)
	assertion.Equal("ETHBTC", swapRepoMock.swapAction.SwapOneSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapOneExternalStatus)
	assertion.Equal("13", *swapRepoMock.swapAction.SwapTwoExternalId)
	assertion.Equal("XRPBTC", swapRepoMock.swapAction.SwapTwoSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapTwoExternalStatus)
	assertion.Equal("14", *swapRepoMock.swapAction.SwapThreeExternalId)
	assertion.Equal("XRPETH", swapRepoMock.swapAction.SwapThreeSymbol)
	assertion.Equal("FILLED_FORCE", *swapRepoMock.swapAction.SwapThreeExternalStatus)
}
