#!/usr/bin/env bash
# Self-test for scripts/detector-power.sh (P7 #2b detector-power / regeneration-trust gate).
# Auto-discovered by `make hooks-test` (and therefore part of `make check`).
#
# Law 3: a hook's self-test proves it FIRES. Each DP-1..DP-8 invariant (spec §5) gets a load-bearing
# control: a NEUTRAL fixture that MUST be accepted, and a POSITIVE control — one field mutated, the
# rest left valid — that MUST be rejected AND rejected for the INTENDED reason (its DP tag must appear
# in the failure message), so a control cannot pass by tripping some other check. Fully hermetic:
# synthetic fixtures in a temp dir, no real mutation sweep, no repo package source.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
gate=scripts/detector-power.sh
pkg=internal/fixturepkg

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
pkgdir="$tmp/pkg"; mkdir -p "$pkgdir"
mutdir="$tmp/mutants"; mkdir -p "$mutdir"
printf 'package fixturepkg\nfunc F() int { return 1 }\n' >"$pkgdir/f.go"
printf 'package fixturepkg\nimport "testing"\nfunc TestF(t *testing.T){ if F()!=1 {t.Fatal("x")} }\n' >"$pkgdir/f_test.go"
printf 'package main\nfunc main(){}\n' >"$mutdir/main.go"
printf 'module mutants\n' >"$mutdir/go.mod"

allowed="$tmp/allowed-survivors.txt"
cat >"$allowed" <<EOF
# fixture allowlist
$pkg/f.go:10#1	string	literal -> ""  # EQUIVALENT: log string only {dp:class=E}
$pkg/f.go:20#2	binary	<= -> <  # redundantly pinned {dp:class=R}
EOF

sweep="$tmp/pkg.sweep.txt"
cat >"$sweep" <<EOF
calibration neutral=survived invalid=invalid lethal=killed
$pkg/f.go:5#0 killed
$pkg/f.go:10#1 survived
$pkg/f.go:20#2 survived
$pkg/f.go:30#3 invalid
summary candidates=4 killed=1 invalid=1 survived=2
EOF
sweep_sha=$(sha256sum "$sweep" | awk '{print $1}')

# Base record (lock hash filled in after the probe below).
base="$tmp/base.json"
cat >"$base" <<EOF
{"schema":"proofbound.detector-power.v1","package":"$pkg","swept_at_commit":"deadbeefdeadbeef",
 "lock_inputs_sha256":"$(printf '%064d' 0)","raw_sweep_path":"irrelevant.sweep.txt",
 "raw_sweep_sha256":"$sweep_sha","execution_binding":{"mode":"reproducible"},
 "killed_count":1,"survived_count":2,"invalid_count":1,"candidates_count":4,
 "survivors":[
   {"id":"$pkg/f.go:10#1","class":"E","justification":"equivalent: affects only a log string","pins":null},
   {"id":"$pkg/f.go:20#2","class":"R","justification":"redundantly pinned; TestBoth on the sibling line subsumes this mutated behavior","pins":"f_test.go::TestBoth"}],
 "unproven_latitude_count":0,"detector_power_diagnostic":"1/1","u_budget_tier":"top","trusted":true}
EOF

gate_args=( "$pkg" --sweep "$sweep" --pkg-dir "$pkgdir" --allowed-survivors "$allowed" --mutants-dir "$mutdir" )
run() { bash "$gate" "${gate_args[@]}" --record "$1" "${@:2}"; }

# Discover the correct lock_inputs_sha256 for these fixtures (all earlier checks pass; DP-6 reports it).
err=$(run "$base" 2>&1 >/dev/null || true)
lock=$(sed -n 's/.*current inputs (\([0-9a-f]\{64\}\)).*/\1/p' <<<"$err")
[[ $lock =~ ^[0-9a-f]{64}$ ]] || { echo "could not discover fixture lock hash; DP-6 error was: $err" >&2; exit 1; }

rec="$tmp/rec.json"   # the NEUTRAL, fully-valid record
jq --arg l "$lock" '.lock_inputs_sha256=$l' "$base" >"$rec"

accept() { # $1=message ; remaining args appended to the run
  local msg=$1; shift
  run "$rec" "$@" >/dev/null 2>&1 || { echo "NEUTRAL control rejected: $msg" >&2; exit 1; }
}
reject() { # $1=expected-DP-tag $2=message $3=record-path ; remaining args appended
  local want=$1 msg=$2 r=$3; shift 3
  local out
  if out=$(run "$r" "$@" 2>&1); then
    echo "POSITIVE control ACCEPTED (check not load-bearing): $msg" >&2; exit 1
  fi
  grep -qF "$want" <<<"$out" \
    || { echo "POSITIVE control '$msg' rejected for the WRONG reason — wanted [$want], got: $out" >&2; exit 1; }
}
mut() { jq "$1" "$rec" >"$tmp/mut.json"; echo "$tmp/mut.json"; }   # record-only mutation

# ---- NEUTRAL: the valid fixture is accepted --------------------------------------------------------
accept "valid reproducible record"

# ---- DP-7: calibration must be present ------------------------------------------------------------
grep -v '^calibration ' "$sweep" >"$tmp/nocalib.sweep.txt"
reject "DP-7:" "uncalibrated sweep" "$rec" --sweep "$tmp/nocalib.sweep.txt"

# ---- DP-8: execution binding ----------------------------------------------------------------------
reject "DP-8:" "execution binding mode missing/invalid" "$(mut '.execution_binding.mode="none"')"
reject "DP-8:" "attested binding without a named re-runner (CL-1)" \
  "$(mut '.execution_binding={"mode":"attested","attested_expiry":"2999-01-01"}')"
reject "DP-8:" "attested binding expired (Law 4)" \
  "$(mut '.execution_binding={"mode":"attested","re_runner":"someone-else","attested_expiry":"2000-01-01"}')"
cp "$sweep" "$tmp/tampered.sweep.txt"; printf 'x\n' >>"$tmp/tampered.sweep.txt"
reject "DP-8:" "tampered committed sweep vs pinned digest" "$rec" --sweep "$tmp/tampered.sweep.txt"

# ---- DP-5: counts re-derived from the raw sweep ---------------------------------------------------
reject "DP-5:" "record killed_count disagrees with the raw sweep" "$(mut '.killed_count=99')"
sed 's/killed=1 /killed=2 /' "$sweep" >"$tmp/badsummary.sweep.txt"
bad_sha=$(sha256sum "$tmp/badsummary.sweep.txt" | awk '{print $1}')
jq --arg s "$bad_sha" '.raw_sweep_sha256=$s' "$rec" >"$tmp/badsum.json"
reject "DP-5:" "summary disagrees with tallied status lines" "$tmp/badsum.json" --sweep "$tmp/badsummary.sweep.txt"

# ---- DP-2 / DP-3: bijection, classification, allowlist binding ------------------------------------
reject "DP-2/DP-3:" "an unclassified sweep survivor (missing from record)" "$(mut '.survivors |= [.[0]]')"
reject "DP-2/DP-3:" "an extra record survivor absent from the sweep" \
  "$(mut '.survivors += [{"id":"'$pkg'/f.go:99#9","class":"E","justification":"bogus"}]')"
reject "DP-3:" "class-R survivor without a pinning line in .pins (CL-3)" \
  "$(mut '(.survivors[]|select(.class=="R")).pins=null')"
reject "DP-3:" "class-R survivor without a subsumption argument (CL-3)" \
  "$(mut '(.survivors[]|select(.class=="R")).justification="pinned by another test"')"
reject "DP-3:" "invalid survivor class" "$(mut '(.survivors[0]).class="X"')"
grep -v 'f.go:10#1' "$allowed" >"$tmp/allowed2.txt"
reject "DP-3:" "survivor not declared in allowed-survivors (binding)" "$rec" --allowed-survivors "$tmp/allowed2.txt"
# register class drift: the 20#2 survivor is {dp:class=R} in the register; record relabels it E
reject "DP-3:" "register/record class drift (Law 2 guard)" "$(mut '.survivors[1].class="E"')"

# ---- DP-1: gate keys on absolute class-U count; diagnostic never drives trust ---------------------
reject "DP-1:" "detector_power_diagnostic disagrees with computed K/(K+U)" "$(mut '.detector_power_diagnostic="9/9"')"
reject "DP-1:" "unproven_latitude_count disagrees with counted class-U survivors" "$(mut '.unproven_latitude_count=1')"

# ---- DP-4: trust gate respects the per-tier U budget ----------------------------------------------
reject "DP-4:" "trusted=false while U=0<=budget (trusted iff U<=budget)" "$(mut '.trusted=false')"
reject "DP-4:" "unknown u_budget_tier (no budget defined — fail closed)" "$(mut '.u_budget_tier="bogus"')"
# Relabelling the 10#1 survivor U requires the register to agree (else the DP-3 drift guard fires
# first). Use a register variant that classes 10#1 as U so DP-4 is the check under test.
sed 's/{dp:class=E}/{dp:class=U}/' "$allowed" >"$tmp/allowed_u.txt"
reject "DP-4:" "class-U survivor above budget still marked trusted" \
  "$(mut '.survivors[0].class="U" | .unproven_latitude_count=1 | .detector_power_diagnostic="1/2"')" \
  --allowed-survivors "$tmp/allowed_u.txt"
# ...and the SAME shape with trusted=false is the honest, accepted form (U reported, not trusted).
# The register variant changes the lock inputs, so recompute the variant's lock before accepting.
jq '.survivors[0].class="U" | .unproven_latitude_count=1 | .detector_power_diagnostic="1/2" | .trusted=false | .lock_inputs_sha256="'"$(printf '%064d' 0)"'"' "$rec" >"$tmp/untrusted0.json"
err_u=$(bash "$gate" "${gate_args[@]}" --allowed-survivors "$tmp/allowed_u.txt" --record "$tmp/untrusted0.json" 2>&1 >/dev/null || true)
lock_u=$(sed -n 's/.*current inputs (\([0-9a-f]\{64\}\)).*/\1/p' <<<"$err_u")
[[ $lock_u =~ ^[0-9a-f]{64}$ ]] || { echo "could not discover variant lock; got: $err_u" >&2; exit 1; }
jq --arg l "$lock_u" '.lock_inputs_sha256=$l' "$tmp/untrusted0.json" >"$tmp/untrusted.json"
bash "$gate" "${gate_args[@]}" --allowed-survivors "$tmp/allowed_u.txt" --record "$tmp/untrusted.json" >/dev/null 2>&1 \
  || { echo "NEUTRAL control rejected: honest untrusted U>0 record" >&2; exit 1; }

# ---- DP-6: freshness over the lock inputs ---------------------------------------------------------
printf '\n// drift\n' >>"$pkgdir/f.go"   # mutate a package input after the record was pinned
reject "DP-6:" "stale record: package source changed" "$rec"
printf 'package fixturepkg\nfunc F() int { return 1 }\n' >"$pkgdir/f.go"   # restore for later checks
accept "record is fresh again after restoring inputs"

# ---- DP-8 reproducible --refresh: a live re-run must reproduce the pinned digest ------------------
{ echo 'calibration neutral=survived invalid=invalid lethal=killed'
  echo "/any/host/path/kernel/$pkg/f.go:5#0 killed"
  echo "/any/host/path/kernel/$pkg/f.go:10#1 survived"
  echo "/any/host/path/kernel/$pkg/f.go:20#2 survived"
  echo "/any/host/path/kernel/$pkg/f.go:30#3 invalid"
  echo 'summary candidates=4 killed=1 invalid=1 survived=2'
} >"$tmp/refresh_raw.txt"
accept "reproducible refresh reproduces the pinned digest" --refresh --sweep-cmd "cat $tmp/refresh_raw.txt"
sed 's/#3 invalid/#3 survived/' "$tmp/refresh_raw.txt" >"$tmp/refresh_bad.txt"
reject "DP-8" "reproducible refresh that does not reproduce the digest" "$rec" --refresh --sweep-cmd "cat $tmp/refresh_bad.txt"

echo 'detector-power.test: ok'
