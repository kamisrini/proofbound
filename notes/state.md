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

**Exact next action — an owner decision, then one of two paths:** P7's non-runtime work is now
complete up to the runtime gate. The remaining P7 ladder (a *verified* runtime `ReplayIsolated`
lock, L12/L17, clean-clone proof, a shippable embedded-DB release) is blocked on the owner's
**PostgreSQL provenance/licensing go/no-go** (`DEF-P7-001`). The decision brief is
[`docs/plans/P7-postgres-provenance-go-no-go-brief.md`](../docs/plans/P7-postgres-provenance-go-no-go-brief.md):
option **A** keep dev/demo-only (no further runtime work; close out), **B** move to a
provenance-verifiable Postgres source (recommended **B1**: build from the official signed source —
then the agent scopes an authorized setup/build + runtime-fixture slice), or **C** formally accept
zonky (blocked on the upstream maintainer supplying a verifying key). The agent cannot make this
call. Until the owner picks, no runtime fixtures, binary validation, L12/L17, PostgreSQL approval,
or `DEF-P7-001` activation. The snapshot provider stays skipped (Gate 1) absent a new lawful-feed
decision. The `twin replay` CLI / demo tests use embedded-postgres ports 55461-55463.

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

- **`make check` expected GREEN at this non-governed closeout HEAD** (run it to confirm — it is slow
  and not run by the SessionStart hook). The governed implementation `6dc9372` carries its `Intent:`
  trailers; this closeout is non-governed, so `delivery-enforce.test.sh` (via `hooks-test`) sees a
  non-governed HEAD. Package count stays 17 (new files share the `twin/manifest` dir), invariant
  gates pass (the manifest SPEC uses the table form, which does not feed `invariants.lock`), and the
  `@generated` verdict-integrity audit artifact was refreshed with `--write`.
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

- `DEF-P7-001` OPEN/BLOCKED — the pre-GA PostgreSQL provenance item above; now has a decision brief.
- `DEF-P7-002` RESOLVED (Go-module evidence corpus).
- `DEF-P7-003` RESOLVED (twin/manifest accepted as the 17th package, 2026-10-31).
- `DEF-METATAX-001` RESOLVED (meta-tax is a diagnostic, not a gate).

## Branch status

Branch `main`, tracking `origin/main`. This session added governed implementation `6dc9372`
(case-manifest + behavior-lock contracts, both `Intent:` trailers) followed by this non-governed
closeout (ROADMAP P7 note, journal, this state note, the verdict, the refreshed verdict-integrity
audit artifact, and the Postgres brief). Prior session wrapped at `0f44df9`, pushed to `origin/main`.
**Not yet pushed this session** — push is an outward action left for the owner / the wrap step.

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
