package event

// Source identifies what kind of underlying exchange event produced a
// MarketTick. The strategy reads it to know e.g. that a `Heartbeat`
// MarketTick has no fresh trade information, just a timekeeping pulse.
type Source uint8

// The Source enum values cover every event kind the watcher can emit.
// Trade and BookTop are the live trading inputs; Liquidation /
// UserData / Heartbeat ride on the same wire to keep the trader's
// consumption loop uniform.
const (
	SourceTrade Source = iota + 1
	SourceBookTop
	SourceLiquidation
	SourceUserData
	SourceHeartbeat
)

func (s Source) String() string {
	switch s {
	case SourceTrade:
		return "trade"
	case SourceBookTop:
		return "book_top"
	case SourceLiquidation:
		return "liquidation"
	case SourceUserData:
		return "user_data"
	case SourceHeartbeat:
		return "heartbeat"
	}
	return "unknown"
}
