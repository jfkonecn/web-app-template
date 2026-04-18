#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

if [[ -f "${repo_root}/.env" ]]; then
	set -a
	# Load local development environment variables.
	. "${repo_root}/.env"
	set +a
fi

export APP_ENV="${APP_ENV:?APP_ENV must be set}"
export APP_HOST="${APP_HOST:?APP_HOST must be set}"
export APP_PORT="${APP_PORT:?APP_PORT must be set}"
export SESSION_SECRET="${SESSION_SECRET:?SESSION_SECRET must be set}"
export POSTGRES_HOST="${POSTGRES_HOST:?POSTGRES_HOST must be set}"
export POSTGRES_PORT="${POSTGRES_PORT:?POSTGRES_PORT must be set}"
export POSTGRES_USER="${POSTGRES_USER:?POSTGRES_USER must be set}"
export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:?POSTGRES_PASSWORD must be set}"
export POSTGRES_DB="${POSTGRES_DB:?POSTGRES_DB must be set}"
export POSTGRES_SSLMODE="${POSTGRES_SSLMODE:?POSTGRES_SSLMODE must be set}"
export OIDC_BASE_URL="${OIDC_BASE_URL:?OIDC_BASE_URL must be set}"
export OIDC_CLIENT_ID="${OIDC_CLIENT_ID:?OIDC_CLIENT_ID must be set}"
export OIDC_CLIENT_SECRET="${OIDC_CLIENT_SECRET:?OIDC_CLIENT_SECRET must be set}"
export OIDC_CALLBACK_URL="${OIDC_CALLBACK_URL:?OIDC_CALLBACK_URL must be set}"

cd "${repo_root}"

exec go run ./cmd/server
