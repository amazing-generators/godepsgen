#!/usr/bin/env bash

set -Eeuo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# В этом репозитории один go-модуль, но скрипт оставлен общим для CI и hooks.
cd "$root_dir"

if [ -f go.work ]; then
    go work sync
fi

while IFS= read -r mod_file; do
    mod_dir="$(dirname "$mod_file")"
    echo "tidy: $mod_dir"
    (
        cd "$mod_dir"
        go mod tidy
    )
done < <(find "$root_dir" -name 'go.mod' -not -path '*/vendor/*' | sort)
