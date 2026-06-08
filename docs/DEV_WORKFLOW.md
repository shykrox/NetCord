# NetCord Dev Workflow

## Backend

```bash
cd backend/api
cp .env.example .env
nano .env
go mod tidy
go test ./...
go run ./cmd/server
```

Useful scripts from the repo root:

```bash
./scripts/migrate-up.sh
./scripts/run-api-dev.sh
./scripts/build-api.sh
./scripts/backup-db.sh
```

`./scripts/reset-db-dev.sh` is destructive and asks for confirmation.

## Desktop Client

Open `client/desktop-qt/CMakeLists.txt` in Qt Creator with `Desktop Qt 6.11.1 MinGW 64-bit`, or run:

```powershell
.\scripts\build-windows-client.ps1
```

The login screen stores the API URL and JWT with `QSettings`. The client keeps a small SQLite cache for servers, channels, and messages.

## Smoke Test

1. Register/login.
2. Create server.
3. Create text channel.
4. Send message.
5. Upload and send an attachment.
6. Open a second client and verify realtime message/typing.
7. Edit/delete/search messages.
8. Create a voice channel and verify the voice UI.
9. Create invite, role, and friend request.
10. Send `/ask hello` or use `/ai/ask`.
