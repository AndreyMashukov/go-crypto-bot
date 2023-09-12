package model

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
