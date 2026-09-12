#!/usr/bin/env bash
set -euo pipefail
command -v jq >/dev/null || { echo 'block-secrets: jq is required' >&2; exit 2; }
input=$(cat)
command_text=$(jq -er '.tool_input.command // ""' <<<"$input") || { echo 'block-secrets: malformed hook input' >&2; exit 2; }
if rg -q -- 'git[[:space:]]+push.*(--force|-f([[:space:]]|$))|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{20,}|BEGIN [A-Z ]*PRIVATE KEY' <<<"$command_text"; then
  echo 'block-secrets: blocked secret-shaped text or force push' >&2; exit 2
fi
if [[ $command_text == *'docs/decisions/INDEX.md'* ]] && rg -q -- '>|tee|sed[[:space:]]+-i' <<<"$command_text"; then
  echo 'block-secrets: generated decision index must be regenerated' >&2; exit 2
fi
