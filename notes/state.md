# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-03

## Resume — 2026-11-03

**STANDING AUTHORIZATION IN FORCE** ([[VD-agent-standing-authorization-5ouz7q]], 2026-10-31): the
owner (Srinivas Kamisetty) authorized a product push. The agent self-authorizes governed commits
under the owner's name, works ACROSS the normally-dormant P7/deferral gates, and stops only at
genuine forks or irreversible/outward actions (shipping/redistributing a third-party binary, a
license/provenance go/no-go, deletion, publishing). Owner remains the non-author acceptor (Law 9);
adversarial rounds run via dispatched non-author reviewers. **This overrides the "P7 and OPEN
deferrals are dormant until owner authorizes" stop for the duration of the push** — do not re-apply
it blindly.

**Exact next action — review/approve the B1 unit-3 spec, then implement it.** A spec-first draft
(Law 6) now exists: [`docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md`](../docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md).
It defines the prepared-offline runtime fixture that runs the from-source server under
`ReplayIsolated`, witnesses the run (launched-binary / zero-fetch / ledger-preserved / cleanup —
new TWIN-INV-18..22), and binds the real `vera.replay.v1` proof so `EvaluateLock` → LOCKED and
`DeriveCaseStatus` → PASS. **Before implementing:** the spec's open decisions (§8) still need owner/non-author sign-off:
(2) symlink dereference in the archive prep, (3) the launched-binary + (4) zero-fetch witness
mechanisms, and (5) a **change intent (CI-\*)** since the fixture is governed `kernel/*` code.
**§8.1 (highest risk) is RESOLVED 2026-11-03:** a throwaway probe proved `fergusstrange/embedded-postgres`
v1.31.0 runs our from-source 16.4 tree via `BinariesPath` end-to-end (initdb + start on 127.0.0.1 +
`select version()` = PostgreSQL 16.4 + clean stop), with **no download** (the `bin/pg_ctl` stat
short-circuits the fetch) and a short `/tmp` socket. The staged `bin/`+`share/` subset still needs
confirming, but the core approach is validated — the probe was deleted (never committed). The
unit also carries the P1 package-acceptance bar (`make mutants` + non-author ACCEPTABLE).
B1 unit 2 (build + license/notice inventory) is **non-author ACCEPTED 2026-11-03** (round-2
ACCEPTABLE at `a4642fe`). 2026-11-03: the owner installed the build toolchain
interactively (`build-essential bison flex libreadline-dev zlib1g-dev`; gcc 11.4.0) and approved the
**minimal-permissive** config `--without-icu --without-readline` (zlib default-on). Source re-verified
first (unit-1 verifier OK + self-test green), built outside the clone (configure 43s / make 35s /
install 4s); `ldd bin/postgres` links only zlib + glibc; functional proof = initdb (libc) + loopback
start + `select version()` = PostgreSQL 16.4 + clean stop + temp-root removal. The license inventory
survived an adversarial non-author review: **round-1 NEEDS_CHANGES** caught a real MED undercount
(the `ldd`-only method is blind to third-party source **vendored and statically compiled** into the
binary — Ryu/Boost-1.0, hyperloglog/MIT, Spencer regex, ISC, Gifford SHA-2, etc.), which I fixed by
rewriting the inventory to cover **two linkage dimensions** (dynamic + static vendored, enumerated
from a per-file header scan, verbatim notices for the obligation-bearing ones); **round-2
ACCEPTABLE**. The all-permissive / no-copyleft conclusion held throughout. Three LOW/cosmetic
findings recorded as optional cleanup (truncated HLL MIT tail; slightly over-inclusive CJK row;
unlisted public-domain IANA tz). Evidence: `postgres-source-16.4-build-20261103.md` +
`postgres-16.4-license-notice-inventory.md`; verdicts `…unit2-build-round{1,2}-20261103.md`.
**Unit-3 constraint (surfaced + independently re-confirmed during the build):** the replay fixture
MUST use a SHORT Unix-socket dir — the deep scratchpad/prefix path (>107 chars) overran PostgreSQL's
107-byte socket limit; TCP bind on 127.0.0.1 was fine (not a build defect).
B1 unit 1 (source provenance) is COMPLETE and ACCEPTED, and its two LOW findings have been applied
(`d9590da`) and the fix non-author-accepted (`10b4204`, round-2 ACCEPTABLE at `95dfb47`). The owner RESOLVED `DEF-P7-001` on
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
only — Law-2 drift seed removed). The first non-author re-verify of `d9590da` returned **NEEDS-CHANGES** and caught a real Law-3
harness defect: the symlink positive control did NOT isolate the `-L` guard (`stat -c %s` on a
symlink reports the target-path length, so the size check rejected the fixture on its own → the
guard could be deleted with the self-test staying green; the "test that could not fail" class). The
runtime behaviour was correct (no false-accept). Fixed in `10b4204`: the control now makes the
symlink target-path length equal the content size so only the guard can reject it — proven
load-bearing (guard stripped from a copy → fixture ACCEPTED). The **round-2 non-author re-verify
returned ACCEPTABLE** at `95dfb47`
(`docs/verification/verdicts/p7-postgres-source-verify-hardening-round2-20261101.md`; the round-1
NEEDS_WORK verdict is recorded alongside as `...-hardening-round1-...`): the reviewer independently
proved the control load-bearing (stripped-guard self-test exits non-zero) and found no false-accept
across an 11-case battery. One LOW/cosmetic finding (a leftover `download_sha256` in the test's
throwaway synthetic `mkman` manifest — not a drift seed, never read) is recorded as optional cleanup.
**The unit-1 hardening is ACCEPTED.** **Unit 2 (build + license/notice inventory)** is **non-author
ACCEPTED 2026-11-03** (build `a79d53d` + inventory undercount fix `a4642fe`; build host unblocked —
owner installed the toolchain interactively, [[proofbound-p7-b1-build-toolchain-gap]] RESOLVED;
round-1 NEEDS_CHANGES → fix → round-2 ACCEPTABLE). **Unit 3** = prepared-offline runtime fixture binding a real `vera.replay.v1` proof,
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

- **2026-11-03:** unit-2 work is **docs/evidence only** (`a79d53d`, non-governed — `docs/plans/*`
  + `docs/verification/*`, not in the governed path set); `make link-lint` green. No `kernel/*`,
  `scripts/*`, `gates/*`, or Makefile change, so no gate code restales and the last full-run green
  still stands. Full `make check` not re-run this session (nothing gate-affecting changed; the
  `DEF-KERNELCHECK-FLAKE-001` flake makes a full run noisy). Re-run `make check` at next wrap.
- **`make check` VERIFIED GREEN at `6078563`** (full run; golangci-lint 0 issues; this notes pin
  commit sits one past it, non-governed, and does not touch any gate). The 2026-11-01 B1 work is
  docs/notes/evidence + four governed scripts commits (`edef420` INDEX regen; `00af5f3` verifier +
  self-test; `d9590da` symlink-guard hardening; `10b4204` self-test isolation fix) — no `kernel/*`,
  so the route-matrix code-commit pin does not restale; each governed
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
  owner said "do all 3" (push) 2026-11-01; the full batch through this pin commit is pushed to
  `origin/main` this session.
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
  [[VD-p7-postgres-source-route-2026-11-01]]). Unit-1 source provenance ACCEPTED; unit-2 build +
  license/notice inventory **non-author ACCEPTED 2026-11-03** (round-2 ACCEPTABLE at `a4642fe`).
  Stays open until the unit-3 runtime fixture is produced and accepted.
- `DEF-P7-002` RESOLVED (Go-module evidence corpus).
- `DEF-P7-003` RESOLVED (twin/manifest accepted as the 17th package, 2026-10-31).
- `DEF-METATAX-001` RESOLVED (meta-tax is a diagnostic, not a gate).

## Branch status

Branch `main`, tracking `origin/main`. `origin/main` is at `3012110` (the 2026-11-01 EOD wrap).
Local is **ahead by several unpushed commits** — the 2026-11-03 B1 unit-2 batch: `a79d53d` (build
evidence + initial inventory), `caf94c8` (state/journal), `a4642fe` (round-1 verdict + inventory
undercount fix), plus this closeout (round-2 ACCEPTABLE verdict + state/journal). **Not yet pushed**
— left for the owner / next wrap (no unprompted push).

## P7 position

P1–P6 deep-complete (round-C ACCEPTABLE, zero-open census). P7 is a capability phase with no single
mechanical DoD. Status: the replay-lock / behavior-lock line's **non-runtime** contracts are now all
implemented and independently accepted — projector/build manifest + source-toolchain inventory +
evidence closure + skeleton + declaration (prior sessions), and now the case-manifest and
ReplayIsolated behavior-lock. Everything beyond is runtime/provider work gated on the owner
decisions above. The snapshot provider remains deliberately skipped (Gate 1).

## Tomorrow's first action

Run `/proofbound-next`. B1 units 1 and 2 are non-author ACCEPTED; the **unit-3 spec is drafted**
(`docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md`). Next: get the spec reviewed (non-author
+ owner) and settle its §8 open decisions, raise the change intent for the governed kernel work,
then implement spec-first: prepared-offline archive prep (dereference symlinks; bin/+share/ subset)
→ prove embedded-postgres runs from it (highest risk) → runtime fixture + TWIN-INV-18..22 → author
the behavior-lock + case-manifest over the real proof (EvaluateLock LOCKED, case_status PASS) →
`make mutants` + non-author ACCEPTABLE. Only then can `DEF-P7-001` move toward RESOLVED. The
2026-11-03 batch is unpushed — push at wrap or on owner instruction. Optional: three LOW cosmetic
inventory cleanups (see round-2 verdict).
