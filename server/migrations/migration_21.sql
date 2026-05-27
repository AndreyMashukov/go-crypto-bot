ALTER TABLE trade_limit ADD COLUMN sentiment_label CHAR(20) DEFAULT NULL;
ALTER TABLE trade_limit ADD COLUMN sentiment_score double DEFAULT NULL;
