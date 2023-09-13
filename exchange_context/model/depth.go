package model

import "sort"

type Depth struct {
	Symbol    string      `json:"s"`
	Timestamp UnixTime    `json:"T"`
	Bids      [][2]Number `json:"b"`
	Asks      [][2]Number `json:"a"`
}

func (d *Depth) GetBestBid() float64 {
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

func (d *Depth) GetBestAsk() float64 {
	topPrice := 0.00
	priceSum := 0.00
	bidCount := 0.00

	for _, ask := range d.Asks {
		priceSum += ask[0].Value
		bidCount++
	}

	for _, ask := range d.Asks {
		if (0.00 == topPrice || ask[0].Value < topPrice) && ask[0].Value <= (priceSum/bidCount)*1.2 {
			topPrice = ask[0].Value
		}
	}

	return topPrice
}

func (d *Depth) GetAvgAsk() float64 {
	sum := 0.00
	amount := 0.00

	for _, ask := range d.Asks {
		sum += ask[0].Value
		amount++
	}

	return sum / amount
}

func (d *Depth) GetAvgBid() float64 {
	sum := 0.00
	amount := 0.00

	for _, bid := range d.Bids {
		sum += bid[0].Value
		amount++
	}

	return sum / amount
}

func (d *Depth) GetMaxQtyAsk() float64 {
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

func (d *Depth) GetMaxQtyBid() float64 {
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
func (d *Depth) GetAvgVolAsk() float64 {
	sumVolume := 0.00
	sumQty := 0.00

	for _, ask := range d.Asks {
		sumVolume += ask[0].Value * ask[1].Value
		sumQty += ask[1].Value
	}

	return sumVolume / sumQty
}

func (d *Depth) GetAvgVolBid() float64 {
	sumVolume := 0.00
	sumQty := 0.00

	for _, bid := range d.Bids {
		sumVolume += bid[0].Value * bid[1].Value
		sumQty += bid[1].Value
	}

	return sumVolume / sumQty
}

func (d *Depth) GetBidVolume() float64 {
	volume := 0.00

	for _, bid := range d.Bids {
		volume += bid[0].Value * bid[1].Value
	}

	return volume
}

func (d *Depth) GetAskVolume() float64 {
	volume := 0.00

	for _, ask := range d.Asks {
		volume += ask[0].Value * ask[1].Value
	}

	return volume
}

func (d *Depth) GetBidPosition(price float64) (int, [2]Number) {
	bids := make([][2]Number, len(d.Bids))
	copy(bids, d.Bids)
	sort.SliceStable(bids, func(i int, j int) bool {
		return bids[i][0].Value > bids[j][0].Value
	})

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

func (d *Depth) GetAskPosition(price float64) (int, [2]Number) {
	asks := make([][2]Number, len(d.Asks))
	copy(asks, d.Asks)
	sort.SliceStable(asks, func(i int, j int) bool {
		return asks[i][0].Value < asks[j][0].Value
	})

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
