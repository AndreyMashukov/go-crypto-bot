ALTER TABLE swap_action CHANGE swap_one_external_id swap_one_external_id CHAR(36) DEFAULT NULL;
ALTER TABLE swap_action CHANGE swap_two_external_id swap_two_external_id CHAR(36) DEFAULT NULL;
ALTER TABLE swap_action CHANGE swap_three_external_id swap_three_external_id CHAR(36) DEFAULT NULL;
ALTER TABLE orders CHANGE external_id external_id CHAR(36) DEFAULT NULL;
ALTER TABLE bots ADD COLUMN exchange enum('binance', 'bybit') DEFAULT 'binance';
