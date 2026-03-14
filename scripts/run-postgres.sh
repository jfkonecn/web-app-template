#!/usr/bin/env bash

set -euo pipefail
compose_file="docker-dev/docker-compose.yml"

if docker compose version >/dev/null 2>&1; then
  compose_cmd=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  compose_cmd=(docker-compose)
else
  echo "Error: docker compose is not installed." >&2
  exit 1
fi

"${compose_cmd[@]}" -f "$compose_file" up postgres
