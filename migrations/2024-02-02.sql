ALTER table trade_limit ADD COLUMN extra_charge_options JSON
UPDATE trade_limit SET extra_charge_options = JSON_ARRAY() where id > 0
ALTER table orders ADD COLUMN extra_charge_options JSON
UPDATE orders SET extra_charge_options = JSON_ARRAY() where id > 0
