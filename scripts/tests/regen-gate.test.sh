#!/usr/bin/env bash
# Self-test for scripts/regen-gate.sh (P7 #2b regeneration-as-repair gate, gate-only slice).
# Law 3: a hook's self-test proves it FIRES. A NEUTRAL fixture MUST AUTHORIZE; a POSITIVE control per
# RL — one field mutated, the rest valid — MUST be DENIED AND for the INTENDED reason (its RL tag must
# appear), so a control cannot pass by tripping a different check. Fully hermetic: synthetic fixtures
# in a temp dir, a stubbed detector-power command, no real sweep, no repo lock.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
gate=scripts/regen-gate.sh

tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT

# L@r (revision 1) and L@r+1 (revision 2, parent = BehaviorLockDigest(L@r) = sha256 of L@r bytes).
printf '{"lock_revision":{"subject":"internal/twin/manifest","number":1,"parent":"genesis"}}\n' >"$tmp/lr.json"
lr_digest=$(sha256sum "$tmp/lr.json" | awk '{print $1}')
printf '{"lock_revision":{"subject":"internal/twin/manifest","number":2,"parent":"%s"}}\n' "$lr_digest" >"$tmp/lr1.json"
lr1_digest=$(sha256sum "$tmp/lr1.json" | awk '{print $1}')

cat >"$tmp/admission.json" <<EOF
{"lock_revision_digest":"$lr1_digest",
 "counterexample":{"id":"CE-1","reproduced_violation":"vera.replay.v1 yielded changed where preserved required on the pre-repair impl"},
 "identities":{"lock_author":"alice","regenerator":"bob","verdict_reviewer":"carol"},
 "parent_detector_power_swept_at_commit":"0000000000000000000000000000000000000001"}
EOF
cat >"$tmp/dp.json" <<EOF
{"lock_revision_digest":"$lr1_digest","swept_at_commit":"0000000000000000000000000000000000000002"}
EOF
printf '{"representativeness":"operator basis spans the regenerator failure classes (bounded-advisory, genesis)"}\n' >"$tmp/da.json"
cat >"$tmp/db.json" <<EOF
{"lock_revision_digest":"$lr1_digest","isolation":"candidate derived from the contract, not the measured tests",
 "checked_by":"dave","declared_limitations":"single-context D-b seam; operator-unreachable blind spots"}
EOF

# detector-power stub (hermetic): echoes a controllable verdict line, exit 0.
ok_dp="printf 'regen-gate-test K=330 U=0 detector_power=330/330 trusted=true\\n'"

base=( --lr "$tmp/lr.json" --lr1 "$tmp/lr1.json" --admission "$tmp/admission.json"
       --dp-record "$tmp/dp.json" --da "$tmp/da.json" --db "$tmp/db.json"
       --subject-pkg internal/twin/manifest )

run() { bash "$gate" "${base[@]}" --dp-cmd "$ok_dp" "$@"; }
reject() { # $1=tag $2=msg ; remaining args appended to run
  local tag=$1 msg=$2; shift 2
  local out
  if out=$(run "$@" 2>&1); then echo "CONTROL ACCEPTED (not load-bearing): $msg" >&2; exit 1; fi
  grep -qF "$tag" <<<"$out" || { echo "CONTROL '$msg' denied for WRONG reason — wanted [$tag], got: $out" >&2; exit 1; }
}
mutate() { local dst=$tmp/mut.json; jq "$2" "$1" >"$dst"; echo "$dst"; }

# ---- NEUTRAL: valid evidence AUTHORIZEs -----------------------------------------------------------
run >/dev/null 2>&1 || { echo "NEUTRAL control DENIED (should AUTHORIZE)" >&2; exit 1; }

# ---- RL-1: new revision / ancestry ----------------------------------------------------------------
reject "RL-1:" "L@r+1.number not L@r.number+1"  --lr1 "$(mutate "$tmp/lr1.json" '.lock_revision.number=5')"
printf '{"lock_revision":{"subject":"internal/twin/manifest","number":2,"parent":"deadbeef"}}\n' >"$tmp/badparent.json"
reject "RL-1:" "L@r+1.parent not BehaviorLockDigest(L@r)"  --lr1 "$tmp/badparent.json"

# ---- RL-11: digest identity across evidence -------------------------------------------------------
reject "RL-11:" "admission names a different L@r+1 digest"  --admission "$(mutate "$tmp/admission.json" '.lock_revision_digest="feedface"')"
reject "RL-11:" "detector-power record names a different L@r+1 digest"  --dp-record "$(mutate "$tmp/dp.json" '.lock_revision_digest="feedface"')"

# ---- RL-9: counterexample lethality ---------------------------------------------------------------
reject "RL-9:" "no reproduced violation (X not shown lethal)"  --admission "$(mutate "$tmp/admission.json" 'del(.counterexample.reproduced_violation)')"

# ---- RL-10: three distinct named identities -------------------------------------------------------
reject "RL-10:" "regenerator == lock_author (roles collapse)"  --admission "$(mutate "$tmp/admission.json" '.identities.regenerator="alice"')"
reject "RL-10:" "an identity unnamed"  --admission "$(mutate "$tmp/admission.json" '.identities.verdict_reviewer=""')"

# ---- RL-4: D-a representativeness present ----------------------------------------------------------
printf '{}\n' >"$tmp/emptyda.json"
reject "RL-4:" "D-a representativeness empty"  --da "$tmp/emptyda.json"

# ---- RL-5: D-b non-author isolation attestation ---------------------------------------------------
reject "RL-5:" "D-b checked_by == regenerator (self-attestation)"  --db "$(mutate "$tmp/db.json" '.checked_by="bob"')"

# ---- RL-8: declared limitations recorded ----------------------------------------------------------
reject "RL-8:" "declared limitations absent"  --db "$(mutate "$tmp/db.json" 'del(.declared_limitations)')"

# ---- RL-2: L@r+1 detector power + no inheritance --------------------------------------------------
reject "RL-2:" "L@r+1 not U=0"  --dp-cmd "printf 'K=9 U=1 trusted=false\\n'"
reject "RL-2:" "detector-power verification failed (non-zero exit)"  --dp-cmd "false"
reject "RL-2:" "inherits L@r's swept commit (no re-measurement)"  --dp-record "$(mutate "$tmp/dp.json" '.swept_at_commit="0000000000000000000000000000000000000001"')"

echo "regen-gate.test: ok"
