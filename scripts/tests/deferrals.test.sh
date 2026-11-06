#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

# Capture then grep a here-string rather than `producer | grep -Fq`: grep -q closes
# the pipe on first match, the producer dies of SIGPIPE, and pipefail promotes 141
# to the pipeline's exit. The `out=$(...)` assignment still fails the test under
# set -e if the producer errors, so producer-exit-status coverage is preserved.
out=$(scripts/deferrals.sh --check); grep -Fq 'deferrals: PASS' <<<"$out"
# DEF-P7-001 was RESOLVED 2026-11-05 (owner go/no-go, VD-p7-def001-resolve-2026-11-05), so it is no
# longer an OPEN/BLOCKED row. The live OPEN rows are DEF-KERNELCHECK-FLAKE-001 (DEFERRED) and
# DEF-P7-REGENLOOP-001 (PENDING_PREREQUISITE); assert the lister/filter against a current OPEN row.
out=$(scripts/deferrals.sh --open); grep -Fq $'DEF-KERNELCHECK-FLAKE-001\tOPEN\tDEFERRED' <<<"$out"
out=$(scripts/deferrals.sh --hold-type DEFERRED); grep -Fq 'DEF-KERNELCHECK-FLAKE-001' <<<"$out"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/docs" "$tmp/scripts"
cp scripts/deferrals.sh "$tmp/scripts/deferrals.sh"
printf '%s\n' '# fixture evidence' >"$tmp/docs/evidence.md"
printf '%s\n' '# register' >"$tmp/docs/deferrals.tsv"
printf '%s\n' 'id	status	hold_type	owner	created	phase	subject	reason	review_trigger	unblock_condition	evidence' >>"$tmp/docs/deferrals.tsv"
printf '%s\n' 'DEF-X-001	OPEN	AWAITING_OWNER	fixture	2026-10-17	P7	fixture	needs owner choice	owner decision	owner choice	docs/evidence.md' >>"$tmp/docs/deferrals.tsv"
"$tmp/scripts/deferrals.sh" --root "$tmp" --check >/dev/null

sed -i 's/AWAITING_OWNER/NOT_A_STATUS/' "$tmp/docs/deferrals.tsv"
if "$tmp/scripts/deferrals.sh" --root "$tmp" --check >/dev/null 2>&1; then
  echo 'invalid hold_type was accepted' >&2
  exit 1
fi

echo 'ok deferrals'
