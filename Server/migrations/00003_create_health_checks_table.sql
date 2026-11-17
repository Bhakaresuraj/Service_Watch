
-- +goose Up

CREATE TABLE health_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    service_id UUID NOT NULL
        REFERENCES services(id) ON DELETE CASCADE,

    status VARCHAR(20) NOT NULL,

    http_status_code INTEGER,

    response_time_ms BIGINT,

    error_message TEXT,

    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_health_check_status
        CHECK (status IN ('UP', 'DOWN')),

    CONSTRAINT chk_response_time
        CHECK (response_time_ms IS NULL OR response_time_ms >= 0),

    CONSTRAINT chk_http_status_code
        CHECK (
            http_status_code IS NULL
            OR http_status_code BETWEEN 100 AND 599
        )
);

CREATE INDEX idx_health_checks_service_id
ON health_checks(service_id);

CREATE INDEX idx_health_checks_checked_at
ON health_checks(checked_at);


-- +goose Down

DROP TABLE health_checks;
