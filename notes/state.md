# Proofbound — Current State

> THE resume note. Rewrite in place; never append or create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

As of: 2026-10-23

## Resume — 2026-10-23

Exact next action: skip resolved `DEF-P7-002` and dormant `DEF-P7-001`; continue only with separately
authorized non-deferred work.
Do not add runtime fixture acceptance or binary validation. Keep `DEF-P7-001` dormant; do not treat
the pinned Linux/amd64 PostgreSQL 16.4.0 Maven JAR as approved.

Immediate blockers: the complete projector/build manifest, source/toolchain binding, runtime fixture
corpus, no-fetch and cleanup witnesses, L12/L17 runtime result, behavior-lock acceptance, and
independent runtime verdict are absent. `DEF-P7-001` remains owner-postponed because the pinned
PostgreSQL JAR's detached signature did not verify with the public key found; publisher identity and
bundled license/notice coverage remain unverified. `DEF-P7-002` is resolved by the owner-approved
scope decision and independent non-author corpus review. `DEF-P7-001` remains dormant.

## Branch and local ownership

- `main` tracks `origin/main` and is ahead by thirty-one commits after the P7 corpus review verdict;
  push was not attempted or authorized.
- Preserve the unrelated user-owned untracked paths `docs/plans/P5-intent-provenance-plan.md` and
  `kernel/cmd/p6dbprobe/`; do not stage, delete, or relabel them as complete.
- No PostgreSQL binary, database directory, downloaded toolchain, or built executable was added to
  Git. The selected `.mod`/`.zip` corpus is bound under the accepted content-addressed route; nested
  test-data members are covered by the owner-approved scope decision.

## Completed

- P7 documentary planning, readiness-manifest shape, setup-source documentation, companion API/SPEC
  proposal, fixture plan, and review packets have their recorded non-author acceptances where
  applicable.
- The verified-archive preflight/staging seam is implemented and independently accepted. Synthetic
  archive-control tests are implemented and independently accepted with no findings.
- The first non-PostgreSQL Phase-A Go/projector build probe is recorded as explicitly `UNVERIFIED`.
- The selected Go module source corpus is bound: 77 modules, 154 `.mod`/`.zip` artifacts, totaling
  171,215,534 bytes under the accepted content-addressed route. The owner-approved nested payload
  scope is recorded in `docs/decisions/VD-p7-go-module-source-archive-payload-scope-2026-10-23.md`.
- The canonical deferral process remains active: `DEF-P7-001` is postponed and dormant; `DEF-P7-002`
  is resolved with the recorded independent non-author review.

## Open, blocked, and unverified

- Runtime fixtures, real PostgreSQL execution, no-fetch/cleanup witnesses, L12/L17, behavior-lock
  acceptance, clean-clone proof, and the complete build/projector evidence remain unverified.
- PostgreSQL publisher identity, detached-signature resolution, and component license/notice
  coverage remain unresolved under dormant `DEF-P7-001`.
- The 171 MB module corpus is repository-bound and `DEF-P7-002` is resolved by the recorded
  independent non-author review; this does not change any runtime or PostgreSQL blocker.
- The restricted Snap sandbox cannot run the capability-dependent mutation self-test. The pinned
  Go 1.26.7 toolchain and installed linter completed substantive checks outside that restriction.
- Existing process debt remains around fail-closed hook-test behavior and advisory-expiry
  enforcement; neither is being expanded during this closeout.

## Verification

- Targeted checks: link lint, index check, deferral check (`rows=2 open=1`), P6
  artifact-integrity, state freshness, route-matrix freshness, and whitespace checks passed.
- The pinned clean-tree `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH make check` passed
  with `0 issues` outside the restricted Snap sandbox. Expected `index stale; run make index`,
  `INERT cleanroom`, and temporary bundle-verification lines were negative/self-test output, not
  failures. The route matrix was regenerated before the passing run.
- A dated Git bundle fallback is created after this final closeout commit and independently
  verified; its exact path is reported in the closeout handoff.

## Institutionalized improvement

The EOD prompt now requires a queryable deferral row for every newly postponed, blocked, or
owner-pending item, plus a direct P6 artifact-integrity freshness check after documentary verdicts.
The current closeout records the selected module corpus as repository-bound with its independent
review verdict; runtime and PostgreSQL evidence remain explicitly unverified.

## Current checkpoint handoff

- Accepted implementation and synthetic-control evidence remains bounded to preflight/staging and
  archive controls; it is not runtime or PostgreSQL acceptance.
- Push status: not attempted and not authorized; `main` remains thirty-one commits ahead.
- Verification: final pinned full gate passed with `0 issues`; no PostgreSQL or P7 runtime
  acceptance was implied.
- Next action: skip resolved `DEF-P7-002` and dormant `DEF-P7-001`; continue only with separately
  authorized non-deferred work.
