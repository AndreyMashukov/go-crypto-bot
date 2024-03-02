ALTER TABLE orders ADD column executed_quantity double unsigned default '0.00';
UPDATE orders set executed_quantity = quantity WHERE id > 0;
ALTER TABLE orders ADD column closes_order int default null;
UPDATE orders o1 INNER JOIN orders o2 ON o2.closed_by = o1.id SET o1.closes_order = o2.id WHERE o1.id > 0;
UPDATE orders o1 INNER JOIN orders o2 ON o1.closed_by = o2.id SET o1.closes_order = o2.id WHERE o1.quantity = 0;
ALTER TABLE orders add constraint order_closes_order_fk foreign key (closes_order) references orders (id);
ALTER TABLE orders DROP FOREIGN KEY order_closed_by_fk;
ALTER TABLE orders DROP column closed_by;
