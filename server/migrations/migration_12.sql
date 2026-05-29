alter table trade_limit add column extra_charge_options jsonb;
update trade_limit set extra_charge_options = '[]'::jsonb where id > 0;
alter table orders add column extra_charge_options jsonb;
update orders set extra_charge_options = '[]'::jsonb where id > 0;
