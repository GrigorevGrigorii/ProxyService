BEGIN;

ALTER TABLE targets ADD COLUMN query TEXT NOT NULL DEFAULT '';
ALTER TABLE targets DROP CONSTRAINT targets_pkey;
ALTER TABLE targets ADD PRIMARY KEY (service_name, path, method, query);

ALTER TABLE targets ADD COLUMN cache_interval INTEGER CHECK (cache_interval > 0);

COMMIT;
