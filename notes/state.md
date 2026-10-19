# Proofbound — Current State

> THE resume note. Rewrite in place; never append or create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

As of: 2026-10-19

## Resume — 2026-10-19

Exact next action: obtain an independent non-author verdict against
`docs/verification/p7-replay-lock-companion-archive-controls-review-packet-20261019.md`, bound to
implementation commit `bd5459592cc2ea53a95845a0c61fc8bb6f5e3e93`. Do not add runtime fixture
acceptance or binary validation. Keep
deferral `DEF-P7-001` dormant: do not treat the pinned Linux/amd64 PostgreSQL 16.4.0 Maven JAR as
approved, and do not resume its provenance/licensing go/no-go until the owner explicitly says they
are ready, asks to resume it, or names it in a work request. Do not ask, remind, or pressure the
owner about any open deferral; skip all open deferrals until explicitly activated.

Immediate blockers: the setup-source manifest and README have fresh non-author
documentary `ACCEPT` for exact commit `d07f3ffee70d1acd5f917d4fa06362aaac594718`
([verdict](../docs/verification/verdicts/p7-replay-setup-sources-round2-20261016.md)), but
the pinned PostgreSQL JAR's detached signature did not verify with the public key found;
publisher identity and bundled license/notice coverage remain unverified. There is no
case-manifest binding, projector/build manifest, complete Go module byte corpus, runtime fixture
corpus, L12/L17 runtime result, or runtime acceptance. The full `make check`
gate has passed all substantive checks outside the restricted sandbox; the latest attempt stopped
only at stale durable state before this rewrite. The pinned Go 1.26.7 toolchain is available;
the default sandbox remains capability-limited. The provenance
and licensing check is intentionally deferred, not accepted or passed.

## Branch and local ownership

- `main` tracks `origin/main`. The current implementation checkpoint is commit
  `849b1e7` (`chore: refresh P6 route matrix after lint fix`), and the acceptance checkpoint is
  `bd54595`; `main` is ahead of `origin/main` by twenty-four commits. Push was not attempted: it is not authorized and the
  remote was not DNS-reachable from this environment. Tracked files are clean after commit; two
  unrelated user-owned untracked paths keep the worktree non-clean.
- Preserve the unrelated untracked `docs/plans/P5-intent-provenance-plan.md`
  (SHA-256 `eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439`)
  and `kernel/cmd/p6dbprobe/main.go`
  (SHA-256 `f2a747ed588141723ce8f0932aa07814fcca63f9901b7dca0d71119b93ee15a0`).
  They are not part of this checkpoint; do not delete, stage, or relabel them as complete.
- No database directory, downloaded toolchain/Postgres distribution, or built executable
  is to be committed. Reviewed `.mod`/`.zip` module source bytes remain eligible under
  the accepted external-prerequisites amendment.

## Completed

- The P7 replay-lock documentary plan, addendum, fixture matrix, bounded first-slice scope,
  and external-prerequisites overlay have their recorded non-author planning verdicts.
  None is a product implementation or behavior-lock acceptance.
- The proposed two-input canonical setup manifest and human README were amended after
  round-1 `AMEND` and accepted by a different non-author reviewer in round 2 for their
  exact 2026-10-16 bytes. The manifest's SHA-256 is
  `e312bf28cea20d516e3b21704a53ba3ebc8ac84ff988eccb7da555317b900b33`.
  The verdict did not re-run live downloads or authenticate publishers.
- On 2026-10-17, regenerated the derived P6 artifact-integrity packet after the two new
  documentary verdicts; its direct check now passes. Added the direct check to the canonical
  EOD prompt as a lightweight repeatable closeout step. A fail-closed hook-loop repair was
  trialed but withdrawn when the intent-delivery gate required a relevant `Intent:` claim;
  no product code or runtime fixture was changed.
- On 2026-10-17, aligned `scripts/p6-task1-close.sh` with the current canonical README status
  wording. The targeted P6 closure test, deferral-register check, and artifact-integrity check
  pass. The ignored `notes/overall-status.md` was refreshed for the same closeout.
- On 2026-10-17, made open deferrals dormant by policy: preserve and query them, but do not ask,
  remind, or resume them until the owner explicitly activates one.
- On 2026-10-17, added the planning-only P7 readiness-manifest draft and linked it from the
  roadmap. It organizes the future case/evidence index without authorizing code, fixtures, or
  binary validation.
- On 2026-10-17, settled the planning shape in
  `docs/decisions/VD-p7-readiness-manifest-2026-10-17.md`: exact top-level fields, no-null rules,
  derived case status, outcome vocabulary, content references, and explicit pending review state.
  Fresh non-author review was subsequently completed with `ACCEPT` for documentary planning only.
- On 2026-10-17, froze commit `334b33822cfe2cf35071ee936f260ed079b8a6c1` into
  `docs/verification/p7-readiness-manifest-review-packet-20261017.md` with the exact corpus
  digests and review questions. The packet is a review request, not a verdict.
- On 2026-10-19, recorded Srini K's independent non-author `ACCEPT` against the exact planning
  commit in `docs/verification/verdicts/p7-readiness-manifest-round1-20261019.md`. This accepts
  documentary planning only; implementation remains separately gated.
- On 2026-10-19, prepared the proposed documentation-only companion API/SPEC amendment and
  fixture plan for separate review. The proposal makes archive-byte binding, safe staging,
  `BinariesDir` routing, no-fetch behavior, cleanup, ledger preservation, and negative controls
  explicit without changing the live API or authorizing implementation.
- On 2026-10-19, froze the exact planning corpus in
  `docs/verification/p7-replay-lock-companion-review-packet-20261019.md`. The packet SHA-256 is
  `5bbfea2a51393c20e27a052a5e6761bd2d8419c9320b509ce4d60596b2a6c5ff`; no verdict has been
  recorded.
- On 2026-10-19, recorded Srini K's independent non-author `ACCEPT` with no findings for the
  exact documentary packet and commit. That acceptance remained planning-only; no live contract,
  runtime fixture, binary validation, or implementation was authorized by it.
- On 2026-10-19, the owner authorized implementation of the accepted narrow companion slice.
  The live `twin` SPEC and code now contain only the verified-archive preflight/staging seam;
  runtime fixtures and binary provenance/licensing remain pending.
- On 2026-10-19, implemented `ReplayIsolatedWithVerifiedPostgres` with exact archive-byte and
  inventory binding, safe regular-file extraction, executable checks, cancellation-aware staging,
  `BinariesDir` routing, cleanup through the existing isolated replay path, and focused negative
  tests. The relevant intent is `CI-p7-companion-1a2b3c`; the implementation checkpoint is
  `849b1e7` after generated route-matrix refresh.
- On 2026-10-19, froze the independent implementation-review packet at
  `docs/verification/p7-replay-lock-companion-implementation-review-packet-20261019.md` with
  SHA-256 `2d89d69f4dfa81a5d6e12249aa66228c238fb120cb1dbf4bcf571ef68f44f900`.
- On 2026-10-19, recorded Srini K's independent non-author implementation `ACCEPT` with no
  findings for the verified-archive preflight/staging slice. Runtime fixtures, PostgreSQL
  execution, behavior-lock evidence, and binary provenance/licensing remain separately gated.
- On 2026-10-19, added synthetic archive-control coverage for inventory ordering, duplicate and
  mode/required-file rejection, unsafe paths, staged-file drift, and cancellation before staging.
  The live SPEC and planning/prerequisite documents now distinguish this tested preflight boundary
  from still-unverified runtime evidence. The pinned full `make check` passed with `0 issues`.
- On 2026-10-19, froze the independent review packet at
  `docs/verification/p7-replay-lock-companion-archive-controls-review-packet-20261019.md` for
  implementation commit `bd5459592cc2ea53a95845a0c61fc8bb6f5e3e93`. No verdict is recorded yet.

## Open and unverified

- Publisher signing identity, component license/notice and distribution clearance, exact
  Go toolchain patch suitability, complete Phase-A projector/module/toolchain/executable
  evidence, source-manifest binding in a case manifest, runtime fixture corpus, zero-fetch and
  cleanup witnesses, L12/L17 `PASS`, requirement satisfaction, and
  runtime acceptance and the remaining implementation evidence are open. The old empty-cache/no-network
  clean-clone claim remains ineligible for `PASS` under the accepted overlay.
- Deferral `DEF-P7-001` covers the third-party binary provenance and licensing go/no-go. Its open
  findings remain unresolved and must not be relabeled `PASS` or `NO-GO` until the check
  is resumed or an alternate source is explicitly selected.
- The default restricted sandbox's Snap Go lacks `cap_dac_override`; its mutation self-test
  remains unavailable there. Outside that restriction, the repository-pinned Go 1.26.7
  toolchain was available through the Go toolchain cache and, with `/home/thamm/go/bin` on
  `PATH`, the installed `golangci-lint` completed successfully. No PostgreSQL validation was
  performed.
  Native Windows acceptance and P7 mutation detector power are also unverified.
- Process debt: the gate registry describes advisory-expiry enforcement, but the inspected
  `scripts/link-lint.sh` checks decision-index membership only; the advisory rows dated
  2026-10-16 need a separate policy decision and a real enforcement check. Do not treat a
  green link-lint as proof of advisory expiry compliance.
- The existing `make hooks-test` loop can mask an intermediate failing self-test by returning
  the last script's status. A fail-closed repair needs a relevant accepted change intent
  before a governed Makefile change is committed. Do not cite the unrelated P5 intent.
- The P6 closure checker had retained historical README wording and failed against the aligned
  roadmap text; this is repaired and verified. Keep closure assertions tied to canonical current
  status vocabulary when roadmap wording changes.

## Verification

- The final bare `make check` on 2026-10-17 passed the deferral, generated/documentary,
  P6 closure, package, and hostile-fixture checks, then failed at live `kernel-check`
  because the installed Snap Go runtime lacks `cap_dac_override`. `index stale; run make index`
  was an expected negative-test message, not an actual stale decision index; temporary bundle
  verification lines were also self-test output, not a failure.
- Direct `scripts/p6-artifact-integrity.sh --check` found 19 versus 21 documentary
  verdicts; `--write` regenerated the derived packet and direct `--check` now passes.
  `make index-check` and `scripts/invariant-lint.sh` pass.
- A temporary fail-closed loop control, mutation-harness self-test, and full hook suite
  passed outside Snap confinement. That loop change was withdrawn; the trial checkpoint's
  bare `make check` had stopped at the delivery self-test because its Makefile change
  lacked a relevant `Intent:` trailer. After the documentation-only correction, bare
  `make check` passed generated/documentary checks but stopped at sandbox Snap Go.
  Outside confinement, a first bare run stopped at missing linter `PATH`; with the
  installed linter on `PATH`, bare `make check` reached lint and failed on a Go
  1.27-versus-1.26 loader panic. That old compatibility result is superseded by the pinned
  toolchain rerun below. `INERT cleanroom: no readable external pattern file` leaves that
  control unproved.

- Latest environment rerun through commit `849b1e7`:
  `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH make check` passed all substantive checks
  outside the restricted sandbox and reached only the stale-state freshness gate. The plain
  restricted-sandbox invocation remains unavailable only at the Snap capability-dependent
  mutation self-test; this is an execution-environment distinction, not a P7 or PostgreSQL
  acceptance result.

## Institutionalized improvement

The canonical EOD prompt now requires a direct P6 artifact-integrity freshness check after
documentary verdicts and a queryable deferral-register update. The P6 closure assertion was
aligned to canonical README status wording after the full gate exposed the stale expectation.
These are lightweight closeout controls, not a claim that the existing hook-test loop is
fail-closed. Preserve reviewed setup bytes; record changing acceptance status in verdicts and
this state note instead of silently changing reviewed content.

## Current checkpoint handoff

- Current accepted implementation checkpoint: `c5f65a4`; independent implementation verdict is
  recorded at `docs/verification/verdicts/p7-replay-lock-companion-implementation-round1-20261019.md`.
  The synthetic archive-control expansion is implementation commit `bd54595`, pending independent
  review.
- Push status: not attempted; not authorized, and GitHub DNS was unavailable from this
  environment. `main` is ahead of `origin/main` by twenty-four commits.
- Verification: the pinned full `make check` passed with `golangci-lint` reporting `0 issues`;
  expected `index stale; run make index` output was a negative-test message. Focused archive
  control tests pass outside the restricted Snap sandbox.
  The restricted sandbox remains capability-limited; expected negative-test output was not a
  live failure.
- Existing bundle fallback remains the prior closeout bundle; no new EOD bundle was created for
  this mid-day planning checkpoint.
- Next action: obtain the independent implementation verdict against the frozen synthetic
  archive-control review packet. Keep `DEF-P7-001` dormant; do not perform runtime binary
  validation or ask about any open deferral.
