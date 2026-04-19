#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${ROOT_DIR}/release/dist/windows"

mkdir -p "${OUTPUT_DIR}"

GOCACHE="${ROOT_DIR}/.cache/go-build" \
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
go build -o "${OUTPUT_DIR}/content-bot.exe" "${ROOT_DIR}/cmd/content-bot"

cp "${ROOT_DIR}/release/config.example.ini" "${OUTPUT_DIR}/config.ini"
cp "${ROOT_DIR}/release/README-run.txt" "${OUTPUT_DIR}/README-run.txt"

echo "Windows release created at ${OUTPUT_DIR}"
