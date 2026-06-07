CREATE TABLE servers (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    icon_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE server_members (
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, user_id),
    CONSTRAINT server_members_role_check CHECK (role IN ('owner', 'member'))
);

CREATE TABLE channels (
    id UUID PRIMARY KEY,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'text',
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT channels_type_check CHECK (type = 'text'),
    CONSTRAINT channels_id_server_unique UNIQUE (id, server_id),
    CONSTRAINT channels_server_name_unique UNIQUE (server_id, name)
);

CREATE TABLE messages (
    id UUID PRIMARY KEY,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT messages_channel_server_fk FOREIGN KEY (channel_id, server_id)
        REFERENCES channels(id, server_id) ON DELETE CASCADE
);

CREATE INDEX server_members_user_id_idx ON server_members (user_id);
CREATE INDEX channels_server_id_idx ON channels (server_id);
CREATE INDEX messages_channel_created_at_idx ON messages (channel_id, created_at);
CREATE INDEX messages_author_id_idx ON messages (author_id);
