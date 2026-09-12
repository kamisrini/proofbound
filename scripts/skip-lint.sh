#!/usr/bin/env bash
set -euo pipefail
repo=${SKIP_LINT_ROOT:-$(git rev-parse --show-toplevel)}
log=${1:-${SKIP_LINT_LOG:-}}
allow=${SKIP_LINT_ALLOWLIST:-$repo/docs/allowed-skips.txt}
[[ -n $log && -s $log ]] || { echo 'skip-lint: nonempty Go test log required' >&2; exit 1; }
rg -q '^(ok|FAIL|\?|--- (PASS|FAIL|SKIP):)' "$log" || { echo 'skip-lint: log contains no Go test result' >&2; exit 1; }
declare -A allowed=()
while IFS= read -r line; do
  [[ $line =~ ^([A-Za-z0-9_/-]+)[[:space:]]+#[[:space:]]+(.+)$ ]] || continue
  allowed[${BASH_REMATCH[1]}]=1
done <"$allow"
failed=0
while IFS= read -r name; do
  [[ -n ${allowed[$name]+x} ]] || { echo "skip-lint: undeclared skip $name" >&2; failed=1; }
done < <(sed -nE 's/^[[:space:]]*--- SKIP: ([^ ]+).*/\1/p' "$log" | sort -u)
exit "$failed"
