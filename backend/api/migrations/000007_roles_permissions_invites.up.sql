CREATE TABLE roles (
    id UUID PRIMARY KEY,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    color TEXT,
    position INT NOT NULL DEFAULT 0,
    permissions BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(server_id, name)
);

CREATE TABLE member_roles (
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(server_id, user_id, role_id),
    CONSTRAINT member_roles_member_fk FOREIGN KEY (server_id, user_id)
        REFERENCES server_members(server_id, user_id) ON DELETE CASCADE
);

CREATE TABLE channel_permission_overwrites (
    id UUID PRIMARY KEY,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    allow BIGINT NOT NULL DEFAULT 0,
    deny BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT channel_permission_overwrites_target_check CHECK (
        (role_id IS NOT NULL AND user_id IS NULL)
        OR (role_id IS NULL AND user_id IS NOT NULL)
    )
);

CREATE TABLE server_invites (
    id UUID PRIMARY KEY,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code TEXT NOT NULL UNIQUE,
    max_uses INT,
    uses INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT server_invites_max_uses_check CHECK (max_uses IS NULL OR max_uses > 0)
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    target_type TEXT,
    target_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX roles_server_id_idx ON roles (server_id);
CREATE INDEX member_roles_user_idx ON member_roles (user_id);
CREATE INDEX channel_permission_overwrites_channel_idx ON channel_permission_overwrites (channel_id);
CREATE INDEX server_invites_server_id_idx ON server_invites (server_id);
CREATE INDEX audit_logs_server_created_idx ON audit_logs (server_id, created_at DESC);
