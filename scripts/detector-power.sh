#!/usr/bin/env bash
# detector-power.sh — the P7 #2b detector-power / regeneration-trust gate (BR-regeneration-trust-5f8a03
# obligation O-1), per the ACCEPTED design spec docs/plans/P7-2b-detector-power-spec-draft1.md and
# change intent CI-p7-detector-power-1a2b3c. Owner §6 decisions realized here: a SHELL gate (not a Go
# package) + measure-first provisional U=0 trust bound.
#
# It does NOT re-sweep on `make check` (§7): the default mode VERIFIES committed, calibrated,
# freshness- and execution-bound evidence and FAILS CLOSED on any inconsistency. `--refresh` performs
# the out-of-band reproducible re-run (DP-8 reproducible mode) and asserts the sweep reproduces.
#
# The committed <slug>.sweep.txt is mutants.sh stdout with the host-specific repo-root prefix stripped
# to kernel-relative survivor keys (e.g. internal/twin/manifest/declaration.go:189#30). This makes
# raw_sweep_sha256 host-independent (a precondition of DP-8 reproducibility) and makes the survivor
# keys identical to docs/allowed-survivors.txt. The normalization is `s#^.*/kernel/##` on each line;
# calibration/summary lines carry no path and pass through unchanged.
#
# DP-1..DP-8 (spec §5) are each proven load-bearing by scripts/tests/detector-power.test.sh.
set -euo pipefail

die() { echo "detector-power: $*" >&2; exit 1; }
command -v jq >/dev/null || die 'jq is required'
command -v sha256sum >/dev/null || die 'sha256sum is required'

# The single home of the per-tier U_BUDGET. Owner decision 2026-11-05 (CI-p7-detector-power-1a2b3c §6):
# measure-first, provisional U=0. The top tier requires zero unproven-latitude survivors. Additional
# tiers with a NON-ZERO budget are an owner decision not yet taken (spec §3/§6); until then an unknown
# tier fails closed. A non-zero budget additionally requires each class-U survivor to argue why the
# un-pinned behavior is safe for a regenerated implementation to vary (CL-4), enforced below.
u_budget_for_tier() {
  case $1 in
    top) echo 0 ;;
    *) return 1 ;;
  esac
}

pkg=''
record=''
sweep=''
pkg_dir=''
allowed_survivors=''
mutants_dir=''
spec=''
refresh=0
sweep_cmd=''   # override the sweep command for a hermetic reproducible-mode self-test
while (($#)); do
  case $1 in
    --refresh) refresh=1 ;;
    --record) shift; (($#)) || die '--record requires a path'; record=$1 ;;
    --sweep) shift; (($#)) || die '--sweep requires a path'; sweep=$1 ;;
    --pkg-dir) shift; (($#)) || die '--pkg-dir requires a path'; pkg_dir=$1 ;;
    --allowed-survivors) shift; (($#)) || die '--allowed-survivors requires a path'; allowed_survivors=$1 ;;
    --mutants-dir) shift; (($#)) || die '--mutants-dir requires a path'; mutants_dir=$1 ;;
    --spec) shift; (($#)) || die '--spec requires a path'; spec=$1 ;;
    --sweep-cmd) shift; (($#)) || die '--sweep-cmd requires a command'; sweep_cmd=$1 ;;
    -*) die "unknown option: $1" ;;
    *) [[ -z $pkg ]] || die 'exactly one <pkg> positional is allowed'; pkg=$1 ;;
  esac
  shift
done
[[ -n $pkg ]] || die 'usage: detector-power.sh <kernel-relative-package> [--refresh] [overrides]'

# Resolve defaults from the repo layout unless overridden (overrides keep the self-test hermetic).
root=$(git rev-parse --show-toplevel 2>/dev/null || true)
slug=${pkg//\//-}
: "${record:=${root:+$root/}docs/verification/detector-power/$slug.json}"
: "${pkg_dir:=${root:+$root/}kernel/$pkg}"
: "${allowed_survivors:=${root:+$root/}docs/allowed-survivors.txt}"
: "${mutants_dir:=${root:+$root/}tools/mutants}"
if [[ -z $spec ]]; then
  cand="${root:+$root/}kernel/$pkg/SPEC.md"
  [[ -f $cand ]] && spec=$cand
fi

[[ -f $record ]] || die "record not found: $record"
jq -e . "$record" >/dev/null 2>&1 || die "record is not valid JSON: $record"

# --- read the record -------------------------------------------------------------------------------
schema=$(jq -r '.schema // ""' "$record")
[[ $schema == proofbound.detector-power.v1 ]] || die "unexpected schema: $schema"
rec_pkg=$(jq -r '.package // ""' "$record")
[[ $rec_pkg == "$pkg" ]] || die "record package ($rec_pkg) != requested package ($pkg)"
raw_sweep_path=$(jq -r '.raw_sweep_path // ""' "$record")
[[ -n $raw_sweep_path ]] || die 'record missing raw_sweep_path'
raw_sweep_sha256=$(jq -r '.raw_sweep_sha256 // ""' "$record")
[[ $raw_sweep_sha256 =~ ^[0-9a-f]{64}$ ]] || die "raw_sweep_sha256 is not 64 lowercase hex: $raw_sweep_sha256"
lock_inputs_sha256=$(jq -r '.lock_inputs_sha256 // ""' "$record")
[[ $lock_inputs_sha256 =~ ^[0-9a-f]{64}$ ]] || die "lock_inputs_sha256 is not 64 lowercase hex: $lock_inputs_sha256"
mode=$(jq -r '.execution_binding.mode // ""' "$record")
tier=$(jq -r '.u_budget_tier // ""' "$record")
rec_trusted=$(jq -r '.trusted // false' "$record")
rec_k=$(jq -r '.killed_count // -1' "$record")
rec_s=$(jq -r '.survived_count // -1' "$record")
rec_v=$(jq -r '.invalid_count // -1' "$record")
rec_c=$(jq -r '.candidates_count // -1' "$record")
rec_u=$(jq -r '.unproven_latitude_count // -1' "$record")
rec_diag=$(jq -r '.detector_power_diagnostic // ""' "$record")

# Locate the raw sweep. If an explicit --sweep was given it must equal the recorded raw_sweep_path's
# basename resolution; otherwise resolve raw_sweep_path against the repo root.
if [[ -z $sweep ]]; then
  sweep=${root:+$root/}$raw_sweep_path
fi
[[ -f $sweep ]] || die "raw sweep not found: $sweep"

# --- DP-7: calibration must be present -------------------------------------------------------------
grep -qxF 'calibration neutral=survived invalid=invalid lethal=killed' "$sweep" \
  || die "DP-7: raw sweep lacks the required calibration line (uncalibrated sweep rejected): $sweep"

# --- DP-8: execution binding -----------------------------------------------------------------------
# The committed sweep is already normalized, so its own digest is the pin.
sweep_hash=$(sha256sum "$sweep" | awk '{print $1}')
[[ $sweep_hash == "$raw_sweep_sha256" ]] \
  || die "DP-8: committed raw sweep digest ($sweep_hash) != pinned raw_sweep_sha256 ($raw_sweep_sha256)"
case $mode in
  reproducible)
    : # the hash pin above fixes the bytes; `--refresh` reproduces them from a live sweep (below)
    ;;
  attested)
    re_runner=$(jq -r '.execution_binding.re_runner // ""' "$record")
    attested_expiry=$(jq -r '.execution_binding.attested_expiry // ""' "$record")
    [[ -n $re_runner ]] || die 'DP-8: attested binding requires execution_binding.re_runner (a named non-author who performed an independent re-run — CL-1)'
    [[ $attested_expiry =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] || die "DP-8: attested binding requires execution_binding.attested_expiry as an ISO date (Law 4): $attested_expiry"
    today=$(date -u +%Y-%m-%d)
    [[ $attested_expiry > $today || $attested_expiry == "$today" ]] \
      || die "DP-8: attested binding expired ($attested_expiry < $today) — re-run and re-attest (Law 4)"
    ;;
  *)
    die "DP-8: execution_binding.mode must be 'reproducible' or 'attested', got: '$mode' (a sweep bound to neither fails closed)"
    ;;
esac

# --- DP-5: re-derive C/K/V/S from the committed raw sweep stdout -----------------------------------
# Tally the per-mutant status lines, independent of the summary, then require the summary to agree.
der_k=$(grep -cE ' killed$' "$sweep" || true)
der_s=$(grep -cE ' survived$' "$sweep" || true)
der_v=$(grep -cE ' invalid$' "$sweep" || true)
summary=$(grep -E '^summary candidates=' "$sweep" | tail -1 || true)
[[ -n $summary ]] || die 'DP-5: raw sweep has no summary line'
sum_c=$(sed -n 's/.*candidates=\([0-9]*\).*/\1/p' <<<"$summary")
sum_k=$(sed -n 's/.*killed=\([0-9]*\).*/\1/p' <<<"$summary")
sum_v=$(sed -n 's/.*invalid=\([0-9]*\).*/\1/p' <<<"$summary")
sum_s=$(sed -n 's/.*survived=\([0-9]*\).*/\1/p' <<<"$summary")
[[ $der_k == "$sum_k" && $der_s == "$sum_s" && $der_v == "$sum_v" ]] \
  || die "DP-5: summary (k=$sum_k v=$sum_v s=$sum_s) disagrees with the tallied status lines (k=$der_k v=$der_v s=$der_s)"
[[ $sum_c == $((der_k + der_s + der_v)) ]] \
  || die "DP-5: summary candidates=$sum_c != killed+survived+invalid=$((der_k + der_s + der_v))"
[[ $rec_k == "$der_k" && $rec_s == "$der_s" && $rec_v == "$der_v" && $rec_c == "$sum_c" ]] \
  || die "DP-5: record counts (k=$rec_k v=$rec_v s=$rec_s c=$rec_c) disagree with the raw sweep (k=$der_k v=$der_v s=$der_s c=$sum_c)"

# --- DP-2 / DP-3: survivor<->record bijection, classification, allowed-survivors binding -----------
# Sweep survivor keys (the first field of each ` survived` line), normalized/sorted.
mapfile -t sweep_survivors < <(grep -E ' survived$' "$sweep" | awk '{print $1}' | LC_ALL=C sort)
mapfile -t record_survivors < <(jq -r '.survivors[].id' "$record" | LC_ALL=C sort)
# Any duplicate id in the record is itself a defect.
dups=$(jq -r '.survivors[].id' "$record" | LC_ALL=C sort | uniq -d)
[[ -z $dups ]] || die "DP-3: duplicate survivor id(s) in record: $dups"
diff_out=$(diff <(printf '%s\n' "${sweep_survivors[@]}") <(printf '%s\n' "${record_survivors[@]}") || true)
[[ -z $diff_out ]] || die "DP-2/DP-3: survivor set is not an exact bijection (sweep '<' vs record '>'):
$diff_out"

# Per-survivor class rules + allowed-survivors binding.
n=$(jq '.survivors | length' "$record")
computed_u=0
for ((i = 0; i < n; i++)); do
  id=$(jq -r ".survivors[$i].id" "$record")
  class=$(jq -r ".survivors[$i].class" "$record")
  just=$(jq -r ".survivors[$i].justification // \"\"" "$record")
  pins=$(jq -r ".survivors[$i].pins // \"\"" "$record")
  [[ -n $just ]] || die "DP-3: survivor $id has an empty justification"
  # Bind to the declared-survivor allowlist: every classified survivor must be a declared entry, and
  # the register's {dp:class=...} annotation MUST agree with the record's class. The register and the
  # record both name the class (spec §4/§8), so the gate enforces agreement rather than trusting it —
  # the alternative is a silent drift seed (Law 2). `pins` lives once, in the record (Law 2: linked
  # from the register by the key + the class=R marker, never copied).
  reg_line=$(grep -E "^${id//./\\.}([[:space:]]|$)" "$allowed_survivors" | head -1 || true)
  [[ -n $reg_line ]] \
    || die "DP-3: survivor $id is not declared in $allowed_survivors (classification must be bound to the allowlist)"
  reg_class=$(sed -n 's/.*{dp:class=\([ERU]\)[};].*/\1/p' <<<"$reg_line")
  [[ -n $reg_class ]] \
    || die "DP-3: survivor $id is in the register but carries no {dp:class=E|R|U} annotation (register must declare the class)"
  [[ $reg_class == "$class" ]] \
    || die "DP-3: survivor $id class DRIFT — register says {dp:class=$reg_class} but record says class=$class (Law 2 drift guard)"
  case $class in
    E) : ;;  # equivalence argument carried in justification (non-empty, checked above)
    R)
      [[ -n $pins && $pins != null ]] || die "DP-3: class-R survivor $id must name the pinning test/line in .pins (CL-3)"
      shopt -s nocasematch
      [[ $just == *subsum* ]] || { shopt -u nocasematch; die "DP-3: class-R survivor $id must argue that the pinning line SUBSUMES the mutated behavior (CL-3); justification does not mention subsumption"; }
      shopt -u nocasematch
      ;;
    U) computed_u=$((computed_u + 1)) ;;
    *) die "DP-3: survivor $id has an invalid class '$class' (must be E, R, or U)" ;;
  esac
done

# --- DP-1: gate keys on the ABSOLUTE class-U count; diagnostic never drives trust; K+U=0 rejected ---
[[ $rec_u == "$computed_u" ]] || die "DP-1: record unproven_latitude_count ($rec_u) != class-U survivors counted ($computed_u)"
u=$computed_u
k=$der_k
denom=$((k + u))
[[ $denom -gt 0 ]] || die "DP-1: K+U=0 — detector_power is undefined and the record is rejected (F7)"
want_diag="$k/$denom"
[[ $rec_diag == "$want_diag" ]] || die "DP-1: detector_power_diagnostic ($rec_diag) != computed K/(K+U) ($want_diag)"

# --- DP-4: trust gate respects the per-tier U budget ----------------------------------------------
budget=$(u_budget_for_tier "$tier") || die "DP-4: unknown u_budget_tier '$tier' (no budget defined — owner decision pending, spec §6)"
# CL-4: a non-zero budget authorizes regeneration against a known un-pinned surface, so each class-U
# survivor must argue why that behavior is safe for a regenerated implementation to vary.
if [[ $budget -gt 0 ]]; then
  for ((i = 0; i < n; i++)); do
    class=$(jq -r ".survivors[$i].class" "$record")
    [[ $class == U ]] || continue
    id=$(jq -r ".survivors[$i].id" "$record")
    just=$(jq -r ".survivors[$i].justification // \"\"" "$record")
    shopt -s nocasematch
    [[ $just == *vary* || $just == *"safe to"* ]] || { shopt -u nocasematch; die "DP-4/CL-4: budgeted class-U survivor $id must argue why the un-pinned behavior is safe for a regenerated implementation to vary"; }
    shopt -u nocasematch
  done
fi
if [[ $u -le $budget ]]; then expect_trusted=true; else expect_trusted=false; fi
[[ $rec_trusted == "$expect_trusted" ]] \
  || die "DP-4: trusted=$rec_trusted but U=$u vs budget($tier)=$budget requires trusted=$expect_trusted (trusted iff U<=budget)"

# --- DP-6: freshness — recompute lock_inputs over the CURRENT inputs and reject a stale record ------
compute_lock_inputs() {
  # A deterministic manifest of component digests; the whole is hashed. Components (spec §4, N2/CL-2):
  #   - every *.go in the package dir (non-test source AND _test.go — the lock IS the tests)
  #   - this package's docs/allowed-survivors.txt entries (keys prefixed by "<pkg>/")
  #   - the package SPEC.md (its invariant/lock rows; whole file is a stricter superset)
  #   - the tools/mutants operator-set (every file under the mutants dir) — an operator-basis change
  #     invalidates stale attested evidence (CL-2)
  {
    if [[ -d $pkg_dir ]]; then
      while IFS= read -r f; do
        printf 'go %s %s\n' "${f##*/}" "$(sha256sum "$f" | awk '{print $1}')"
      done < <(find "$pkg_dir" -maxdepth 1 -name '*.go' -type f | LC_ALL=C sort)
    fi
    if [[ -f $allowed_survivors ]]; then
      as_hash=$(grep -E "^${pkg//./\\.}/" "$allowed_survivors" | LC_ALL=C sort | sha256sum | awk '{print $1}')
      printf 'allowed-survivors %s\n' "$as_hash"
    fi
    if [[ -n $spec && -f $spec ]]; then
      printf 'spec %s\n' "$(sha256sum "$spec" | awk '{print $1}')"
    else
      printf 'spec none\n'
    fi
    if [[ -d $mutants_dir ]]; then
      while IFS= read -r f; do
        printf 'mutants %s %s\n' "${f##*/}" "$(sha256sum "$f" | awk '{print $1}')"
      done < <(find "$mutants_dir" -type f | LC_ALL=C sort)
    fi
  } | sha256sum | awk '{print $1}'
}
current_lock=$(compute_lock_inputs)
[[ $current_lock == "$lock_inputs_sha256" ]] \
  || die "DP-6: record is STALE — pinned lock_inputs_sha256 ($lock_inputs_sha256) != current inputs ($current_lock). Re-sweep and refresh the record."

# --- optional reproducible re-run (DP-8 reproducible mode, out-of-band; never in make check) --------
if ((refresh)); then
  [[ $mode == reproducible ]] || die "--refresh only applies to execution_binding.mode=reproducible (this record is '$mode')"
  cmd=${sweep_cmd:-"bash ${root:+$root/}scripts/mutants.sh $pkg"}
  echo "detector-power: refreshing — re-running: $cmd" >&2
  raw=$(mktemp); tmp=$(mktemp)
  trap 'rm -f "$raw" "$tmp"' EXIT
  # scripts/mutants.sh EXITS 1 whenever survivors exist (the normal case for any detector-power
  # package) and 2 on a tool error — so accept exit <=1 and reject a real failure, rather than
  # treating the expected survivors-present exit as a failure. Capture first (no pipe), then
  # normalize the host-specific repo-root prefix to kernel-relative and compare the digest.
  set +e; eval "$cmd" >"$raw" 2>/dev/null; rc=$?; set -e
  [[ $rc -le 1 ]] || die "refresh: the sweep command failed (exit $rc)"
  sed 's#^.*/kernel/##' "$raw" >"$tmp"
  grep -qxF 'calibration neutral=survived invalid=invalid lethal=killed' "$tmp" \
    && grep -qE '^summary candidates=' "$tmp" \
    || die 'refresh: the sweep produced malformed or empty output (no calibration or summary line)'
  new_hash=$(sha256sum "$tmp" | awk '{print $1}')
  [[ $new_hash == "$raw_sweep_sha256" ]] \
    || die "DP-8 refresh: re-run did NOT reproduce raw_sweep_sha256 (got $new_hash, pinned $raw_sweep_sha256) — the sweep is not reproducible at the pinned lock inputs"
  echo "detector-power: refresh OK — re-run reproduced raw_sweep_sha256"
fi

echo "detector-power: OK pkg=$pkg K=$k U=$u detector_power=$want_diag tier=$tier budget=$budget trusted=$rec_trusted mode=$mode"
