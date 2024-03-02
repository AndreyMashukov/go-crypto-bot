alter table trade_limit add min_profit_percent double not null;
alter table trade_limit add is_enabled tinyint not null;
UPDATE trade_limit SET is_enabled = 1, min_profit_percent = 0.6 WHERE id > 0;
alter table trade_limit add usdt_extra_budget double not null;
alter table trade_limit add buy_on_fall_percent double not null;
alter table orders add used_extra_budget double not null;