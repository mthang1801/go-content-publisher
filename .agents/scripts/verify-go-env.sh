#!/usr/bin/env bash
set -euo pipefail

checks=(go git rg)
status=0

for cmd in "${checks[@]}"; do
  if command -v "$cmd" >/dev/null 2>&1; then
    echo "ok: $cmd"
  else
    echo "missing: $cmd" >&2
    status=1
  fi
done

if command -v go >/dev/null 2>&1; then
  go version
fi

exit "$status"
