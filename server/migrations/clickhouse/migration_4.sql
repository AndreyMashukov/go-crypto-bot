CREATE TABLE IF NOT EXISTS default.market_tick
(
    exchange     LowCardinality(String),
    symbol       LowCardinality(String),
    event_time   DateTime64(3, 'UTC'),
    ingested_at  DateTime64(3, 'UTC'),
    sequence_id  UInt64,
    source       UInt8,
    price        Decimal64(8),
    best_bid     Decimal64(8),
    best_ask     Decimal64(8),
    spread_bps   Int32,
    volume_24h   Decimal64(8)
) ENGINE = MergeTree
PARTITION BY toYYYYMM(event_time)
ORDER BY (exchange, symbol, event_time)
SETTINGS index_granularity = 8192;
