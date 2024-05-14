CREATE TABLE default.trades_2(
    symbol String,
    timestamp DateTime64(3, 'Europe/London'),
    bot_id UUID,
    price Float64,
    buy_qty Float64,
    sell_qty Float64,
    buy_volume Float64,
    sell_volume Float64,
    trade_count Int64,
    max_pcs Float64,
    min_pcs Float64,
    open Float64,
    close Float64,
    high Float64,
    low Float64,
    volume Float64,
    order_book_buy_length Int64,
    order_book_sell_length Int64,
    order_book_buy_qty_sum Float64,
    order_book_sell_qty_sum Float64,
    order_book_buy_volume_sum Float64,
    order_book_sell_volume_sum Float64,
    order_book_buy_iceberg_qty Float64,
    order_book_buy_iceberg_price Float64,
    order_book_sell_iceberg_qty Float64,
    order_book_sell_iceberg_price Float64,
    order_book_buy_first_qty Float64,
    order_book_buy_first_price Float64,
    order_book_sell_first_qty Float64,
    order_book_sell_first_price Float64,
    exchange Enum('binance', 'bybit') DEFAULT 'binance'
)
ENGINE = MergeTree()
PRIMARY KEY (symbol, timestamp, exchange);

INSERT INTO default.trades_2 (*) SELECT * FROM default.trades;
RENAME TABLE default.trades TO default.trades_3, default.trades_2 TO default.trades;


