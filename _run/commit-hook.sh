#!/usr/bin/env bash

set -Eeuo pipefail

echo "[HOOK]" "Commit"

run_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_path="$(cd "$run_dir/.." && pwd)"

#############################################################################

VERSION="$(go run github.com/amazing-generators/gometagen/cmd/gometagen@latest version print -source "$run_dir/values.yml")"
NAME="$(go run github.com/amazing-generators/gometagen/cmd/gometagen@latest git branch -source "$root_path")"

echo -e "$NAME [$VERSION] \n" "$(cat "$1")" > "$1"

#############################################################################

(
    cd "$root_path" || exit 1
    go test -v ./...
)

#############################################################################

exit 0
