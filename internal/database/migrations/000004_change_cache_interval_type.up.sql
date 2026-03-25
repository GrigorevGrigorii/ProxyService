BEGIN;

ALTER TABLE targets DROP COLUMN cache_interval;
ALTER TABLE targets ADD COLUMN cache_interval VARCHAR(8);

COMMIT;
