#!/usr/bin/env bash
# SessionStart orientation: prints the resume note, derived git facts, and cheap
# freshness nudges so a fresh session can answer "what's next" without asking.
# Fast and fail-safe by construction: it runs no gate, touches nothing, and always
# exits 0 — a broken orientation print must never block a session from starting.
set -uo pipefail

root=$(git rev-parse --show-toplevel 2>/dev/null) || exit 0
cd "$root" || exit 0

echo '── Proofbound session orientation ───────────────────────────────'

# Derived git facts (single home: scripts/gen-state.sh).
if [[ -x scripts/gen-state.sh ]]; then
  scripts/gen-state.sh 2>/dev/null || true
fi
echo

# The first Resume block of the durable state note — the exact next action lives here.
if [[ -f notes/state.md ]]; then
  echo '── notes/state.md · Resume ──────────────────────────────────────'
  awk '
    /^## Resume/ { if (started) exit; started=1 }
    started && /^## / && !/^## Resume/ { exit }
    started { print }
  ' notes/state.md
  echo
fi

# Cheap, honest nudges (no gate run).
if [[ -n $(git status --porcelain --untracked-files=all 2>/dev/null) ]]; then
  echo '⚠  worktree is dirty — uncommitted work is present (git status -sb below).'
fi
if [[ -x scripts/state-freshness.sh ]] && ! STATE_FRESHNESS_ROOT="$root" scripts/state-freshness.sh >/dev/null 2>&1; then
  echo '⚠  notes/state.md is stale — HEAD has moved past its last update.'
fi

echo '── git status -sb ───────────────────────────────────────────────'
git status -sb 2>/dev/null | head -20
echo '─────────────────────────────────────────────────────────────────'
echo 'Next: run /proofbound-next to orient and pick the next authorized task.'
echo 'The gate is on demand — run `make check` (it is not run here; it is slow).'

exit 0
