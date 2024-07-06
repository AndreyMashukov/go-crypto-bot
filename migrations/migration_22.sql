ALTER TABLE swap_action ADD COLUMN swap_one_side char(4) default null;
ALTER TABLE swap_action ADD COLUMN swap_one_quantity double not null;
ALTER TABLE swap_action ADD COLUMN swap_two_side char(4) default null;
ALTER TABLE swap_action ADD COLUMN swap_two_quantity double not null;
ALTER TABLE swap_action ADD COLUMN swap_three_side char(4) default null;
ALTER TABLE swap_action ADD COLUMN swap_three_quantity double not null;
