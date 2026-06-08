# NetCord API

Base local URL:

```text
http://127.0.0.1:8080
```

Most endpoints require:

```text
Authorization: Bearer <token>
```

Errors use:

```json
{"error":{"code":"validation_error","message":"request validation failed","fields":{}}}
```

## Auth and Users

- `GET /health`
- `POST /auth/register` with `username`, `email`, `password`
- `POST /auth/login` with `email`, `password`
- `GET /users/me`
- `PATCH /users/me/presence` with `status` (`online`, `idle`, `dnd`, `offline`) and optional `custom_status`

## Servers and Channels

- `POST /servers` with `name`, optional `description`
- `GET /servers`
- `GET /servers/{server_id}`
- `POST /servers/{server_id}/channels` with `name`, optional `type` (`text` or `voice`)
- `GET /servers/{server_id}/channels`

The server creator is inserted as owner. Users can only read servers/channels where they are members.

## Messages

- `GET /channels/{channel_id}/messages?before=<message_id>&after=<message_id>&limit=50`
- `GET /channels/{channel_id}/messages/search?q=<query>&limit=50`
- `POST /channels/{channel_id}/messages` with `content` and optional `attachments`
- `PATCH /messages/{message_id}` with `content`
- `DELETE /messages/{message_id}`

Only text channels accept messages. Delete is soft delete; deleted messages are hidden from list/search. Only the author can edit/delete for now.

## Files

- `POST /files/upload` as `multipart/form-data` field `file`
- `GET /files/{id}`

Files are stored privately in MinIO/S3. API responses expose only attachment IDs and `/files/{id}` download URLs.

## Friends and DMs

- `POST /friends/requests` with `recipient_id`
- `GET /friends/requests`
- `POST /friends/requests/{id}/accept`
- `POST /friends/requests/{id}/decline`
- `GET /friends`
- `DELETE /friends/{user_id}`
- `POST /dm` with `member_ids`, optional `name`
- `GET /dm`
- `GET /dm/{conversation_id}/messages?limit=50`
- `POST /dm/{conversation_id}/messages` with `content`

DM access is restricted to conversation members. Blocked users cannot be added to a DM.

## Roles, Permissions, Invites

- `POST /servers/{id}/roles`
- `GET /servers/{id}/roles`
- `PATCH /roles/{id}`
- `DELETE /roles/{id}`
- `PUT /servers/{id}/members/{user_id}/roles/{role_id}`
- `DELETE /servers/{id}/members/{user_id}/roles/{role_id}`
- `POST /servers/{id}/invites`
- `GET /invites/{code}`
- `POST /invites/{code}/join`

Owners have all permissions. `ADMINISTRATOR` bypasses checks. Role management requires `MANAGE_ROLES`; invite creation requires `CREATE_INVITE`.

## Voice

- `POST /voice/join` with `channel_id`
- `POST /voice/leave` with `channel_id`

Voice join requires a `voice` channel, membership, `CONNECT_VOICE`, and LiveKit env config. The join response includes `url`, `token`, `room`, `server_id`, `channel_id`, `user_id`.

## AI Commands

Posting `/ask <prompt>` or `/draw <prompt>` in a text channel creates an `ai_jobs` queue row and broadcasts `job.progress`. Execution workers and ComfyUI output attachment are foundation work for the next implementation step.
