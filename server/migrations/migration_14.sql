ALTER TABLE trade_limit ADD COLUMN profit_options JSON;
UPDATE trade_limit tl SET profit_options = CAST(CONCAT('[{"index": 0, "isTriggerOption": true, "optionUnit": "h", "optionValue": 1, "optionPercent": ', tl.min_profit_percent, '}]') as JSON) WHERE tl.id > 0;
ALTER TABLE trade_limit DROP column min_profit_percent;
ALTER TABLE orders ADD COLUMN profit_options JSON;
UPDATE orders o INNER JOIN trade_limit tl ON tl.symbol = o.symbol SET o.profit_options = tl.profit_options WHERE o.id > 0;
