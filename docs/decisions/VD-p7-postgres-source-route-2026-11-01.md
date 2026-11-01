# VD-p7-postgres-source-route-2026-11-01: build PostgreSQL from signed source for P7 replay

**Status:** Accepted
**Date:** 2026-11-01
**Context:** P7's non-runtime contracts (case-manifest, behavior-lock, and the earlier
projector/build-manifest slice) are complete and independently accepted, but every remaining P7
step that turns a `ReplayIsolated` lock from `UNVERIFIED` into a verified, shippable claim was
blocked on one owner decision: which PostgreSQL the replay is allowed to treat as evidence
(`DEF-P7-001`). The pinned zonky.io embedded-postgres v16.4.0 binary fails provenance on two counts
— its detached GPG signature does not verify (the signing key ID resolves to an encryption-only
subkey that expired January 2020, before the August 2024 signature, and the maintainer's published
GitHub key lists a different ID), and the JAR plus its 1,058-member nested `txz` carry no
license/notice files despite bundling PostgreSQL, OpenSSL, ICU, libxml2, libxslt, zlib, liblzma, and
OSSP-UUID. The forcing question posed to the owner in
[the go/no-go brief](../plans/P7-postgres-provenance-go-no-go-brief.md) was: keep the current binary
dev/demo-only (A), move to a provenance-verifiable source (B), or formally accept the zonky binary
with a self-produced inventory and residual risk (C).
**Decision:** Adopt **Option B, sub-route B1**: PostgreSQL built from the official signed
postgresql.org source tarball is the provenance-verifiable embedded-database route for P7 replay.
The build obtains the exact upstream source tarball and its detached signature in an explicit
network-permitted preparation phase, verifies the tarball against a published PGDG release key and a
reviewed digest before any compilation, produces the server binary in the P7 setup step, and
generates a complete component license/notice inventory from the known source components. The zonky
binary is retired as the *shippable* path and remains dev/demo-only
([`kernel/cmd/replaydemo`](../../kernel/cmd/replaydemo)); nothing in this decision stops that
existing dev/demo use. Option C is rejected because its authenticity blocker depends on the upstream
zonky maintainer supplying a verifying key — an external party, not this project. This decision
authorizes the B1 *route*; it does not by itself accept any produced binary, runtime fixture, L12/L17
result, or requirement satisfaction — each of those remains separately gated behind spec-first work
and a non-author verdict, and `DEF-P7-001` stays explicit until the provenance + license evidence is
in hand.
**Consequences:** Makes easy — a real *verified* (not `UNVERIFIED`) runtime `ReplayIsolated` lock,
L12/L17 evidence, a clean-clone runtime proof, and a redistributable embedded-DB release become
reachable; provenance is verifiable against bytes and a key this project can check today, and the
license/notice inventory is derivable from a known component set. Makes hard / cost accepted — a
bounded setup/build spike (fetch, signature + digest verification, compile, inventory) now sits on
the P7 critical path, more build work than adopting a prebuilt distributor (B2, not chosen because
no distributor with a verifying signature chain was identified — zonky's does not verify today). We
give up the convenience of the prebuilt zonky JAR as a shippable artifact and accept owning a source
build. The frozen `ReplayIsolated` subject, the `vera.replay.v1` proof shape, the authored/observed
boundary, and the no-new-provider/feed/event exclusions do not change; the deliberately-skipped
snapshot provider (Gate 1, [[VD-p7-snapshot-provider-skip-2026-10-01]]) stays skipped absent a
separate lawful-feed decision.
**Revisit when:** the official postgresql.org source distribution begins (or ceases) publishing a
detached signature for source tarballs, changing what authentication is possible; or a prebuilt
distributor is found whose signature chain and license/notice set verify, making B2 cheaper than
owning the build; or the zonky upstream maintainer publishes a key whose signature over the pinned
artifact verifies, which would reopen C as a lower-effort alternative.

**Amended 2026-11-01 (premise correction + mechanism chosen):** the Context above assumed
"postgresql.org tarballs are signed by a published PGDG key." That premise is **false for the source
tarballs**: the official v16.4 source folder ships SHA-256 and MD5 checksum files only — no detached
`.asc` signature and no `KEYS` file in the source tree (PostgreSQL historically published PGP-signed
*checksums* via the announce list, not a detached signature in the download folder). B1's goal
(provenance-verifiable source) stands, but its mechanism cannot be "verify a detached signature
against a published key." Presented the fork to the owner; the owner chose the **least-expensive,
light-touch mechanism**: pin the **published SHA-256** and verify a pristine-upstream tarball
**fetched over HTTPS** from the official domain. Honest guarantee: byte integrity against the
published checksum + official-domain TLS trust — **not** an offline signer-identity signature (the
checksum and tarball share an origin). This residual is owner-accepted and is materially stronger
than the retired zonky route (whose detached signature actively failed), while weaker than a true
signature chain. This is the authenticity half of `DEF-P7-001` for the source route; the licensing
half (component license/notice inventory) is unchanged and still owed at build time. First
verification evidence (digest + size match, no redirect):
[`postgres-source-16.4-verify-20261101`](../verification/p7-replay-lock-prerequisites/postgres-source-16.4-verify-20261101.md).
