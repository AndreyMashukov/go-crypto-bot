ALTER TABLE default.trades ADD COLUMN exchange Enum('binance', 'bybit') DEFAULT NULL;
UPDATE default.trades SET exchange = 'binance' WHERE exchange is NULL;
