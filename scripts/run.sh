#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

export APP_ENV="${APP_ENV:-development}"
export APP_HOST="${APP_HOST:-0.0.0.0}"
export APP_PORT="${APP_PORT:-8080}"
export OIDC_BASE_URL="http://localhost:5556/dex"
export OIDC_CLIENT_ID="example-app"
export OIDC_CLIENT_SECRET="ZXhhbXBsZS1hcHAtc2VjcmV0"
export OIDC_CALLBACK_URL="http://localhost:8080/callback"

cd "${repo_root}"

exec go run ./cmd/server
