#!/usr/bin/env bash
set -euo pipefail
command -v jq >/dev/null || { echo 'block-generated-edit: jq is required' >&2; exit 2; }
input=$(cat)
path=$(jq -er '.tool_input.file_path // .tool_input.path // empty' <<<"$input") || { echo 'block-generated-edit: file path is required' >&2; exit 2; }
root=$(git rev-parse --show-toplevel)
case $path in /*) target=$path;; *) target=$root/$path;; esac
if [[ -f $target ]] && rg -q '@generated' "$target"; then echo 'block-generated-edit: generated files are not hand-edited' >&2; exit 2; fi
