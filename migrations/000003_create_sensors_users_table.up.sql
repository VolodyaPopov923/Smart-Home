create table sensors_users
(
    id          bigserial   not null,
    sensor_id   bigint      not null,
    user_id     bigint      not null
);
