package exchange_context

type KLine struct {
	Symbol    string   `json:"s"`
	Open      float64  `json:"o,string"`
	Close     float64  `json:"c,string"`
	Low       float64  `json:"l,string"`
	High      float64  `json:"h,string"`
	Interval  string   `json:"i"`
	Timestamp UnixTime `json:"T"`
	Volume    float64  `json:"v,string"`
	// 		"t": 1694420820000,
	//      "T": 1694420879999,
	//      "s": "BTCUSDT",
	//      "i": "1m",
	//      "f": 4075396873,
	//      "L": 4075398403,
	//      "o": "25760.40",
	//      "c": "25756.30",
	//      "h": "25760.40",
	//      "l": "25753.60",
	//      "v": "95.902",
	//      "n": 1531,
	//      "x": false,
	//      "q": "2470213.54580",
	//      "V": "51.693",
	//      "Q": "1331459.78300",
	//      "B": "0"
}
