-- +goose Up

CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY
        REFERENCES users(id) ON DELETE CASCADE,

    email_notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    down_alerts_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    recovery_alerts_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- +goose Down

DROP TABLE user_settings;
