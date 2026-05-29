alter table trade_limit add column trade_filters_buy jsonb;
update trade_limit set trade_filters_buy = '[]'::jsonb where id > 0;
alter table trade_limit add column trade_filters_sell jsonb;
update trade_limit set trade_filters_sell = '[]'::jsonb where id > 0;
alter table trade_limit add column trade_filters_extra_charge jsonb;
update trade_limit set trade_filters_extra_charge = '[]'::jsonb where id > 0;
