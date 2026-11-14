#!/usr/bin/env bash
# Proofbound — FULL VALUE TOUR on a sample project ("ledgerloom").
#
# Runs every Release-1 capability you can show on an ordinary application repo, end to end, with
# plain-English narration: the permanent trust record, decisions bound to commits, the full
# decision->requirement->change->commit intent chain, independent review verdicts (including a
# red-verdict chain: a fix that introduced a new defect), and deterministic rebuild.
#
# PREREQUISITES
#   - `proofbound` on your PATH (build it per DEMO.md step 1), or set PROOFBOUND_BIN=/path/to/proofbound
#   - Git and Go 1.26+; outbound internet on first run (downloads the embedded PostgreSQL once)
#   - RUN THIS ON A LINUX FILESYSTEM (WSL home dir), NOT a Windows /mnt/d path — the embedded
#     PostgreSQL needs real Linux file semantics.
#
# It is self-contained and idempotent: it (re)creates its sample repo under $DEMO_DIR and exits clean.
set -euo pipefail

BIN="${PROOFBOUND_BIN:-proofbound}"
command -v "$BIN" >/dev/null 2>&1 || { echo "ERROR: '$BIN' not found. Build it (DEMO.md step 1) or set PROOFBOUND_BIN." >&2; exit 1; }
DEMO_DIR="${DEMO_DIR:-$HOME/pb-full-tour}"
case "$DEMO_DIR" in /mnt/*) echo "WARNING: $DEMO_DIR is on a Windows drive; the embedded PostgreSQL may fail. Use a Linux path." >&2;; esac

# Hide the embedded PostgreSQL's startup/shutdown + one-time initdb log noise so only Proofbound's
# lines show. (initdb runs once, on the very first command, while it builds ./.proofbound.)
PGNOISE='LOG:|server (started|stopped)|listening|database system|starting up|shutting down|background|pg_|checkpoint|redo|MultiXact|autovacuum|FATAL|terminating|This user must|database cluster|locale|encoding|initialized|^The |^Data page|creating (directory|subdirectories|configuration)|selecting (dynamic|default)|running bootstrap|performing post-bootstrap|syncing data|Success\.|You can now|^ done$|^$'
pb() { "$BIN" "$@" 2>&1 | grep -vE "$PGNOISE" || true; }
say() { printf '\n\033[1m== %s\033[0m\n' "$*"; }
note() { printf '   %s\n' "$*"; }
zeros=$(printf '0%.0s' {1..64})

rm -rf "$DEMO_DIR"; mkdir -p "$DEMO_DIR"; cd "$DEMO_DIR"
git init -q; git config user.email dev@example.com; git config user.name "Sample Dev"
printf '.proofbound/\n' > .gitignore
printf 'module example.com/ledgerloom\ngo 1.21\n' > go.mod

cat <<'INTRO'
========================================================================
 Proofbound — full value tour on a sample project ("ledgerloom")
========================================================================
We build a tiny invoice-tax library through a realistic history — a
decision, a requirement, a change that implements it, two independent
reviews that catch defects — and then read it back as a durable trust
record. Proof IDs (proof=01M...) are time-based and will differ per run.
INTRO

# ---------------------------------------------------------------------------------------------------
say "1. A first commit, then record the history"
printf 'package ledgerloom\n\n// Tax returns cents of tax for an amount of cents at a whole-percent rate.\nfunc Tax(amountCents, ratePct int) int { return amountCents * ratePct / 100 }\n' > tax.go
git add -A; git commit -q -m "feat: tax computes whole-percent tax"
note "proofbound sync git  — turn git history into permanent, content-addressed records:"
pb sync git | sed 's/^/   /'

# ---------------------------------------------------------------------------------------------------
say "2. An intent chain: decision -> requirement -> change -> the commit that implements it"
note "These three records state WHY the work exists and WHAT it must guarantee. Each is"
note "content-addressed; each child cites its parent's exact digest (a tamper-evident chain)."
mkrec() { mkdir -p "$(dirname "$1")"; { printf '```proofbound-json\n%s\n```\n\n%s\n' "$2" "$3"; } > "$1"; }
digest_in_place() { local f=$1 d; d=$(sha256sum "$f" | awk '{print $1}'); sed -i "s/$zeros/$d/" "$f"; echo "$d"; }

BD=BD-charge-invoice-tax-7a1b2c; BDp=docs/intent/records/business-decisions/$BD/0001.md
mkrec "$BDp" '{"schema":"proofbound.business-decision.v1","decision_id":"'$BD'","status":"accepted","outcome":"Charge sales tax on every invoice line.","declared_owner":"product-lead","declared_approver":"product-lead","decided_at":"2026-11-10T00:00:00Z","supersedes":[],"artifact_path":"'$BDp'","artifact_sha256":"'$zeros'"}' "# $BD — charge invoice tax (business decision)"
BDd=$(digest_in_place "$BDp")

BR=BR-tax-never-negative-4d5e6f; BRp=docs/intent/records/requirements/$BR/0001.md
mkrec "$BRp" '{"schema":"proofbound.requirement.v1","requirement_id":"'$BR'","status":"proposed","authorized_by":[{"record_kind":"business_decision","source":"intent.records","record_id":"'$BD'","artifact_sha256":"'$BDd'","relation":"authorizes"}],"declared_owner":"product-lead","obligations":[{"id":"O-1","statement":"Computed tax must never be negative for any invoice line.","state":"active"}],"supersedes":[],"artifact_path":"'$BRp'","artifact_sha256":"'$zeros'"}' "# $BR — tax never negative (requirement, obligation O-1)"
BRd=$(digest_in_place "$BRp")

CI=CI-clamp-negative-tax-8c9d0e; CIp=docs/intent/records/change-intents/$CI/0001.md
mkrec "$CIp" '{"schema":"proofbound.change-intent.v1","intent_id":"'$CI'","status":"accepted","declared_sponsor":"product-lead","targets":[{"record_kind":"requirement","source":"intent.records","record_id":"'$BR'","artifact_sha256":"'$BRd'","relation":"implements","obligation_ids":["O-1"]}],"constrained_by":[],"supersedes":[],"artifact_path":"'$CIp'","artifact_sha256":"'$zeros'"}' "# $CI — clamp negative tax (change intent)"
CId=$(digest_in_place "$CIp")
git add -A; git commit -q -m "docs(intent): charge-tax decision chain (BD/BR/CI)"

# The implementing commit cites the change intent via a standalone Git trailer.
printf 'package ledgerloom\n\n// Tax returns cents of tax; never negative (clamped).\nfunc Tax(amountCents, ratePct int) int {\n\tt := amountCents * ratePct / 100\n\tif t < 0 {\n\t\treturn 0\n\t}\n\treturn t\n}\n' > tax.go
git commit -q -am "fix(tax): clamp negative tax to zero

Clamp computed tax to zero so an invoice line can never show negative tax.

Intent: $CI"
IMPL=$(git rev-parse HEAD)
# ORDER MATTERS: intent records must enter the ledger BEFORE the commit that cites them, or the
# reference reads as "dangling". Always sync intent before sync git.
pb sync intent all >/dev/null; pb sync git >/dev/null

note "proofbound intent check --commit <impl>  — is this commit bound to an authorized intent?"
pb intent check --commit "$IMPL" | sed 's/^/   /'
note ""
note "proofbound report intent $CI  — the whole chain, with proof IDs and HONEST status:"
pb report intent "$CI" | sed 's/^/   /'
note "(Notice review=UNREVIEWED, outcome=UNVERIFIED, deployment=missing — Proofbound separates"
note " 'implemented' from 'verified'. The next section adds the missing reviews.)"

# ---------------------------------------------------------------------------------------------------
say "3. Independent reviews that catch defects — including a fix that introduced a new one"
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
# A buggy "rounding" change + a NEEDS_WORK review of it, synced. Then a fix + a NEEDS_WORK review of
# the FIX, synced. Interleaving the syncs puts them in ledger order red -> change -> red, which is
# exactly what a red-verdict CHAIN is (a defect introduced by a fix, caught by the next review).
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
note "Both reviews are now in the ledger; one reviewed a fix that introduced a NEW defect."

# ---------------------------------------------------------------------------------------------------
say "4. A simple decision bound to the commit that implements it"
mkdir -p docs/decisions
printf '# VD-rounding-policy-3f2a1b — Round tax half-up\n\nStatus: accepted\nDate: 2026-11-12\n\nTax rounds half-up to the nearest cent.\n' > docs/decisions/VD-rounding-policy-3f2a1b.md
git add -A; git commit -q -m "docs: record rounding-policy decision"
printf 'package ledgerloom\n\n// Tax rounds half up (overflow-safe).\nfunc Tax(amountCents, ratePct int) int {\n\tt := int((int64(amountCents)*int64(ratePct) + 50) / 100)\n\tif t < 0 {\n\t\treturn 0\n\t}\n\treturn t\n}\n' > tax.go
git commit -q -am "fix(tax): overflow-safe half-up rounding

Implements the rounding policy decided in VD-rounding-policy-3f2a1b."
pb sync git >/dev/null

# ---------------------------------------------------------------------------------------------------
say "5. Read the whole trust record"
note "proofbound report week  — commits with proof IDs, a bound decision, and the reviews:"
pb report week | grep -E '^(week|commits|- |checks|sessions|reviews|red_)' | sed 's/^/   /'
note ""
note "Look for:  decisions=VD-rounding-policy-3f2a1b on the last commit (the WHY, bound to the WHAT);"
note "           reviews count=2 (what was FOUND, not just what ran);"
note "           red_verdict_chains count=1 (a fix that introduced a new defect — Proofbound's whole"
note "           reason for preferring rebuild-from-rules over patching)."

# ---------------------------------------------------------------------------------------------------
say "6. The record is reproducible, not a snapshot"
B=$(pb report week | grep -E '^- .*whole-percent' | grep -oE 'proof=[0-9A-Z]+')
pb rebuild >/dev/null
A=$(pb report week | grep -E '^- .*whole-percent' | grep -oE 'proof=[0-9A-Z]+')
note "Dropped all projections and rebuilt from the raw ledger."
note "First commit's proof ID  before: $B"
note "                          after:  $A"
if [ "$B" = "$A" ]; then note "IDENTICAL — the trust record rebuilds exactly. This is the point."; else note "CHANGED — unexpected."; fi

cat <<OUTRO

========================================================================
 What a sample project does NOT show (by design)
========================================================================
 - proofbound verify   : checks Proofbound's OWN delivery boundary; on an app
                         repo it says "no make check witness found in spool" —
                         expected. Your trust surface is sync + report + rebuild.
 - proofbound gates ... : evaluate rules against recorded CI/check witnesses,
                         which a plain app repo doesn't produce. Needs check-run data.
 - proofbound sync github / report github : need a real repo + GITHUB_TOKEN.
 - proofbound twin replay : the advanced replay engine (spins a private PostgreSQL).

 Sample repo left at: $DEMO_DIR   (delete with: rm -rf "$DEMO_DIR")
OUTRO
