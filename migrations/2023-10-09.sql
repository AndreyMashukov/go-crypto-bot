create table trade_limit
(
    id         int auto_increment primary key,
    symbol     CHAR(20) not null,
    usdt_limit double   not null
);
