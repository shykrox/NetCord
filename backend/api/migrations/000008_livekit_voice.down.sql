DROP TABLE IF EXISTS voice_sessions;

ALTER TABLE channels
    DROP CONSTRAINT channels_type_check,
    ADD CONSTRAINT channels_type_check CHECK (type = 'text');
