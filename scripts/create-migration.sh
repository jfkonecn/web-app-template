#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <migration_name>" >&2
  exit 1
fi

name="$1"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$script_dir/.." && pwd)"
migrate_bin="$project_root/bin/migrate"
migrations_dir="$project_root/db/migrations"

if [ ! -x "$migrate_bin" ]; then
  echo "Error: migrate binary not found or not executable at $migrate_bin" >&2
  echo "Run ./scripts/install-migrate.sh first." >&2
  exit 1
fi

mkdir -p "$migrations_dir"
"$migrate_bin" create -ext sql -dir "$migrations_dir" -seq "$name"
