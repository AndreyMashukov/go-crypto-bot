create table `orders`
(
    id          int auto_increment
        primary key,
    symbol      CHAR(20)                                      not null,
    quantity    double                                        not null,
    price       double                                        not null,
    created_at  datetime                                      not null,
    sell_volume double                                        not null,
    buy_volume  double                                        not null,
    sma_value   double                                        not null,
    operation   enum ('sell', 'buy')                          not null,
    status      enum ('new', 'opened', 'closed', 'cancelled') not null,
    external_id bigint                                        null,
    closed_by   int                                           null,
    constraint order_closed_by_fk
        foreign key (closed_by) references `orders` (id)
);
