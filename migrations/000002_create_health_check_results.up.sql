CREATE TABLE health_check_results (
    id BIGSERIAL PRIMARY KEY,

    service_id BIGINT NOT NULL,

    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    status_code INTEGER,

    latency_ms BIGINT NOT NULL,

    success BOOLEAN NOT NULL,

    error_message TEXT,

    CONSTRAINT health_check_results_service_fk
        FOREIGN KEY (service_id)
        REFERENCES services(id)
        ON DELETE CASCADE,

    CONSTRAINT health_check_results_status_code_valid
        CHECK (
            status_code IS NULL
            OR (
                status_code >= 100
                AND status_code <= 599
            )
        ),

    CONSTRAINT health_check_results_latency_non_negative
        CHECK (latency_ms >= 0)
);

CREATE INDEX idx_health_check_results_service_checked_at
    ON health_check_results (
        service_id,
        checked_at DESC
    );