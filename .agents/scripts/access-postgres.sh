#!/usr/bin/env bash
set -euo pipefail

host="${PGHOST:-localhost}"
port="${PGPORT:-5432}"
user="${PGUSER:-postgres}"
db="${PGDATABASE:-postgres}"

if command -v psql >/dev/null 2>&1; then
  exec psql "host=$host port=$port user=$user dbname=$db"
fi

echo "psql is not installed" >&2
exit 1
