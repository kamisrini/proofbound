#!/usr/bin/env bash
set -euo pipefail

# Read the repository the way the kernel's adapter does (gitcmd: repositoryGitEnv +
# --no-replace-objects): drop every inherited GIT_* variable and ignore replace refs, so a
# `git replace` ref or a stray GIT_DIR cannot show this gate a different commit than the kernel
# binds (gate review round 1, F3). Exported so intent-applicability.sh inherits it.
while IFS= read -r variable; do unset "$variable"; done < <(compgen -e | grep '^GIT_' || true)
export GIT_NO_REPLACE_OBJECTS=1

usage() { echo 'usage: intent-delivery-applicability.sh [--repo <repository>] [--commit <commit>]' >&2; exit 2; }
repo=$(git rev-parse --show-toplevel)
commit=
while (($#)); do
  case $1 in
    --repo) (($# >= 2)) || usage; repo=$2; shift 2 ;;
    --commit) (($# >= 2)) || usage; commit=$2; shift 2 ;;
    *) usage ;;
  esac
done
repo=$(cd "$repo" && pwd)

# Strict UTF-8 exactly as the kernel judges it (Go utf8.Valid): surrogates, code points above
# U+10FFFF, and overlong forms are invalid. glibc iconv accepts some of these (gate review round 2,
# F4), so Perl's strict "UTF-8" decoder is used instead.
valid_utf8() { perl -MEncode -e 'local $/; my $b = <STDIN>; eval { Encode::decode("UTF-8", $b, Encode::FB_CROAK); 1 } or exit 1'; }

# check_commit <sha>: 0 when the commit needs no Intent claim or carries one the kernel binds.
check_commit() {
  local sha=$1 result applicable message field values value parent
  result=$("$repo/scripts/intent-applicability.sh" --repo "$repo" --commit "$sha")
  applicable=$(printf '%s\n' "$result" | sed -n 's/^applicable=//p')
  if [[ $applicable != true ]]; then
    printf 'applicable=false commit=%s\n' "$sha"
    return 0
  fi

  # The kernel adapter rejects a commit outright when any field it reads is not valid UTF-8 —
  # identities and subject/body (gitcmd parseScalars) and touched paths (gitcmd nulUTF8Strings) —
  # so no claim of it can bind (gate review rounds 1-2, F4).
  for field in '%an' '%ae' '%cn' '%ce' '%B'; do
    if ! git -C "$repo" show -s --format="$field" "$sha" | valid_utf8; then
      printf 'intent-delivery-applicability: applicable commit %s has a %s field that is not valid UTF-8, which the kernel rejects outright, so no Intent claim can bind\n' "$sha" "$field" >&2
      return 1
    fi
  done
  parent=$(git -C "$repo" show -s --format=%P "$sha" | awk '{print $1}')
  if [[ -z $parent ]]; then
    paths_ok=$(git -C "$repo" diff-tree --root --ignore-submodules=none --no-ext-diff --no-textconv --no-commit-id --name-only -r -z "$sha" | valid_utf8 && echo yes || echo no)
  else
    paths_ok=$(git -C "$repo" diff --ignore-submodules=none --no-ext-diff --no-textconv --name-only -z "$parent" "$sha" | valid_utf8 && echo yes || echo no)
  fi
  if [[ $paths_ok != yes ]]; then
    printf 'intent-delivery-applicability: applicable commit %s touches a path that is not valid UTF-8, which the kernel rejects outright, so no Intent claim can bind\n' "$sha" >&2
    return 1
  fi

  # The Intent claim must be one the KERNEL binds, not merely a line somewhere in the message
  # (DEF-INTENT-TRAILER-BLOCK-001). This is a port of intentTrailers() in
  # kernel/internal/connector/git/gitcmd/gitcmd.go: drop trailing empty lines; require >= 3 lines;
  # walk back from the end while lines match `^[A-Za-z0-9-]+: .+$`; that run must be preceded by
  # an empty line; collect the `Intent` values. Both sides are tested against one corpus
  # (scripts/tests/fixtures/intent-trailers/) so they cannot drift apart.
  message=$(git -C "$repo" show -s --format=%B "$sha")
  # LC_ALL=C: byte semantics and fixed character ranges, like Go regexp. Defence in depth:
  # invalid UTF-8 is rejected above, so dropping it is an equivalent mutant today.
  values=$(printf '%s\n' "$message" | LC_ALL=C awk '
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
    if [[ $message == *$'\r'* ]] && printf '%s\n' "$message" | grep -q '^Intent:'; then
      printf 'intent-delivery-applicability: applicable commit %s has CRLF line endings; the kernel reads the carriage returns as content, so its Intent line binds no claim — rewrite the message with LF endings\n' "$sha" >&2
    elif printf '%s\n' "$message" | grep -q '^Intent:'; then
      printf 'intent-delivery-applicability: applicable commit %s has an Intent line OUTSIDE the final trailer block, so the kernel binds no claim; put Intent: directly above Co-Authored-By with no blank line\n' "$sha" >&2
    else
      printf 'intent-delivery-applicability: applicable commit %s has no explicit Intent trailer\n' "$sha" >&2
    fi
    return 1
  fi
  # Same value rule the kernel applies at resolution (connector.go resolvedCommit + intentIDRE):
  # bare CI-* resolves as `records`; a qualifier must be exactly records: or specdir:.
  local id_re='^((records|specdir):)?CI-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6}$'
  while IFS= read -r value; do
    [[ $value =~ $id_re ]] || {
      printf 'intent-delivery-applicability: applicable commit %s has a malformed Intent trailer %q\n' "$sha" "$value" >&2
      return 1
    }
  done <<<"$values"
  printf 'applicable=true commit=%s explicit-intent=true intents=%s\n' "$sha" "$(LC_ALL=C sort -u <<<"$values" | paste -sd, -)"
}

if [[ -n $commit ]]; then
  git -C "$repo" cat-file -e "$commit^{commit}" 2>/dev/null || {
    printf 'intent-delivery-applicability: invalid commit %s\n' "$commit" >&2
    exit 1
  }
  check_commit "$(git -C "$repo" rev-parse "$commit")"
  exit
fi

# Default: every commit not yet on the upstream, plus HEAD (gate review round 2, F2). Checking only
# HEAD let this repository's commit-then-refresh-the-pin pattern bury each governed commit under a
# non-governed one. Without an upstream (detached HEAD, fresh clone), HEAD alone.
head=$(git -C "$repo" rev-parse HEAD)
commits=()
if upstream=$(git -C "$repo" rev-parse --verify --quiet '@{upstream}' 2>/dev/null); then
  mapfile -t commits < <(git -C "$repo" rev-list --reverse "$upstream..$head")
fi
[[ " ${commits[*]} " == *" $head "* ]] || commits+=("$head")
failed=0
for sha in "${commits[@]}"; do
  check_commit "$sha" || failed=1
done
exit "$failed"
