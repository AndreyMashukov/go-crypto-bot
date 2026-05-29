alter table trade_limit add column min_notional double precision not null default 0 check (min_notional >= 0);
