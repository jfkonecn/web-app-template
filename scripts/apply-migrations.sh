#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 0 ]; then
  echo "Usage: $0" >&2
  exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$script_dir/.." && pwd)"
migrate_bin="$project_root/bin/migrate"
migrations_dir="$project_root/db/migrations"

db_host="${POSTGRES_HOST:-localhost}"
db_port="${POSTGRES_PORT:-5432}"
db_user="${POSTGRES_USER:-postgres}"
db_password="${POSTGRES_PASSWORD:-postgres}"
db_name="${POSTGRES_DB:-web_app_template}"

if [ ! -x "$migrate_bin" ]; then
  echo "Error: migrate binary not found or not executable at $migrate_bin" >&2
  echo "Run ./scripts/install-migrate.sh first." >&2
  exit 1
fi

if [ ! -d "$migrations_dir" ]; then
  echo "Error: migrations directory not found at $migrations_dir" >&2
  echo "Create one with ./scripts/create-migration.sh <name>." >&2
  exit 1
fi

if ! compgen -G "$migrations_dir/*.up.sql" > /dev/null; then
  echo "Error: no up migration files found in $migrations_dir" >&2
  exit 1
fi

database_url="postgres://${db_user}:${db_password}@${db_host}:${db_port}/${db_name}?sslmode=disable"
"$migrate_bin" -path "$migrations_dir" -database "$database_url" up
