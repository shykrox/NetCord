ALTER TABLE messages
    ADD COLUMN edited_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE TABLE message_edits (
    id UUID PRIMARY KEY,
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    old_content TEXT NOT NULL,
    new_content TEXT NOT NULL,
    edited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    edited_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX message_edits_message_id_idx ON message_edits (message_id, edited_at);
CREATE INDEX messages_channel_active_created_at_idx
    ON messages (channel_id, created_at, id)
    WHERE deleted_at IS NULL;
CREATE INDEX messages_channel_content_search_idx
    ON messages USING GIN (to_tsvector('simple', content))
    WHERE deleted_at IS NULL;
