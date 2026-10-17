#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

scripts/deferrals.sh --check | grep -Fq 'deferrals: PASS'
scripts/deferrals.sh --open | grep -Fq $'DEF-P7-001\tOPEN\tPOSTPONED'
scripts/deferrals.sh --hold-type POSTPONED | grep -Fq 'DEF-P7-001'

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
