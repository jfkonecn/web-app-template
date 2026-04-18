#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"
migrations_dir="${repo_root}/db/migrations"

if [[ -f "${repo_root}/.env" ]]; then
	set -a
	. "${repo_root}/.env"
	set +a
fi

export POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
export POSTGRES_PORT="${POSTGRES_PORT:-5432}"
export POSTGRES_USER="${POSTGRES_USER:-postgres}"
export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
export POSTGRES_DB="${POSTGRES_DB:-web_app_template}"
export POSTGRES_SSLMODE="${POSTGRES_SSLMODE:-disable}"

cd "${repo_root}"

exec go run ./cmd/migrate-up "${migrations_dir}"
