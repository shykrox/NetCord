#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_DIR="$ROOT_DIR/backend/api"

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

for migration in migrations/*.up.sql; do
  echo "Applying $migration"
  psql "$NETCORD_DATABASE_URL" -v ON_ERROR_STOP=1 -f "$migration"
done
