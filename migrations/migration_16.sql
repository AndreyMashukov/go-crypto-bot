ALTER TABLE bots ADD COLUMN swap_config JSON;
UPDATE bots b SET b.swap_config = CAST('{"swapMinPercent": 2.00, "swapOrderProfitTrigger": -5.00, "orderTimeTrigger": 36000, "useSwapCapital": true, "historyInterval": "1d", "historyPeriod": 14}' as JSON) WHERE b.id > 0;
