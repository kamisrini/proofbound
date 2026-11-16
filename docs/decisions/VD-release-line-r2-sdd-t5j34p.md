# VD-release-line-r2-sdd-t5j34p: Release 1 is shipped; Release 2 is the SDD Warranty Connector; the revenue organs are Release 3

**Status:** Accepted
**Date:** 2026-11-16

**Context:** Release 1 ("the Trust Record") reached its acceptance gate — the wrap-up set (R1-1..R1-4)
is closed and the clean-clone build+`make check` is proven — with only the owner's acceptance
outstanding. The prior release line [`VD-release-line-kwdh4v`](VD-release-line-kwdh4v.md) deferred
*everything else* to a single Release 2 (the revenue organs + turnkey redistribution). The owner has
chosen to insert a new, nearer product ahead of the organs: a **framework-agnostic connector that
lets Proofbound warrant the output of any spec-driven-development tool** (OpenSpec, Spec Kit, and
others). The forcing question: what is the smallest honest shippable cut *after* the Trust Record —
the heavy revenue organs, or a connector that points the existing warranty at an external build
workflow?

**Decision:** **Accept Release 1 as shipped** (owner, 2026-11-16) and **redraw the release line into
three cuts:**
- **Release 2 = the SDD Warranty Connector.** One framework-agnostic **intent provider**
  (`intent.sdd`, a sibling of the P5 `records` and `specdir` providers) that maps spec-driven-dev
  artifacts onto the **existing canonical intent model** — governance/ADRs → business decisions,
  requirements → requirements with obligations, change proposals/tasks → change intents — through
  declarative per-kit *profiles* (JSON, no new dependency), with a code/API provider as the escape
  hatch the intent SPEC already anticipates. It binds tenant commits and CI check-witnesses to those
  obligations through the existing `Intent:` trailer resolution, extends `delivery-enforce` with a
  spec→code traceability gate, and surfaces spec drift as references bound to a superseded exact
  digest. The mapping is designed against the richest producer (**OpenSpec**, reference);
  **Spec Kit** is a subset-conformance profile. **No new intermediate model:** a kit that cannot map
  onto BD/BR/CI without fabricating a segment is unmappable by design (`ErrUnmappable`), not a reason
  to widen the model. It **carries all of Release 1**: nothing bypasses the intent → verdict →
  `delivery-enforce` path — the SDD tool is a new *source* feeding the same Trust Record. Proofbound
  does **not** author specs or decisions; those stay home in the SDD tool (one home per datum). The
  build plan is [`sdd-warranty-connector-plan.md`](../plans/sdd-warranty-connector-plan.md).
- **Release 3 = the former Release 2 scope, unchanged.** The revenue organs — #2b witnessed
  regeneration's full loop, #3 Twin-as-product, #4 Governor, #5 Trust Engine (portable signed
  warranties, verifier identities/signing) — plus broader connectors and any embedded-DB / turnkey
  redistribution. Every item stays gated exactly as it was; only the label moves from R2 to R3.

The maintained scope split lives in [`RELEASE-PLAN.md`](../plans/RELEASE-PLAN.md); this VD is the
authority for the line itself and **supersedes `VD-release-line-kwdh4v`**.

**Consequences:** Makes easy — a shipped R1 behind us and a clear, nearer next product that proves
self-hosting *outward* (Proofbound warranting an external build workflow, its first non-self tenant
shape) before the heavier organs; a sharpened boundary rule — **R1 records + verifies, R2 warrants
SDD output, R3 regenerates / predicts / governs / redistributes.** Makes hard / gave up — the
revenue-bearing organs (the warranty business the vision defines, #5) move one release later, so
Release 2 deliberately ships **no** new runtime organ — it extends the record to a new source rather
than adding regeneration, prediction, governance, or redistribution. A three-release line is also
more surface to keep consistent, carried by this governed intent chain and `delivery-enforce` rather
than by memory.

**Amended 2026-11-16:** the Decision's mapping clause "governance/ADRs → business decisions" is
withdrawn. The R2 M0 mapping spike (`docs/verification/r2-m0-mapping-spike.md`) found that no SDD
kit records an approver, so mapping governance to a business decision would invent one; governance
maps to nothing and requirements render `authorization: undeclared`. The release line itself is
unchanged. Raised as finding F10 in `docs/verification/verdicts/r2-m0-review-round1-20261116.md`.

**Revisit when:** Release 2 (the SDD Warranty Connector) wrap-up is complete and the owner accepts it
as shipped — then open a Release 3 scope decision; OR the owner pulls a Release-3 organ forward into
R2; OR a shift in the spec-driven-dev ecosystem makes "warrant any SDD kit" the wrong Release 2
boundary.
