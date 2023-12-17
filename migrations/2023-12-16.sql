alter table swap_pair change column last_price buy_price double default 0.00;
alter table swap_pair add column sell_price double default 0.00;
UPDATE swap_transition SET operation = 'S' WHERE operation = 'BUY';
UPDATE swap_transition SET operation = 'BUY' WHERE operation = 'SELL';
UPDATE swap_transition SET operation = 'SELL' WHERE operation = 'S';
UPDATE swap_chain SET type = 'SSB' WHERE type = 'BBS';
# -----
ALTER TABLE swap_chain ADD COLUMN max_percent double default 0.00;
ALTER TABLE swap_chain ADD COLUMN max_percent_timestamp int unsigned default null;

UPDATE swap_transition SET operation = 'S' WHERE operation = 'BUY'
UPDATE swap_transition SET operation = 'BUY' WHERE operation = 'SELL'
UPDATE swap_transition SET operation = 'SELL' WHERE operation = 'S'
UPDATE swap_chain SET type = 'SSB' WHERE type = 'BBS'
UPDATE swap_transition SET type = 'SSB' WHERE type = 'BBS'
# ------
DELETE FROM swap_chain WHERE id NOT IN (select swap_chain_id from swap_action)
DELETE IGNORE FROM swap_transition WHERE id > 0