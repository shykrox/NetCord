ALTER TABLE users
    ADD COLUMN is_bot BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE ai_jobs (
    id UUID PRIMARY KEY,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    command TEXT NOT NULL,
    prompt TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    progress INT NOT NULL DEFAULT 0,
    result_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    CONSTRAINT ai_jobs_command_check CHECK (command IN ('ask', 'draw')),
    CONSTRAINT ai_jobs_status_check CHECK (status IN ('queued', 'running', 'completed', 'failed', 'cancelled')),
    CONSTRAINT ai_jobs_progress_check CHECK (progress >= 0 AND progress <= 100)
);

CREATE INDEX ai_jobs_channel_created_idx ON ai_jobs (channel_id, created_at DESC);
CREATE INDEX ai_jobs_status_idx ON ai_jobs (status);
