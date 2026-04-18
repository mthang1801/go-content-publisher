#!/usr/bin/env bash
set -euo pipefail

root="${1:-$(pwd)}"

mkdir -p \
  "$root/cmd/api" \
  "$root/cmd/worker" \
  "$root/cmd/migrate" \
  "$root/cmd/cli" \
  "$root/internal" \
  "$root/pkg" \
  "$root/docs/plan" \
  "$root/changelogs/changes"

echo "initialized Go service skeleton directories under $root"
