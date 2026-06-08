#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_DIR="$ROOT_DIR/backend/api"
BACKUP_DIR="$ROOT_DIR/backups"

cd "$API_DIR"
set -a
if [ -f .env ]; then
  . ./.env
fi
set +a

if [ -z "${NETCORD_DATABASE_URL:-}" ]; then
  echo "NETCORD_DATABASE_URL is required" >&2
  exit 1
fi

mkdir -p "$BACKUP_DIR"
file="$BACKUP_DIR/netcord-$(date +%Y%m%d-%H%M%S).sql"
pg_dump "$NETCORD_DATABASE_URL" > "$file"
echo "Backup written to $file"
