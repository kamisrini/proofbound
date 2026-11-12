#!/usr/bin/env bash
# Blocking: every gate definition in gates/ must carry an unexpired ISO `expires` date.
#
# `proofbound gates enforce` fail-closes on an expired gate definition, but `make check`
# does NOT run enforce (it needs a ledger + witnesses), so an expired enforce gate can rot
# silently while the daily gate stays green. Seen 2026-11-12: the seven enforce-mode gates
# all expired 2026-10-16 unnoticed (self-hosting probe). This compiles that lesson into
# `make check` so a lapsed gate definition can never again pass unseen.
#
# Fails CLOSED: a missing gates directory, no gate files, or a gate without a parseable ISO
# `expires` date is a failure, not a pass. Boundary is inclusive (expiring today is valid today).
# Self-test: scripts/tests/gate-expiry.test.sh
set -euo pipefail

dir=${GATE_EXPIRY_DIR:-gates}
today=${GATE_EXPIRY_TODAY:-$(date -u +%F)}
[[ $today =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] || { echo "gate-expiry: invalid today date: $today" >&2; exit 2; }
[[ -d $dir ]] || { echo "gate-expiry: gates directory not found: $dir" >&2; exit 2; }

shopt -s nullglob
files=("$dir"/*.yaml)
[[ ${#files[@]} -gt 0 ]] || { echo "gate-expiry: no gate definitions found in $dir" >&2; exit 2; }

bad=0
for g in "${files[@]}"; do
  date=$(grep -oE '"expires"[[:space:]]*:[[:space:]]*"[0-9]{4}-[0-9]{2}-[0-9]{2}"' "$g" \
           | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}' | head -n1 || true)
  if [[ -z $date ]]; then
    echo "gate-expiry: $g has no ISO \"expires\" date (Law 4)" >&2
    bad=1
    continue
  fi
  # ISO dates sort lexicographically == chronologically; inclusive boundary.
  if [[ $date < $today ]]; then
    echo "gate-expiry: $g expired $date (today $today) — renew the gate definition or retire it" >&2
    bad=1
  fi
done
exit "$bad"
