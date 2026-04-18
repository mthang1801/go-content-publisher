#!/usr/bin/env bash
set -euo pipefail

root="${1:-$(pwd)}"

if ! command -v go >/dev/null 2>&1; then
  echo "go is not installed; skipping gofmt check"
  exit 0
fi

mapfile -t files < <(find "$root" -type f -name '*.go' -not -path '*/vendor/*')
if [[ "${#files[@]}" -eq 0 ]]; then
  echo "no Go files found; skipping gofmt check"
  exit 0
fi

unformatted="$(gofmt -l "${files[@]}")"
if [[ -n "$unformatted" ]]; then
  echo "unformatted Go files detected:"
  echo "$unformatted"
  exit 1
fi

echo "gofmt check passed"
