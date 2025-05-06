ALTER TABLE sensors ADD CONSTRAINT sensors_pkey PRIMARY KEY (id);
ALTER TABLE sensors ADD CONSTRAINT sensors_serial_number_key UNIQUE (serial_number);