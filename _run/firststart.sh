#!/usr/bin/env bash

set -Eeuo pipefail

run_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dir="$(cd "$run_dir/.." && pwd)"

"$run_dir/scripts/git.sh" --add_commit
"$run_dir/scripts/git.sh" --add_push

cd "$root_dir"

if [ -f go.work ]; then
  go work sync
fi

# go generate .  # Нет codegen в этом репозитории.
./_run/scripts/go_tidy_all.sh
