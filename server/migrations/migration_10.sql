alter table swap_pair change column last_price buy_price double default 0.00;
alter table swap_pair add column sell_price double default 0.00;
UPDATE swap_transition SET operation = 'S' WHERE operation = 'BUY';
UPDATE swap_transition SET operation = 'BUY' WHERE operation = 'SELL';
UPDATE swap_transition SET operation = 'SELL' WHERE operation = 'S';
UPDATE swap_chain SET type = 'SSB' WHERE type = 'BBS';
# -----
ALTER TABLE swap_chain ADD COLUMN max_percent double default 0.00;
ALTER TABLE swap_chain ADD COLUMN max_percent_timestamp int unsigned default null;
DELETE FROM swap_chain WHERE id NOT IN (select swap_chain_id from swap_action);
DELETE IGNORE FROM swap_transition WHERE id > 0;
# -----
ALTER TABLE orders ADD CONSTRAINT order_external_id_symbol UNIQUE (external_id,symbol);
# -----
alter table swap_pair add column sell_volume double unsigned default 0.00;
alter table swap_pair add column buy_volume double unsigned default 0.00;
alter table swap_pair add column daily_percent double default 0.00;
