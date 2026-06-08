CREATE TABLE friend_requests (
    id UUID PRIMARY KEY,
    requester_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    responded_at TIMESTAMPTZ,
    CONSTRAINT friend_requests_not_self CHECK (requester_id <> recipient_id),
    CONSTRAINT friend_requests_status_check CHECK (status IN ('pending', 'accepted', 'declined'))
);

CREATE UNIQUE INDEX friend_requests_pending_unique
    ON friend_requests (requester_id, recipient_id)
    WHERE status = 'pending';

CREATE TABLE friends (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, friend_id),
    CONSTRAINT friends_not_self CHECK (user_id <> friend_id)
);

CREATE TABLE blocked_users (
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (blocker_id, blocked_id),
    CONSTRAINT blocked_users_not_self CHECK (blocker_id <> blocked_id)
);

CREATE TABLE dm_conversations (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT dm_conversations_type_check CHECK (type IN ('dm', 'group'))
);

CREATE TABLE dm_members (
    conversation_id UUID NOT NULL REFERENCES dm_conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE dm_messages (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES dm_conversations(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    edited_at TIMESTAMPTZ
);

CREATE INDEX friend_requests_recipient_status_idx ON friend_requests (recipient_id, status);
CREATE INDEX friends_friend_id_idx ON friends (friend_id);
CREATE INDEX blocked_users_blocked_id_idx ON blocked_users (blocked_id);
CREATE INDEX dm_members_user_id_idx ON dm_members (user_id);
CREATE INDEX dm_messages_conversation_created_at_idx ON dm_messages (conversation_id, created_at);
