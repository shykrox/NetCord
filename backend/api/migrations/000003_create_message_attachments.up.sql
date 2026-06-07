CREATE TABLE message_attachments (
    id UUID PRIMARY KEY,
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    channel_id UUID,
    bucket TEXT NOT NULL,
    object_key TEXT NOT NULL UNIQUE,
    original_filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT message_attachments_size_check CHECK (size_bytes > 0),
    CONSTRAINT message_attachments_channel_server_fk FOREIGN KEY (channel_id, server_id)
        REFERENCES channels(id, server_id) ON DELETE CASCADE,
    CONSTRAINT message_attachments_attached_consistency CHECK (
        (message_id IS NULL AND server_id IS NULL AND channel_id IS NULL)
        OR
        (message_id IS NOT NULL AND server_id IS NOT NULL AND channel_id IS NOT NULL)
    )
);

CREATE INDEX message_attachments_uploader_id_idx ON message_attachments (uploader_id);
CREATE INDEX message_attachments_message_id_idx ON message_attachments (message_id);
CREATE INDEX message_attachments_server_id_idx ON message_attachments (server_id);
