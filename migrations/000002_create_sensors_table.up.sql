create type sensor_type as enum ('cc', 'adc');

create table sensors
(
    id            bigserial     not null,
    serial_number text,
    type          sensor_type   not null,
    current_state bigint,
    description   text,
    is_active     boolean,
    registered_at timestamp,
    last_activity timestamp
);
