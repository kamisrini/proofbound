#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"; tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT; mkdir -p "$tmp/notes/journal" "$tmp/scripts"
printf 'LESSON: class=repeat\nLESSON: class=repeat\n' >"$tmp/notes/journal/a.md"
if LESSON_ROOT=$tmp scripts/lesson-recurrence.sh >/dev/null 2>&1; then exit 1; fi
printf 'LESSON: class=repeat response=scripts/fix.sh\nLESSON: class=repeat\n' >"$tmp/notes/journal/a.md"; touch "$tmp/scripts/fix.sh"
LESSON_ROOT=$tmp scripts/lesson-recurrence.sh
printf 'LESSON: class=retired retired=not-applicable\nLESSON: class=retired\n' >"$tmp/notes/journal/a.md"; LESSON_ROOT=$tmp scripts/lesson-recurrence.sh
