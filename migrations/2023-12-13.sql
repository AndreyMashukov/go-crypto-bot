# // SwapTransition Entity
# type SwapTransition struct {
# 	Id            int64            `json:"id"`
# 	Type          string           `json:"type"`
# 	BaseAsset     string           `json:"baseAsset"`
# 	QuoteAsset    string           `json:"quoteAsset"`
# 	Operation     string           `json:"operation"`
# 	Quantity      float64          `json:"baseQuantity"`
# 	Price         float64          `json:"price"`
# 	Level         int64            `json:"level"`
# }
create table `swap_transition`
(
    id          int auto_increment primary key,
    type        CHAR(3)                                     not null,
    symbol      CHAR(10)                                    not null,
    base_asset  CHAR(10)                                    not null,
    quote_asset CHAR(10)                                    not null,
    operation   CHAR(4)                                     not null,
    quantity    double                                      not null,
    price       double                                      not null,
    level       double                                      not null
);

create table `swap_chain`
(
    id              int auto_increment primary key,
    title           CHAR(255)                                   not null,
    type            CHAR(3)                                     not null,
    hash            CHAR(32)                                    not null,
    swap_one        int                                         not null,
    swap_two        int                                         not null,
    swap_three      int                                         not null,
    percent         double                                      not null,
    timestamp       bigint                                      not null,
    constraint swap_transition_one_fk foreign key (swap_one) references `swap_transition` (id),
    constraint swap_transition_two_fk foreign key (swap_two) references `swap_transition` (id),
    constraint swap_transition_three_fk foreign key (swap_three) references `swap_transition` (id)
);
ALTER TABLE swap_chain ADD CONSTRAINT swap_chain_hash_uniq UNIQUE (hash);
# --------
alter table orders add swap tinyint(1) not null default 0;
create table `swap_action`
(
    id                         int auto_increment primary key,
    order_id                   int                                               not null,
    bot_id                     int unsigned                                      not null,
    swap_chain_id              int                                               not null,
    asset                      char(10)                                          not null,
    status                     enum('pending', 'process', 'canceled', 'success') not null,
    start_timestamp            int                                               not null,
    start_quantity             double                                            not null,
    end_timestamp              int                                               default null,
    end_quantity               int                                               default null,
    swap_one_external_id       bigint                                            default null,
    swap_one_external_status   char(20)                                          default null,
    swap_one_symbol            char(10)                                          not null,
    swap_one_price             double                                            not null,
    swap_one_timestamp         int                                               default null,
    swap_two_external_id       bigint                                            default null,
    swap_two_external_status   char(20)                                          default null,
    swap_two_symbol            char(10)                                          not null,
    swap_two_price             double                                            not null,
    swap_two_timestamp         int                                               default null,
    swap_three_external_id     bigint                                            default null,
    swap_three_external_status char(20)                                          default null,
    swap_three_symbol          char(10)                                          not null,
    swap_three_price           double                                            not null,
    swap_three_timestamp       int                                               default null,
    constraint swap_action_order_fk foreign key (order_id) references `orders` (id),
    constraint swap_action_swap_chain_fk foreign key (swap_chain_id) references `swap_chain` (id),
    constraint swap_action_bot_fk foreign key (bot_id) references `bots` (id)
);
