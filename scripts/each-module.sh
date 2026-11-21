#!/usr/bin/env bash
# Run a command in every Go module at or below the directory mise was invoked
# from. Modules listed in SKIP_MODULES (space-separated) are skipped.
# Usage: scripts/each-module.sh <command> [args...]
set -uo pipefail

prefix=$(cd "${MISE_ORIGINAL_CWD:-.}" && git rev-parse --show-prefix)
prefix=${prefix%/}
cd "$(git rev-parse --show-toplevel)" || exit 1

rc=0
for mod in $(git ls-files '*go.mod'); do
  dir=$(dirname "$mod")
  if [[ -n "$prefix" && "$dir" != "$prefix" && "$dir" != "$prefix"/* ]]; then
    continue
  fi
  [[ " ${SKIP_MODULES:-} " == *" $dir "* ]] && continue
  echo ":: $dir"
  (cd "$dir" && "$@") || rc=1
done
exit $rc
