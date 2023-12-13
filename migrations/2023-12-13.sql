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