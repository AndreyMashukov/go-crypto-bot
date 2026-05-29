alter table bots add column swap_config jsonb;
update bots set swap_config = '{"swapMinPercent": 2.00, "swapOrderProfitTrigger": -5.00, "orderTimeTrigger": 36000, "useSwapCapital": true, "historyInterval": "1d", "historyPeriod": 14}'::jsonb where id > 0;
