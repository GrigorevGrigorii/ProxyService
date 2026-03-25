BEGIN;

ALTER TABLE targets DROP COLUMN cache_interval;

ALTER TABLE targets DROP CONSTRAINT targets_pkey;
ALTER TABLE targets ADD PRIMARY KEY (service_name, path, method);
ALTER TABLE targets DROP COLUMN query;

COMMIT;
