ALTER TABLE channels
    DROP CONSTRAINT channels_type_check,
    ADD CONSTRAINT channels_type_check CHECK (type IN ('text', 'voice'));

CREATE TABLE voice_sessions (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_name TEXT NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    left_at TIMESTAMPTZ,
    UNIQUE (channel_id, user_id)
);

CREATE INDEX voice_sessions_channel_active_idx
    ON voice_sessions (channel_id)
    WHERE left_at IS NULL;
