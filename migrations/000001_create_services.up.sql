CREATE TABLE services (
    id BIGSERIAL PRIMARY KEY,

    name VARCHAR(120) NOT NULL,

    url TEXT NOT NULL,

    expected_status INTEGER NOT NULL DEFAULT 200,

    check_interval_seconds INTEGER NOT NULL DEFAULT 30,

    timeout_seconds INTEGER NOT NULL DEFAULT 3,

    enabled BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT services_name_not_blank
        CHECK (LENGTH(TRIM(name)) > 0),

    CONSTRAINT services_expected_status_valid
        CHECK (
            expected_status >= 100
            AND expected_status <= 599
        ),

    CONSTRAINT services_check_interval_positive
        CHECK (check_interval_seconds > 0),

    CONSTRAINT services_timeout_positive
        CHECK (timeout_seconds > 0)
);