# NetCord Database

Migrations live in `backend/api/migrations`.

## Current Tables

- `users`: accounts, bcrypt password hash, profile fields, `banner_url`, `is_bot`.
- `user_presence`: status, custom status, last seen.
- `servers`: server metadata and owner.
- `server_members`: membership and legacy owner/member role marker.
- `channels`: text and voice channels.
- `messages`: text messages, `edited_at`, `deleted_at` for soft delete.
- `message_edits`: edit history.
- `message_attachments`: private MinIO/S3 attachment metadata.
- `friend_requests`: pending/accepted/declined requests.
- `friends`: symmetric friend edges.
- `blocked_users`: block edges used by DM validation.
- `dm_conversations`: direct and group DM containers.
- `dm_members`: DM membership.
- `dm_messages`: DM message history.
- `roles`: per-server permission roles.
- `member_roles`: role assignments.
- `channel_permission_overwrites`: foundation table for channel-level allow/deny.
- `server_invites`: invite codes, use counts, expiration.
- `audit_logs`: admin action logs.
- `voice_sessions`: foundation table for voice presence/session audit.
- `ai_jobs`: queued `/ask` and `/draw` jobs.

## Notes

- Message deletion is soft delete with `messages.deleted_at`.
- Text search currently uses basic SQL matching and excludes deleted messages.
- Attachments store internal bucket/object keys; clients only see `/files/{id}`.
- Owners have all permissions; other users get default member permissions plus assigned roles.
- Voice channels are represented by `channels.type = 'voice'`.
- AI jobs are queued by chat commands; an external worker can process `ai_jobs` later.
