#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
ROOT_DIR="$(CDPATH= cd -- "${SCRIPT_DIR}/.." && pwd)"
GO_CMD="${GO_CMD:-go}"

cd "${ROOT_DIR}"

run_cli() {
  "${GO_CMD}" run ./cmd/cli "$@"
}

run_cli settings-set enable_twitter_crawler true
run_cli settings-set enable_telegram_crawler true
run_cli settings-set enable_rewrite_processor true
run_cli settings-set auto_publish true
run_cli settings-set crawl_interval_seconds 300
run_cli settings-set process_interval_seconds 30
run_cli settings-set publish_interval_seconds 10
run_cli settings-set twitter_publish_interval_seconds 600
run_cli settings-set rewrite_provider gemini
run_cli settings-set enable_twitter_publish_vi false
run_cli settings-set enable_twitter_publish_en false

echo "Applied preset: demo"
echo "Restart API/worker unless you only changed auto_publish."
