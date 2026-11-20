# VD-vertical-mvp-first-ai3dhn: Finish one vertical MVP path before deepening any organ

**Status:** Accepted
**Date:** 2026-11-20

**Context:** R1, R2 and R3 slice 1 (`VD-r3-warranty-first-s59ktv`) are accepted, but every result
still lives on the builder's machine: nobody else can install Proofbound without building it from
source, nothing appears where developers work, the warranty carries only the exporter's signature,
there is nothing to share with a buyer, and there is no way to pay. The Release 3 scope list
(`VD-release-line-r2-sdd-t5j34p`, [RELEASE-PLAN.md](../plans/RELEASE-PLAN.md)) goes organ by organ
(#2b, #3, #4, #5). The forcing question: build the organs, or make one thin path usable, demoable and
payable end to end first?

**Decision:** Release 3 builds this vertical path first, in this order, each step ending in one
command or link the owner can show to someone (Law 11). The owner approved it on 2026-11-20 ("finish R3
Slice 1 and then go with steps listed one after other"):
1. **Installable:** per-OS release archives with checksums, an install script, and a first PASS on
   a fresh repository in under 5 minutes. The CLI keeps its current database behaviour (`DATABASE_URL`
   when set, otherwise the embedded database fetched at first run), so this ships no PostgreSQL
   binary and does not decide the gated embedded-database redistribution item.
2. **Pull-request check:** a GitHub Action that runs sync, gate and drift on every PR and reports
   warranted / not warranted with reasons.
3. **Reviewer-signed warranty:** warranty slice 2. A reviewer's signature is required for VALID.
4. **Shareable proof:** a warranty attached to each release, a README badge, and a one-page "how to
   verify this".
5. **Commercial step:** planned and tracked outside this repository (owner's decision,
   2026-11-20).

Until step 5 reports back, #2b (full regeneration loop), #3 (Twin product), #4 (Governor), the
verifier marketplace and broader connectors are frozen: not started unless real users need them.
MVP steps use one spec round and one code review each; a mutation sweep only where a step touches
signing or verification.

**Consequences:** The fastest route to real users; every later feature lands on a
path people already use. Given up: the organs the vision calls the business model wait, and the
what users value is guessed, not validated, until real users talk back. The embedded database is
downloaded at first run from a Maven artifact whose signature did not verify (`DEF-P7-001`); the
install docs must say so and offer `DATABASE_URL`.

**Revisit when:** the first external users report back (re-rank by what they asked for), or a step
turns out to need the gated embedded-database redistribution decision after all.
