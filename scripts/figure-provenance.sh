#!/usr/bin/env bash
set -euo pipefail
repo=${FIGURE_ROOT:-$(git rev-parse --show-toplevel)}
file=$repo/notes/state.md
[[ -f $file ]] || { echo 'figure-provenance: notes/state.md missing' >&2; exit 1; }
failed=0
while IFS= read -r line; do
  [[ $line =~ [0-9] ]] || continue
  [[ $line =~ [0-9]{4}-[0-9]{2}-[0-9]{2} || $line == *'`'* || $line =~ ([Ss]ample|[Tt]rial|[Ii]nvocation|[Rr]ound|not-reproducible) ]] || { echo "figure-provenance: unsupported numeric claim: $line" >&2; failed=1; }
done < <(awk '/^## Blockers/{inside=1; next} inside && /^## /{exit} inside{print}' "$file")
exit "$failed"
