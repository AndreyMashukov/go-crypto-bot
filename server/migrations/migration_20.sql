create table `object_storage`
(
    storage_key CHAR(255) primary key,
    object      JSON         not null,
    created_at  datetime     not null,
    updated_at  datetime     not null,
    bot_id      int unsigned not null,
    constraint object_storage_bot_id_fk foreign key (bot_id) references `bots` (id)
);
