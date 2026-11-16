#!/usr/bin/env bash
set -euo pipefail

repo=$(git rev-parse --show-toplevel)
commit=HEAD
while (($#)); do
  case $1 in
    --repo) (($# >= 2)) || { echo 'usage: intent-delivery-applicability.sh [--repo <repository>] [--commit <commit>]' >&2; exit 2; }; repo=$2; shift 2 ;;
    --commit) (($# >= 2)) || { echo 'usage: intent-delivery-applicability.sh [--repo <repository>] [--commit <commit>]' >&2; exit 2; }; commit=$2; shift 2 ;;
    *) echo 'usage: intent-delivery-applicability.sh [--repo <repository>] [--commit <commit>]' >&2; exit 2 ;;
  esac
done
repo=$(cd "$repo" && pwd)
git -C "$repo" cat-file -e "$commit^{commit}" 2>/dev/null || {
  printf 'intent-delivery-applicability: invalid commit %s\n' "$commit" >&2
  exit 1
}

result=$($repo/scripts/intent-applicability.sh --repo "$repo" --commit "$commit")
applicable=$(printf '%s\n' "$result" | sed -n 's/^applicable=//p')
if [[ $applicable != true ]]; then
  printf 'applicable=false commit=%s\n' "$(git -C "$repo" rev-parse "$commit")"
  exit 0
fi

# The Intent claim must be one the KERNEL binds, not merely a line somewhere in the message
# (DEF-INTENT-TRAILER-BLOCK-001). This is a port of intentTrailers() in
# kernel/internal/connector/git/gitcmd/gitcmd.go: drop trailing empty lines; require >= 3 lines;
# walk back from the end while lines match `^[A-Za-z0-9-]+: .+$`; that run must be preceded by an
# empty line; collect the `Intent` values. Both parsers are tested against one corpus
# (scripts/tests/fixtures/intent-trailers) so they cannot drift apart.
sha=$(git -C "$repo" rev-parse "$commit")
message=$(git -C "$repo" show -s --format=%B "$commit")
values=$(printf '%s\n' "$message" | awk '
  { line[NR] = $0 }
  END {
    n = NR
    while (n > 0 && line[n] == "") n--
    if (n < 3) exit 0
    i = n; k = 0
    while (i >= 1 && line[i] ~ /^[A-Za-z0-9-]+: .+$/) {
      if (substr(line[i], 1, 8) == "Intent: ") v[++k] = substr(line[i], 9)
      i--
    }
    if (i < 1 || line[i] != "") exit 0
    for (j = 1; j <= k; j++) print v[j]
  }')
if [[ -z $values ]]; then
  if printf '%s\n' "$message" | grep -q '^Intent:'; then
    printf 'intent-delivery-applicability: applicable commit %s has an Intent line OUTSIDE the final trailer block, so the kernel binds no claim; put Intent: directly above Co-Authored-By with no blank line\n' "$sha" >&2
  else
    printf 'intent-delivery-applicability: applicable commit %s has no explicit Intent trailer\n' "$sha" >&2
  fi
  exit 1
fi
# Same value rule the kernel applies at resolution (connector.go resolvedCommit + intentIDRE):
# bare CI-* resolves as `records`; a qualifier must be exactly records: or specdir:.
id_re='^((records|specdir):)?CI-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6}$'
while IFS= read -r value; do
  [[ $value =~ $id_re ]] || {
    printf 'intent-delivery-applicability: applicable commit %s has a malformed Intent trailer %q\n' "$sha" "$value" >&2
    exit 1
  }
done <<<"$values"
printf 'applicable=true commit=%s explicit-intent=true intents=%s\n' "$sha" "$(LC_ALL=C sort -u <<<"$values" | paste -sd, -)"
