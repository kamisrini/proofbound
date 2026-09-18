#!/usr/bin/env bash
set -euo pipefail

die() { echo "p6-census: $*" >&2; exit 1; }

mode=''
root=''
while (($#)); do
  case $1 in
    --write|--check|--render)
      [[ -z $mode ]] || die 'choose exactly one mode'
      mode=$1
      ;;
    --root)
      shift
      (($#)) || die '--root requires a path'
      root=$1
      ;;
    *) die "unknown option: $1" ;;
  esac
  shift
done
[[ -n $mode ]] || die 'usage: scripts/p6-census.sh --write|--check|--render [--root PATH]'
if [[ -z $root ]]; then
  root=$(git rev-parse --show-toplevel)
fi
root=$(cd "$root" && pwd)
registry=$root/docs/plans/p6-census-rows.tsv
canonical=$root/docs/plans/p6-census.md
[[ -f $registry ]] || die "missing registry: $registry"
git -C "$root" rev-parse --git-dir >/dev/null 2>&1 || die "not a Git worktree: $root"
input_commit=$(git -C "$root" log -1 --format=%H -- scripts/p6-census.sh 2>/dev/null || true)
[[ -n $input_commit ]] || input_commit=uncommitted

profile=$(awk -F '\t' '$1 == "# universe-profile" {print $2}' "$registry")
[[ $profile == proofbound || $profile == fixture ]] || die 'registry must declare one universe-profile'

declare -A row_ids=() subjects=() categories=() rows=()
header_seen=0
row_count=0
while IFS= read -r raw || [[ -n $raw ]]; do
  [[ -n $raw ]] || continue
  [[ $raw == '# '* ]] && continue
  IFS=$'\t' read -r row_id category subject observation probe disposition decision evidence extra <<<"$raw"
  if ((header_seen == 0)); then
    [[ $row_id == row_id && $category == category && $subject == subject && $observation == observation && $probe == probe && $disposition == disposition && $decision == decision && $evidence == evidence && -z ${extra:-} ]] || die 'invalid registry header'
    header_seen=1
    continue
  fi
  [[ -z ${extra:-} ]] || die "row $row_id has more than eight fields"
  [[ -n ${evidence:-} ]] || die "row $row_id has fewer than eight fields"
  [[ $row_id =~ ^C([1-8])-([0-9]{3})$ ]] || die "invalid row_id: $row_id"
  [[ $category == "C${BASH_REMATCH[1]}" ]] || die "row/category mismatch: $row_id $category"
  [[ $subject != *'|'* && $observation != *'|'* && $probe != *'|'* && $decision != *'|'* && $evidence != *'|'* ]] || die "pipe is forbidden in row $row_id"
  [[ -z ${row_ids[$row_id]+x} ]] || die "duplicate row_id: $row_id"
  key=$category:$subject
  [[ -z ${subjects[$key]+x} ]] || die "duplicate category/subject: $key"
  case $disposition in close-in-P6|defer-to-P7|wontfix) ;; *) die "invalid disposition in $row_id: $disposition" ;; esac
  if [[ $disposition == close-in-P6 ]]; then
    [[ $decision == '—' ]] || die "close-in-P6 row $row_id must use decision —"
  else
    [[ $decision == docs/decisions/VD-*.md ]] || die "defer/wontfix row $row_id lacks an exact VD path"
  fi
  row_ids[$row_id]=1
  subjects[$key]=$row_id
  categories[$category]=$(( ${categories[$category]:-0} + 1 ))
  rows[$row_id]=$raw
  row_count=$((row_count + 1))
done <"$registry"
((header_seen == 1)) || die 'registry header is missing'
for n in {1..8}; do
  [[ ${categories[C$n]:-0} -gt 0 ]] || die "category C$n has zero rows"
done

tracked_path() {
  local path=$1
  [[ -e $root/$path ]] && git -C "$root" ls-files --error-unmatch -- "$path" >/dev/null 2>&1
}

make_target() {
  local target=$1
  awk -v wanted="$target" '
    /^[^.#%[:space:]][^=%]*:/ {
      line=$0; sub(/:.*/, "", line); n=split(line, names, /[[:space:]]+/)
      for (i=1; i<=n; i++) if (names[i] == wanted) found=1
    }
    END {exit !found}
  ' "$root/Makefile"
}

accepted_vd() {
  local path=$1
  tracked_path "$path" && rg -q '^\*\*Status:\*\*[[:space:]]+Accepted([[:space:]]*)$' "$root/$path"
}

evidence_resolves() {
  local value=$1 item
  [[ $value != '—' ]] || return 1
  IFS=',' read -ra items <<<"$value"
  for item in "${items[@]}"; do
    item=${item#"${item%%[![:space:]]*}"}
    item=${item%"${item##*[![:space:]]}"}
    case $item in
      docs/*|scripts/*|kernel/*|ROADMAP.md|README.md|CLAUDE.md|Makefile)
        tracked_path "$item" || return 1
        ;;
      commit:*) git -C "$root" cat-file -e "${item#commit:}^{commit}" 2>/dev/null || return 1 ;;
      test:*) [[ ${item#test:} =~ ^[^:]+::Test[A-Za-z0-9_]+$ ]] || return 1 ;;
      *) return 1 ;;
    esac
  done
}

run_probe() {
  local probe=$1 decision=$2 subject=$3 kind arg rest
  kind=${probe%%:*}
  arg=${probe#*:}
  case $kind in
    path) tracked_path "$arg" ;;
    make-target) make_target "$arg" ;;
    make-useful-short)
      make_target short && ! awk '/^short:/ && $0 ~ /hooks-test[[:space:]]*$/ {only=1} END {exit !only}' "$root/Makefile"
      ;;
    shell-test) tracked_path "$arg" && bash "$root/$arg" >/dev/null ;;
    artifact) tracked_path "$arg" ;;
    decision-skip)
      local vd=${arg%%:*} named=${arg#*:}
      [[ $decision == "$vd" && $subject == "$named" ]] && accepted_vd "$vd" && rg -qi --fixed-strings "${named//-/ }" "$root/$vd"
      ;;
    acceptance)
      row_probe=$(awk -F '\t' -v wanted="package:$arg" '$3 == wanted {print $5; exit}' "$registry")
      row_evidence=$(awk -F '\t' -v wanted="package:$arg" '$3 == wanted {print $8; exit}' "$registry")
      [[ $row_probe == "acceptance:$arg" ]] || return 1
      [[ ",$row_evidence," == *",docs/verification/p6-task8-package-results-e4c8e77.tsv,"* ]] || return 1
      [[ ",$row_evidence," == *",docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md,"* ]] || return 1
      bash "$root/scripts/p6-task8-package-acceptance.sh" --package "$arg" --root "$root" >/dev/null
      ;;
    connector|route|intent-history|p5-result|fresh-clone)
      return 1
      ;;
    *) die "unknown probe form in $subject: $probe" ;;
  esac
}

require_subject() {
  local category=$1 subject=$2
  [[ -n ${subjects[$category:$subject]+x} ]] || die "unclassified scan subject: $category $subject"
}

declare -a live_files=()
for path in CLAUDE.md README.md ROADMAP.md notes/state.md docs/gates.md; do
  tracked_path "$path" && live_files+=("$root/$path")
done
while IFS=$'\t' read -r marker path; do
  [[ $marker == '# accepted-plan' ]] || continue
  tracked_path "$path" || die "accepted-plan path does not resolve: $path"
  live_files+=("$root/$path")
done <"$registry"
while IFS= read -r path; do live_files+=("$root/$path"); done < <(
  git -C "$root" ls-files 'docs/decisions/VD-*.md' | while IFS= read -r path; do
    rg -q '^\*\*Status:\*\*[[:space:]]+Accepted' "$root/$path" && printf '%s\n' "$path"
  done
)
while IFS= read -r path; do live_files+=("$root/$path"); done < <(git -C "$root" ls-files 'kernel/internal/**/SPEC.md')

if ((${#live_files[@]})); then
  while IFS= read -r token; do
    token=${token#\`}
    target=${token#make }
    make_target "$target" || require_subject C1 "mechanism:$token"
  done < <(rg -o --no-filename '`make [A-Za-z0-9][A-Za-z0-9_-]*' "${live_files[@]}" | LC_ALL=C sort -u || true)
  while IFS= read -r token; do
    token=${token#\`}
    [[ $token == */ ]] && continue
    [[ $token == *'<'* || $token == *'>'* || $token == *'*'* || $token == *'?'* ]] && continue
    tracked_path "$token" || require_subject C1 "mechanism:$token"
  done < <(rg -o --no-filename '`(scripts|kernel/scripts|\.claude/hooks|\.github/workflows|gates|kernel/internal)/[A-Za-z0-9_./<>?*-]+' "${live_files[@]}" | LC_ALL=C sort -u || true)
fi

if [[ $profile == proofbound ]]; then
  for subject in \
    mechanism:scripts/commit-cadence.sh mechanism:scripts/cleanroom-lint.sh \
    mechanism:scripts/lesson-recurrence.sh mechanism:scripts/figure-provenance.sh \
    mechanism:scripts/skip-lint.sh mechanism:scripts/prescription-lint.sh \
    mechanism:scripts/state-freshness.sh mechanism:scripts/gen-state.sh \
    mechanism:.claude/hooks/block-secrets.sh mechanism:.claude/hooks/block-generated-edit.sh \
    mechanism:.claude/hooks/lint-on-write.sh mechanism:.claude/hooks/stop-check.sh \
    mechanism:make\ backup mechanism:make\ meta-tax mechanism:make\ wrap-verify \
    mechanism:make\ laws-lock mechanism:make\ state mechanism:make\ vera; do
    require_subject C1 "$subject"
  done
  for subject in make-short-useful full-invariant-citation-resolution legacy-alias-due; do require_subject C2 "$subject"; done

  public_targets=(check check-witnessed delivery-enforce verify gates-canary gates-enforce short hooks-test index index-check invariants-lock identity-inventory mutants kernel-check)
  for target in "${public_targets[@]}"; do
    if make_target "$target" && ! rg -q --fixed-strings "make $target" "$root/README.md" "$root/CLAUDE.md"; then
      require_subject C1 "undocumented-make:$target"
    fi
  done
  while IFS= read -r check_name; do
    require_subject C2 "gate-row:$check_name"
  done < <(awk -F'|' '/^\|/ {x=$2; gsub(/^[[:space:]]+|[[:space:]]+$/, "", x); if (x != "Check" && x !~ /^---/) print x}' "$root/docs/gates.md")
  while IFS= read -r gate_path; do
    rg -q --fixed-strings "$gate_path" "$root/docs/gates.md" || require_subject C2 "undocumented-gate:$gate_path"
  done < <(git -C "$root" ls-files 'gates/*.yaml')

  while IFS= read -r dir; do
    pkg=${dir#kernel/}
    require_subject C3 "package:$pkg"
  done < <(git -C "$root" ls-files 'kernel/internal/**/*.go' | rg -v '_test\.go$' | sed 's#/[^/]*$##' | LC_ALL=C sort -u)

  while IFS= read -r dir; do
    name=${dir#kernel/internal/connector/}
    require_subject C4 "connector:$name"
  done < <(git -C "$root" ls-files 'kernel/internal/connector/**/*.go' | rg -v '_test\.go$' | sed 's#/[^/]*$##' | LC_ALL=C sort -u)
  for subject in sessions-live-corpus github-narrowness delivery-boundary-order; do require_subject C4 "$subject"; done

  pairs=(
    git,commit.recorded checks,check.run sessions,session.observed reviews,review.verdict
    reviews,requirement.reviewed github,github.workflow_run github,github.deployment
    intent.records,business_decision.recorded intent.records,requirement.recorded
    intent.records,change_intent.recorded intent.specdir,business_decision.recorded
    intent.specdir,requirement.recorded intent.specdir,change_intent.recorded
  )
  consumers=(projection verify twin gates)
  for pair in "${pairs[@]}"; do
    for consumer in "${consumers[@]}"; do
      require_subject C5 "route:${pair%,*}:${pair#*,}:$consumer"
    done
  done

  while IFS= read -r sha; do require_subject C6 "intent-commit:$sha"; done < <(git -C "$root" rev-list --reverse "f426ca8..$input_commit")
  for subject in active-requirements active-obligations exact-revision-reviews self-hosted-delivery-chain; do require_subject C6 "$subject"; done

  measurements=(artifact-authoring-minutes independent-review-field-change-rate ambiguous-obligations-preimplementation false-or-missing-commit-intent-links verdict-evidence-mismatches gate-canary-false-block-pass empty-store-reconstruction-time provider-mapping-conformance requirement-review-findings)
  falsifiers=(chain-maintenance-cost artifact-role-confusion subjective-obligation-outcomes old-meaning-reconstruction applicability-boundary foreign-provider-schema-change requirement-review-rubber-stamp)
  for subject in "${measurements[@]}"; do require_subject C7 "measurement:$subject"; done
  for subject in "${falsifiers[@]}"; do require_subject C7 "falsifier:$subject"; done

  for subject in fresh-clone-linux platform-windows artifact-integrity vision-progress-assessment snapshot-provider; do require_subject C8 "$subject"; done
fi

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
open=0
closed=0
declare -A rendered=()
for row_id in "${!rows[@]}"; do
  IFS=$'\t' read -r _ category subject observation probe disposition decision evidence <<<"${rows[$row_id]}"
  if run_probe "$probe" "$decision" "$subject"; then state=closed; else state=open; fi
  if [[ $state == closed ]]; then
    evidence_resolves "$evidence" || die "closed row $row_id has missing or unresolved evidence"
    closed=$((closed + 1))
  else
    [[ $evidence == '—' ]] || die "open row $row_id claims evidence"
    [[ $disposition == close-in-P6 ]] || die "defer/wontfix row $row_id is not closed by its decision"
    open=$((open + 1))
  fi
  rendered[$row_id]="| $row_id | $category | $subject | $observation | $probe | $state | $disposition | $decision | $evidence |"
done

{
  echo '<!-- @generated by scripts/p6-census.sh --write; never hand-edit (Law 1) -->'
  echo '# P6 debt census'
  echo
  echo "Generation input commit: \`$input_commit\`."
  echo "Counts: total=$row_count open=$open closed=$closed unclassified=0."
  printf 'Categories:'
  for n in {1..8}; do printf ' C%d=%d' "$n" "${categories[C$n]}"; done
  echo '.'
  echo
  echo '| row_id | category | subject | observation | probe | state | disposition | decision | evidence |'
  echo '|---|---|---|---|---|---|---|---|---|'
  while IFS= read -r row_id; do echo "${rendered[$row_id]}"; done < <(printf '%s\n' "${!rows[@]}" | LC_ALL=C sort -t- -k1.2,1n -k2,2n)
} >"$tmp"

case $mode in
  --render) cat "$tmp" ;;
  --write)
    mkdir -p "$(dirname "$canonical")"
    cp "$tmp" "$canonical"
    ;;
  --check)
    [[ -f $canonical ]] || die "canonical result is missing: $canonical"
    cmp -s "$tmp" "$canonical" || die 'canonical census is stale; run scripts/p6-census.sh --write'
    ;;
esac
