create table object_storage
(
    storage_key char(255)   primary key,
    object      jsonb       not null,
    created_at  timestamptz not null,
    updated_at  timestamptz not null,
    bot_id      bigint      not null,
    constraint object_storage_bot_id_fk foreign key (bot_id) references bots (id)
);
