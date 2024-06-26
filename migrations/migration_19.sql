ALTER TABLE swap_action CHANGE swap_one_external_id swap_one_external_id CHAR(36) DEFAULT NULL;
ALTER TABLE swap_action CHANGE swap_two_external_id swap_two_external_id CHAR(36) DEFAULT NULL;
ALTER TABLE swap_action CHANGE swap_three_external_id swap_three_external_id CHAR(36) DEFAULT NULL;
ALTER TABLE orders CHANGE external_id external_id CHAR(36) NOT NULL;
ALTER TABLE orders ADD COLUMN exchange enum('binance', 'bybit') DEFAULT NULL;
UPDATE orders SET exchange = 'binance' WHERE id > 0;
ALTER TABLE orders CHANGE exchange exchange enum('binance', 'bybit') NOT NULL;
ALTER TABLE bots ADD COLUMN exchange enum('binance', 'bybit') DEFAULT NULL;
UPDATE bots SET exchange = 'binance' WHERE id > 0;
ALTER TABLE bots CHANGE exchange exchange enum('binance', 'bybit') NOT NULL;
ALTER TABLE swap_pair ADD COLUMN exchange enum('binance', 'bybit') NOT NULL;
UPDATE swap_pair SET exchange = 'binance' WHERE id > 0;
ALTER TABLE swap_chain ADD COLUMN exchange enum('binance', 'bybit') NOT NULL;
UPDATE swap_chain SET exchange = 'binance' WHERE id > 0;
