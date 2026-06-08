CREATE TABLE user_presence (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'offline',
    custom_status TEXT,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_presence_status_check CHECK (status IN ('online', 'idle', 'dnd', 'offline'))
);

CREATE INDEX user_presence_status_idx ON user_presence (status);
