#!/usr/bin/env bash
# regen-gate.sh — the P7 #2b regeneration-as-repair GATE (gate-only first slice), per the ACCEPTED
# design spec docs/plans/P7-2b-regeneration-loop-spec-draft1.md (§5 SETTLED: gate-only) and change
# intent CI-regen-loop-gate-1a2b3c (implements BR-regeneration-trust-5f8a03 O-2).
#
# It does NOT regenerate (step 3) or verify a candidate (step 4) — those are manual / out of this
# slice. It enforces the loop's AUTHORIZE/DENY gate over COMMITTED evidence and FAILS CLOSED:
#   RL-1  new revision:  L@r+1.number == L@r.number+1 ; parent == BehaviorLockDigest(L@r) ; same subject
#   RL-11 digest identity: the L@r+1 canonical digest (sha256 of its canonical bytes) is the one the
#         admission, the detector-power record, and the D-b attestation all name
#   RL-9  counterexample lethality: admission records a reproduced violation on the pre-repair impl
#   RL-2  L@r+1 detector power: a fresh (not inherited) detector-power record for the subject package
#         clears U=0/trusted, verified by reusing scripts/detector-power.sh
#   RL-4  D-a representativeness record present (argument or unexpired Law-4 advisory)
#   RL-5  D-b isolation attestation present AND made/checked by a NAMED NON-AUTHOR of the regeneration
#   RL-10 three distinct NAMED identities: lock_author, regenerator, verdict_reviewer
#   RL-8  the declared Law-4 limitations are recorded
# AUTHORIZE (exit 0) only if every applicable control passes; any failure => DENY (exit 1).
#
# DESIGN SEAM (for the non-author verdict): O-1 detector-power records are keyed per PACKAGE, while a
# behavior lock is a per-revision artifact. This slice binds them through the admission evidence
# (which names the L@r+1 digest AND the detector-power record), and checks that binding structurally;
# it does not re-derive the package<->lock mapping. Carried as a declared limitation (RL-8).
set -euo pipefail

die_deny() { echo "regen-gate: DENY: $*" >&2; exit 1; }   # fail closed
command -v jq >/dev/null || die_deny 'jq is required'
command -v sha256sum >/dev/null || die_deny 'sha256sum is required'

lr='' lr1='' admission='' dp_record='' da='' db='' subject_pkg='' dp_cmd=''
while (($#)); do
  case $1 in
    --lr) shift; lr=${1:?}; ;;
    --lr1) shift; lr1=${1:?}; ;;
    --admission) shift; admission=${1:?}; ;;
    --dp-record) shift; dp_record=${1:?}; ;;
    --da) shift; da=${1:?}; ;;
    --db) shift; db=${1:?}; ;;
    --subject-pkg) shift; subject_pkg=${1:?}; ;;
    --dp-cmd) shift; dp_cmd=${1:?}; ;;   # override the detector-power invocation (hermetic self-test)
    *) die_deny "unknown argument: $1" ;;
  esac
  shift
done
for v in lr lr1 admission dp_record da db subject_pkg; do
  [[ -n ${!v} ]] || die_deny "missing required --${v//_/-}"
  [[ -f ${!v} || $v == subject_pkg ]] || die_deny "file not found for --${v//_/-}: ${!v}"
done
jq -e . "$lr" >/dev/null 2>&1 || die_deny "L@r is not valid JSON"
jq -e . "$lr1" >/dev/null 2>&1 || die_deny "L@r+1 is not valid JSON"
jq -e . "$admission" >/dev/null 2>&1 || die_deny "admission is not valid JSON"

# --- lock digests (BehaviorLockDigest = sha256 of the canonical lock bytes; the committed file IS canonical) ---
lr_digest=$(sha256sum "$lr" | awk '{print $1}')
lr1_digest=$(sha256sum "$lr1" | awk '{print $1}')

# --- RL-1: new revision, ancestry, same subject ----------------------------------------------------
lr_num=$(jq -r '.lock_revision.number // empty' "$lr")
lr1_num=$(jq -r '.lock_revision.number // empty' "$lr1")
lr_subj=$(jq -r '.lock_revision.subject // empty' "$lr")
lr1_subj=$(jq -r '.lock_revision.subject // empty' "$lr1")
lr1_parent=$(jq -r '.lock_revision.parent // empty' "$lr1")
[[ $lr_num =~ ^[0-9]+$ && $lr1_num =~ ^[0-9]+$ ]] || die_deny "RL-1: lock_revision.number must be integers"
[[ $lr1_num -eq $((lr_num + 1)) ]] || die_deny "RL-1: L@r+1.number ($lr1_num) must be L@r.number+1 ($((lr_num + 1)))"
[[ -n $lr_subj && $lr1_subj == "$lr_subj" ]] || die_deny "RL-1: subject mismatch (L@r=$lr_subj L@r+1=$lr1_subj)"
[[ $lr1_parent == "$lr_digest" ]] || die_deny "RL-1: L@r+1.parent ($lr1_parent) must be BehaviorLockDigest(L@r) ($lr_digest)"

# --- RL-11: the L@r+1 digest is the one every piece of evidence names ------------------------------
adm_lr1=$(jq -r '.lock_revision_digest // empty' "$admission")
[[ $adm_lr1 == "$lr1_digest" ]] || die_deny "RL-11: admission lock_revision_digest ($adm_lr1) != L@r+1 digest ($lr1_digest)"
db_lr1=$(jq -r '.lock_revision_digest // empty' "$db" 2>/dev/null || true)
[[ $db_lr1 == "$lr1_digest" ]] || die_deny "RL-11: D-b attestation lock_revision_digest ($db_lr1) != L@r+1 digest ($lr1_digest)"
dp_lr1=$(jq -r '.lock_revision_digest // empty' "$dp_record" 2>/dev/null || true)
[[ $dp_lr1 == "$lr1_digest" ]] || die_deny "RL-11: detector-power record lock_revision_digest ($dp_lr1) != L@r+1 digest ($lr1_digest)"

# --- RL-9: counterexample lethality (reproduced violation on the pre-repair impl) ------------------
ce_id=$(jq -r '.counterexample.id // empty' "$admission")
ce_lethal=$(jq -r '.counterexample.reproduced_violation // empty' "$admission")
[[ -n $ce_id ]] || die_deny "RL-9: admission.counterexample.id is empty"
[[ -n $ce_lethal && $ce_lethal != null ]] || die_deny "RL-9: admission.counterexample.reproduced_violation absent — X not shown lethal on the pre-repair impl"

# --- RL-10: three distinct NAMED identities --------------------------------------------------------
la=$(jq -r '.identities.lock_author // empty' "$admission")
rg=$(jq -r '.identities.regenerator // empty' "$admission")
vr=$(jq -r '.identities.verdict_reviewer // empty' "$admission")
[[ -n $la && -n $rg && -n $vr ]] || die_deny "RL-10: identities.{lock_author,regenerator,verdict_reviewer} must all be named"
[[ $la != "$rg" && $la != "$vr" && $rg != "$vr" ]] || die_deny "RL-10: lock_author/regenerator/verdict_reviewer must be three DISTINCT identities (got la=$la rg=$rg vr=$vr)"

# --- RL-4: D-a representativeness record present ---------------------------------------------------
[[ -f $da ]] || die_deny "RL-4: D-a representativeness record not found: $da"
da_stmt=$(jq -r '.representativeness // empty' "$da" 2>/dev/null || grep -c . "$da" 2>/dev/null || echo '')
[[ -n $da_stmt && $da_stmt != 0 ]] || die_deny "RL-4: D-a representativeness record is empty: $da"

# --- RL-5: D-b isolation attestation by a NAMED NON-AUTHOR of the regeneration ---------------------
db_checker=$(jq -r '.checked_by // empty' "$db")
db_stmt=$(jq -r '.isolation // empty' "$db")
[[ -n $db_stmt && $db_stmt != null ]] || die_deny "RL-5: D-b isolation attestation is empty"
[[ -n $db_checker ]] || die_deny "RL-5: D-b attestation has no named checked_by"
[[ $db_checker != "$rg" ]] || die_deny "RL-5: D-b attestation checker ($db_checker) is the regenerator — must be a non-author of the regeneration (O-1 CL-1)"

# --- RL-8: declared limitations recorded ----------------------------------------------------------
db_lims=$(jq -r '.declared_limitations // empty' "$db")
[[ -n $db_lims && $db_lims != null ]] || die_deny "RL-8: the declared Law-4 limitations are not recorded in the D-b attestation"

# --- RL-2: L@r+1 detector power clears U=0/trusted (fresh, not inherited), via detector-power.sh ----
# The dp-record must be for the subject package; verify via detector-power.sh (reused). A --dp-cmd
# override keeps the self-test hermetic. Freshness: the dp-record must NOT inherit L@r's swept commit.
dp_swept=$(jq -r '.swept_at_commit // empty' "$dp_record" 2>/dev/null || true)
parent_dp_swept=$(jq -r '.parent_detector_power_swept_at_commit // empty' "$admission")
[[ -n $dp_swept ]] || die_deny "RL-2: detector-power record missing swept_at_commit"
[[ -n $parent_dp_swept ]] || die_deny "RL-2: admission missing parent_detector_power_swept_at_commit (the no-inheritance anchor)"
[[ $dp_swept != "$parent_dp_swept" ]] || die_deny "RL-2: L@r+1 detector-power record inherits L@r's swept commit ($dp_swept) — a fresh re-measurement is required (no inheritance)"
cmd=${dp_cmd:-"bash $(dirname "$0")/detector-power.sh $subject_pkg --record $dp_record"}
set +e; dp_out=$(eval "$cmd" 2>&1); dp_rc=$?; set -e
[[ $dp_rc -eq 0 ]] || die_deny "RL-2: detector-power verification failed (exit $dp_rc): $dp_out"
grep -Eq 'U=0( |$)' <<<"$dp_out" || die_deny "RL-2: L@r+1 detector power is not U=0: $dp_out"
grep -Eq 'trusted=true( |$)' <<<"$dp_out" || die_deny "RL-2: L@r+1 is not detector-power-trusted: $dp_out"

echo "regen-gate: AUTHORIZE pkg=$subject_pkg L@r=$lr_num->L@r+1=$lr1_num lr1_digest=$lr1_digest U=0 trusted=true counterexample=$ce_id"
