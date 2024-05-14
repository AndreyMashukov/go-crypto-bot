ALTER TABLE default.trades ADD COLUMN exchange Enum('binance', 'bybit') DEFAULT 'binance';
