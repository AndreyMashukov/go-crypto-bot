alter table swap_action alter column swap_one_external_id type char(36) using swap_one_external_id::text::char(36);
alter table swap_action alter column swap_two_external_id type char(36) using swap_two_external_id::text::char(36);
alter table swap_action alter column swap_three_external_id type char(36) using swap_three_external_id::text::char(36);
alter table orders alter column external_id type char(36) using external_id::text::char(36);
alter table orders alter column external_id set not null;

create type exchange_name as enum ('binance', 'bybit');

alter table orders add column exchange exchange_name default null;
update orders set exchange = 'binance' where id > 0;
alter table orders alter column exchange set not null;
alter table bots add column exchange exchange_name default null;
update bots set exchange = 'binance' where id > 0;
alter table bots alter column exchange set not null;
alter table swap_pair add column exchange exchange_name;
update swap_pair set exchange = 'binance' where id > 0;
alter table swap_pair alter column exchange set not null;
alter table swap_chain add column exchange exchange_name;
update swap_chain set exchange = 'binance' where id > 0;
alter table swap_chain alter column exchange set not null;
