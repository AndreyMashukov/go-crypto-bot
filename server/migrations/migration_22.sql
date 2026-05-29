alter table swap_action add column swap_one_side char(4) default null;
alter table swap_action add column swap_one_quantity double precision not null default 0;
alter table swap_action alter column swap_one_quantity drop default;
alter table swap_action add column swap_two_side char(4) default null;
alter table swap_action add column swap_two_quantity double precision not null default 0;
alter table swap_action alter column swap_two_quantity drop default;
alter table swap_action add column swap_three_side char(4) default null;
alter table swap_action add column swap_three_quantity double precision not null default 0;
alter table swap_action alter column swap_three_quantity drop default;
