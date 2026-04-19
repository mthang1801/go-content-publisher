#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${ROOT_DIR}/release/dist/windows"
VERSION="$(git -C "${ROOT_DIR}" describe --tags --always --dirty 2>/dev/null || echo dev)"
COMMIT="$(git -C "${ROOT_DIR}" rev-parse --short=12 HEAD 2>/dev/null || echo unknown)"
BUILD_TIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
LDFLAGS="-X go-content-bot/pkg/buildinfo.Version=${VERSION} -X go-content-bot/pkg/buildinfo.Commit=${COMMIT} -X go-content-bot/pkg/buildinfo.BuildTime=${BUILD_TIME}"

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

GOCACHE="${ROOT_DIR}/.cache/go-build" \
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
go build -ldflags "${LDFLAGS}" -o "${OUTPUT_DIR}/content-bot.exe" "${ROOT_DIR}/cmd/content-bot"
GOCACHE="${ROOT_DIR}/.cache/go-build" \
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
go build -ldflags "${LDFLAGS}" -o "${OUTPUT_DIR}/cli.exe" "${ROOT_DIR}/cmd/cli"

cp "${ROOT_DIR}/release/config.example.ini" "${OUTPUT_DIR}/config.ini"
cp "${ROOT_DIR}/release/README.md" "${OUTPUT_DIR}/README.md"
cp "${ROOT_DIR}/release/run-content-bot.bat" "${OUTPUT_DIR}/run-content-bot.bat"
cp "${ROOT_DIR}/release/run-cli.bat" "${OUTPUT_DIR}/run-cli.bat"
mkdir -p "${OUTPUT_DIR}/db/migrations"
cp "${ROOT_DIR}"/db/migrations/*.sql "${OUTPUT_DIR}/db/migrations/"

echo "Windows release created at ${OUTPUT_DIR}"
