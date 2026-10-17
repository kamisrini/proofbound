# Proofbound — Current State

> THE resume note. Rewrite in place; never append or create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

As of: 2026-10-17

## Resume — 2026-10-17

Exact next action: perform a read-only provenance go/no-go for the pinned Linux/amd64
PostgreSQL 16.4.0 Maven JAR. Seek a publisher-authenticated signing key or equivalent
artifact provenance and a component-license/notice inventory for its bundled libraries;
if unavailable, document `NO-GO` and propose an alternate authenticated source for owner
review. Do not start companion, behavior-lock, regeneration, snapshot-provider, or
external-feed implementation from the documentary setup acceptance.

Immediate blockers: the setup-source manifest and README have fresh non-author
documentary `ACCEPT` for exact commit `d07f3ffee70d1acd5f917d4fa06362aaac594718`
([verdict](../docs/verification/verdicts/p7-replay-setup-sources-round2-20261016.md)), but
the pinned PostgreSQL JAR's detached signature did not verify with the public key found;
publisher identity and bundled license/notice coverage remain unverified. There is no
case-manifest binding, projector/build manifest, complete Go module byte corpus, reviewed
companion API/live `twin` SPEC or fixture corpus, L12/L17 runtime result, or independent
implementation verdict. Go 1.26.7 version suitability remains unreviewed. The full
`make check` gate has not completed on this host; see Verification below.

## Branch and local ownership

- `main` tracks `origin/main`. At the initial 2026-10-17 inspection, HEAD was
  `724c365e66cb97d95c201c74bbb4c4cc5d4611c8` with no ahead/behind count and no
  commits dated today. The documentation-only checkpoint is
  `de2e3be0b2f0ed53586948bf3b97e5fa69198899` (ahead by one at this writing);
  the final closeout commit and push are pending. `origin/main` was reachable and
  still at `724c365e66cb97d95c201c74bbb4c4cc5d4611c8` before final commit.
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

## Open and unverified

- Publisher signing identity, component license/notice and distribution clearance, exact
  Go toolchain patch suitability, complete Phase-A projector/module/toolchain/executable
  evidence, source-manifest binding in a case manifest, companion API/live SPEC and fixture
  review, zero-fetch and cleanup witnesses, L12/L17 `PASS`, requirement satisfaction, and
  an independent implementation review are open. The old empty-cache/no-network
  clean-clone claim remains ineligible for `PASS` under the accepted overlay.
- Sandbox Snap Go lacks `cap_dac_override`; outside the sandbox the installed
  `golangci-lint` v2.13.2 built with Go 1.26.7 panicked while loading a file
  requiring Go 1.27. A compatible-toolchain lint result remains unverified.
  Native Windows acceptance and P7 mutation detector power are also unverified.
- Process debt: the gate registry describes advisory-expiry enforcement, but the inspected
  `scripts/link-lint.sh` checks decision-index membership only; the advisory rows dated
  2026-10-16 need a separate policy decision and a real enforcement check. Do not treat a
  green link-lint as proof of advisory expiry compliance.
- The existing `make hooks-test` loop can mask an intermediate failing self-test by returning
  the last script's status. A fail-closed repair needs a relevant accepted change intent
  before a governed Makefile change is committed. Do not cite the unrelated P5 intent.

## Verification

- Initial bare `make check` on 2026-10-17 reached `commit-cadence` and stopped because
  preserved untracked paths made the old HEAD dirty. The old hook-test loop had masked a
  real stale P6 packet and a sandbox Snap Go failure. `index stale; run make index` was an
  expected negative-test message, not an actual stale decision index.
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
  1.27-versus-1.26 loader panic. Overall gate: **FAIL/UNVERIFIED**, not `PASS`.
  `INERT cleanroom: no readable external pattern file` leaves that control unproved.

## Institutionalized improvement

The canonical EOD prompt now requires a direct P6 artifact-integrity freshness check after
documentary verdicts. This is an interim closeout control, not a claim that the existing
hook-test loop is fail-closed. Preserve reviewed setup bytes; record changing acceptance
status in verdicts and this state note instead of silently changing reviewed content.
