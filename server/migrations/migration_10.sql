alter table swap_pair rename column last_price to buy_price;
alter table swap_pair alter column buy_price set default 0.00;
alter table swap_pair add column sell_price double precision default 0.00;
update swap_transition set operation = 'S' where operation = 'BUY';
update swap_transition set operation = 'BUY' where operation = 'SELL';
update swap_transition set operation = 'SELL' where operation = 'S';
update swap_chain set type = 'SSB' where type = 'BBS';

alter table swap_chain add column max_percent double precision default 0.00;
alter table swap_chain add column max_percent_timestamp bigint default null;
delete from swap_chain where id not in (select swap_chain_id from swap_action);
delete from swap_transition where id > 0;

alter table orders add constraint order_external_id_symbol unique (external_id, symbol);

alter table swap_pair add column sell_volume double precision default 0.00;
alter table swap_pair add column buy_volume double precision default 0.00;
alter table swap_pair add column daily_percent double precision default 0.00;
