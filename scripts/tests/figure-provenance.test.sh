#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"; tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT; mkdir -p "$tmp/notes"
printf '# state\n## Blockers\n- 7 defects\n' >"$tmp/notes/state.md"
if FIGURE_ROOT=$tmp scripts/figure-provenance.sh >/dev/null 2>&1; then exit 1; fi
printf '# state\n## Blockers\n- 7 defects in round 2\n' >"$tmp/notes/state.md"; FIGURE_ROOT=$tmp scripts/figure-provenance.sh
