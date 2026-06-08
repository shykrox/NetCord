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

go run ./cmd/server
