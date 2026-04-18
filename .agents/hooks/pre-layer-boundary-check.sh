#!/usr/bin/env bash
set -euo pipefail

root="${1:-$(pwd)}"
status=0

domain_dir="$root/internal/domain"
if [[ -d "$domain_dir" ]]; then
  if rg -n 'gin|gorm|redis|kafka|rabbitmq|amqp' "$domain_dir" >/dev/null 2>&1; then
    echo "forbidden infrastructure or framework reference detected under internal/domain" >&2
    status=1
  fi
fi

pkg_dir="$root/pkg"
if [[ -d "$pkg_dir" ]]; then
  if rg -n 'internal/' "$pkg_dir" >/dev/null 2>&1; then
    echo "pkg must not depend on internal packages" >&2
    status=1
  fi
fi

exit "$status"
