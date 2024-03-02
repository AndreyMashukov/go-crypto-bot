alter table trade_limit add column min_price_minutes_period int default 200;
alter table trade_limit add column frame_interval char(5) default '2h';
alter table trade_limit add column frame_period int default 20;
alter table trade_limit add column buy_price_history_check_interval char(5) default '1d';
alter table trade_limit add column buy_price_history_check_period int default 14;
