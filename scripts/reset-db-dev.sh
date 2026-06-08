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

echo "This will drop and recreate all NetCord dev tables."
read -r -p "Type RESET to continue: " answer
if [ "$answer" != "RESET" ]; then
  echo "Aborted"
  exit 1
fi

for migration in $(ls migrations/*.down.sql | sort -r); do
  echo "Reverting $migration"
  psql "$NETCORD_DATABASE_URL" -v ON_ERROR_STOP=1 -f "$migration" || true
done

"$ROOT_DIR/scripts/migrate-up.sh"
