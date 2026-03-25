BEGIN;

ALTER TABLE targets DROP COLUMN cache_interval;
ALTER TABLE targets ADD COLUMN cache_interval INTEGER CHECK (cache_interval > 0);

COMMIT;
