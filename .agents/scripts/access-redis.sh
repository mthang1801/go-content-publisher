#!/usr/bin/env bash
set -euo pipefail

host="${REDIS_HOST:-127.0.0.1}"
port="${REDIS_PORT:-6379}"

if command -v redis-cli >/dev/null 2>&1; then
  exec redis-cli -h "$host" -p "$port"
fi

echo "redis-cli is not installed" >&2
exit 1
