#!/usr/bin/env bash
# Law 4: nothing advisory without an owner and an ISO expiry date.
# Every ADVISORY-tier row in the gate registry (docs/gates.md) must carry an ISO
# expiry date that has not passed. A dateless or expired advisory fails the build;
# the owner graduates it to blocking or deletes it — an advisory may not linger.
#
# Reads ONLY the Tier column to decide what is advisory, so a BLOCKING row whose
# notes happen to mention the word "advisory" (e.g. "graduated from advisory …")
# is never enforced. Self-test: scripts/tests/advisory-expiry.test.sh
set -euo pipefail

gates=${ADVISORY_GATES_FILE:-docs/gates.md}
[[ -f $gates ]] || { echo "advisory-expiry: gate registry not found: $gates" >&2; exit 2; }
today=${ADVISORY_TODAY:-$(date -u +%F)}

bad=0
while IFS=$'\t' read -r lineno tier expiry; do
  date=$(printf '%s\n' "$expiry" | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}' | head -n1 || true)
  if [[ -z $date ]]; then
    echo "$gates:$lineno ADVISORY row has no ISO expiry date (Law 4): [$tier]" >&2
    bad=1
    continue
  fi
  # ISO dates sort lexicographically == chronologically; boundary is inclusive
  # (an advisory expiring today is still valid today).
  if [[ $date < $today ]]; then
    echo "$gates:$lineno ADVISORY expired $date (today $today) — graduate to blocking or delete (Law 4)" >&2
    bad=1
  fi
done < <(awk -F'|' '
  /^[[:space:]]*\|/ {
    tier = $3; gsub(/^[[:space:]]+|[[:space:]]+$/, "", tier)
    if (toupper(tier) ~ /ADVISORY/) {
      expiry = $(NF - 1); gsub(/^[[:space:]]+|[[:space:]]+$/, "", expiry)
      printf "%d\t%s\t%s\n", NR, tier, expiry
    }
  }' "$gates")

exit "$bad"
