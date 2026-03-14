#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 0 ]; then
  echo "Usage: $0" >&2
  exit 1
fi

bin_dir="bin"
destination="$bin_dir"

archives=( "$bin_dir"/*migrate*.tar.gz "$bin_dir"/*migrate*.tgz )
matches=()

for candidate in "${archives[@]}"; do
  if [ -f "$candidate" ]; then
    matches+=( "$candidate" )
  fi
done

if [ "${#matches[@]}" -eq 0 ]; then
  echo "Error: no migrate archive found in $bin_dir" >&2
  exit 1
fi

if [ "${#matches[@]}" -gt 1 ]; then
  echo "Error: multiple migrate archives found in $bin_dir:" >&2
  printf ' - %s\n' "${matches[@]}" >&2
  exit 1
fi

archive="${matches[0]}"
tar xvz -f "$archive" -C "$destination"
