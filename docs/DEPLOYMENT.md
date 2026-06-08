# NetCord Deployment

Target VM path:

```text
/opt/netcord
```

## Update

```bash
cd /opt/netcord
git pull
mkdir -p backups
./scripts/backup-db.sh
```

## Configure

```bash
cd /opt/netcord/backend/api
cp .env.example .env
nano .env
```

Required values:

- `NETCORD_DATABASE_URL`
- `NETCORD_JWT_SECRET`
- `NETCORD_MINIO_ENDPOINT`
- `NETCORD_MINIO_ACCESS_KEY`
- `NETCORD_MINIO_SECRET_KEY`

Optional values:

- `NETCORD_LIVEKIT_URL`
- `NETCORD_LIVEKIT_API_KEY`
- `NETCORD_LIVEKIT_API_SECRET`
- `NETCORD_COMFYUI_URL`

## Migrate and Test

```bash
cd /opt/netcord
./scripts/migrate-up.sh

cd backend/api
set -a
. .env
set +a
go mod tidy
go test ./...
```

## Run

```bash
cd /opt/netcord/backend/api
set -a
. .env
set +a
go run ./cmd/server
```

For production, build a binary:

```bash
cd /opt/netcord
./scripts/build-api.sh
./bin/netcord-api
```

## Systemd Sketch

```ini
[Unit]
Description=NetCord API
After=network.target docker.service

[Service]
WorkingDirectory=/opt/netcord/backend/api
EnvironmentFile=/opt/netcord/backend/api/.env
ExecStart=/opt/netcord/bin/netcord-api
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```
