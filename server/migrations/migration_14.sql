alter table trade_limit add column profit_options jsonb;
update trade_limit tl set profit_options =
    ('[{"index": 0, "isTriggerOption": true, "optionUnit": "h", "optionValue": 1, "optionPercent": ' || tl.min_profit_percent || '}]')::jsonb
    where tl.id > 0;
alter table trade_limit drop column min_profit_percent;
alter table orders add column profit_options jsonb;
update orders o set profit_options = tl.profit_options from trade_limit tl where tl.symbol = o.symbol and o.id > 0;
