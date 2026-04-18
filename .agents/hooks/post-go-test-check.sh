#!/usr/bin/env bash
set -euo pipefail

root="${1:-$(pwd)}"
target="${2:-./...}"

if ! command -v go >/dev/null 2>&1; then
  echo "go is not installed; skipping go test check"
  exit 0
fi

cd "$root"
go test "$target"
