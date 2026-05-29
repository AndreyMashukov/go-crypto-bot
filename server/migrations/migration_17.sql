create type trade_stack_sorting as enum ('percent', 'diff');
alter table bots add column trade_stack_sorting trade_stack_sorting default 'percent';
