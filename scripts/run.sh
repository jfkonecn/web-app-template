#!/usr/bin/env bash

set -euo pipefail

export APP_ENV="${APP_ENV:-development}"
export APP_HOST="${APP_HOST:-0.0.0.0}"
export APP_PORT="${APP_PORT:-8080}"

exec go run ./cmd/server
