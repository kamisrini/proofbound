#!/usr/bin/env bash
# Proves the Law 4 advisory-expiry gate fires. Calibrated: positive controls that
# MUST fail, negative controls that MUST pass, and a neutral control (a BLOCKING
# row mentioning "advisory") that MUST be ignored — so a green verdict means the
# check discriminated, not that it never looked.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
g=$tmp/gates.md
hdr=$'| Check | Tier | Enforced by | Self-test | Expiry |\n|---|---|---|---|---|'

# positive control 1: an expired advisory MUST fail
printf '%s\n| x | ADVISORY | m | t | 2026-10-16 (past) |\n' "$hdr" >"$g"
if ADVISORY_GATES_FILE=$g ADVISORY_TODAY=2026-11-01 scripts/advisory-expiry.sh >/dev/null 2>&1; then
  echo "FAIL: expired advisory passed" >&2; exit 1; fi

# positive control 2: a dateless advisory MUST fail
printf '%s\n| x | ADVISORY | m | t | becomes a threshold later |\n' "$hdr" >"$g"
if ADVISORY_GATES_FILE=$g ADVISORY_TODAY=2026-11-01 scripts/advisory-expiry.sh >/dev/null 2>&1; then
  echo "FAIL: dateless advisory passed" >&2; exit 1; fi

# negative control 1: a future-dated advisory MUST pass
printf '%s\n| x | ADVISORY (measured) | m | t | 2026-12-31 (plan) |\n' "$hdr" >"$g"
ADVISORY_GATES_FILE=$g ADVISORY_TODAY=2026-11-01 scripts/advisory-expiry.sh >/dev/null

# negative control 2: expiry == today is still valid (boundary is inclusive)
printf '%s\n| x | ADVISORY | m | t | 2026-11-01 |\n' "$hdr" >"$g"
ADVISORY_GATES_FILE=$g ADVISORY_TODAY=2026-11-01 scripts/advisory-expiry.sh >/dev/null

# neutral control: a BLOCKING row whose note mentions an advisory date is NOT enforced
printf '%s\n| x | BLOCKING | m | t | — (graduated from advisory 2026-08-08) |\n' "$hdr" >"$g"
ADVISORY_GATES_FILE=$g ADVISORY_TODAY=2026-11-01 scripts/advisory-expiry.sh >/dev/null

# live control: the real registry must be clean under the enforced rule
scripts/advisory-expiry.sh >/dev/null
