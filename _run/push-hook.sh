#!/usr/bin/env bash

set -Eeuo pipefail

echo "[HOOK]" "Push"

run_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_path="$(cd "$run_dir/.." && pwd)"

#############################################################################

(
    cd "$root_path" || exit 1
    export CGO_ENABLED=1

    go mod tidy

    old_ver="$(go run github.com/amazing-generators/gometagen/cmd/gometagen@latest version print -source "$run_dir/values.yml")"
    version="$(go run github.com/amazing-generators/gometagen/cmd/gometagen@latest version patch -source "$run_dir/values.yml")"

    echo "Updated patch-ver: $old_ver >> $version"

    echo "==> Running tests with race detector..."
    go test -race -v ./...

    echo ""
    echo "==> Running benchmarks..."
    go test -bench=. -run=NONE -benchmem -v ./...

    echo ""
    echo "[HOOK] All tests, benchmarks and generators passed"
)

#############################################################################

exit 0
