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

func TestSwapSellBuyBuy(t *testing.T) {
	exchangeRepoMock := new(ExchangeRepositoryMock)

	b, err := os.ReadFile("swap_pair_sbb.json") // just pass the file name
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

	options3 := make([]model.SwapPair, 0)
	options3 = append(options3, options[0])
	options3 = append(options3, options[2])

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "SOL").Return(options3)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "ETH").Return(options0)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)
	//exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)

	swapManager := exchange.SBBSwapFinder{
		Formatter:          &utils.Formatter{},
		ExchangeRepository: exchangeRepoMock,
	}

	chain := swapManager.Find("SOL").BestChain
	assertion := assert.New(t)
	assertion.Equal(3.98, chain.Percent.Value())
	assertion.Equal("SBB", chain.Type)
	assertion.Equal("SOL sell-> GBP buy-> ETH buy-> SOL", chain.Title)
	assertion.Equal("SOLGBP", chain.SwapOne.Symbol)
	assertion.Equal(58.53, chain.SwapOne.Price)
	assertion.Equal("ETHGBP", chain.SwapTwo.Symbol)
	assertion.Equal(1783.04, chain.SwapTwo.Price)
	assertion.Equal("SOLETH", chain.SwapThree.Symbol)
	assertion.Equal(0.03138, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price/chain.SwapThree.Price, 104.6)

	// validate
	swapRepoMock := new(SwapRepositoryMock)

	swapRepoMock.On("GetSwapPairBySymbol", "SOLETH").Return(options0[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "ETHGBP").Return(options2[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "SOLGBP").Return(options2[1], nil)

	binance := new(ExchangePriceAPIMock)

	binance.On("GetKLinesCached", "SOLGBP", "1d", int64(14)).Return([]model.KLine{
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
	})
	binance.On("GetKLinesCached", "ETHGBP", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
	})
	binance.On("GetKLinesCached", "SOLETH", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
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

	balanceServiceMock.On("GetAssetBalance", "SOL", false).Times(1).Return(order.ExecutedQuantity, nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	swapRepoMock.On("GetActiveSwapAction", order).Return(model.SwapAction{
		Id:              990,
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

	binanceMock.On("LimitOrder", "SOLGBP", 100.00, 58.53, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             "19",
		Symbol:              "SOLGBP",
		ExecutedQty:         0.00,
		OrigQty:             100.00,
		Price:               58.53,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "SOLGBP", "19").Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             "19",
		ExecutedQty:         80.00,
		OrigQty:             100.00,
		Symbol:              "SOLGBP",
		Price:               58.53,
		Side:                "SELL",
		CummulativeQuoteQty: 80 * 58.53,
	}, nil)
	binanceMock.On("QueryOrder", "SOLGBP", "19").Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             "19",
		ExecutedQty:         100.00,
		OrigQty:             100.00,
		Symbol:              "SOLGBP",
		Price:               58.53,
		Side:                "SELL",
		CummulativeQuoteQty: 100 * 58.53,
	}, nil)

	gbpInitialBalance := 50.99
	balanceServiceMock.On("GetAssetBalance", "GBP", false).Return(5856.00+gbpInitialBalance, nil)

	binanceMock.On("LimitOrder", "ETHGBP", 3.2825, 1783.04, "BUY", "GTC").Return(model.BinanceOrder{
		Status:      "NEW",
		OrderId:     "20",
		Symbol:      "ETHGBP",
		ExecutedQty: 0.00,
		OrigQty:     3.2825,
		Price:       1783.04,
	}, nil)
	binanceMock.On("QueryOrder", "ETHGBP", "20").Times(1).Return(model.BinanceOrder{
		Status:      "PARTIALLY_FILLED",
		OrderId:     "20",
		Symbol:      "ETHGBP",
		ExecutedQty: 1.272,
		OrigQty:     3.2825,
		Price:       1783.04,
	}, nil)
	binanceMock.On("QueryOrder", "ETHGBP", "20").Times(2).Return(model.BinanceOrder{
		Status:      "FILLED",
		OrderId:     "20",
		Symbol:      "ETHGBP",
		ExecutedQty: 3.2825,
		OrigQty:     3.2825,
		Price:       1783.04,
	}, nil)
	ethInitialBalance := 2.99
	balanceServiceMock.On("GetAssetBalance", "ETH", false).Return(3.282+ethInitialBalance, nil)

	binanceMock.On("LimitOrder", "SOLETH", 104.604, 0.03138, "BUY", "GTC").Return(model.BinanceOrder{
		Status:      "NEW",
		OrderId:     "21",
		Symbol:      "SOLETH",
		ExecutedQty: 0.00,
		OrigQty:     104.604,
		Price:       0.03138,
	}, nil)
	binanceMock.On("QueryOrder", "SOLETH", "21").Times(1).Return(model.BinanceOrder{
		Status:      "PARTIALLY_FILLED",
		OrderId:     "21",
		Symbol:      "SOLETH",
		ExecutedQty: 12.00,
		OrigQty:     104.604,
		Price:       0.03138,
	}, nil)
	binanceMock.On("QueryOrder", "SOLETH", "21").Times(2).Return(model.BinanceOrder{
		Status:      "FILLED",
		OrderId:     "21",
		Symbol:      "SOLETH",
		ExecutedQty: 104.604,
		OrigQty:     104.604,
		Price:       0.03138,
	}, nil)

	orderRepositoryMock.On("Update", mock.Anything).Once().Return(nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	balanceServiceMock.On("InvalidateBalanceCache", "SOL").Once()
	solInitialBalance := 50.00
	balanceServiceMock.On("GetAssetBalance", "SOL", false).Times(2).Return(104.72+solInitialBalance, nil)

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

	assertion.Equal(model.SwapActionStatusSuccess, swapRepoMock.swapAction.Status)
	assertion.Equal(104.604, *swapRepoMock.swapAction.EndQuantity)
	assertion.Equal("19", *swapRepoMock.swapAction.SwapOneExternalId)
	assertion.Equal("SOLGBP", swapRepoMock.swapAction.SwapOneSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapOneExternalStatus)
	assertion.Equal("20", *swapRepoMock.swapAction.SwapTwoExternalId)
	assertion.Equal("ETHGBP", swapRepoMock.swapAction.SwapTwoSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapTwoExternalStatus)
	assertion.Equal("21", *swapRepoMock.swapAction.SwapThreeExternalId)
	assertion.Equal("SOLETH", swapRepoMock.swapAction.SwapThreeSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapThreeExternalStatus)
}

func TestSwapSellBuyBuyRollback(t *testing.T) {
	exchangeRepoMock := new(ExchangeRepositoryMock)

	b, err := os.ReadFile("swap_pair_sbb.json") // just pass the file name
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

	options3 := make([]model.SwapPair, 0)
	options3 = append(options3, options[0])
	options3 = append(options3, options[2])

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "SOL").Return(options3)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "ETH").Return(options0)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)
	//exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)

	swapManager := exchange.SBBSwapFinder{
		Formatter:          &utils.Formatter{},
		ExchangeRepository: exchangeRepoMock,
	}

	chain := swapManager.Find("SOL").BestChain
	assertion := assert.New(t)
	assertion.Equal(3.98, chain.Percent.Value())
	assertion.Equal("SBB", chain.Type)
	assertion.Equal("SOL sell-> GBP buy-> ETH buy-> SOL", chain.Title)
	assertion.Equal("SOLGBP", chain.SwapOne.Symbol)
	assertion.Equal(58.53, chain.SwapOne.Price)
	assertion.Equal("ETHGBP", chain.SwapTwo.Symbol)
	assertion.Equal(1783.04, chain.SwapTwo.Price)
	assertion.Equal("SOLETH", chain.SwapThree.Symbol)
	assertion.Equal(0.03138, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price/chain.SwapThree.Price, 104.6)

	// validate
	swapRepoMock := new(SwapRepositoryMock)

	swapRepoMock.On("GetSwapPairBySymbol", "SOLETH").Times(1).Return(options0[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "ETHGBP").Times(1).Return(options2[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "SOLGBP").Times(1).Return(options2[1], nil)

	binance := new(ExchangePriceAPIMock)

	binance.On("GetKLinesCached", "SOLGBP", "1d", int64(14)).Return([]model.KLine{
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
	})
	binance.On("GetKLinesCached", "ETHGBP", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
	})
	binance.On("GetKLinesCached", "SOLETH", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
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

	swapRepoMock.On("GetSwapPairBySymbol", "SOLETH").Times(2).Return(model.SwapPair{
		Symbol:      "SOLETH",
		SellPrice:   0.03131,
		MinNotional: 0.001,
		MinQuantity: 0.001,
		MinPrice:    0.00001,
	}, nil)
	swapRepoMock.On("GetSwapPairBySymbol", "ETHGBP").Times(2).Return(model.SwapPair{
		Symbol:      "ETHGBP",
		SellPrice:   1712.96,
		BuyPrice:    1711.04,
		MinNotional: 5,
		MinQuantity: 0.0001,
		MinPrice:    0.01,
	}, nil)
	swapRepoMock.On("GetSwapPairBySymbol", "SOLGBP").Times(2).Return(model.SwapPair{
		Symbol:      "SOLGBP",
		BuyPrice:    58.30,
		MinNotional: 5,
		MinQuantity: 0.01,
		MinPrice:    0.01,
	}, nil)

	balanceServiceMock.On("GetAssetBalance", "SOL", false).Times(1).Return(order.ExecutedQuantity, nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	swapRepoMock.On("GetActiveSwapAction", order).Return(model.SwapAction{
		Id:              990,
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

	gbpInitialBalance := 50.99
	balanceServiceMock.On("GetAssetBalance", "GBP", false).Return(5856.00+gbpInitialBalance, nil)

	binanceMock.On("LimitOrder", "SOLGBP", 100.00, 58.53, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             "19",
		Symbol:              "SOLGBP",
		ExecutedQty:         0.00,
		OrigQty:             100.00,
		Price:               58.53,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "SOLGBP", "19").Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             "19",
		ExecutedQty:         100.00,
		OrigQty:             100.00,
		Symbol:              "SOLGBP",
		Price:               58.53,
		Side:                "SELL",
		CummulativeQuoteQty: 100 * 58.53,
	}, nil)

	binanceMock.On("LimitOrder", "ETHGBP", 3.4205, 1711.14, "BUY", "GTC").Return(model.BinanceOrder{
		Status:      "NEW",
		OrderId:     "20",
		Symbol:      "ETHGBP",
		ExecutedQty: 0.00,
		OrigQty:     3.4205,
		Price:       1711.14,
	}, nil)
	binanceMock.On("QueryOrder", "ETHGBP", "20").Times(1).Return(model.BinanceOrder{
		Status:      "NEW",
		OrderId:     "20",
		Symbol:      "ETHGBP",
		ExecutedQty: 0.00,
		OrigQty:     3.4205,
		Price:       1711.14,
	}, nil)

	orderRepositoryMock.On("Update", mock.Anything).Once().Return(nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	balanceServiceMock.On("InvalidateBalanceCache", "SOL").Once()
	solInitialBalance := 50.00
	balanceServiceMock.On("GetAssetBalance", "SOL", false).Times(2).Return(104.72+solInitialBalance, nil)

	timeServiceMock := new(TimeServiceMock)
	timeServiceMock.On("WaitSeconds", int64(5)).Times(3)
	timeServiceMock.On("WaitSeconds", int64(15)).Times(1)
	timeServiceMock.On("WaitSeconds", int64(7)).Times(3)
	timeServiceMock.On("GetNowDiffMinutes", mock.Anything).Return(50.00)

	swapRepoMock.On("GetSwapPairBySymbol", "SOLGBP").Times(3).Return(model.SwapPair{
		Symbol:      "SOLGBP",
		BuyPrice:    57.38,
		MinNotional: 5,
		MinQuantity: 0.01,
		MinPrice:    0.01,
	}, nil)

	binanceMock.On("CancelOrder", "ETHGBP", "20").Return(model.BinanceOrder{
		Status:      "CANCELED",
		OrderId:     "20",
		Symbol:      "ETHGBP",
		ExecutedQty: 0.00,
		OrigQty:     3.4205,
		Price:       1711.14,
	}, nil)

	binanceMock.On("LimitOrder", "SOLGBP", 101.98, 57.39, "BUY", "IOC").Return(model.BinanceOrder{
		Status:      "FILLED",
		OrderId:     "21",
		Symbol:      "SOLGBP",
		ExecutedQty: 101.98,
		OrigQty:     101.98,
		Price:       57.39,
	}, nil)

	executor := exchange.SwapExecutor{
		SwapRepository:  swapRepoMock,
		OrderRepository: orderRepositoryMock,
		BalanceService:  balanceServiceMock,
		Binance:         binanceMock,
		TimeService:     timeServiceMock,
		Formatter:       &utils.Formatter{},
	}

	executor.Execute(order)

	assertion.Equal(model.SwapActionStatusSuccess, swapRepoMock.swapAction.Status)
	assertion.Equal(101.98, *swapRepoMock.swapAction.EndQuantity)
	assertion.Equal("19", *swapRepoMock.swapAction.SwapOneExternalId)
	assertion.Equal(58.53, swapRepoMock.swapAction.SwapOnePrice)
	assertion.Equal("SOLGBP", swapRepoMock.swapAction.SwapOneSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapOneExternalStatus)
	assertion.Equal("21", *swapRepoMock.swapAction.SwapTwoExternalId)
	assertion.Equal(57.39, swapRepoMock.swapAction.SwapTwoPrice)
	assertion.Equal("SOLGBP", swapRepoMock.swapAction.SwapTwoSymbol)
	assertion.Equal("FILLED_RB", *swapRepoMock.swapAction.SwapTwoExternalStatus)
	assertion.Nil(swapRepoMock.swapAction.SwapThreeExternalId)
	assertion.Equal(0.03138, swapRepoMock.swapAction.SwapThreePrice)
	assertion.Equal("SOLETH", swapRepoMock.swapAction.SwapThreeSymbol)
	assertion.Nil(swapRepoMock.swapAction.SwapThreeExternalStatus)
}

func TestSwapSellBuyBuyForceSwap(t *testing.T) {
	exchangeRepoMock := new(ExchangeRepositoryMock)

	b, err := os.ReadFile("swap_pair_sbb.json") // just pass the file name
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

	options3 := make([]model.SwapPair, 0)
	options3 = append(options3, options[0])
	options3 = append(options3, options[2])

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "SOL").Return(options3)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "ETH").Return(options0)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)
	//exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)

	swapManager := exchange.SBBSwapFinder{
		Formatter:          &utils.Formatter{},
		ExchangeRepository: exchangeRepoMock,
	}

	chain := swapManager.Find("SOL").BestChain
	assertion := assert.New(t)
	assertion.Equal(3.98, chain.Percent.Value())
	assertion.Equal("SBB", chain.Type)
	assertion.Equal("SOL sell-> GBP buy-> ETH buy-> SOL", chain.Title)
	assertion.Equal("SOLGBP", chain.SwapOne.Symbol)
	assertion.Equal(58.53, chain.SwapOne.Price)
	assertion.Equal("ETHGBP", chain.SwapTwo.Symbol)
	assertion.Equal(1783.04, chain.SwapTwo.Price)
	assertion.Equal("SOLETH", chain.SwapThree.Symbol)
	assertion.Equal(0.03138, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price/chain.SwapThree.Price, 104.6)

	// validate
	swapRepoMock := new(SwapRepositoryMock)

	swapRepoMock.On("GetSwapPairBySymbol", "SOLETH").Times(1).Return(options0[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "ETHGBP").Times(1).Return(options2[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "SOLGBP").Times(1).Return(options2[1], nil)

	binance := new(ExchangePriceAPIMock)

	binance.On("GetKLinesCached", "SOLGBP", "1d", int64(14)).Return([]model.KLine{
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
		{
			High: 58.57,
		},
	})
	binance.On("GetKLinesCached", "ETHGBP", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
		{
			Low: 1782.95,
		},
	})
	binance.On("GetKLinesCached", "SOLETH", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
		},
		{
			Low: 0.03131,
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

	swapRepoMock.On("GetSwapPairBySymbol", "SOLETH").Times(2).Return(model.SwapPair{
		Symbol:      "SOLETH",
		SellPrice:   0.03131,
		BuyPrice:    0.03128,
		MinNotional: 0.001,
		MinQuantity: 0.001,
		MinPrice:    0.00001,
	}, nil)
	swapRepoMock.On("GetSwapPairBySymbol", "ETHGBP").Times(2).Return(model.SwapPair{
		Symbol:      "ETHGBP",
		SellPrice:   1712.96,
		BuyPrice:    1711.04,
		MinNotional: 5,
		MinQuantity: 0.0001,
		MinPrice:    0.01,
	}, nil)
	swapRepoMock.On("GetSwapPairBySymbol", "SOLGBP").Times(2).Return(model.SwapPair{
		Symbol:      "SOLGBP",
		BuyPrice:    58.30,
		MinNotional: 5,
		MinQuantity: 0.01,
		MinPrice:    0.01,
	}, nil)

	balanceServiceMock.On("GetAssetBalance", "SOL", false).Times(1).Return(order.ExecutedQuantity, nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	swapRepoMock.On("GetActiveSwapAction", order).Return(model.SwapAction{
		Id:              990,
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

	gbpInitialBalance := 50.99
	balanceServiceMock.On("GetAssetBalance", "GBP", false).Return(5856.00+gbpInitialBalance, nil)

	binanceMock.On("LimitOrder", "SOLGBP", 100.00, 58.53, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             "19",
		Symbol:              "SOLGBP",
		ExecutedQty:         0.00,
		OrigQty:             100.00,
		Price:               58.53,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "SOLGBP", "19").Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             "19",
		ExecutedQty:         100.00,
		OrigQty:             100.00,
		Symbol:              "SOLGBP",
		Price:               58.53,
		Side:                "SELL",
		CummulativeQuoteQty: 100 * 58.53,
	}, nil)

	binanceMock.On("LimitOrder", "ETHGBP", 3.4205, 1711.14, "BUY", "GTC").Return(model.BinanceOrder{
		Status:      "NEW",
		OrderId:     "20",
		Symbol:      "ETHGBP",
		ExecutedQty: 0.00,
		OrigQty:     3.4205,
		Price:       1711.14,
	}, nil)
	binanceMock.On("QueryOrder", "ETHGBP", "20").Times(1).Return(model.BinanceOrder{
		Status:      "FILLED",
		OrderId:     "20",
		Symbol:      "ETHGBP",
		ExecutedQty: 3.4205,
		OrigQty:     3.4205,
		Price:       1711.14,
	}, nil)

	ethInitialBalance := 2.99
	balanceServiceMock.On("GetAssetBalance", "ETH", false).Return(3.4205+ethInitialBalance, nil)

	orderRepositoryMock.On("Update", mock.Anything).Once().Return(nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	balanceServiceMock.On("InvalidateBalanceCache", "SOL").Once()
	solInitialBalance := 50.00
	balanceServiceMock.On("GetAssetBalance", "SOL", false).Times(2).Return(104.72+solInitialBalance, nil)

	timeServiceMock := new(TimeServiceMock)
	timeServiceMock.On("WaitSeconds", int64(5)).Times(3)
	timeServiceMock.On("WaitSeconds", int64(7)).Times(3)
	timeServiceMock.On("GetNowDiffMinutes", mock.Anything).Times(1).Return(50.00)

	swapRepoMock.On("GetSwapPairBySymbol", "SOLETH").Times(3).Return(model.SwapPair{
		Symbol:      "SOLETH",
		BuyPrice:    0.03128,
		SellPrice:   0.03131,
		MinNotional: 0.001,
		MinQuantity: 0.001,
		MinPrice:    0.00001,
	}, nil)

	binanceMock.On("LimitOrder", "SOLETH", 109.002, 0.03138, "BUY", "GTC").Return(model.BinanceOrder{
		Status:      "NEW",
		OrderId:     "21",
		Symbol:      "SOLETH",
		ExecutedQty: 0.00,
		OrigQty:     109.002,
		Price:       0.03138,
	}, nil)
	binanceMock.On("QueryOrder", "SOLETH", "21").Times(3).Return(model.BinanceOrder{
		Status:      "NEW",
		OrderId:     "21",
		Symbol:      "SOLETH",
		ExecutedQty: 12.00,
		OrigQty:     109.002,
		Price:       0.03138,
	}, nil)

	timeServiceMock.On("GetNowDiffMinutes", mock.Anything).Times(2).Return(50.00)

	timeServiceMock.On("WaitSeconds", int64(15)).Times(1)

	binanceMock.On("CancelOrder", "SOLETH", "21").Return(model.BinanceOrder{
		Status:      "CANCELED",
		OrderId:     "21",
		Symbol:      "SOLETH",
		ExecutedQty: 0.00,
		OrigQty:     109.002,
		Price:       0.03138,
	}, nil)

	binanceMock.On("LimitOrder", "SOLETH", 109.316, 0.03129, "BUY", "IOC").Return(model.BinanceOrder{
		Status:      "FILLED",
		OrderId:     "21",
		Symbol:      "SOLETH",
		ExecutedQty: 109.316,
		OrigQty:     109.316,
		Price:       0.03129,
	}, nil)

	executor := exchange.SwapExecutor{
		SwapRepository:  swapRepoMock,
		OrderRepository: orderRepositoryMock,
		BalanceService:  balanceServiceMock,
		Binance:         binanceMock,
		TimeService:     timeServiceMock,
		Formatter:       &utils.Formatter{},
	}

	executor.Execute(order)

	assertion.Equal(model.SwapActionStatusSuccess, swapRepoMock.swapAction.Status)
	assertion.Equal(109.316, *swapRepoMock.swapAction.EndQuantity)
	assertion.Equal("19", *swapRepoMock.swapAction.SwapOneExternalId)
	assertion.Equal(58.53, swapRepoMock.swapAction.SwapOnePrice)
	assertion.Equal("SOLGBP", swapRepoMock.swapAction.SwapOneSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapOneExternalStatus)
	assertion.Equal("20", *swapRepoMock.swapAction.SwapTwoExternalId)
	assertion.Equal(1711.14, swapRepoMock.swapAction.SwapTwoPrice)
	assertion.Equal("ETHGBP", swapRepoMock.swapAction.SwapTwoSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapTwoExternalStatus)
	assertion.Equal("21", *swapRepoMock.swapAction.SwapThreeExternalId)
	assertion.Equal(0.03129, swapRepoMock.swapAction.SwapThreePrice)
	assertion.Equal("SOLETH", swapRepoMock.swapAction.SwapThreeSymbol)
	assertion.Equal("FILLED_FORCE", *swapRepoMock.swapAction.SwapThreeExternalStatus)
}
