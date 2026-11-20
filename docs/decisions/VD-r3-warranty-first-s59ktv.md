# VD-r3-warranty-first-s59ktv: Release 3 starts with a portable signed warranty, ahead of witnessed regeneration

**Status:** Accepted
**Date:** 2026-11-19

**Context:** Release 2 shipped (`v0.2.0`). It warrants a commit against its spec, but the evidence
lives only in one machine's Postgres ledger, so nobody else can check it. The roadmap put #2b
(witnessed regeneration, full loop) first in Release 3 (`VD-release-line-r2-sdd-t5j34p`). The forcing
question: what is the next thing a user can run and hand to someone else?

**Decision:** Release 3's first slice is the portable warranty of
[`R3-warranty-export-spec.md`](../plans/R3-warranty-export-spec.md): `proofbound keys init`,
`warranty export <commit>` (gate PASS only), and `warranty verify <file> [--repo]` (no database, no
network). The owner chose all three recommended options on 2026-11-19 ("go"):
- **D1:** the exporter's ed25519 key signs. Reviewer signatures are slice 2.
- **D2:** this goes ahead of #2b.
- **D3:** a key is trusted when the repository commits it under `.proofbound/keys/`.

Signing uses Go's standard `crypto/ed25519`, so there is no new dependency (Law 8).

**Consequences:** R2's output becomes a file a buyer can check against a clone. #2b moves back one
slot. Trust anchored in the repository is weak: anyone who can commit to the repository can add a key,
and the release notes must say so. No revocation, rotation, or transparency log yet.

**Revisit when:** a second party (buyer, auditor) actually consumes warranties, which raises
anchoring beyond the repository; or when reviewer independence must be proven rather than declared,
which is D1 slice 2.
