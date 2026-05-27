package service

import "strings"

type CoinDetector struct {
}

func (c *CoinDetector) DetectCoins(text string) []string {
	coinMap := map[string][2]string{
		"BTC":   {"BTC", "Bitcoin"},
		"NEO":   {"NEO", "Neo"},
		"PERP":  {"PERP", "Perpetual Protocol"},
		"ETH":   {"ETH", "Ethereum"},
		"SOL":   {"SOL", "Solana"},
		"LTC":   {"LTC", "Litecoin"},
		"XRP":   {"XRP", "Ripple"},
		"BNB":   {"BNB", "Binance Coin"},
		"TRX":   {"TRX", "TRON"},
		"AVAX":  {"AVAX", "Avalanche"},
		"ADA":   {"ADA", "Cardano"},
		"DOGE":  {"DOGE", "Dogecoin"},
		"BCH":   {"BCH", "Bitcoin Cash"},
		"LINK":  {"LINK", "Chainlink"},
		"MATIC": {"MATIC", "Polygon"},
		"DOT":   {"DOT", "Polkadot"},
		"UNI":   {"UNI", "Uniswap"},
		"ETC":   {"ETC", "Ethereum Classic"},
		"XLM":   {"XLM", "Stellar"},
		"ATOM":  {"ATOM", "Cosmos"},
		"NEAR":  {"NEAR", "NEAR Protocol"},
		"ZEC":   {"ZEC", "Zcash"},
		"SHIB":  {"SHIB", "Shiba Inu"},
		"TON":   {"TON", "TON Crystal"},
		"PEPE":  {"PEPE", "PepeCoin"},
		"ICP":   {"ICP", "Internet Computer"},
		"DASH":  {"DASH", "Dash"},
		"IMX":   {"IMX", "ImpactCoin"},
	}

	coins := make([]string, 0)

	for code, data := range coinMap {
		if strings.Contains(text, data[0]) || strings.Contains(text, data[1]) {
			coins = append(coins, code)
		}
	}

	return coins
}
