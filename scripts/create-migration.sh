#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <migration-name>" >&2
  exit 1
fi

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"
migrations_dir="${repo_root}/db/migrations"

cd "${repo_root}"

exec go run ./cmd/create-migration "${migrations_dir}" "$1"
