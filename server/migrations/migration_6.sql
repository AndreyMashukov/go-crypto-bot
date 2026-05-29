alter table orders add column executed_quantity double precision not null default 0;
update orders set executed_quantity = quantity where id > 0;
alter table orders add constraint orders_executed_quantity_nonneg check (executed_quantity >= 0);
alter table orders add column closes_order bigint default null;
update orders o1 set closes_order = o2.id from orders o2 where o2.closed_by = o1.id and o1.id > 0;
update orders o1 set closes_order = o2.id from orders o2 where o1.closed_by = o2.id and o1.quantity = 0;
alter table orders add constraint order_closes_order_fk foreign key (closes_order) references orders (id);
alter table orders drop constraint order_closed_by_fk;
alter table orders drop column closed_by;
