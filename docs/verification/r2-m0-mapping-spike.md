# R2 M0 — mapping spike: can SDD kits map onto BD/BR/CI without fabrication?

**Date:** 2026-11-16 · **Plan:** [`sdd-warranty-connector-plan.md`](../plans/sdd-warranty-connector-plan.md)
precondition 2 · **Author:** the plan's author (this is evidence for the non-author review, not a
verdict).

**Verdict: GO — with one owner decision (accountable identity) and two scoped limits.**

## Evidence (read-only, shallow clones; nothing executed)

| Kit | Source | What it contained |
|---|---|---|
| OpenSpec | `Fission-AI/OpenSpec` (the tool uses its own format) | 44 capability specs, 5 active changes, 108 archived changes; `config.yaml` |
| Spec Kit | `github/spec-kit` | constitution + templates only (no feature specs) |
| Spec Kit | `pamburus/hl` (a real project) | 9 feature folders; 223 `FR-` / 29 `SC-` / 8 `NFR-` lines; 2 of 9 features have `tasks.md` |

Canonical model rules applied (intent SPEC §3, `intent.go` validation): business decision needs
non-empty `outcome`, `declared_owner`, `declared_approver`; requirement needs non-empty
`declared_owner` and ≥ 1 obligation; change intent needs non-empty `declared_sponsor` and ≥ 1
target; only `authorized_by` may be empty (`authorization: undeclared`). Statuses and relations are
closed vocabularies.

## Mapping, kind by kind

**Requirement → `requirement` + obligations — GO, both kits.**
- OpenSpec `specs/<capability>/spec.md`: each `### Requirement: <name>` (SHALL statement +
  scenarios) is one obligation. There are **no IDs**: obligation IDs must be derived from the
  heading slug (hash-suffixed past 31 chars), **never from position** — positional numbering
  renumbers on insert, which is the stale-index class this repo already paid for once.
- Spec Kit `specs/NNN-feature/spec.md`: `- **FR-001**` / `SC-` / `NFR-` lines carry **stable
  explicit IDs** → `O-FR-001` etc. (valid against `^O-[A-Z0-9][A-Z0-9-]{0,31}$`). Better than
  OpenSpec here.

**Change → `change_intent` — GO for OpenSpec (strong), COARSE for Spec Kit.**
- OpenSpec `changes/<name>/specs/*/spec.md` deltas are headed `## ADDED | MODIFIED | REMOVED
  Requirements` — these map **directly onto the closed relation vocabulary**: `implements` /
  `modifies` / `retires`. Active change → `proposed`; archived → `accepted`. This is the best fit
  in the whole spike.
- Spec Kit: tasks (`T001`) cite **user stories** (`US1`), not requirement IDs, and most features
  have no `tasks.md`. A change intent can only be the feature folder implementing all of its
  `FR-` obligations, which is coarse. Per-task targeting is **unmappable without inventing a link**
  → not done. (This confirms Spec Kit as the subset profile.)

**Governance → `business_decision` — NO-GO, both kits.** OpenSpec `config.yaml` has no owner,
approver or date. Spec Kit's constitution has a ratification date but no named approver. Mapping
either to a business decision would invent the approver. **Resolution, already sanctioned by the
SPEC:** emit no BD; requirements carry empty `authorized_by` → `authorization: undeclared`. A tenant
that wants governance warranted authors real BDs in Proofbound's own records format.

## Cross-cutting findings

1. **Accountable identity — BLOCKING, owner decision.** `declared_owner` (requirement) and
   `declared_sponsor` (change intent) must be non-empty, and **neither kit records an owner or
   sponsor anywhere.** `specdir` solves this by requiring `Owner:` / `Sponsor:` lines in its own
   format; SDD kits have none. Options:
   - **(a) Profile-declared identity (recommended).** The tenant's committed profile declares the
     owner/sponsor per path pattern (e.g. `specs/** → owner: team-payments`). Committed at the tree,
     so it is a real declaration, resolvable at any commit, and reviewable in a PR. No invented
     content.
   - (b) Git author of the artifact's last commit. Observed, not declared — the canonical field is
     named `declared_*`; this would quietly change its meaning. Rejected unless relabelled.
   - (c) Require tenants to add `Owner:` / `Sponsor:` lines to their specs. Honest but adds friction
     to every spec and breaks "works on an unmodified OpenSpec repo".
2. **Status is free text and drifts.** Real Spec Kit specs use `**Status**: Draft` and
   `**Status:** Existing Implementation (Documented)`, with bold placement and trailing whitespace
   varying. The profile maps an **explicit list** of observed strings onto the closed status set;
   an unlisted status is `ErrUnmappable` (fail closed), never guessed.
3. **Format drift is real.** Header fields differ from the template (`**Feature Name:**`,
   `**Branch:**`). Grammars must tolerate the known variants. The digest stays the exact committed
   bytes (P5 contract), unaffected.
4. **Record IDs.** `BR-<slug>-<6>` needs a 6-character suffix. Derive it deterministically from the
   artifact path (a stable function, stated in the SPEC). This derives identity; it does not invent
   content.
5. **Law 8 holds.** The YAML files (`config.yaml`, `.openspec.yaml`) carry nothing the mapping
   needs, so no YAML dependency is required. Profiles stay JSON.
6. **Granularity.** OpenSpec is per-capability (fine); Spec Kit per-feature (one requirement record
   per feature). Drift detection is correspondingly finer for OpenSpec.

## What this changes downstream

- **Step 2 (`BR-sdd-warranty-*` obligations):** add an explicit obligation for **accountable
  identity** (whichever option the owner picks) and one for **fail-closed status mapping**.
- **Step 3 (package SPEC):** invariants — no BD emitted; obligation IDs never positional;
  unlisted status is `ErrUnmappable`; Spec Kit change intents are feature-level only.
- **Plan:** governance → BD is removed from the mapping ("Mechanism" section) and moved to
  non-goals for R2.
