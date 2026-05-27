create table trade_limit
(
    id           int auto_increment primary key,
    symbol       char(20) not null,
    usdt_limit   double   not null,
    min_price    double   null,
    min_quantity double   not null
);
