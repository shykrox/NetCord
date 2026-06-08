DROP INDEX IF EXISTS messages_channel_content_search_idx;
DROP INDEX IF EXISTS messages_channel_active_created_at_idx;
DROP INDEX IF EXISTS message_edits_message_id_idx;
DROP TABLE IF EXISTS message_edits;

ALTER TABLE messages
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS edited_at;
