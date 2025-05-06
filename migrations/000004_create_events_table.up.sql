create table events
(
    timestamp               timestamp   not null,
    sensor_serial_number    text        not null,
    sensor_id               bigint      not null,
    payload                 bigint      not null
);