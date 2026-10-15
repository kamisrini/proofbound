# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads CLAUDE.md, then this file, before acting.

As of: 2026-10-15

## Resume — 2026-10-15

Exact next action: obtain owner scope approval for
`docs/plans/P7-replay-lock-first-implementation-slice-proposal-draft1.md`, then obtain a fresh
non-author review of that exact proposal and record `ACCEPT`, `AMEND`, or `REJECT` before any
product code. The proposed first slice is the offline Postgres `BinariesDir` provisioning seam
and L17 evidence, not a full L12 pass. Keep behavior-lock, regeneration, snapshot-provider,
and external-feed implementation gated.

Immediate blockers: documentary planning was accepted by fresh round-4 non-author review, but
runtime fixtures, the L12 offline replay result, first-slice scope approval/review,
implementation authorization and verdict, and
requirement satisfaction are absent. The full gate cannot complete because the installed linter
crashes on a Go 1.27 version marker. Push
authorization/authentication remain unverified after the prior GitHub DNS failure.

## Branch and repository status

- Branch: `main`, tracking `origin/main`, ahead of the locally known remote-tracking ref.
  The remote tip is unverified.
- Preserved untracked user artifacts: `docs/plans/P5-intent-provenance-plan.md` and
  `kernel/cmd/p6dbprobe/main.go`. They are untouched and excluded from this checkpoint.
- The P7 addendum and fixture matrix are accepted for documentary planning only; no runtime
  fixture, product implementation, or independent implementation verdict is accepted.
- A bounded first implementation slice proposal is drafted but not yet owner-approved or
  independently reviewed. It proposes only the offline Postgres provisioning seam and L17
  evidence; L12 and the full lock evaluator remain outside this first slice.

## Completed

- P6 consolidation and the P7 snapshot-provider skip remain accepted.
- Fresh round-3 non-author review accepted the amended P7 `ReplayIsolated` planning subject and
  contract; the verdict is `docs/verification/verdicts/p7-behavior-lock-plan-round3-20261013.md`.
- On 2026-10-14, the owner authorized drafting the lock-schema/addendum and fixture-planning slice.
  Drafts are `docs/plans/P7-replay-lock-spec-addendum-draft1.md` and
  `docs/plans/P7-replay-lock-fixture-matrix-draft1.md`. No product code, provider, feed, wire
  identity, event kind, or persistence surface changed.
- The owner requested a recommendation for the pilot's intent binding. The recommended path is a
  genuine P5 requirement, retaining `UNDECLARED` as a negative control. The owner authorized a
  draft proposal, then approved its business outcome and `O-1` wording and selected existing
  declared roles. Formal records are `BD-replay-evidence-7a91c4` and
  `BR-replay-evidence-6e3d2b/0001`; the active requirement binds the exact decision digest.
- Fresh non-author reviewer `Fermat` judged the exact active P5 `O-1` `VERIFIABLE`; the P7 addendum
  and fixture planning slice received `AMEND` with one high and three medium findings. The
  requirement review and documentary slice verdict are recorded under `docs/verification/verdicts/`.
- Amended the addendum and fixture matrix for all four round-1 findings: complete event evidence
  beyond P4's source digest, whole-ledger before/after witnesses on success and failure, pinned
  projector/build/runtime and proof evidence encodings with a byte-exact proof control, and
  independent candidate payload-hash verification. These are proposed contracts, not accepted
  behavior or runtime fixtures.
- Fresh non-author reviewer `Mendel` reviewed exact commit `730a474` and returned `AMEND` for the
  later addendum and fixture slice. The three round-1 source/preservation/candidate findings and
  proof encoding were judged resolved in the proposal; snapshot-map round-trip and executable/
  dependency byte identity remain blocking. The accepted round-3 P7 subject is unaffected.
- On owner approval, amended those two round-2 findings: snapshot evidence preserves exactly the
  keys present in `Snapshot.Tables`, with empty-map and injected-key controls; the build manifest
  now binds retrievable `.mod`, `.zip`, and exact executable bytes and rejects missing or changed
  inputs. These are proposed contracts, not runtime fixtures or implementation acceptance.
- Fresh non-author reviewer `Confucius` reviewed exact commit `16c6189` and returned `AMEND` for
  the later addendum and fixture slice. The round-2 snapshot and executable/module-byte findings
  were judged resolved in the proposal; committed Go toolchain and Postgres bytes and exact
  nested toolchain manifest encoding remain blocking. The accepted P7 subject is unaffected.
- On 2026-10-15, amended the two round-3 findings in the proposed documents: the nested Go
  toolchain manifest has an exact canonical JSON shape, and both the Go and embedded-Postgres
  archives require verified Git-tracked bytes in the reviewed evidence revision. L12 remains
  `UNVERIFIED` until a later reviewed implementation can supply the Postgres archive to P4's
  temporary store offline. No archives, runtime fixtures, or product code were created.
- Fresh non-author reviewer `Gibbs` reviewed exact commit `d37a874` and returned `ACCEPT` for the
  documentary addendum and fixture matrix, with no blocking finding. The reviewer noted a low
  implementation caution: verified Postgres binaries must go through `BinariesDir`, not P4's
  cleared `RuntimeDir`. The acceptance does not extend to L12, a runtime corpus, requirement
  satisfaction, full-gate success, or implementation.
- Drafted `docs/plans/P7-replay-lock-first-implementation-slice-proposal-draft1.md` to present
  a bounded offline Postgres seam, L17 fixture evidence, archive-size/provenance go/no-go,
  and explicit stop points. This is a proposal only; no product code or runtime corpus changed.
- The generated P6 requirement-review packet now includes both active requirements, three active
  obligations, and two exact reviews. The generated P6 artifact-integrity packet now enumerates
  26 schema-bearing verdicts and 15 documentary verdict artifacts after the round-4 verdict.

## Verification

- Bare `make check` ran on 2026-10-14. The `index stale; run make index` output was an expected
  negative control; generated freshness passed. The hook loop's exit status is not sufficient to
  claim every self-test passed, as discovered later. The actual stop was
  `commit-cadence: dirty worktree outlived 90m commit window`, so the main DAG did not reach
  kernel checks. A Snap Go confinement diagnostic in a self-test remains an environment issue.
- After documentation commit `7ca242b`, bare `make check` was rerun. Cadence and the repository,
  generated freshness, P6 package/verdict, and Windows contract checks passed. It reached
  `kernel-check` and stopped because this host's Snap Go confinement lacks
  `cap_dac_override` (`Function not implemented`). Go build/tests/lint remain unverified here;
  no source or generated-artifact failure was observed.
- After formal P5 record commit `9e0621f`, both normalized self-digests and the exact decision
  reference were verified; JSON metadata parses. The P5 records package tests passed with Go
  outside the sandbox. Bare `make check` passed repository checks and Go build/tests, then stopped
  because `golangci-lint` was absent from PATH. With its installed directory on PATH, the gate
  again passed repository checks and Go build/tests, then the linter panicked: a file requires
  Go 1.27 while the installed linter was built with Go 1.26.7. Lint remains unverified; this is
  a tooling failure, not evidence of a source defect.
- For the amended P7 draft, `git diff --check`, `make link-lint`, `make index-check`,
  `scripts/p6-task5-review.sh --check`, and `scripts/p6-artifact-integrity.sh --check` passed.
  The proof encoding control hashed to its documented SHA-256. After commit `210830e`, bare
  `make check` passed repository/generated checks but stopped at `kernel-check` because sandboxed
  Snap Go lacks `cap_dac_override`. With the installed linter on PATH outside that confinement,
  the same gate passed repository checks and Go build/tests, then the linter panicked on a file
  marked Go 1.27 while it was built with Go 1.26.7. The `index stale` line was an expected
  negative-test message, not a freshness failure. Full lint remains unverified.
- For the round-2 amendments, link, index, P6 artifact-integrity, exact requirement-review,
  state-freshness, and whitespace checks passed; both snapshot byte-control hashes reproduced.
  The canonical manifest shape parsed and its key order matched canonical ordering. Bare
  `make check` again passed repository and generated checks but stopped at `kernel-check` because
  sandboxed Snap Go lacks `cap_dac_override`. The `index stale` message was the expected negative
  control. No generated artifact needed regeneration. Go build/tests/lint were not rerun outside
  confinement for this documentation-only edit; the earlier compatible-toolchain lint blocker
  remains unresolved.
- After recording the round-3 review, link, index, P6 artifact-integrity, exact
  requirement-review, census, state-freshness, and whitespace checks passed. The derived
  artifact-integrity packet was regenerated for the new documentary verdict. Bare `make check`
  was not rerun for this review-record-only checkpoint; its prior Snap Go/linter blockers remain.
- On 2026-10-15, the round-3 planning amendment passed link, index, P6 artifact-integrity,
  exact requirement-review, state-freshness, and whitespace checks. The main manifest example
  parsed with canonical key order. Bare `make check` after documentation commit `e4ebf73`
  passed repository and generated-freshness checks, then stopped at `kernel-check` because the
  sandboxed Snap Go lacks `cap_dac_override`. `index stale` was an expected negative-test
  message; no generated artifact was stale. Go build/tests/lint were not rerun outside the
  sandbox for this documentation-only change; compatible-toolchain lint remains unverified.
- After recording the round-4 documentary acceptance, link, index, P6 artifact-integrity,
  exact requirement-review, census, state-freshness, and whitespace checks passed. The derived
  artifact-integrity packet was regenerated for the new verdict. Bare `make check` was not
  rerun for this review-record-only checkpoint; the earlier same-day Snap Go and known linter
  blockers remain unresolved.

## Open, blocked, and unverified

- Open: obtain owner approval and fresh non-author review of the proposed bounded first slice;
  if accepted, resolve archive budget and exact SPEC/API before implementation. Later build
  committed runtime fixtures and evidence manifest, demonstrate L12 offline replay, and obtain
  a separate non-author review of any implementation. The `hooks-test` Makefile loop can
  mask a failing intermediate self-test because it returns only the last test's status; repair
  this independently of the P7 contract work.
- Blocked: behavior-lock, regeneration, snapshot-provider, and external-feed implementation until
  the applicable authorization, runtime evidence, and independent implementation gate are
  satisfied. Push remains unverified.
- Unverified: satisfaction of the new requirement, complete clean-clone replay, detector power
  for a P7 mutation set, lint with a compatible toolchain, native Windows
  acceptance, and remote-tip freshness.

## Institutionalized improvement

When preserved user-owned untracked artifacts trigger cadence, keep them untouched, record their
ownership, commit only a coherent documentation checkpoint, and rerun the full gate. For replay
evidence, check both complete event fields and complete-ledger preservation independently of
existing P4 digests. Fixture matrices must distinguish absent map keys from present empty values
and bind build inputs separately from source revision. An offline replay claim requires both an
exact digest and retrievable bytes for each runtime input. No new automation is justified.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, pinned vectors, and
  historical artifacts.
- Planning acceptance and draft review do not authorize behavior-lock or regeneration code.
