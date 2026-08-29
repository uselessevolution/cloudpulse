CREATE TABLE incidents (
    id BIGSERIAL PRIMARY KEY,

    service_id BIGINT NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',

    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    resolved_at TIMESTAMPTZ,

    failure_count INTEGER NOT NULL DEFAULT 0,

    last_error_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT incidents_service_fk
        FOREIGN KEY (service_id)
        REFERENCES services(id)
        ON DELETE CASCADE,

    CONSTRAINT incidents_status_valid
        CHECK (
            status IN (
                'OPEN',
                'RESOLVED'
            )
        ),

    CONSTRAINT incidents_failure_count_non_negative
        CHECK (
            failure_count >= 0
        )
);

CREATE INDEX idx_incidents_service_status
    ON incidents (
        service_id,
        status
    );

CREATE INDEX idx_incidents_started_at
    ON incidents (
        started_at DESC
    );