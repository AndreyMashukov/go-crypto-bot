ALTER TABLE trade_limit ADD COLUMN trade_filters_buy JSON;
UPDATE trade_limit SET trade_filters_buy = json_array() WHERE id > 0
ALTER TABLE trade_limit ADD COLUMN trade_filters_sell JSON;
UPDATE trade_limit SET trade_filters_sell = json_array() WHERE id > 0
ALTER TABLE trade_limit ADD COLUMN trade_filters_extra_charge JSON;
UPDATE trade_limit SET trade_filters_extra_charge = json_array() WHERE id > 0
