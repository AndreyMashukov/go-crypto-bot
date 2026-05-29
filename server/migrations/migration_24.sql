-- Phase J: closed-perimeter single-bot seed.
-- Pins a fixed BOT_UUID so the watcher and trader containers join the
-- same bot row instead of creating their own on first boot, and seeds
-- five default trade pairs the bot starts watching immediately.

-- Enforce uniqueness on uuid so re-running the seed cannot duplicate
-- the bot row (the original schema in migration_1 only had a primary
-- key on id, leaving uuid free to be inserted N times).
do $$ begin
    if not exists (select 1 from pg_indexes where indexname = 'bots_uuid_uniq') then
        alter table bots add constraint bots_uuid_uniq unique (uuid);
    end if;
    if not exists (select 1 from pg_indexes where indexname = 'trade_limit_bot_symbol_uniq') then
        alter table trade_limit add constraint trade_limit_bot_symbol_uniq unique (bot_id, symbol);
    end if;
end $$;

-- char(20) right-pads symbols with spaces; the bot's WS stream-URL
-- builder concatenates `BTCUSDT             ` straight into the
-- subscribe URL and Binance refuses the handshake. Convert to varchar
-- so reads return the canonical symbol with no padding.
alter table trade_limit alter column symbol type varchar(20) using rtrim(symbol);

insert into bots (uuid, exchange, is_master_bot, is_swap_enabled, swap_config, trade_stack_sorting)
values (
    '00000000-0000-0000-0000-000000000001',
    'binance',
    false,
    false,
    '{}'::jsonb,
    'percent'
)
on conflict (uuid) do nothing;

-- Unique seed key — re-running the migration is a no-op because each
-- (bot_id, symbol) is already there.
insert into trade_limit (
    bot_id,
    symbol,
    usdt_limit,
    min_quantity,
    min_notional,
    min_price,
    is_enabled,
    profit_options,
    trade_filters_buy,
    trade_filters_sell,
    trade_filters_extra_charge,
    extra_charge_options
)
select
    b.id,
    sym.symbol,
    100.00,
    0.0001,
    10.00,
    0.0001,
    true,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb,
    '[]'::jsonb
from bots b
cross join (values
    ('BTCUSDT'),
    ('ETHUSDT'),
    ('SOLUSDT'),
    ('BNBUSDT'),
    ('TONUSDT')
) as sym(symbol)
where b.uuid = '00000000-0000-0000-0000-000000000001'
  and not exists (
      select 1 from trade_limit tl
      where tl.bot_id = b.id and tl.symbol = sym.symbol
  );

-- Patch any pre-existing rows missing min_price so the bot's float64 Scan
-- doesn't panic. Safe no-op on rows already populated.
update trade_limit set min_price = 0.0001 where min_price is null;
