# Proofbound — Current State

> THE resume note. Rewrite in place; never append or create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

As of: 2026-10-24

## Resume — 2026-10-24

Exact next action: remain stopped at the completed `P7-SPEC-ADOPT-1` slice. If P7 resumes, obtain a
separately recorded authorization for the next non-runtime evidence/implementation slice; do not
start runtime fixtures or binary validation, and do not activate `DEF-P7-001`. Continue only with
separately authorized non-deferred work.

Immediate blockers: the complete projector/build manifest, source/toolchain binding, runtime fixture
corpus, no-fetch and cleanup witnesses, L12/L17 runtime result, behavior-lock acceptance, clean-clone
proof, and independent runtime verdict are absent. `DEF-P7-001` remains owner-postponed because the
pinned PostgreSQL JAR's detached signature did not verify with the public key found; publisher
identity and bundled license/notice coverage remain unverified. `DEF-P7-001` is dormant; `DEF-P7-002`
is resolved.

Do not treat the accepted live SPEC/API wording as runtime acceptance. Do not treat the pinned
Linux/amd64 PostgreSQL 16.4.0 Maven JAR as approved. The accepted P7-SPEC-ADOPT-1 slice changes
only the live SPEC/API wording and preserves the future runtime gate.

## Branch and local ownership

- `main` tracks `origin/main`; local commits are ahead and push was not attempted or authorized.
- Preserve the unrelated user-owned untracked paths `docs/plans/P5-intent-provenance-plan.md` and
  `kernel/cmd/p6dbprobe/main.go`; do not stage, delete, or relabel them as complete.
- No PostgreSQL binary, database directory, downloaded toolchain, built executable, runtime fixture,
  or binary-validation artifact was added to Git.

## Completed

- The P7 companion API/SPEC documentary boundary received three independent review rounds; the
  amended boundary was accepted in
  `docs/verification/verdicts/p7-replay-lock-companion-api-round3-20261024.md`.
- Owner authorization `P7-SPEC-ADOPT-1` was recorded in
  `docs/intent/records/change-intents/CI-p7-spec-adopt-1a2b3c/0001.md`; the live twin SPEC now
  adopts the staged archive/inventory/executable-binding boundary and explicitly marks launched
  executable observation as future runtime evidence.
- The first SPEC-adoption review rejected an invalid native intent self-digest; the digest was
  corrected and the fresh independent review accepted the exact corrected commit in
  `docs/verification/verdicts/p7-spec-adopt-round2-20261024.md`.
- The verified-archive preflight/staging seam, synthetic archive controls, and selected Go module
  corpus remain independently accepted within their recorded scopes. `DEF-P7-002` remains resolved.
- The artifact-integrity summary was regenerated after today’s documentary verdicts.

## Open, blocked, and unverified

- Runtime fixtures, real PostgreSQL execution, process/network/no-fetch/cleanup witnesses, L12/L17,
  behavior-lock acceptance, clean-clone proof, complete projector/build evidence, and an independent
  runtime verdict remain unverified.
- PostgreSQL publisher identity, detached-signature resolution, and component license/notice
  coverage remain unresolved under dormant `DEF-P7-001`.
- The restricted Snap sandbox cannot run the capability-dependent mutation self-test. This remains
  an environment limitation, not acceptance evidence.
- Existing process debt around fail-closed hook-test behavior and advisory-expiry enforcement is
  unchanged and not expanded during this closeout.

## Verification

- `scripts/p6-artifact-integrity.sh --check` passes after refreshing its generated documentary count.
- The native `P7-SPEC-ADOPT-1` intent self-digest now reproduces exactly under the repository’s
  normalization rule; link lint and diff whitespace checks passed.
- The authoritative pinned bare `make check` result and dated bundle verification are recorded in
  the final closeout journal after this state refresh.
- No push was attempted or authorized.

## Institutionalized improvement

- After adding documentary verdicts, run `scripts/p6-artifact-integrity.sh --check`; use `--write`
  only for the expected generated-count drift, then recheck before the closeout commit.
- For every native intent record, verify the declared self-digest with the repository’s exact
  zero-field normalization rule before staging. Do not use an ad-hoc shell backreference for this
  calculation; today’s independent rejection caught that failure mode.
- Keep the exact next action and immediate blockers in this first resume section, and distinguish
  expected negative-test output from actual gate failures.

## Current checkpoint handoff

- `P7-SPEC-ADOPT-1` is complete and independently accepted; it is live SPEC/API adoption only.
- The P7 documentary and preflight/staging acceptances do not establish runtime or PostgreSQL
  acceptance.
- `DEF-P7-001` is still dormant and the pinned PostgreSQL JAR remains unapproved; `DEF-P7-002` is
  resolved.
- Tomorrow’s first action: remain at this stop point and select work only from a separately recorded
  authorization, with runtime/binary work excluded unless explicitly authorized.
