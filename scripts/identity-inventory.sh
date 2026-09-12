#!/usr/bin/env bash
set -euo pipefail

require_no_live_aliases=false
if [[ ${1:-} == --require-no-live-aliases ]]; then
  require_no_live_aliases=true
  shift
fi
root=${1:-$(git rev-parse --show-toplevel)}
needle='v''era'
upper=${needle^^}
regex="(^|[^[:alnum:]])${needle}([^[:alnum:]]|$)"
unclassified=0
live_aliases=0
declare -A counts=([frozen-wire]=0 [frozen-storage]=0 [frozen-history]=0 [deprecated-alias]=0 [baseline-quote]=0)

classify() {
  local path=$1 text=$2 lower=${2,,} category=''
  case $path in
    docs/decisions/*|docs/verification/verdicts/*|notes/journal/*|docs/verification/p2-*|docs/verification/p3-*|docs/verification/task*|docs/plans/P1-*|GENESIS-PROMPT.md:*|MIGRATION-PROMPT.md:*)
      category=frozen-history
      ;;
    docs/plans/P5-*|docs/plans/p6-census-rows.tsv:*|docs/plans/p6-census.md:*|scripts/p6-census.sh:*|docs/verification/p5-ratification-baseline.md:*|scripts/tests/identity-inventory.test.sh:*)
      category=baseline-quote
      ;;
    *)
      if [[ $path == kernel/internal/store/store.go:* || $path == kernel/internal/store/surface_test.go:* || ( $path == docs/proofbound-identity-migration.md:* && $lower == *"${needle}-v1"* ) ]]; then
        category=frozen-storage
      elif [[ $lower == *"${needle}.witness.v1"* || $lower == *"${needle}.verdict.v1"* || $lower == *"${needle}.replay.v1"* ]]; then
        category=frozen-wire
      elif [[ $text == *"${upper}_"* || $lower == *".${needle}"* || $lower == *"cmd/${needle}"* || $lower == *"make ${needle}"* || $lower == *"\"${needle}\""* || $lower == *"\`${needle}\`"* || $lower == *" ${needle} "* || $lower == *"${needle}:"* || $lower == *"usage: ${needle}"* || $lower == *"/${needle}?"* || $lower == *"${needle}-v1"* ]]; then
        category=deprecated-alias
      fi
      ;;
  esac
  if [[ -z $category ]]; then
    printf 'unclassified\t%s\t%s\n' "$path" "$text"
    unclassified=$((unclassified + 1))
    return
  fi
  counts[$category]=$((counts[$category] + 1))
  if [[ $category == deprecated-alias && ( $path == kernel/cmd/* || $path == kernel/internal/* || $path == kernel/scripts/* || $path == .gitignore:* ) ]]; then
    live_aliases=$((live_aliases + 1))
  fi
  printf '%s\t%s\t%s\n' "$category" "$path" "$text"
}

while IFS= read -r match; do
  [[ -n $match ]] || continue
  absolute=${match%%:*}
  rest=${match#*:}
  line=${rest%%:*}
  content=${rest#*:}
  path=${absolute#"$root"/}
  path=${path#./}
  classify "$path:$line" "$content"
done < <(rg -n -i --no-heading --color never --hidden -g '!.git/**' "$regex" "$root" || true)

while IFS= read -r path; do
  relative=${path#"$root"/}
  if [[ ${relative##*/} == "$needle" ]]; then
    counts[deprecated-alias]=$((counts[deprecated-alias] + 1))
    if [[ $relative == kernel/cmd/* || $relative == kernel/internal/* || $relative == kernel/scripts/* ]]; then
      live_aliases=$((live_aliases + 1))
    fi
    printf 'deprecated-alias\t%s\tpath component\n' "$relative"
  fi
done < <(find "$root" -path "$root/.git" -prune -o -print)

printf 'summary\tfrozen-wire=%d frozen-storage=%d frozen-history=%d deprecated-alias=%d baseline-quote=%d unclassified=%d live-aliases=%d\n' \
  "${counts[frozen-wire]}" "${counts[frozen-storage]}" "${counts[frozen-history]}" "${counts[deprecated-alias]}" "${counts[baseline-quote]}" "$unclassified" "$live_aliases"
((unclassified == 0))
if $require_no_live_aliases; then
  ((live_aliases == 0))
fi
