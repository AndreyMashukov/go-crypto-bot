create table `swap_pair`
(
    id              int auto_increment primary key,
    source_symbol   CHAR(20)                                    not null,
    symbol          CHAR(20)                                    not null,
    base_asset      CHAR(10)                                    not null,
    quote_asset     CHAR(10)                                    not null,
    last_price      double                                      not null,
    price_timestamp bigint unsigned                             not null,
    min_notional    double                                      not null,
    min_quantity    double                                      not null,
    min_price       double                                      not null
);
