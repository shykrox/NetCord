#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_DIR="$ROOT_DIR/backend/api"
OUT_DIR="$ROOT_DIR/bin"

mkdir -p "$OUT_DIR"
cd "$API_DIR"
go mod tidy
go test ./...
go build -o "$OUT_DIR/netcord-api" ./cmd/server

echo "Built $OUT_DIR/netcord-api"
