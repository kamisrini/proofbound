# P7 PostgreSQL provenance/licensing — owner go/no-go brief

**For:** the project owner (the decision-maker on `DEF-P7-001`).
**Date:** 2026-11-01.
**Why now:** the P7 non-runtime contract work is complete up to the runtime gate (case-manifest and
behavior-lock contracts landed and independently reviewed). Every remaining P7 step that would turn
a `ReplayIsolated` lock from `UNVERIFIED` into a verified, shippable claim is blocked on one decision
that is yours, not the agent's: the provenance/licensing go/no-go on the embedded PostgreSQL the
replay runs against. This brief states that decision, the facts, the options, and a recommendation.
It changes nothing on its own. See [`DEF-P7-001`](../deferrals.tsv) and the
[prerequisites record](../verification/p7-replay-lock-prerequisites/README.md).

## The decision in one line

Pick the PostgreSQL route P7 is allowed to treat as **evidence**: keep the current binary
**dev/demo-only** (A), move to a **provenance-verifiable source** (B), or **formally accept the
current binary** with a self-produced license/notice inventory and residual risk (C).

## What is true today

- **Replay works for dev/demo.** `kernel/cmd/replaydemo` runs `ReplayIsolated` against
  zonky.io embedded-postgres v16.4.0 (pinned via `kernel/go.mod` `embedded-postgres v1.31.0`,
  PostgreSQL 16.4.0, Linux/amd64). The tests and demo exercise the real projector in a disposable
  store. Nothing in this brief stops that.
- **The lock and case contracts are ready and honest about this.** The new
  `proofbound.p7.behavior-lock.v1` and `proofbound.p7.case-manifest.v1` contracts derive
  `UNVERIFIED` for any case that depends on a runtime input that has not cleared provenance. They
  cannot be made to read `PASS`/`LOCKED` by absence of evidence. So the gap is recorded truthfully;
  it is not hidden.
- **What is blocked:** a *verified* (not `UNVERIFIED`) runtime `ReplayIsolated` lock, L12/L17
  evidence, a clean-clone runtime proof, and any customer-shippable release that embeds or
  redistributes PostgreSQL.

## The two blockers, precisely

1. **Provenance (authenticity).** The pinned Maven JAR's published checksum and nested `txz` matched
   the direct download, but **the detached JAR signature does not verify**. The signing key ID
   `6B02ED0ED53F9FD3` resolves to an *encryption-only* subkey that **expired January 2020** — before
   the August 2024 signature — and the maintainer's published GitHub GPG key lists a *different* key
   ID. GPG reports `Can't check signature: No public key`. No valid signature and no independently
   authenticated publisher identity has been established.
2. **Licensing/notice coverage.** The JAR and its 1,058-member nested `txz` contain **no
   license/copyright/notice files**, despite bundling PostgreSQL plus OpenSSL, ICU, libxml2, libxslt,
   zlib, liblzma, and OSSP-UUID (confirmed from the publisher's tagged Debian build script). The
   Maven POM declares Apache-2.0 for the *wrapper*, which says nothing about the bundled components.
   Redistribution permission is therefore unverified.

Neither blocker affects dev/demo use. Both block redistribution and any "verified" runtime claim.

## Options

### A — Keep dev/demo-only (status quo; no decision cost)
- **Do:** nothing. `DEF-P7-001` stays OPEN/BLOCKED. Replay stays a dev/demo capability.
- **Unblocks:** nothing new. Runtime lock cases remain permanently `UNVERIFIED`; no GA that embeds
  PostgreSQL.
- **Cost/risk:** none added. The ceiling is that P7's headline promise (a *verified* behavior-lock
  on real replay) is never demonstrated end-to-end, and Proofbound cannot ship a product that embeds
  this database.
- **Pick this if:** the near-term goal is the contracts and the demo story, not a shippable
  embedded-DB release.

### B — Move to a provenance-verifiable PostgreSQL source (recommended)
- **Do:** stop treating the zonky binary as the shippable route. Adopt a source whose bytes you can
  authenticate and whose component licenses you can inventory. Two sub-routes, in rough order of
  effort:
  - **B1 — build PostgreSQL from the official signed source tarball** (postgresql.org tarballs are
    signed by a published PGDG key; versions and component libraries are known), produce the binary
    in the P7 setup step, and generate a complete license/notice inventory from the known
    components. Highest provenance confidence; most build work.
  - **B2 — adopt a distributor that publishes an authenticated signature chain and license/notice
    set** for the exact binary. Less build work; depends on finding one whose signature actually
    verifies (zonky's does not today).
- **Unblocks:** a real verified `ReplayIsolated` runtime lock, L12/L17, clean-clone runtime proof,
  and a shippable path — i.e. the whole remaining P7 runtime ladder.
- **Cost/risk:** a bounded setup/build spike plus one authorized runtime slice and its non-author
  verdict. Needs a short decision from you on B1 vs B2 before the agent picks the exact source.
- **Pick this if:** you want P7 to actually demonstrate its headline claim and keep a shippable
  option open.

### C — Formally accept the current zonky binary
- **Do:** obtain a trusted publisher key (or other authenticated provenance) for the zonky artifact
  *and* produce/commit a complete component-license/notice inventory for the exact runtime binary,
  then record an owner decision accepting it with documented residual risk.
- **Unblocks:** the same runtime ladder as B, *if* authenticated provenance can be obtained — which
  it cannot today (the signature does not verify and the maintainer's published key does not match).
- **Cost/risk:** you would be accepting a binary whose signature does not currently authenticate. The
  missing-signature problem is the hard part; the license inventory is doable but does not fix
  authenticity. **Not recommended** unless the maintainer supplies a verifying key.

## What each option lets the agent do next (no new decision from you beyond the pick)

- **A:** nothing further on runtime; P7 non-runtime work is done. The agent closes out.
- **B:** on your B1/B2 pick, the agent scopes the authorized setup/build + runtime-fixture slice
  (spec-first, non-author verdict), keeping `DEF-P7-001` explicit until the provenance+license
  evidence is in hand.
- **C:** the agent can produce the license/notice inventory, but the authenticity blocker needs a
  verifying publisher key that only the upstream maintainer can supply — so this stays blocked on an
  external party, not on the agent.

## Recommendation

**Option B**, sub-route **B1** (build from the official signed PostgreSQL source), because it is the
only route that both (a) gives provenance you can actually verify today and (b) keeps a shippable
product on the table — which is the point of the P7 runtime ladder. If the near-term goal is only
the contracts and the dev/demo story, **A** is a legitimate, zero-cost hold and nothing is lost that
B cannot pick up later. **C** is only viable if the upstream maintainer provides a key whose
signature verifies; until then it is blocked on someone outside this project.

This remains your call: it is a licensing/redistribution and authenticity judgment, and the agent
will not make it for you.
