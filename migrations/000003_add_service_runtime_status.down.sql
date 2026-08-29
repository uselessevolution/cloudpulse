ALTER TABLE services
DROP CONSTRAINT IF EXISTS services_runtime_status_valid,
DROP CONSTRAINT IF EXISTS services_consecutive_failures_non_negative,
DROP CONSTRAINT IF EXISTS services_consecutive_successes_non_negative;

ALTER TABLE services
DROP COLUMN IF EXISTS runtime_status,
DROP COLUMN IF EXISTS consecutive_failures,
DROP COLUMN IF EXISTS consecutive_successes,
DROP COLUMN IF EXISTS last_checked_at;