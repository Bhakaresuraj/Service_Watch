
-- +goose Up

CREATE TABLE incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    service_id UUID NOT NULL
        REFERENCES services(id) ON DELETE CASCADE,

    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',

    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    resolved_at TIMESTAMPTZ,

    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_incident_status
        CHECK (status IN ('OPEN', 'RESOLVED')),

    CONSTRAINT chk_incident_resolution
        CHECK (
            (status = 'OPEN' AND resolved_at IS NULL)
            OR
            (status = 'RESOLVED' AND resolved_at IS NOT NULL)
        )
);

-- Only one OPEN incident per service
CREATE UNIQUE INDEX idx_incidents_one_open_per_service
ON incidents(service_id)
WHERE status = 'OPEN';

-- Retrieve incidents belonging to a service
CREATE INDEX idx_incidents_service_id
ON incidents(service_id);

-- Retrieve incidents by creation time
CREATE INDEX idx_incidents_created_at
ON incidents(created_at);


-- +goose Down

DROP TABLE incidents;
