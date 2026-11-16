#!/usr/bin/env bash
# Proofbound — regeneration-repair SAFETY GATE demo (Release 3 preview, #2b gate-only slice).
#
# What this shows, in one run: when a rule ("behavior lock") is broken and someone wants to REBUILD
# the code that broke it, Proofbound's gate decides whether that rebuild can be TRUSTED — and refuses
# ("DENY") when the rule is too weak, the failing example is a strawman, or one person is grading
# their own work. When everything checks out it says "AUTHORIZE" (safe to rebuild).
#
# This is the gate-only slice: the gate JUDGES the rebuild; a person still DOES the rebuild by hand.
# Nothing here invokes an AI or rewrites code.
#
# HONEST NOTE ON THE STUB: one input to the gate is a "detector-power" measurement — how good the
# rule is at catching wrong code, from a real mutation sweep. That sweep is slow, so this demo feeds
# the gate a *pre-canned* measurement via --dp-cmd (exactly how the gate's own self-test runs). Every
# OTHER check below is the real gate running on real evidence files. The stubbed line is labeled.
set -euo pipefail

repo=$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)
gate="$repo/scripts/regen-gate.sh"
[[ -x $gate || -f $gate ]] || { echo "cannot find $gate" >&2; exit 1; }

tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
hr() { printf '%s\n' "------------------------------------------------------------------------"; }
note() { printf '  %s\n' "$*"; }

# ---------------------------------------------------------------------------------------------------
# Build a small, valid set of evidence files describing ONE repair cycle:
#   L@r      the rule as it was        (revision 1)
#   L@r+1    the rule after folding in the failing example as a new pinned case (revision 2)
#   admission   the counterexample + who played which role + the "no-inheritance" anchor
#   dp          the detector-power record naming the new rule revision
#   da / db     the two anti-cheating records (representativeness; no-peeking-at-the-tests)
# ---------------------------------------------------------------------------------------------------
printf '{"lock_revision":{"subject":"internal/twin/manifest","number":1,"parent":"genesis"}}\n' >"$tmp/lr.json"
lr_digest=$(sha256sum "$tmp/lr.json" | awk '{print $1}')
printf '{"lock_revision":{"subject":"internal/twin/manifest","number":2,"parent":"%s"}}\n' "$lr_digest" >"$tmp/lr1.json"
lr1_digest=$(sha256sum "$tmp/lr1.json" | awk '{print $1}')

cat >"$tmp/admission.json" <<EOF
{"lock_revision_digest":"$lr1_digest",
 "counterexample":{"id":"CE-1","reproduced_violation":"replay yielded 'changed' where the rule requires 'preserved' — reproduced on the OLD code"},
 "identities":{"lock_author":"alice","regenerator":"bob","verdict_reviewer":"carol"},
 "parent_detector_power_swept_at_commit":"0000000000000000000000000000000000000001"}
EOF
printf '{"lock_revision_digest":"%s","swept_at_commit":"0000000000000000000000000000000000000002"}\n' "$lr1_digest" >"$tmp/dp.json"
printf '{"representativeness":"operator basis spans the known rebuild failure classes (bounded-advisory, genesis)"}\n' >"$tmp/da.json"
cat >"$tmp/db.json" <<EOF
{"lock_revision_digest":"$lr1_digest","isolation":"candidate derived from the rule's contract, not the test bodies",
 "checked_by":"dave","declared_limitations":"single-context no-peeking seam; operator-unreachable blind spots"}
EOF

# A pre-canned "strong rule" detector-power line (stub; the real thing is a mutation sweep).
STRONG_RULE="printf 'DEMO-STUB K=330 U=0 detector_power=330/330 trusted=true\\n'"
# A pre-canned "weak rule" line: U=1 means one blind spot the rule can't catch.
WEAK_RULE="printf 'DEMO-STUB K=330 U=1 detector_power=329/330 trusted=true\\n'"

base=( --lr "$tmp/lr.json" --lr1 "$tmp/lr1.json" --admission "$tmp/admission.json"
       --dp-record "$tmp/dp.json" --da "$tmp/da.json" --db "$tmp/db.json"
       --subject-pkg internal/twin/manifest )

mutate() { jq "$2" "$1" >"$tmp/mut.json"; echo "$tmp/mut.json"; }

show() { # $1=title  $2=plain-English expectation ; remaining args appended to the gate call
  local title=$1 expect=$2; shift 2
  hr; printf 'SCENARIO: %s\n' "$title"; note "$expect"; echo
  if out=$(bash "$gate" "${base[@]}" "$@" 2>&1); then
    printf '  >> %s\n' "$out"
    printf '  RESULT: ✅ AUTHORIZE — the gate says this rebuild can be trusted.\n'
  else
    printf '  >> %s\n' "$out"
    printf '  RESULT: 🛑 DENY — the gate refuses to trust this rebuild (fails closed).\n'
  fi
  echo
}

cat <<'INTRO'
========================================================================
 Proofbound — "is this rebuild safe to trust?" gate (live demo)
========================================================================
A rule was broken. Someone wants to throw out the broken code and rebuild
it. Before trusting that rebuild, Proofbound checks the evidence. Watch it
say YES once, then refuse three unsafe situations.
INTRO
echo

show "A clean repair — everything checks out" \
     "Strong rule, a real failing example, three different people in the three roles." \
     --dp-cmd "$STRONG_RULE"

show "The rule is too weak to trust a rebuild" \
     "Same setup, but the rule has a blind spot (it can't catch one kind of wrong code)." \
     --dp-cmd "$WEAK_RULE"

show "The failing example was a strawman" \
     "The 'counterexample' was never shown to actually break the old code." \
     --dp-cmd "$STRONG_RULE" \
     --admission "$(mutate "$tmp/admission.json" 'del(.counterexample.reproduced_violation)')"

show "One person is grading their own work" \
     "The person who wrote the rule is also the one who rebuilt the code (roles collapse)." \
     --dp-cmd "$STRONG_RULE" \
     --admission "$(mutate "$tmp/admission.json" '.identities.regenerator="alice"')"

hr
cat <<'OUTRO'
Takeaway: the gate only says AUTHORIZE when the rule is strong enough, the
failure is real, and no one graded their own work. Otherwise it fails
closed. A person still does the actual rebuild — the gate decides whether
that rebuild can be trusted.
OUTRO
