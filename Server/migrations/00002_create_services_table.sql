
-- +goose Up

CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,
    url TEXT NOT NULL,

    monitoring_interval INTEGER NOT NULL DEFAULT 60,
    timeout_seconds INTEGER NOT NULL DEFAULT 10,

    failure_threshold INTEGER NOT NULL DEFAULT 3,
    recovery_threshold INTEGER NOT NULL DEFAULT 2,

    current_status VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN',

    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    last_checked_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_monitoring_interval
        CHECK (monitoring_interval >= 10),

    CONSTRAINT chk_timeout_seconds
        CHECK (timeout_seconds > 0),

    CONSTRAINT chk_failure_threshold
        CHECK (failure_threshold > 0),

    CONSTRAINT chk_recovery_threshold
        CHECK (recovery_threshold > 0),

    CONSTRAINT chk_current_status
        CHECK (current_status IN ('UNKNOWN', 'UP', 'DOWN'))
);

CREATE INDEX idx_services_user_id ON services(user_id);

-- +goose Down

DROP TABLE services;
