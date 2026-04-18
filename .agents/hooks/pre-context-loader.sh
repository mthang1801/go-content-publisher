#!/usr/bin/env bash
set -euo pipefail

root="${1:-$(pwd)}"

required=(
  "$root/docs/plan/progress.md"
  "$root/changelogs/CHANGELOG.md"
  "$root/.agents/project.md"
  "$root/.agents/agents/architecture.md"
)

for file in "${required[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "missing required context file: $file" >&2
    exit 1
  fi
done

echo "core context files present"
