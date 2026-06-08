# Known Limitations

## Working

- Auth register/login/JWT.
- `/users/me`, profile update, public user lookup.
- Servers, text/voice channel creation, server/channel list.
- Text messages with attachments.
- Message edit/delete/search/pagination.
- WebSocket message, typing, presence, friend, DM, voice, and AI job events.
- MinIO-backed upload/download through the API.
- Friends, friend requests, DM list/messages.
- Roles, basic permissions, invites.
- Qt client login, server/channel/chat, upload, cache, typing, edit/delete/search.
- Qt 6.11.1 MinGW configure/build.

## Scaffold Only

- LiveKit audio in the Qt client. Backend token generation exists, UI flow exists, but native audio capture/playback is not integrated.
- ComfyUI execution. `/ask`, `/draw`, and `ai_jobs` exist, but no worker consumes jobs and posts generated attachments yet.
- Advanced role assignment UI. Backend endpoints exist; client only has basic role creation/list display.
- Native OS notifications. Current notifications are in-app status/toast messages.
- Channel permission overwrites resolution. Table exists; full allow/deny resolution is future work.
- Read markers/unread counts are not persisted yet.

## Required Environment

- PostgreSQL with all migrations applied.
- MinIO/S3 settings for file upload.
- LiveKit env for successful `/voice/join`.
- ComfyUI URL for future AI worker execution.

## Risks

- `scripts/migrate-up.sh` blindly applies all `.up.sql` files. Use it on clean/dev DBs or with a migration tracking tool in production.
- Soft-deleted messages remain in the database.
- API broadcasts are in-memory per process; multiple API instances need Redis pub/sub later.
