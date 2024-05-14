package model

import "sort"

const IcebergSideSell = "SELL"
const IcebergSideBuy = "BUY"

type Iceberg struct {
	Side     string  `json:"side"`
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

type OrderBookStat struct {
	BuyVolumeSum   float64 `json:"buyVolumeSum"`
	SellVolumeSum  float64 `json:"sellVolumeSum"`
	BuyQtySum      float64 `json:"buyQtySum"`
	SellQtySum     float64 `json:"sellQtySum"`
	BuyLength      int64   `json:"buyLength"`
	SellLength     int64   `json:"sellLength"`
	SellIceberg    Iceberg `json:"sellIceberg"`
	BuyIceberg     Iceberg `json:"buyIceberg"`
	FirstBuyQty    float64 `json:"firstBuyQty"`
	FirstBuyPrice  float64 `json:"firstBuyPrice"`
	FirstSellQty   float64 `json:"firstSellQty"`
	FirstSellPrice float64 `json:"firstSellPrice"`
}

type OrderBookModel struct {
	Symbol    string      `json:"s"`
	Timestamp int64       `json:"T,int"`
	Bids      [][2]Number `json:"b"`
	Asks      [][2]Number `json:"a"`
	UpdatedAt int64       `json:"updatedAt"`
}

func (d *OrderBookModel) IsEmpty() bool {
	return len(d.Asks) == 0 && len(d.Bids) == 0
}

func (d *OrderBookModel) GetFirstBuyQty() float64 {
	bids := d.GetBids()

	if len(bids) > 0 {
		return bids[0][1].Value
	}

	return 0.00
}

func (d *OrderBookModel) GetFirstSellQty() float64 {
	asks := d.GetAsks()

	if len(asks) > 0 {
		return asks[0][1].Value
	}

	return 0.00
}

func (d *OrderBookModel) GetStat() OrderBookStat {
	bids := d.GetBids()
	asks := d.GetAsks()

	firstBuyPrice := 0.00
	if len(bids) > 0 {
		firstBuyPrice = bids[0][0].Value
	}

	firstSellPrice := 0.00
	if len(asks) > 0 {
		firstSellPrice = asks[0][0].Value
	}

	return OrderBookStat{
		BuyLength:      int64(len(bids)),
		SellLength:     int64(len(asks)),
		BuyQtySum:      d.GetQtySumBid(),
		SellQtySum:     d.GetQtySumAsk(),
		BuyVolumeSum:   d.GetBidVolume(),
		SellVolumeSum:  d.GetAskVolume(),
		SellIceberg:    d.GetIcebergSell(),
		BuyIceberg:     d.GetIcebergBuy(),
		FirstBuyPrice:  firstBuyPrice,
		FirstBuyQty:    d.GetFirstBuyQty(),
		FirstSellPrice: firstSellPrice,
		FirstSellQty:   d.GetFirstSellQty(),
	}
}

func (d *OrderBookModel) GetBestBid() float64 {
	topPrice := 0.00
	priceSum := 0.00
	bidCount := 0.00

	for _, bid := range d.Bids {
		priceSum += bid[0].Value
		bidCount++
	}

	for _, bid := range d.Bids {
		if (0.00 == topPrice || bid[0].Value > topPrice) && bid[0].Value >= (priceSum/bidCount)/1.2 {
			topPrice = bid[0].Value
		}
	}

	return topPrice
}

func (d *OrderBookModel) GetIcebergSell() Iceberg {
	iceberg := Iceberg{
		Side:     IcebergSideSell,
		Price:    0.00,
		Quantity: 0.00,
	}

	for _, ask := range d.Asks {
		if iceberg.Quantity < ask[1].Value {
			iceberg.Price = ask[0].Value
			iceberg.Quantity = ask[1].Value
		}
	}

	return iceberg
}

func (d *OrderBookModel) GetIcebergBuy() Iceberg {
	iceberg := Iceberg{
		Side:     IcebergSideBuy,
		Price:    0.00,
		Quantity: 0.00,
	}

	for _, bid := range d.Bids {
		if iceberg.Quantity < bid[1].Value {
			iceberg.Price = bid[0].Value
			iceberg.Quantity = bid[1].Value
		}
	}

	return iceberg
}

func (d *OrderBookModel) GetBestAsk() float64 {
	topPrice := 0.00
	priceSum := 0.00
	askCount := 0.00

	for _, ask := range d.Asks {
		priceSum += ask[0].Value
		askCount++
	}

	for _, ask := range d.Asks {
		if (0.00 == topPrice || ask[0].Value < topPrice) && ask[0].Value <= (priceSum/askCount)*1.2 {
			topPrice = ask[0].Value
		}
	}

	return topPrice
}

func (d *OrderBookModel) GetAvgAsk() float64 {
	sum := 0.00
	amount := 0.00

	for _, ask := range d.Asks {
		sum += ask[0].Value
		amount++
	}

	return sum / amount
}

func (d *OrderBookModel) GetAvgBid() float64 {
	sum := 0.00
	amount := 0.00

	for _, bid := range d.Bids {
		sum += bid[0].Value
		amount++
	}

	return sum / amount
}

func (d *OrderBookModel) GetMaxQtyAsk() float64 {
	value := 0.00
	qty := 0.00

	for _, ask := range d.Asks {
		if 0.00 == qty || ask[1].Value > qty {
			qty = ask[1].Value
			value = ask[0].Value
		}
	}

	return value
}

func (d *OrderBookModel) GetMaxQtyBid() float64 {
	value := 0.00
	qty := 0.00

	for _, bid := range d.Bids {
		if 0.00 == qty || bid[1].Value > qty {
			qty = bid[1].Value
			value = bid[0].Value
		}
	}

	return value
}

func (d *OrderBookModel) GetAvgVolAsk() float64 {
	sumVolume := 0.00
	sumQty := 0.00

	for _, ask := range d.Asks {
		sumVolume += ask[0].Value * ask[1].Value
		sumQty += ask[1].Value
	}

	return sumVolume / sumQty
}

func (d *OrderBookModel) GetAvgVolBid() float64 {
	sumVolume := 0.00
	sumQty := 0.00

	for _, bid := range d.Bids {
		sumVolume += bid[0].Value * bid[1].Value
		sumQty += bid[1].Value
	}

	return sumVolume / sumQty
}

func (d *OrderBookModel) GetBidVolume() float64 {
	volume := 0.00

	for _, bid := range d.Bids {
		volume += bid[0].Value * bid[1].Value
	}

	return volume
}

func (d *OrderBookModel) GetAskVolume() float64 {
	volume := 0.00

	for _, ask := range d.Asks {
		volume += ask[0].Value * ask[1].Value
	}

	return volume
}

func (d *OrderBookModel) GetBidPosition(price float64) (int, [2]Number) {
	bids := d.GetBids()

	for index, bid := range bids {
		if bid[0].Value >= price && len(bids) > index+1 && bids[index+1][0].Value < price {
			return index, bid
		}
	}

	if 0 == len(bids) {
		return 100, [2]Number{{Value: 0.00}, {Value: 0.00}}
	}

	lastPosition := bids[len(d.Bids)-1]

	return len(bids) - 1, lastPosition
}

func (d *OrderBookModel) GetAskPosition(price float64) (int, [2]Number) {
	asks := d.GetAsks()

	for index, ask := range asks {
		if ask[0].Value <= price && len(asks) > index+1 && asks[index+1][0].Value > price {
			return index, ask
		}
	}

	if 0 == len(asks) {
		return 100, [2]Number{{Value: 0.00}, {Value: 0.00}}
	}

	lastPosition := asks[len(asks)-1]

	return len(asks) - 1, lastPosition
}

func (d *OrderBookModel) GetBestAvgBid() float64 {
	priceSum := 0.00
	bidCount := 0.00

	for _, bid := range d.Bids {
		priceSum += bid[0].Value
		bidCount++
	}

	bestPriceSum := 0.00
	bestPriceAmount := 0.00

	for _, bid := range d.Bids {
		if bid[0].Value >= (priceSum/bidCount)/1.2 {
			bestPriceSum += bid[0].Value
			bestPriceAmount++
		}
	}

	return bestPriceSum / bestPriceAmount
}

func (d *OrderBookModel) GetBestAvgAsk() float64 {
	priceSum := 0.00
	askCount := 0.00

	for _, ask := range d.Asks {
		priceSum += ask[0].Value
		askCount++
	}

	bestPriceSum := 0.00
	bestPriceAmount := 0.00

	for _, ask := range d.Asks {
		if ask[0].Value <= (priceSum/askCount)*1.2 {
			bestPriceSum += ask[0].Value
			bestPriceAmount++
		}
	}

	return bestPriceSum / bestPriceAmount
}

func (d *OrderBookModel) GetBids() [][2]Number {
	bids := make([][2]Number, len(d.Bids))
	copy(bids, d.Bids)
	sort.SliceStable(bids, func(i int, j int) bool {
		return bids[i][0].Value > bids[j][0].Value
	})

	return bids
}

func (d *OrderBookModel) GetAsks() [][2]Number {
	asks := make([][2]Number, len(d.Asks))
	copy(asks, d.Asks)
	sort.SliceStable(asks, func(i int, j int) bool {
		return asks[i][0].Value < asks[j][0].Value
	})

	return asks
}

func (d *OrderBookModel) GetAsksReversed() [][2]Number {
	asks := make([][2]Number, len(d.Asks))
	copy(asks, d.Asks)
	sort.SliceStable(asks, func(i int, j int) bool {
		return asks[i][0].Value > asks[j][0].Value
	})

	return asks
}

func (d *OrderBookModel) GetQtySumAsk() float64 {
	qty := 0.00

	for _, ask := range d.Asks {
		qty += ask[1].Value
	}

	return qty
}

func (d *OrderBookModel) GetQtySumBid() float64 {
	qty := 0.00

	for _, bid := range d.Bids {
		qty += bid[1].Value
	}

	return qty
}

type ByBitOrderBookModel struct {
	Symbol    string      `json:"s"`
	Bids      [][2]Number `json:"b"`
	Asks      [][2]Number `json:"a"`
	Timestamp int64       `json:"ts"`
	UpdateId  int         `json:"u"`
	Seq       int64       `json:"seq"`
}
