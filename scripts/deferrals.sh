#!/usr/bin/env bash
set -euo pipefail

die() { echo "deferrals: $*" >&2; exit 1; }

mode=check
status_filter=''
hold_filter=''
root=${DEFERRALS_ROOT:-}

while (($#)); do
  case $1 in
    --check|--all|--open)
      [[ $mode == check || $mode == all || $mode == open ]] || die 'choose one query mode'
      mode=${1#--}
      ;;
    --status)
      shift
      (($#)) || die '--status requires a value'
      [[ -z $status_filter ]] || die '--status may be supplied once'
      status_filter=$1
      mode=query
      ;;
    --hold-type)
      shift
      (($#)) || die '--hold-type requires a value'
      [[ -z $hold_filter ]] || die '--hold-type may be supplied once'
      hold_filter=$1
      mode=query
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

if [[ -z $root ]]; then
  root=$(git rev-parse --show-toplevel)
fi
root=$(cd "$root" && pwd)
registry=$root/docs/deferrals.tsv
[[ -f $registry ]] || die "missing registry: $registry"

header=$'id\tstatus\thold_type\towner\tcreated\tphase\tsubject\treason\treview_trigger\tunblock_condition\tevidence'
declare -A ids=()
rows=()
header_seen=0
row_count=0
open_count=0

valid_status() {
  case $1 in OPEN|RESOLVED|REJECTED|RETIRED) return 0 ;; *) return 1 ;; esac
}

valid_hold_type() {
  case $1 in DEFERRED|POSTPONED|AWAITING_OWNER|PENDING_PREREQUISITE|AWAITING_REVIEW|BLOCKED|NONE) return 0 ;; *) return 1 ;; esac
}

valid_date() { [[ $1 =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; }

valid_evidence() {
  local value=$1 item
  local -a items
  IFS=',' read -ra items <<<"$value"
  ((${#items[@]} > 0)) || return 1
  for item in "${items[@]}"; do
    [[ -n $item && $item != /* && $item != .* && $item != *'..'* ]] || return 1
    [[ -e $root/$item ]] || return 1
  done
}

while IFS= read -r raw || [[ -n $raw ]]; do
  [[ -n $raw ]] || continue
  [[ $raw == '# '* ]] && continue
  IFS=$'\t' read -r id status hold_type owner created phase subject reason review_trigger unblock_condition evidence extra <<<"$raw"
  if ((header_seen == 0)); then
    [[ $raw == "$header" ]] || die 'invalid registry header'
    header_seen=1
    continue
  fi
  [[ -z ${extra:-} ]] || die "row $id has too many fields"
  [[ -n $id && -n $status && -n $hold_type && -n $owner && -n $created && -n $phase && -n $subject && -n $reason && -n $review_trigger && -n $unblock_condition && -n $evidence ]] || die "row $id has an empty field"
  [[ $id =~ ^DEF-[A-Z0-9-]+$ ]] || die "invalid id: $id"
  [[ -z ${ids[$id]+x} ]] || die "duplicate id: $id"
  valid_status "$status" || die "invalid status for $id: $status"
  valid_hold_type "$hold_type" || die "invalid hold_type for $id: $hold_type"
  valid_date "$created" || die "invalid created date for $id: $created"
  valid_evidence "$evidence" || die "invalid evidence for $id"
  if [[ $status == OPEN ]]; then
    [[ $hold_type != NONE ]] || die "open item $id must have a hold_type"
    open_count=$((open_count + 1))
  fi
  ids[$id]=1
  rows+=("$raw")
  row_count=$((row_count + 1))
done <"$registry"

((header_seen == 1)) || die 'registry header is missing'

matches() {
  local raw=$1 id status hold_type owner created phase subject reason review_trigger unblock_condition evidence
  IFS=$'\t' read -r id status hold_type owner created phase subject reason review_trigger unblock_condition evidence <<<"$raw"
  [[ -z $status_filter || $status == "$status_filter" ]] || return 1
  [[ -z $hold_filter || $hold_type == "$hold_filter" ]] || return 1
  [[ $mode != open || $status == OPEN ]] || return 1
}

case $mode in
  check)
    [[ -z $status_filter && -z $hold_filter ]] || die 'filters require a query mode'
    printf 'deferrals: PASS rows=%d open=%d\n' "$row_count" "$open_count"
    ;;
  all|open|query)
    printf '%s\n' "$header"
    for row in "${rows[@]}"; do
      matches "$row" && printf '%s\n' "$row"
    done
    ;;
esac
