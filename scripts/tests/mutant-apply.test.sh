#!/usr/bin/env bash
# Self-test for scripts/mutant-apply.sh: a no-op or comment-only substitution must be refused and
# leave the file byte-identical; a code substitution must apply.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
src=$tmp/subject.sh
printf '#!/usr/bin/env bash\n# limit: (n < 3) in this comment\nif (n < 3) exit 0\n// also (n < 3) in a C-style comment\n' >"$src"
cp "$src" "$tmp/orig"

expect() { # $1 = wanted exit status, $2 = perl expression
  local status=0
  scripts/mutant-apply.sh "$src" "$2" 2>/dev/null || status=$?
  [[ $status -eq $1 ]] || { echo "mutant-apply: '$2' exited $status, want $1" >&2; exit 1; }
}

expect 3 's/no-such-pattern/x/'                              # no-op
cmp -s "$src" "$tmp/orig" || { echo 'mutant-apply: no-op mutant modified the file' >&2; exit 1; }
expect 4 's/^# limit: \(n < 3\)/# limit: (n < 0)/m'          # comment only (#)
expect 4 's/^\/\/ also \(n < 3\)/\/\/ also (n < 0)/m'        # comment only (//)
cmp -s "$src" "$tmp/orig" || { echo 'mutant-apply: comment-only mutant was not restored' >&2; exit 1; }
expect 0 's/^if \(n < 3\)/if (n < 0)/m'                      # real code mutant
grep -q '^if (n < 0) exit 0$' "$src" || { echo 'mutant-apply: code mutant was not applied' >&2; exit 1; }

echo 'ok mutant-apply'
