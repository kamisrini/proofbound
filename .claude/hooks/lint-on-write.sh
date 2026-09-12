#!/usr/bin/env bash
set -euo pipefail
command -v jq >/dev/null || { echo 'lint-on-write: jq is required' >&2; exit 2; }
input=$(cat)
path=$(jq -er '.tool_input.file_path // .tool_input.path // empty' <<<"$input") || exit 0
[[ $path == *.md ]] || exit 0
root=$(git rev-parse --show-toplevel)
(cd "$root" && scripts/link-lint.sh)
