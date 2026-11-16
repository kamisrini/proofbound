#!/usr/bin/env bash
# Apply one mutation (a Perl substitution) to a file, refusing mutants that cannot count.
#
#   scripts/mutant-apply.sh <file> '<perl s/// expression>'
#
# Exit 0: the file changed in at least one non-comment line — a real mutant, safe to count.
# Exit 3: the substitution changed nothing (a no-op mutant: a "survivor" that proves nothing).
# Exit 4: every changed line is a comment ("#" or "//"), so the code is untouched.
# On exit 3 or 4 the file is restored byte-for-byte.
#
# Compiled response for LESSON class=fixture-not-what-it-claims (2026-11-16): a mutation whose
# pattern matched a comment instead of the code was read as a surviving mutant. A mutation harness
# must route every substitution through this script and never count an exit-3/4 mutant.
set -euo pipefail
(($# == 2)) || { echo 'usage: scripts/mutant-apply.sh <file> <perl-substitution>' >&2; exit 2; }
file=$1 expr=$2
[[ -f $file ]] || { echo "mutant-apply: no such file: $file" >&2; exit 2; }
backup=$(mktemp); trap 'rm -f "$backup"' EXIT
cp "$file" "$backup"
perl -0pi -e "$expr" "$file"
if cmp -s "$file" "$backup"; then
  echo "mutant-apply: substitution changed nothing (no-op mutant): $expr" >&2
  exit 3
fi
# diff exits 1 whenever the files differ, so capture it outside any pipefail pipeline.
added=$(diff "$backup" "$file" | grep '^>' || true)
if ! grep -vqE '^>[[:space:]]*(#|//)' <<<"$added"; then
  cp "$backup" "$file"
  echo "mutant-apply: substitution only changed comments (invalid mutant): $expr" >&2
  exit 4
fi
