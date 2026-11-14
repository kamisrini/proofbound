#!/usr/bin/env bash
# Proofbound — seed a self-contained DEMO APP that shows the FULL Release-1 value.
#
# Creates a throwaway sample application ("ledgerloom") pre-loaded with a realistic history and all
# the records needed to exercise every Release-1 capability, then records it into Proofbound so the
# full trust record is live and explorable:
#   * a permanent, content-addressed trust record (proof IDs bound to commits)
#   * a decision bound to the commit that implements it
#   * the full decision -> requirement -> change -> commit INTENT CHAIN
#   * independent REVIEW VERDICTS, including a RED-VERDICT CHAIN (a fix that introduced a new defect)
#   * a record that rebuilds identically from scratch
#
# It leaves the seeded app in place so you can poke at it afterward (report week, report intent, ...).
#
# USAGE
#   seed-demo-app.sh [TARGET_DIR]       # seed + print a concise summary (default: $HOME/proofbound-demo-app)
#   seed-demo-app.sh --tour [TARGET_DIR]# same, with a narrated walkthrough of what each step proves
#
# PREREQUISITES
#   * `proofbound` on your PATH (build it per DEMO.md §1), or set PROOFBOUND_BIN=/path/to/proofbound
#   * Git and Go 1.26+; outbound internet on the first run (downloads the embedded PostgreSQL once)
#   * Run on a LINUX filesystem (WSL home), NOT a Windows /mnt/d path.
#
# Portable + independent: copy this one file to any machine with the proofbound binary and run it.
set -euo pipefail

TOUR=0
args=()
for a in "$@"; do case "$a" in --tour) TOUR=1;; *) args+=("$a");; esac; done
TARGET="${args[0]:-$HOME/proofbound-demo-app}"

BIN="${PROOFBOUND_BIN:-proofbound}"
command -v "$BIN" >/dev/null 2>&1 || { echo "ERROR: '$BIN' not found. Build it (DEMO.md §1) or set PROOFBOUND_BIN." >&2; exit 1; }
case "$TARGET" in /mnt/*) echo "WARNING: $TARGET is on a Windows drive; the embedded PostgreSQL may fail. Use a Linux path." >&2;; esac

# Hide the embedded PostgreSQL's one-time initdb + start/stop log noise; show only Proofbound's lines.
PGNOISE='LOG:|server (started|stopped)|listening|database system|starting up|shutting down|background|pg_|checkpoint|redo|MultiXact|autovacuum|FATAL|terminating|This user must|database cluster|locale|encoding|initialized|^The |^Data page|creating (directory|subdirectories|configuration)|selecting (dynamic|default)|running bootstrap|performing post-bootstrap|syncing data|Success\.|You can now|^ done$|^$'
pb() { "$BIN" "$@" 2>&1 | grep -vE "$PGNOISE" || true; }
say()  { [ "$TOUR" = 1 ] && printf '\n\033[1m== %s\033[0m\n' "$*" || printf '\n== %s\n' "$*"; }
tour() { [ "$TOUR" = 1 ] && printf '   %s\n' "$*" || true; }   # narration only in --tour mode
out()  { sed 's/^/   /'; }
zeros=$(printf '0%.0s' {1..64})

rm -rf "$TARGET"; mkdir -p "$TARGET"; cd "$TARGET"
git init -q; git config user.email dev@example.com; git config user.name "Sample Dev"
printf '.proofbound/\n' > .gitignore
printf 'module example.com/ledgerloom\ngo 1.21\n' > go.mod

cat <<INTRO
========================================================================
 Proofbound — seeding a full-value demo app at:
   $TARGET
========================================================================
A tiny invoice-tax library with a realistic history (a decision, a
requirement, a change that implements it, two independent reviews that
catch defects), recorded into Proofbound. Proof IDs (proof=01M...) are
time-based and differ every run.
INTRO

# ---------------------------------------------------------------------------------------------------
say "1. A permanent trust record (proof IDs bound to commits)"
printf 'package ledgerloom\n\n// Tax returns cents of tax for an amount of cents at a whole-percent rate.\nfunc Tax(amountCents, ratePct int) int { return amountCents * ratePct / 100 }\n' > tax.go
git add -A; git commit -q -m "feat: tax computes whole-percent tax"
tour "proofbound sync git turns git history into permanent, content-addressed records."
pb sync git | out

# ---------------------------------------------------------------------------------------------------
say "2. The intent chain: decision -> requirement -> change -> the implementing commit"
tour "Three records state WHY the work exists and WHAT it must guarantee; each child cites its"
tour "parent's exact digest (a tamper-evident chain). The commit cites the change intent in a trailer."
mkrec() { mkdir -p "$(dirname "$1")"; { printf '```proofbound-json\n%s\n```\n\n%s\n' "$2" "$3"; } > "$1"; }
dip() { local d; d=$(sha256sum "$1" | awk '{print $1}'); sed -i "s/$zeros/$d/" "$1"; echo "$d"; }

BD=BD-charge-invoice-tax-7a1b2c; BDp=docs/intent/records/business-decisions/$BD/0001.md
mkrec "$BDp" '{"schema":"proofbound.business-decision.v1","decision_id":"'$BD'","status":"accepted","outcome":"Charge sales tax on every invoice line.","declared_owner":"product-lead","declared_approver":"product-lead","decided_at":"2026-11-10T00:00:00Z","supersedes":[],"artifact_path":"'$BDp'","artifact_sha256":"'$zeros'"}' "# $BD — charge invoice tax (business decision)"
BDd=$(dip "$BDp")
BR=BR-tax-never-negative-4d5e6f; BRp=docs/intent/records/requirements/$BR/0001.md
mkrec "$BRp" '{"schema":"proofbound.requirement.v1","requirement_id":"'$BR'","status":"proposed","authorized_by":[{"record_kind":"business_decision","source":"intent.records","record_id":"'$BD'","artifact_sha256":"'$BDd'","relation":"authorizes"}],"declared_owner":"product-lead","obligations":[{"id":"O-1","statement":"Computed tax must never be negative for any invoice line.","state":"active"}],"supersedes":[],"artifact_path":"'$BRp'","artifact_sha256":"'$zeros'"}' "# $BR — tax never negative (requirement, obligation O-1)"
BRd=$(dip "$BRp")
CI=CI-clamp-negative-tax-8c9d0e; CIp=docs/intent/records/change-intents/$CI/0001.md
mkrec "$CIp" '{"schema":"proofbound.change-intent.v1","intent_id":"'$CI'","status":"accepted","declared_sponsor":"product-lead","targets":[{"record_kind":"requirement","source":"intent.records","record_id":"'$BR'","artifact_sha256":"'$BRd'","relation":"implements","obligation_ids":["O-1"]}],"constrained_by":[],"supersedes":[],"artifact_path":"'$CIp'","artifact_sha256":"'$zeros'"}' "# $CI — clamp negative tax (change intent)"
CId=$(dip "$CIp")
git add -A; git commit -q -m "docs(intent): charge-tax decision chain (BD/BR/CI)"

printf 'package ledgerloom\n\n// Tax returns cents of tax; never negative (clamped).\nfunc Tax(amountCents, ratePct int) int {\n\tt := amountCents * ratePct / 100\n\tif t < 0 {\n\t\treturn 0\n\t}\n\treturn t\n}\n' > tax.go
git commit -q -am "fix(tax): clamp negative tax to zero

Clamp computed tax to zero so an invoice line can never show negative tax.

Intent: $CI"
IMPL=$(git rev-parse HEAD)
# ORDER MATTERS: intent records must enter the ledger BEFORE the commit that cites them, or the
# reference reads as "dangling". Always sync intent before sync git.
pb sync intent all >/dev/null; pb sync git >/dev/null
tour "proofbound intent check --commit <impl>  — is this commit bound to an authorized intent?"
pb intent check --commit "$IMPL" | out
tour ""
tour "proofbound report intent $CI  — the whole chain, with proof IDs and HONEST status:"
pb report intent "$CI" | out
tour "(review=UNREVIEWED, outcome=UNVERIFIED: Proofbound separates 'implemented' from 'verified'.)"

# ---------------------------------------------------------------------------------------------------
say "3. Independent reviews — including a fix that introduced a new defect (red-verdict chain)"
mkverdict() { # $1=id $2=commit $3=findingid $4=sev
  local p=docs/verification/verdicts/$1.md
  mkdir -p "$(dirname "$p")"
  cat > "$p" <<EOF
---
schema: vera.verdict.v1
verdict_id: $1
status: NEEDS_WORK
reviewed_commit: $2
findings:
  - finding_id: $3
    severity: $4
    defect_commit: $2
artifact_path: $p
artifact_sha: $zeros
---

# Review verdict $1
EOF
  local d; d=$(sha256sum "$p" | awk '{print $1}'); sed -i "s/artifact_sha: $zeros/artifact_sha: $d/" "$p"
}
# Interleave the syncs so the ledger order is red -> change -> red (that IS a red-verdict chain).
printf 'package ledgerloom\n\n// Tax rounds to nearest cent (claimed).\nfunc Tax(amountCents, ratePct int) int {\n\tt := amountCents * ratePct / 100 // BUG: truncates\n\tif t < 0 {\n\t\treturn 0\n\t}\n\treturn t\n}\n' > tax.go
git commit -q -am "feat(tax): claim rounding to nearest cent"; BUG=$(git rev-parse HEAD)
mkverdict review-tax-rounding-round1 "$BUG" tax-rounding-truncates MED
git add -A; git commit -q -m "review: rounding claim is false (NEEDS_WORK)"
pb sync git >/dev/null; pb sync reviews >/dev/null
printf 'package ledgerloom\n\n// Tax rounds half up.\nfunc Tax(amountCents, ratePct int) int {\n\tt := (amountCents*ratePct + 50) / 100\n\tif t < 0 {\n\t\treturn 0\n\t}\n\treturn t\n}\n' > tax.go
git commit -q -am "fix(tax): round half up instead of truncating"; FIX=$(git rev-parse HEAD)
mkverdict review-tax-overflow-round1 "$FIX" tax-rounding-overflows HIGH
git add -A; git commit -q -m "review: the rounding fix can overflow (NEEDS_WORK)"
pb sync git >/dev/null; pb sync reviews >/dev/null
tour "Two reviews recorded; the second caught a defect the FIX introduced."

# ---------------------------------------------------------------------------------------------------
say "4. A decision bound to the commit that implements it"
mkdir -p docs/decisions
printf '# VD-rounding-policy-3f2a1b — Round tax half-up\n\nStatus: accepted\nDate: 2026-11-12\n\nTax rounds half-up to the nearest cent.\n' > docs/decisions/VD-rounding-policy-3f2a1b.md
git add -A; git commit -q -m "docs: record rounding-policy decision"
printf 'package ledgerloom\n\n// Tax rounds half up (overflow-safe).\nfunc Tax(amountCents, ratePct int) int {\n\tt := int((int64(amountCents)*int64(ratePct) + 50) / 100)\n\tif t < 0 {\n\t\treturn 0\n\t}\n\treturn t\n}\n' > tax.go
git commit -q -am "fix(tax): overflow-safe half-up rounding

Implements the rounding policy decided in VD-rounding-policy-3f2a1b."
pb sync git >/dev/null

# ---------------------------------------------------------------------------------------------------
say "5. CI witnesses — tamper-evident proof that a check actually RAN"
tour "A witness is produced by the check itself, never asserted by an agent: the command, the result,"
tour "the timing, bound to the exact commit. We seed one PASSING and one FAILING check for HEAD —"
tour "a failure recorded as permanent evidence, not a green badge that forgets."
mkdir -p .proofbound/spool
SHA=$(git rev-parse HEAD)
mkwitness() { # $1 run_id(ULID)  $2 "make <target>"  $3 exit_code  $4 duration_ms  $5 started  $6 finished  $7 output
  local osha; osha=$(printf '%s\n' "$7" | sha256sum | awk '{print $1}')
  cat > ".proofbound/spool/$1.json" <<EOF
{"schema":"vera.witness.v1","run_id":"$1","command":"$2","exit_code":$3,"started_at":"$5","finished_at":"$6","duration_ms":$4,"output_sha256":"$osha","git_sha":"$SHA","git_dirty":false,"tool_versions":{"go":"go1.27.4","golangci_lint":"2.13.2","make":"GNU Make 4.3"}}
EOF
}
mkwitness 01HQRST9VWXYZ0123456789ABC "make test"  0 3000 "2026-11-14T10:00:00Z" "2026-11-14T10:00:03Z" "all tests passed"
mkwitness 01HQRST9VWXYZ0123456789ABD "make build" 2 1200 "2026-11-14T10:05:00Z" "2026-11-14T10:05:01Z" "build failed: undefined symbol"
pb sync checks >/dev/null
tour "proofbound sync checks ingested the spool; the pass and the fail are now permanent evidence."
tour "(In production, a one-line wrapper writes these witnesses as your real tests/CI run.)"

# ---------------------------------------------------------------------------------------------------
say "6. The whole trust record"
pb report week | grep -E '^(week|commits|- |checks|sessions|reviews|red_)' | out

# ---------------------------------------------------------------------------------------------------
say "7. Gates — rules over the witnesses that stop the bad thing at the boundary"
tour "A gate is a small declarative rule over the check witnesses. We define two: 'tests-pass' is"
tour "promoted to ENFORCE; 'build-passes' is still in CANARY (canary evaluates without ever blocking)."
EXP=$(date -u -d '+2 years' +%Y-%m-%d 2>/dev/null || date -u -v+2y +%Y-%m-%d)
mkdir -p gates
cat > gates/tests-pass.yaml <<GATE
{"schema":"proofbound.gate.v1","id":"tests-pass","description":"The make test check must succeed","expires":"$EXP","mode":"enforce","source":"checks","kind":"check.run","selector":{"field":"command","equals":"make test"},"condition":{"field":"exit_code","equals":0}}
GATE
cat > gates/build-passes.yaml <<GATE
{"schema":"proofbound.gate.v1","id":"build-passes","description":"The make build check must succeed","expires":"$EXP","mode":"canary","source":"checks","kind":"check.run","selector":{"field":"command","equals":"make build"},"condition":{"field":"exit_code","equals":0}}
GATE
tour "proofbound gates canary  — evaluate ALL gates, block nothing (prove behavior first):"
pb gates canary | out
tour ""
tour "proofbound gates enforce  — only PROMOTED (enforce-mode) gates can block:"
pb gates enforce | out
tour "The canary caught the failing build (would_block=true) WITHOUT blocking anything; the promoted"
tour "tests-pass gate is green. Flip build-passes to mode:enforce once the build is fixed, and"
tour "'proofbound gates enforce' would exit non-zero and stop the pipeline — the boundary stop."

# ---------------------------------------------------------------------------------------------------
say "8. The record is reproducible, not a snapshot"
B=$(pb report week | grep -E '^- .*whole-percent' | grep -oE 'proof=[0-9A-Z]+')
pb rebuild >/dev/null
A=$(pb report week | grep -E '^- .*whole-percent' | grep -oE 'proof=[0-9A-Z]+')
tour "Dropped all projections and rebuilt from the raw ledger."
printf '   first commit proof id  before: %s\n   first commit proof id  after:  %s\n' "$B" "$A"
[ "$B" = "$A" ] && printf '   IDENTICAL — the trust record rebuilds exactly.\n' || printf '   CHANGED — unexpected.\n'

cat <<DONE

========================================================================
 Seeded. Explore your demo app:
========================================================================
   cd "$TARGET"
   proofbound report week                         # the whole trust record
   proofbound report intent $CI   # the decision->commit chain
   proofbound report requirement $BR     # a requirement's obligations
   proofbound intent check --commit $IMPL   # bind a commit to its intent
   proofbound gates canary                        # evaluate the gates (blocks nothing)
   proofbound gates enforce                        # only promoted gates block
   proofbound rebuild && proofbound report week   # prove determinism
   # then: edit gates/build-passes.yaml -> "mode":"enforce" and re-run 'gates enforce' to see it BLOCK

 Not shown on a sample app (by design): verify (Proofbound-only),
 github (needs a token), sessions (agent telemetry), twin replay (advanced).
 See FULL-TOUR.md for why.

 Re-run this script any time; it recreates the app cleanly.
 Remove it with:  rm -rf "$TARGET"
DONE
