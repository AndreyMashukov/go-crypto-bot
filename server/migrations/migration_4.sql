alter table orders add column commission double precision default null;
alter table orders add column commission_asset char(5) default null;
