# Release 2 plan — the SDD Warranty Connector

**What this is:** the build plan for Release 2. The authority for the *release line* is
[`VD-release-line-r2-sdd-t5j34p`](../decisions/VD-release-line-r2-sdd-t5j34p.md); the maintained
scope split is [`RELEASE-PLAN.md`](RELEASE-PLAN.md); the technical authorities underneath remain
[ROADMAP.md](../../ROADMAP.md), the vision build order, and the intent connector contract
[`kernel/internal/connector/intent/SPEC.md`](../../kernel/internal/connector/intent/SPEC.md). This
file is the *how* — phases, dependencies, and non-goals. It is a plan, not an implementation; no
connector code exists yet. *Validated 2026-11-16 against the intent-provider contract; the earlier
"normalized IR" framing was dropped as a Law 2 drift seed (see "Mechanism").*

## Thesis
**Proofbound warrants the output of any spec-driven-development tool.** Deliver one
framework-agnostic **intent provider** that maps SDD artifacts onto the existing canonical intent
model, binds tenant commits and CI check-witnesses to the resulting obligations, enforces delivery,
and surfaces spec drift — proven against **OpenSpec** (reference) and **Spec Kit** (subset
conformance).

## How this carries Release 1
Nothing new bypasses the R1 path, because the connector is **a third provider inside the existing
P5 intent connector** — a sibling of `records` and `specdir`. It emits the same canonical
`Revision`s (business decision / requirement / change intent), goes through the same batch
validation and dependency-topological ordering in `Sync`, carries the same exact-byte
`artifact_sha256` + independent `content_sha`, resolves `Intent:` commit trailers through the same
`Resolve(tree, recordID)` seam the git connector already uses, and is enforced by the same
`delivery-enforce` and reported by the same `report intent|requirement`. That is the guarantee that
R1 is carried: by construction, not by promise.

## Roles & boundary (one home per datum)
| Layer | Owns (source of truth) | Responsibility |
|---|---|---|
| OpenSpec / Spec Kit | the *what* & *why* — specs, ADR/DDR/PDR, relationships | author specs + decisions; model impacts; drive implementation |
| CI / checks | the evidence | run checks → emit check-witnesses |
| `intent.sdd` provider + profile | *nothing* (stateless) | read committed native artifacts at a git tree → map to canonical BD/BR/CI revisions → exact-byte digests → `Resolve` for trailers |
| Proofbound kernel | the delivery record & warranty | validate + order + append (event-sourced); bind witnesses; `delivery-enforce`; drift projection; issue the verdict |

Proofbound never authors specs or decisions. Spec content, decision records, and the relationship /
impact graph stay home in the SDD tool; Proofbound links + digests them, never copies.

## Mechanism: one provider, declarative profiles, no new model
- **Prior art is `specdir`.** It reads committed markdown (`specs/<slug>/{requirements,change-intent}.md`)
  at a tree, parses headers / `- [O-n] …` obligation lines / `- implements <src>:<id>@<digest>#<O>`
  target lines with regexes, and emits canonical BR/CI revisions. The SDD provider is **`specdir`
  with a pluggable grammar**: a per-kit **profile** supplies the path patterns and extraction rules;
  the provider supplies the canonical output.
- **The canonical model is fixed.** Requirements + acceptance criteria → `requirement` with
  `obligations[]`; change proposals → `change_intent` with `targets[]` (OpenSpec deltas map
  `ADDED`/`MODIFIED`/`REMOVED` → `implements`/`modifies`/`retires`). **Governance is not mapped to
  `business_decision`:** the M0 spike ([`r2-m0-mapping-spike.md`](../verification/r2-m0-mapping-spike.md))
  found no kit records an approver, so requirements carry `authorization: undeclared` instead. Design docs are not a
  kind: they appear only as a `constrained_by` reference where the kit makes that link explicit.
  Relationships use the closed relation vocabulary (`authorizes`, `implements`, `modifies`,
  `repairs`, `retires`, `supersedes`, `constrained_by`, `reviews`, `evaluates`).
- **No fabrication, no widening.** Per the intent SPEC, a provider may not fabricate a decision,
  requirement, intent, relation, authority, or obligation; a missing segment stays missing (e.g. a
  kit with no authorization chain yields `authorization: undeclared`). A kit that cannot map without
  fabrication is `ErrUnmappable` — that is the answer, not a reason to add an intermediate model.
- **Profiles are JSON** (stdlib; Law 8 — no YAML dependency without a decision record). A profile
  declares, per canonical kind: path pattern(s), the ID rule, the obligation/target grammars, and
  the status mapping onto the closed status vocabulary.
- **Escape hatch = a code/API provider**, which the intent SPEC already anticipates ("a later API
  provider would record a snapshot observed at ingest"). Reserved for a kit whose source of truth is
  a DB/API rather than committed files. Not built in R2.

## Preconditions
1. Release 1 accepted — **met** (`VD-release-line-r2-sdd-t5j34p`, 2026-11-16).
2. **Mapping spike (go/no-go), on real repos:** take one OpenSpec repo and one Spec Kit repo and
   hand-map a governance doc, a requirement with criteria, and a change/task onto BD/BR/CI *without
   fabricating a segment*. If either kit is `ErrUnmappable` at reasonable granularity, the design
   changes before any SPEC is written. This replaces "confirm formats": it is the acceptance test
   for the whole thesis.

## Decisions to ratify before code (as `VD-*`, with intent chain)
- **Intent home for the R2 build:** author `BD-sdd-warranty-*` → `BR-sdd-warranty-*` with
  obligations as authored in `BR-sdd-warranty-i9miv7` (revision 0002: O-1..O-5, O-8, O-9 active; O-6 / O-7 retired) (that record is the one home for
  their numbering and wording). Every R2 change intent must target a requirement obligation. It stays
  `proposed` until a non-author `proofbound.requirement-review.v1` makes it VERIFIABLE. **Done
  2026-11-16.**
- **Provider, not IR / connector:** the SDD connector is the `intent/sdd` provider package (a
  sibling of `records/` and `specdir/` inside the intent connector), a third production provider. This requires **amending the intent SPEC invariant** "P5 production
  providers are exactly `records` and `specdir`" — a spec amendment with its own non-author review.
- **Agnostic mechanism:** one provider + declarative JSON profiles; code/API provider only as the
  named escape hatch.
- **Mapping design target:** the mapping is designed against the richest producer (OpenSpec);
  Spec Kit is a subset-conformance profile and must not shape the contract.
- **Digests are exact committed bytes — no normalization.** The P5 contract fixes `artifact_sha256`
  as the exact committed file bytes for committed-file providers; "normalize before digest" would
  break it. Regeneration churn is handled by **profile scope** (ingest only intent-bearing files) and
  **granularity** (per-requirement / per-proposal artifacts), not by rewriting bytes. A rewritten
  spec *is* a new revision; that is the drift signal, not noise to suppress.

## Phases (each ends on `make check` green + a non-author verdict, Law 9)

**M0 — Spike, spec, and intent home.** Run the mapping spike (precondition 2). Ratify the decisions
above. Author the `BD`/`BR-sdd-warranty-*` chain. Write the `intent/sdd` package SPEC as a `docs/plans` draft
(Law 6; it lands in the kernel with M1's code — census C2-036): profile schema, the BR/CI mapping, the
`Resolve` seam, invariants (no fabrication; exact-byte digests; closed vocabularies), non-goals,
test derivations. Amend the intent SPEC provider invariant.
*Deliverable:* spike verdict (go/no-go) + ratified package SPEC + intent home, gate green.
*Status 2026-11-16:* spike GO; intent home done. Reviews rejected drafts 1 and 2 (snapshot
designs that real OpenSpec history breaks). **Scope narrowed** (`VD-sdd-change-scoped-r2-k7q3wz`):
R2 warrants changes, read history-natively; the main-spec lifecycle is deferred. Contract at
**draft 3** ([`R2-sdd-provider-spec-draft3.md`](R2-sdd-provider-spec-draft3.md)), measured against
OpenSpec's full history before it was written; round 3 rejected it (2 HIGH: a past fail-closed
commit wedges every sync; the walk was not the git connector's). **M0 CLOSED 2026-11-17** (owner). M1 contract: **draft 10** ([`R2-sdd-provider-spec-draft10.md`](R2-sdd-provider-spec-draft10.md); rounds 4-8 + a targeted re-check, no HIGH; `BR-sdd-warranty-i9miv7` rev 0003 written; M1 narrowed to OpenSpec over a literal-path profile, `VD-sdd-m1-openspec-first-2026-11-17`) after **draft 4**
([`R2-sdd-provider-spec-draft4.md`](R2-sdd-provider-spec-draft4.md)): fail-closed per state, the
git connector's unsimplified walk. Next: M1 (OpenSpec), under an owner-accepted M1 change intent. **This
document's sections below that describe main-spec mapping, governance, and drift against
superseded requirement digests are superseded by draft 3 (now draft 10).**

**M1 — `intent.sdd` provider core.** **OpenSpec kit only** (`VD-sdd-m1-openspec-first-2026-11-17`); Spec Kit mapping moves to M2. Profile loading + validation (JSON); history-native mapping of
every commit the git connector ingests, per distinct SDD state, fail-closed per state with reported
skips → canonical revisions; `Resolve(ctx, commit, name)` (native change name; draft 4 §2, §6) — not
wired into the git connector until M3; `core.Source`
`intent.sdd`; `sync intent sdd` in the CLI; target grammars accept `intent.sdd` as a source.
*Deliverable:* provider conformance against the intent connector's conformance suite, unit +
property tests, **calibration controls stated** (a positive control that must die, a neutral one
that must survive).

**M2 — OpenSpec profile (reference).** **Also the Spec Kit mapping** (moved from M1 by `VD-sdd-m1-openspec-first-2026-11-17`; its starting point is draft 10 Appendix A, and it needs its own spec section and non-author review). The profile: path patterns, ID rule, criteria → obligations,
active changes → change intents + delta requirement records, status mapping; governance maps to
nothing (`authorization: undeclared`); git correlation via `Resolve`.
*Deliverable:* an OpenSpec repo syncs end-to-end, across an archive; BR/CI **with relationships**
land in the record;
`report requirement <id>` reads correctly.

**M3 — Binding + `delivery-enforce` traceability gate (the warranty).** Tenant commits carry
`Intent: sdd:<change-name>` trailers that resolve through the provider (git connector, trailer gate,
and shared corpus extended; decision record for the native-name token); satisfaction is the existing
`obligation-verdict.v2` citing check / workflow evidence; the existing `intent-delivery-readiness`
gate applies to `intent.sdd` chains (change intent → requirement obligation → SATISFIED verdict with
VERIFIABLE requirement review — governance is undeclared, so the chain does not reach it) **without
regressing** existing behavior on this repo; verifier acceptance (Law 9).
*Deliverable:* a change flows build → warrant; gaps blocked; verdict recorded.

**M4 — Spec drift projection (the payoff).** A projection/report over existing data: any change
intent or verdict whose reference binds a requirement `artifact_sha256` that a later revision has
superseded is **stale**; the warranty is flagged until re-verified. Largely read-side — the
exact-digest binding already exists in every `Reference`, and the ledger already models revisions:
an event's idempotency key is `(source, native_id, content_sha)` and `IsRevisionOf` identifies a
later revision of the same subject, so an in-place spec edit becomes a new revision event once
committed (note `sync` reads `HEAD`: uncommitted edits are invisible by design).
*Deliverable:* "edit-spec-orphans-code" is mechanically detected and reported.

**M5 — Spec Kit subset profile (agnosticism proof).** Second profile, same provider, zero kernel
change; feature-local subset, relationship fields absent (not fabricated).
*Deliverable:* two unrelated kits on one provider.

**M6 — Self-hosting demo + acceptance.** Run the provider on a real OpenSpec-driven repo;
end-to-end warranty demo; verification doc; owner acceptance.
*Deliverable:* owner acceptance of Release 2.

**Dependencies:** `M0 → M1 → M2 → M3 → M4`; `M5` needs M3; `M6` needs M4+M5. M4 is where the
differentiated value lands; M5/M6 prove and package it.

## Known gotchas (compiled from this repo's lessons)
- **Route matrix.** M1 touches `kernel/internal/cli` (a routed dir): regenerate the `@generated`
  route matrix *after* committing, or the next `make check` reads stale.
- **Governed commits need an intent.** Every R2 commit that touches governed paths carries
  `Intent: CI-sdd-…` bound to `BR-sdd-warranty-*`; HEAD-only landing fails `delivery-enforce`.
- **Integration-tagged tests do not guard `make check`.** Regression tests for the provider go in
  non-tagged files.
- **The `block-secrets` hook** refuses any command that names `docs/decisions/INDEX.md` together
  with a redirect; run `make index` on its own line.
- **The census guard.** The live docs (`README.md`, `ROADMAP.md`, `notes/state.md`, registered
  accepted plans, accepted VDs, kernel SPECs) may not backtick a kernel/internal path that is not a
  tracked file — `p6-census` dies on an unclassified mechanism. Until a file exists, name the
  package (`intent/sdd`), not the path. (Hit on 2026-11-16 while writing this plan.)

## Non-goals
- Authoring specs/decisions or modeling impacts (the SDD tool's job).
- A new intermediate model, a new top-level connector, or a new event kind.
- Digest normalization of committed bytes.
- API/DB providers until a real DB-backed tenant exists — escape hatch deferred.
- BMAD / Kiro / Tessl profiles — demand-driven, later.
- Rebuilding an SDD tool's relationship/impact engine.

## Top risks
- **Unmappable kits.** A kit may not express authorization or obligations cleanly enough to map
  without fabrication. *(the M0 spike exists to find this first)*
- **Granularity.** If a kit keeps many requirements in one file, drift is file-coarse and every
  edit stales everything. *(profile scope + per-item artifacts; decided in M0)*
- **Provider-invariant amendment.** Widening "exactly `records` and `specdir`" is a spec change with
  review cost. *(scheduled in M0, not discovered in M1)*
- **OpenSpec team maturity** — may be thinner than secondary sources claim. *(the spike runs on a
  real repo)*
- **`Resolve` ambiguity.** A tenant intent ID must resolve to exactly one artifact at a tree, as
  `specdir` requires; kits that reuse IDs across folders need a profile rule. *(M1 invariant)*
