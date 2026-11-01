# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-01

## Resume — 2026-11-01

**STANDING AUTHORIZATION IN FORCE** ([[VD-agent-standing-authorization-5ouz7q]], 2026-10-31): the
owner (Srinivas Kamisetty) authorized a product push. The agent self-authorizes governed commits
under the owner's name, works ACROSS the normally-dormant P7/deferral gates, and stops only at
genuine forks or irreversible/outward actions (shipping/redistributing a third-party binary, a
license/provenance go/no-go, deletion, publishing). Owner remains the non-author acceptor (Law 9);
adversarial rounds run via dispatched non-author reviewers. **This overrides the "P7 and OPEN
deferrals are dormant until owner authorizes" stop for the duration of the push** — do not re-apply
it blindly.

**Exact next action — await the non-author re-verify of the `d9590da` hardening, then unblock B1
unit 2.** Unit 2 is blocked on a build host: the owner authorized `sudo apt` here, but this
non-interactive session cannot supply a sudo password, so the owner must install the build deps in
an interactive shell (command below) or name another build host.
B1 unit 1 (source provenance) is COMPLETE and ACCEPTED; its two LOW findings are now applied in
`d9590da` (non-author re-verify owed). The owner RESOLVED `DEF-P7-001` on
2026-11-01: **Option B1** — build PostgreSQL from upstream source. Recorded as [[VD-p7-postgres-source-route-2026-11-01]] (zonky
retired, dev/demo-only; Option C rejected). **Provenance mechanism corrected same day:** discovered
during unit 1 that postgresql.org ships source tarballs with SHA-256/MD5 checksums only — NO detached
`.asc` signature, no `KEYS` file — so "verify a signature against a published PGDG key" (as first
recorded) is impossible for source. Owner chose the least-expensive, light-touch mechanism: pin the
PUBLISHED SHA-256 and verify a pristine-upstream tarball fetched over HTTPS from the official domain
(integrity + official-domain TLS trust, NOT an offline signature — owner-accepted residual). The VD
and brief carry the dated correction. Slice plan:
[`docs/plans/P7-postgres-b1-source-build-slice-plan.md`](../docs/plans/P7-postgres-b1-source-build-slice-plan.md).

**B1 unit 1 fetch/verify is DONE** (2026-11-01): PostgreSQL 16.4 `.tar.bz2` fetched over HTTPS (HTTP
200, no redirect, requested URL == effective), measured SHA-256
`971766d645aa73e93b9ef4e3be44201b4f45b5477095b049125403f9f3386d6f` == published, size 24765786 ==
listed. Manifest `docs/verification/p7-replay-lock-fixtures/postgres-source-16.4.json`; evidence
`docs/verification/p7-replay-lock-prerequisites/postgres-source-16.4-verify-20261101.md`. The
fail-closed verifier scripts/p7-postgres-source-verify.sh (not backticked here; the census treats a
backticked `scripts/…` token as a mechanism that must already exist) + its scripts/tests self-test
are DONE (self-test green, wired into hooks-test; verifier re-verifies the real download).
**Unit 1 ACCEPTED** by independent non-author verdict
`docs/verification/verdicts/p7-postgres-source-verify-round1-20261101.md` (status ACCEPTABLE at
`652bb5f`; reviewer independently re-fetched the upstream checksum and ran the real verifier; ~20
break attempts, no false-accept). Both LOW, fail-closed findings are now **APPLIED** in `d9590da` (owner said "do all 3"): (1) the
verifier rejects a symlinked tarball up front (was over-reject-only; now fails closed explicitly,
with a positive self-test control that only the new `-L` check can reject — Law 3); (2) the unused
duplicate `download_sha256` was dropped from the manifest (the verifier reads `published_sha256`
only — Law-2 drift seed removed). Author-green; a **light non-author re-verify of the hardening is
owed (Law 9)** before it counts as accepted. **Unit 2 (build + license/notice inventory)** is
blocked: owner authorized "install deps here via sudo apt", but this non-interactive session's
`sudo` requires a password it cannot supply (`sudo: a password is required`), so the deps are NOT
installed — gcc/bison/flex still missing ([[proofbound-p7-b1-build-toolchain-gap]]). Owner must run
`sudo apt-get install -y build-essential bison flex libreadline-dev zlib1g-dev` in an interactive
shell (or name a build host) before unit 2 can start. **Unit 3** = prepared-offline runtime fixture binding a real `vera.replay.v1` proof,
which lets case-manifest/behavior-lock derive a *verified* status. `DEF-P7-001` stays **OPEN/BLOCKED**
until units 1–3 are accepted. Binary-exclusion law unchanged. Snapshot provider stays skipped
(Gate 1). The retired dev/demo `twin replay` path still uses embedded-postgres ports 55461-55463.

**Shipped this session (2026-11-01):** two non-runtime P7 contracts added to the accepted
`internal/twin/manifest` package (production count stays 17) and non-author **ACCEPTABLE** —
`proofbound.p7.case-manifest.v1` (`casemanifest.go`, the `ReplayIsolated` case evidence index) and
`proofbound.p7.behavior-lock.v1` (`behaviorlock.go`, the authored lock + pure `EvaluateLock` over
committed `vera.replay.v1` proof bytes). SPEC extended (CASE-INV-1..6 / LOCK-INV-1..6). Calibrated
mutation sweep green (335/330/0/5, all survivors pre-declared, 0 new; calibration controls behaved).
Implementation `6dc9372` (governed, `Intent: CI-p7-case-manifest-1a2b3c` + `CI-p7-behavior-lock-1a2b3c`,
both bound to `BR-replay-evidence-6e3d2b` O-1). Verdict:
`docs/verification/verdicts/p7-case-behaviorlock-round1-20261101.md`. Owner brief:
`docs/plans/P7-postgres-provenance-go-no-go-brief.md`. See `notes/journal/2026-11-01.md`.

**Immediate blockers / honest gate state:**

- **`make check` VERIFIED GREEN at `7f6ac32`** (full run; golangci-lint 0 issues). The 2026-11-01 B1
  work is docs/notes/evidence + two governed scripts (`edef420` INDEX regen; `00af5f3` verifier +
  self-test) — no `kernel/*`, so the route-matrix code-commit pin does not restale; each governed
  commit is followed by a non-governed one so HEAD stays non-governed. The governed implementation
  `6dc9372` carries its `Intent:` trailers; invariant gates
  pass (the manifest SPEC uses the table form, which does not feed `invariants.lock`); the
  `@generated` verdict-integrity audit artifact and route-matrix pin were refreshed at `c83c807`.
- **FLAKE (seen TWICE 2026-11-01):** `make check` failed at `kernel-check` on two separate full runs
  while `twin/manifest` itself PASSED; `make kernel-check` alone and `go test ./...` standalone both
  pass every kernel package, and each full re-run was green. Likely embedded-postgres port contention
  (55461-55463) or parallel/timing sensitivity in `internal/twin`/`internal/cli` under the full-run
  profile. Not caused by any docs/shell change this session. Registered (Law 7, seen twice) as
  `DEF-KERNELCHECK-FLAKE-001` (OPEN/DEFERRED) — fix is to isolate the offending test and make it
  hermetic so the gate is deterministic.
- **Push:** B1 decision batch pushed to `origin/main` (`0f44df9..7ceb29e`) on owner instruction
  2026-11-01. The B1-unit-1 batch + the `d9590da` hardening after it are committed locally; the
  owner said "do all 3" (push) 2026-11-01, so this batch is being pushed this session.
- `DEF-P7-001` stays **OPEN/BLOCKED** (pre-GA): shippable PostgreSQL provenance/licensing. The pinned
  zonky.io embedded-postgres v16.4.0 Maven artifact fails provenance (detached GPG signature does not
  verify — expired encryption-only subkey, mismatched published key) and ships no license/notice
  coverage despite bundling PostgreSQL + OpenSSL/ICU/libxml2/libxslt/zlib/liblzma/OSSP-UUID. It runs
  for dev/demo (`kernel/cmd/replaydemo`) but is NOT safe to redistribute. See the go/no-go brief and
  [`p7-replay-lock-prerequisites/README.md`](../docs/verification/p7-replay-lock-prerequisites/README.md).
- `commit-cadence` fails whenever the tree is dirty and HEAD is older than 90m; the two preserved
  untracked user paths (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`) keep the
  tree dirty between commits, so run `make check` promptly after a commit.
- Kernel toolchain: Go 1.27.3; `golangci-lint` 2.15.0 at `~/go/bin`, resolved by `kernel-check.sh`
  via a `$(go env GOPATH)/bin` fallback (no PATH edit needed). See [[proofbound-wsl-toolchain-gap]].

## Deferrals (see `docs/deferrals.tsv`)

- `DEF-P7-001` OPEN/BLOCKED — pre-GA PostgreSQL provenance; owner RESOLVED the route go/no-go (B1,
  [[VD-p7-postgres-source-route-2026-11-01]]) and unit-1 source provenance is verified. Stays open
  until the B1 build + license/notice inventory + runtime fixture are produced and accepted.
- `DEF-P7-002` RESOLVED (Go-module evidence corpus).
- `DEF-P7-003` RESOLVED (twin/manifest accepted as the 17th package, 2026-10-31).
- `DEF-METATAX-001` RESOLVED (meta-tax is a diagnostic, not a gate).

## Branch status

Branch `main`, tracking `origin/main`. `origin/main` is at `7ceb29e` (the B1 decision batch, pushed
on owner instruction). Local commits after it — the B1-unit-1 provenance batch (source manifest,
fetch/verify evidence, VD/brief/plan correction, notes) — are **not yet pushed**. Earlier this
session: governed implementation `6dc9372` (case-manifest + behavior-lock, both `Intent:` trailers)
→ non-governed closeout → `edef420` (B1 decision + slice scope) → notes.

## P7 position

P1–P6 deep-complete (round-C ACCEPTABLE, zero-open census). P7 is a capability phase with no single
mechanical DoD. Status: the replay-lock / behavior-lock line's **non-runtime** contracts are now all
implemented and independently accepted — projector/build manifest + source-toolchain inventory +
evidence closure + skeleton + declaration (prior sessions), and now the case-manifest and
ReplayIsolated behavior-lock. Everything beyond is runtime/provider work gated on the owner
decisions above. The snapshot provider remains deliberately skipped (Gate 1).

## Tomorrow's first action

Run `/proofbound-next`. The live fork is the owner's Postgres go/no-go (`DEF-P7-001`) per
`docs/plans/P7-postgres-provenance-go-no-go-brief.md`. On **B**, scope the authorized setup/build +
runtime-fixture slice (spec-first, non-author verdict). On **A**, P7 non-runtime work is done and the
agent holds. Do not activate `DEF-P7-001` or start runtime/binary/PostgreSQL work without the owner's
explicit pick.
