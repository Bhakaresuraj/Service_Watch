
-- +goose Up

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    incident_id UUID NOT NULL
        REFERENCES incidents(id) ON DELETE CASCADE,

    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,

    type VARCHAR(20) NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',

    retry_count INTEGER NOT NULL DEFAULT 0,

    max_retries INTEGER NOT NULL DEFAULT 3,

    next_retry_at TIMESTAMPTZ,

    sent_at TIMESTAMPTZ,

    error_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_notification_type
        CHECK (type IN ('DOWN', 'RECOVERY')),

    CONSTRAINT chk_notification_status
        CHECK (status IN ('PENDING', 'SENT', 'FAILED')),

    CONSTRAINT chk_notification_retry_count
        CHECK (retry_count >= 0),

    CONSTRAINT chk_notification_max_retries
        CHECK (max_retries >= 0)
);

-- Find pending notifications efficiently
CREATE INDEX idx_notifications_pending
ON notifications(next_retry_at)
WHERE status = 'PENDING';

-- Find notifications belonging to a user
CREATE INDEX idx_notifications_user_id
ON notifications(user_id);

-- Find notifications belonging to an incident
CREATE INDEX idx_notifications_incident_id
ON notifications(incident_id);


-- +goose Down

DROP TABLE notifications;
