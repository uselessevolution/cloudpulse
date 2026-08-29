ALTER TABLE services
ADD COLUMN runtime_status VARCHAR(20) NOT NULL DEFAULT 'HEALTHY',
ADD COLUMN consecutive_failures INTEGER NOT NULL DEFAULT 0,
ADD COLUMN consecutive_successes INTEGER NOT NULL DEFAULT 0,
ADD COLUMN last_checked_at TIMESTAMPTZ;

ALTER TABLE services
ADD CONSTRAINT services_runtime_status_valid
CHECK (
    runtime_status IN (
        'HEALTHY',
        'DEGRADED',
        'DOWN'
    )
);

ALTER TABLE services
ADD CONSTRAINT services_consecutive_failures_non_negative
CHECK (consecutive_failures >= 0);

ALTER TABLE services
ADD CONSTRAINT services_consecutive_successes_non_negative
CHECK (consecutive_successes >= 0);