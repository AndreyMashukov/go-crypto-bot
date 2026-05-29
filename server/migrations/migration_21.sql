alter table trade_limit add column sentiment_label char(20) default null;
alter table trade_limit add column sentiment_score double precision default null;
