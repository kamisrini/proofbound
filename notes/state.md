# Proofbound — Current State

> THE resume note. Rewrite in place; never append or create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

As of: 2026-10-27

## Resume — 2026-10-27

Exact next action: obtain independent non-author review of
`docs/verification/p7-evidence-manifest-skeleton-review-packet-20261027.md` against reviewed
revision `c9daba2d4afdef4a2966e54813da81c1b56363db`, then accept, amend, or reject that exact
skeleton slice before any further implementation. Do not start runtime fixtures or binary
validation, do not activate `DEF-P7-001`, and do not claim L12/L17 or behavior-lock acceptance.

Immediate blockers: independent review is pending; complete projector/build manifest,
source/toolchain binding, runtime fixture corpus, process/network/no-fetch/cleanup witnesses,
L12/L17 runtime result, behavior-lock acceptance, clean-clone proof, and independent runtime
verdict are absent. `DEF-P7-001` remains owner-postponed because PostgreSQL publisher identity,
detached-signature resolution, and bundled license/notice coverage remain unresolved.
`DEF-P7-001` is dormant; `DEF-P7-002` is resolved.

Do not treat the accepted live SPEC/API wording as runtime acceptance. Do not treat the pinned
Linux/amd64 PostgreSQL 16.4.0 Maven JAR as approved.

## Branch and local ownership

- `main` tracks `origin/main`; local closeout commits remain unpushed because push was not
  authorized.
- Preserve the unrelated user-owned untracked paths `docs/plans/P5-intent-provenance-plan.md` and
  `kernel/cmd/p6dbprobe/main.go`; do not stage, delete, or relabel them as complete.
- No PostgreSQL binary, database directory, downloaded toolchain, built executable, runtime
  fixture, or binary-validation artifact was added to Git.

## Completed

- P0–P6 remain complete under the committed roadmap and consolidation evidence.
- The P7 companion documentary boundary, verified-archive preflight/staging seam, synthetic archive
  controls, selected Go module corpus, and `P7-SPEC-ADOPT-1` live SPEC/API adoption remain
  independently accepted within their recorded scopes.
- Owner authorization `P7-NONRUNTIME-MANIFEST-1` is recorded in
  `docs/intent/records/change-intents/CI-p7-nonruntime-manifest-1a2b3c/0001.md`. The manifest
  contract, strict shape/revision/digest validation, synthetic tests, P7 decision, and independent
  review are committed in `16fdf7d` through `61e7fab` and accepted in
  `docs/verification/verdicts/p7-nonruntime-manifest-round1-20261026.md`.
- Generated decision index, P6 census classification, and route matrix were refreshed; the new
  package is explicitly deferred to P7 rather than added to the frozen P6 acceptance universe.
- The 2026-10-26 durability inspection found no new implementation work, today-dated commits, or
  session artifacts before this closeout. The deferral register remains canonical with two rows,
  one open owner-postponed item.
- `P7-NONRUNTIME-SOURCE-TOOLCHAIN-INVENTORY-1` is authorized by
  `CI-p7-source-toolchain-inventory-1a2b3c`, implemented in the manifest package, and covered by
  the accepted review packet for revision `987ed9104f8a0c9fad5facd9ee81aa030f57253b`; Srinivas
  Kamisetty returned `ACCEPT` with no findings in
  `docs/verification/verdicts/p7-nonruntime-source-toolchain-inventory-round1-20261026.md`.
- `P7-NONRUNTIME-COMMITTED-EVIDENCE-RESOLUTION-1` is authorized by
  `CI-p7-committed-evidence-resolution-1a2b3c`, implemented with synthetic reader tests, and
  independently accepted with no findings in
  `docs/verification/verdicts/p7-committed-evidence-resolution-round1-20261027.md` for reviewed
  revision `dd497f2ae419f0d38af759c86ffe0029fe012aab`.
- `P7-NONRUNTIME-EVIDENCE-CLOSURE-1` is authorized by
  `CI-p7-evidence-closure-1a2b3c`, implemented as a pure evaluator with synthetic tests, and
  independently accepted with no findings in
  `docs/verification/verdicts/p7-evidence-closure-round1-20261027.md` for reviewed revision
  `3e31c3a191a15113cd26e007caf848c0d2a9b79a`.
- `P7-NONRUNTIME-EVIDENCE-MANIFEST-SKELETON-1` is authorized by
  `CI-p7-evidence-manifest-skeleton-1a2b3c`, implemented as a pure deterministic skeleton with
  synthetic serialization/validation tests, and prepared for independent review in
  `docs/verification/p7-evidence-manifest-skeleton-review-packet-20261027.md`. It is not yet
  accepted.

## Open, blocked, and unverified

- Runtime fixtures, real PostgreSQL execution, process/network/no-fetch/cleanup witnesses, L12/L17,
  behavior-lock acceptance, clean-clone proof, complete projector/build evidence, and an
  independent runtime verdict remain unverified.
- PostgreSQL publisher identity, detached-signature resolution, and component license/notice
  coverage remain unresolved under dormant `DEF-P7-001`.
- The restricted Snap sandbox cannot run the capability-dependent mutation self-test. This remains
  an environment limitation, not acceptance evidence.
- Existing process debt around fail-closed hook-test behavior and advisory-expiry enforcement is
  unchanged and not expanded during this closeout.

## Verification

- `scripts/deferrals.sh --check`, `scripts/p6-artifact-integrity.sh --check`,
  `scripts/p6-route-matrix.sh --check`, and `git diff --check` passed.
- The first bare `make check` reached the repository checks and stopped at `commit-cadence`
  because the preserved untracked paths made the worktree dirty beyond 90 minutes. Generated
  freshness, hostile closure, package, verdict, and Windows-contract checks passed. Expected
  `index stale; run make index` and `INERT cleanroom` diagnostics were not failures; Snap’s
  missing `cap_dac_override` is an environment limitation.
- The pinned unrestricted `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH make check` passed
  with `0 issues`.
- `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH go test ./...` passed. The manifest package
  and intent-record tests passed separately. Independent reviewer `Srinivas Kamisetty` accepted
  exact revision `66cd3441ab4952eb0c50d16a2e39071020035492`; artifact-integrity was refreshed from
  32 to 33 documentary verdict artifacts and rechecked successfully.
- The binding and intent-record tests passed, and the pinned unrestricted gate passed after the
  binding implementation and route-matrix refresh. Its frozen review packet is bound to
  `385274060b62e0dd912b80bd946b494b45277e05`.
- The inventory and intent-record tests passed; deferrals, artifact integrity, route-matrix, and
  whitespace checks passed. After the durable state refresh, the pinned unrestricted gate passed
  with `0 issues`; its expected negative-test and cleanroom diagnostics were non-failures.
- The inventory verdict increased documentary artifact integrity from 34 to 35; the generated
  record was refreshed and rechecked successfully.
- The committed-evidence resolver and intent-record tests passed; deferrals, artifact integrity,
  route-matrix, and whitespace checks passed; the pinned unrestricted gate passed with `0 issues`.
- The resolver verdict increased documentary artifact integrity from 35 to 36; the generated
  record was refreshed and rechecked successfully.
- The closure evaluator and intent-record tests passed; deferrals, artifact integrity,
  route-matrix, and whitespace checks passed; the pinned unrestricted gate passed with `0 issues`.
- The closure verdict increased documentary artifact integrity from 36 to 37; the generated
  record was refreshed and rechecked successfully.
- The skeleton focused tests and intent-record checks pass; deferrals, artifact integrity,
  route-matrix, and whitespace checks pass. The first final gate attempt was blocked only by the
  missing PATH entry for repository-local `golangci-lint`; after adding that path, the gate reached
  the state-freshness check and required this durable refresh. Expected index, negative-test,
  package-count, and cleanroom diagnostics remain non-failures.
- No push was attempted or authorized.

## Institutionalized improvement

- At closeout, commit the durable journal/state refresh before the final bare gate so
  `commit-cadence` measures a fresh HEAD; keep unrelated untracked paths untouched and report them
  separately.
- After documentary verdicts, run `scripts/p6-artifact-integrity.sh --check`; use `--write` only
  for expected generated-count drift, then recheck before committing.
- For every native intent record, verify the declared self-digest with the repository’s exact
  zero-field normalization rule before staging.
- Keep the exact next action and immediate blockers in this first resume section, and distinguish
  expected negative-test/environment output from actual gate failures.
- Run `invariant-table-lint` before the final gate and refresh the generated route matrix after the
  last commit touching `kernel/internal/twin`, so documentary citations and generated identities
  are frozen before review.
- Include the repository-local Go toolchain and `golangci-lint` paths explicitly when the Snap
  wrappers are unavailable; use an isolated writable `GOCACHE` for verification.

## Current checkpoint handoff

- `P7-SPEC-ADOPT-1` is complete and independently accepted; it is live SPEC/API adoption only.
- `P7-NONRUNTIME-MANIFEST-1` is authorized, implemented within scope, and independently accepted;
  it is non-runtime preparation only.
- `P7-NONRUNTIME-BINDING-1` is separately authorized and implemented within scope; its independent
  review is accepted with no findings in
  `docs/verification/verdicts/p7-nonruntime-binding-round1-20261026.md`.
- `P7-NONRUNTIME-SOURCE-TOOLCHAIN-INVENTORY-1` is separately authorized and implemented within
  scope; its independent review is accepted with no findings in
  `docs/verification/verdicts/p7-nonruntime-source-toolchain-inventory-round1-20261026.md`.
- `P7-NONRUNTIME-COMMITTED-EVIDENCE-RESOLUTION-1` is separately authorized and implemented within
  scope; its independent review is accepted with no findings in
  `docs/verification/verdicts/p7-committed-evidence-resolution-round1-20261027.md`.
- `P7-NONRUNTIME-EVIDENCE-CLOSURE-1` is separately authorized and implemented within scope;
  its independent review is accepted with no findings in
  `docs/verification/verdicts/p7-evidence-closure-round1-20261027.md`.
- The P7 documentary and preflight/staging acceptances do not establish runtime or PostgreSQL
  acceptance.
- `DEF-P7-001` remains dormant and the pinned PostgreSQL JAR remains unapproved; `DEF-P7-002` is
  resolved.
- Tomorrow’s first action: obtain and process the independent non-author verdict for the pending
  evidence-manifest skeleton packet; if accepted, record it and refresh artifact integrity before
  selecting any further separately authorized slice. Keep runtime, binary, PostgreSQL, L12/L17,
  and behavior-lock work stopped.
