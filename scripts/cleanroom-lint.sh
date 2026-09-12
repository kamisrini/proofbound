#!/usr/bin/env bash
set -euo pipefail
repo=${CLEANROOM_ROOT:-$(git rev-parse --show-toplevel)}
patterns=${PROOFBOUND_CLEANROOM_PATTERNS:-}
if [[ -z $patterns || ! -r $patterns ]]; then echo 'INERT cleanroom: no readable external pattern file' >&2; exit 0; fi
mapfile -t needles < <(sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' "$patterns" | rg -v '^(#|$)' || true)
((${#needles[@]})) || { echo 'cleanroom: configured pattern file has no active patterns' >&2; exit 1; }
failed=0
for needle in "${needles[@]}"; do
  [[ -n $needle ]] || { echo 'cleanroom: empty pattern' >&2; exit 1; }
  while IFS= read -r path; do echo "cleanroom: forbidden pattern matched tracked path $path" >&2; failed=1; done < <(git -C "$repo" grep -Il -F -e "$needle" -- . 2>/dev/null || true)
done
exit "$failed"
