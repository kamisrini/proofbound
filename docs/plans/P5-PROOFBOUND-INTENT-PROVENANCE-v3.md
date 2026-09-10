# P5 Execution Plan — Intent Provenance (Proofbound) — v3 (ratified)

**Status:** v3 — round-1 adjudication folded (v2), founder ratification 2026-09-09 recorded (p5-founder-ratification.md, travels with this plan): all four open tradeoffs APPROVED, including requirement review in soft-gate form (semantic rule 8). Remaining before code: commit the round-1 adjudication and the ratification record on receipt, mint the semantic VD citing both, amend ROADMAP.md — then execute Task 0.5 onward.

**Product identity:** Proofbound (formerly VERA); the identity migration is Task 0.5.

**Proposed phase:** P5, after the accepted P4 twin spike

**Decision required before execution:** a ratified VD must adopt or amend the semantic model in Task 0; this draft does not amend `ROADMAP.md` and is not itself a product decision.

**Audience:** an independent product/architecture verifier first, then a fresh implementation session after explicit human approval.

## Verifier handoff — repository state before P5

This section is the minimum self-contained context for review on a machine with no prior session history. It describes the accepted implementation through P4; everything later in this document is a P5 proposal, not existing behavior. The snapshot used to prepare this handoff is `main` at `786d6e1` (2026-08-27). Verify the actual checkout with `git status --short --branch` and `git log -1 --oneline` because a transferred checkout may be newer or contain local work.

### Product and trust model already established

Proofbound (called VERA in the older internal documents and CLI) is a warranty layer for software delivery. Its kernel records immutable observations in an append-only PostgreSQL event ledger, derives disposable/rebuildable projections, and renders proof-bearing reports. The governing boundary is intentionally strict:

- connector observations, authored claims, and independent verdicts are different kinds of fact;
- derived tables are never a second writable truth and must rebuild to the same row set;
- ingestion is idempotent on source-native identity plus content identity;
- evidence and verdicts bind to exact commits/artifact bytes rather than mutable names or positions;
- new primitives require a real feed, and accepted packages require both mechanical evidence and an independent non-author verdict.

The load-bearing product decisions are indexed in `docs/decisions/INDEX.md`. Start with `VD-north-star-6io56h`, `VD-stack-go-fid9mi`, `VD-commit-is-durability-6p29qf`, `VD-verdicts-are-artifacts-rl0rab`, `VD-verification-asymmetry-2dyjnd`, and `VD-no-graph-2aa4vz`. `CLAUDE.md` contains repository operating laws; `ROADMAP.md` is phase status; `notes/state.md` is the canonical resume note. If those sources disagree, do not silently reconcile them in code—identify the contradiction in the review.

### Accepted implementation through P4

| Phase | What exists now | Primary acceptance evidence |
|---|---|---|
| P0 | Repository laws, decision records, generated indexes/locks, check scripts, hooks self-test, and the `make check` gate. | `ROADMAP.md`; `docs/gates.md` |
| P1 | Go kernel; PostgreSQL migrations and append/read/transaction store; Git, check-witness, session, and review-verdict connectors; rebuildable projections; weekly proof report; ledger-ordered red-verdict/change/next-verdict chain; blocking package-SPEC coverage. | `docs/plans/P1-flight-recorder-plan.md`; `docs/verification/task6-final-evidence.md` through `task9-final-evidence.md`; accepted verdicts under `docs/verification/verdicts/` |
| P2 | Seven YAML gate definitions; ledger-backed canary evaluation; explicit fail-closed enforcement; expiry checks; serialized `make delivery-enforce` workflow that refreshes witnesses, ingests them, then enforces. | `docs/verification/p2-gate-evidence.md` |
| P3 | Bounded read-only GitHub Actions/deployments connector for an owner plus repository allowlist; exact-commit joined tested-what/deployed-where projection; freshness and event proof in the report. Live acceptance against public `github/docs` synced 200 records in 7 seconds, rendered 106 groups, and replayed with 0 new events in 4 seconds. | `docs/decisions/VD-p3-github-connector-2026-08-26.md`; `docs/verification/p3-github-live-acceptance.md` |
| P4 | `internal/twin` bounded replay; candidate-sequence validation; disposable embedded-PostgreSQL projection; deterministic `vera.replay.v1` proof including payload bytes; source-ledger preservation and cleanup on failure; pure in-memory forecast calibration. The accepted mutation sweep killed 31/31 candidates. | `docs/decisions/VD-p4-twin-replay-calibration-2026-08-26.md`; `docs/verification/verdicts/p4-current-round1.md` |

P4 does **not** persist prediction events. That remains deferred until a real prediction feed has a separate decision and acceptance evidence. Nothing currently implements business decisions (`BD`), behavioral requirements (`BR`), change intents (`CI`), commit claims, or obligation-level verdicts; those are the new semantic objects proposed below.

### Current code and runtime surfaces

- `kernel/internal/core`: event kinds, identities, validation, and clocks.
- `kernel/internal/store`: PostgreSQL ledger, migration `migrations/001_ledger.sql`, append/read, locking, and transaction boundaries. Tests can use embedded PostgreSQL or an external `DATABASE_URL`.
- `kernel/internal/connector/{git,checks,sessions,reviews,github}`: existing observation adapters.
- `kernel/internal/projections`: incremental application, destructive rebuild from the ledger, snapshots, weekly/review chains, and GitHub delivery reporting.
- `kernel/internal/gates`: YAML gate loading, canary/enforce evaluation, expiry, and ledger proof.
- `kernel/internal/twin`: isolated replay and calibration spike accepted in P4.
- `kernel/cmd/vera`: CLI entry point. Current surface is `vera sync {git|checks|sessions|reviews|github|all}`, `vera rebuild`, `vera verify`, `vera report {week|github}`, and `vera gates {canary|enforce}`.

Local runtime state lives under `.vera/` and is not authoritative source code. Check witnesses are strict `vera.witness.v1` JSON spool artifacts; review files use `vera.verdict.v1`; replay proof uses `vera.replay.v1`. Preserve backward compatibility for all three when evaluating the P5 schemas.

### Reproducing the baseline on another machine

Prerequisites are Git, Bash, Go 1.26 or newer, `golangci-lint`, and PostgreSQL. Docker with `postgres:16-alpine` has been used for external-database integration and mutation runs. From the repository root:

```bash
git status --short --branch
git log -1 --oneline
make check
```

`make check` is the full blocking repository gate. Run it bare so its exit status is not hidden by a shell pipeline. `make short` currently runs only the hooks self-test and is not evidence that Go tests passed. `make mutants` is the slower package-acceptance mechanism and is intentionally outside `make check`.

For a ledger-backed verification run, provide a disposable PostgreSQL database through `DATABASE_URL`, create a fresh witnessed check with `make check-witnessed`, then run `make verify`. For the promoted delivery boundary use `make delivery-enforce`. GitHub sync additionally requires `VERA_GITHUB_OWNER` and `VERA_GITHUB_REPOS`; it is network-backed and intentionally limited to the configured repositories and v1 collection bound.

### Known baseline limitations and review cautions

- This is an early-stage CLI/kernel, not a finished end-user product.
- `README.md` still lists P4 work under “Remaining work” even though its next paragraph, `ROADMAP.md`, `notes/state.md`, and the committed independent verdict record P4 as accepted. Treat that list as stale documentation, not as the phase authority.
- Some imported prose describes checks that are absent from this checkout. Only visible scripts, their self-tests, kernel build/test/lint, YAML gates, and committed evidence count as mechanisms.
- `make backup` is absent; `make short` is incomplete; state-freshness, skip-lint, prescription-lint, and full historical invariant-citation resolution remain mechanism debt.
- Session ingestion was accepted with synthetic fixtures because no real session JSONL corpus was available. Do not upgrade that evidence into a claim of live-source acceptance.
- The GitHub connector is read-only and deliberately narrow; unmatched workflow/deployment records render as missing instead of being inferred across commit SHAs.
- Projection tables may be dropped and rebuilt. P5 must extend the ledger-first model rather than make a projection or graph database authoritative.
- Acceptance means the cited frozen revision and evidence passed the repository's rules. It does not make hand-authored authority authentic, prove business intent, or prove more than the tested routes and mutation operators cover.

Round-1 adjudication note: the reviewing machine could not reach main@786d6e1, so the acceptance table above is carried context, not re-audited fact — Task 0 attaches this machine's own gate and verify evidence to the ratifying decision.

## Goal

Build the smallest honest chain from a business choice to shipped software:

```text
business decision revision
  -> authorizes requirement revision and its obligations
  -> targeted by a change-intent revision
  -> claimed by an exact commit
  -> evaluated using ledger evidence by an independent verdict
  -> observed in a deployment
```

The product must answer, with event proof:

1. What business choice authorized this behavior?
2. What exact requirement revision and obligations were in force?
3. Which exact commit claimed to implement, modify, repair, or retire them?
4. What evidence was considered for each obligation?
5. What did an independent verifier conclude?
6. Where was that commit observed deployed?
7. What is missing, contradictory, superseded, or inconclusive?

P5 does **not** prove that a business decision was wise, that prose perfectly captured what the business meant, or that a hand-authored approver field is authentic. It records claims as claims, observations as observations, and verdicts as verdicts.

## Sources of intent — modular by contract (F1)

The chain's IMPLEMENTATION side is already source-neutral — commits, checks, deployments arrive through connectors, whatever produced them. P5 makes the INTENT side equally source-neutral. Business decisions, requirements, and change intents may be authored in Proofbound-native committed records, in file-native spec tools (spec-directory conventions of the GitHub spec-kit / AWS Kiro shape), or in API-backed in-house design and decision systems. Three things are therefore separated:

1. **The canonical semantic model** — BD/BR/CI, stable obligation identities, typed exact-revision relations, tri-state per-obligation verdicts — is the LEDGER's payload contract, provider-independent. It is what reports, gates, and verdicts consume. No provider's file format is the model.
2. **The Intent Provider contract** (pinned in the connector-family SPEC's interface-lock section before any code). A provider must supply: (a) stable record identity in a provider-scoped namespace; (b) an immutable revision identity — the artifact digest over the provider's canonical bytes (git blob bytes for committed files; canonicalized export bytes otherwise); (c) a mapping into the canonical types and typed relations; (d) obligations with stable ids and statements; (e) a point-in-time rule answering "which revision was in force for commit X" — committed-file providers resolve against the commit's own tree; a provider that cannot do better resolves to the revision observed at ingest, and the event says so. A provider that cannot supply a chain segment yields an honest partial chain (`authorization: undeclared`), never a synthesized record — partial chains are first-class gaps, which is the anti-bypass design: teams onboard at the segment they actually author.
3. **Providers in P5 — exactly two real providers.** The repo-native records provider (the reference implementation; the records-connector design below, unchanged in substance) and ONE file-native foreign mapper (`specdir`, parsing a spec-directory convention), which exists to prove the contract discriminates: the same conformance suite runs against every provider, and the canonical model must absorb a format it does not own without bending. The conformance harness additionally carries a synthetic test-double provider — fixtures only, no production mapping, never configurable in a release build — to exercise contract classes no file-native provider can reach (upstream mutation between resolution and ingest; cross-provider ordering races). API-backed providers against mutable upstreams are OUT of P5: the contract pins how they will work — snapshot ingestion, where the ledger becomes the append-only history the upstream lacks; upstream revision metadata is a claim, observed ledger sequence is proof — and implementation waits for a real feed, per the no-primitive-without-a-feed law.

**Two digests, stated once (F3):** relations bind the ARTIFACT digest (the provider's immutable revision bytes); ledger idempotency binds the CANONICAL PAYLOAD digest (RFC 8785 over the normalized event payload — existing machinery). **Replay self-sufficiency (F4):** every id, lifecycle state, typed relation, and obligation statement a verdict can evaluate embeds in the normalized payload; prose bodies bind by digest and remain the source's artifact — so a fresh clone reconstructs every relation and obligation meaning from the ledger alone, for every provider, by construction. **Source naming (F5):** sources are `intent.<provider>` (e.g. `intent.records`, `intent.specdir`); record-id namespaces are provider-scoped; `supersedes` may cross providers and is a typed, tested relation. Provider names are configuration and generic convention names — no vendor coupling in kernel code or schemas.

## Identity migration — VERA → Proofbound (Task 0.5, F2)

The product is **Proofbound**; VERA remains only where history or pinned bytes require it.

FROZEN (never renamed): the wire schema names `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1` and their pinned byte-exact vectors — new schema versions take `proofbound.*` names, as this plan's schemas already do; committed adjudications, journals, decision records, and git history — history is not edited to flatter the present.

LIVE (renamed in Task 0.5, while zero external consumers exist — the cheapest this rename will ever be): the CLI (`proofbound`, with `vera` kept for one phase as a deprecated alias so existing wrappers keep working), the `VERA_*` environment-variable namespace (renamed to `PROOFBOUND_*`, with the old names honored for one phase under the same deprecated-alias advisory), the Go module path and imports, README/ROADMAP/plan prose, the runtime state directory (`.vera/` → `.proofbound/`, a one-time documented move; the derived lock-path rule is unaffected), and every live document.

Mechanical DoD: a classified inventory of every remaining case-insensitive `vera` occurrence — each hit tagged frozen-wire, frozen-history, deprecated-alias, or baseline-quote (a live document deliberately quoting pre-rename state — such as this plan's verifier-handoff section) — with zero unclassified hits; pinned vectors reproduce byte-exact; the full gate green; the alias's removal date recorded as a dated advisory in the gates registry.

## Why separate record types are required

The current `VD-*` records answer engineering questions: why a stack, boundary, or build rule was chosen. A business decision, a behavioral requirement, and a change's implementation claim have different authors, lifetimes, revision rules, and proof semantics. A generic `type` label on the current VD template would still leave those meanings conflated.

P5 introduces separate semantic artifacts while reusing one ingestion and ledger mechanism:

| Artifact | Meaning | Typical authority | Changes at |
|---|---|---|---|
| Engineering decision (`VD`) | Why an engineering constraint was chosen | Technical owner | Architecture-policy cadence |
| Business decision (`BD`) | Why a business outcome or constraint was authorized | Named business authority | Business-decision cadence |
| Requirement (`BR`) | What observable obligations must hold | Domain owner delegated by a BD | Behavior cadence |
| Change intent (`CI`) | What one delivery claims to do to exact BR obligations | Change sponsor/builder | Delivery cadence |
| Verdict | Whether evidence supports each obligation for an exact commit | Independent verifier | Verification cadence |

These remain files during bootstrap, consistent with `VD-no-graph-2aa4vz`. The ledger stores immutable observations of committed artifact revisions; projections and joins are derived. No graph database and no second writable truth are introduced.

## Semantic rules to ratify before code

### 1. A record is not automatically a fact

The connector can observe that committed bytes exist and who Git says authored the commit. It cannot infer that the record is correct or that a named person approved it. Event names and reports must preserve this boundary:

- `business_decision.recorded` means an artifact revision was observed.
- `requirement.recorded` means obligations were declared in an artifact revision.
- `change_intent.recorded` means a delivery claim was declared.
- `approval.observed` is reserved for an authenticated approval source added later.
- `review.verdict` remains a verifier conclusion, not an observed law of nature.

Until an authenticated approval connector exists, repo-native `owner` and `approved_by` fields are displayed as **declared authority**, never “verified approval.” Git signature verification is not in P5 unless separately decided and fed by a real identity source.

### 2. Every relation binds an exact revision

An ID alone is insufficient because record contents can change. Every relation carries:

```json
{
  "record_kind": "requirement",
  "source": "intent.<provider>",
  "record_id": "BR-<slug>-<suffix>",
  "artifact_sha256": "<64 lowercase hex>",
  "relation": "implements",
  "obligation_ids": ["O-1", "O-2"]
}
```

`record_id` resolves within the named provider's namespace; a cross-provider relation names the target's source explicitly — without the `source` field a cross-provider reference cannot be expressed at all.

`artifact_sha256` identifies the exact canonical artifact bytes. A later revision produces a new ledger event and cannot reinterpret an old commit, verdict, or deployment.

### 3. Relations are typed

Closed relation values for P5:

- A requirement is `authorized_by` a business decision.
- A change intent `implements`, `modifies`, `repairs`, or `retires` requirement obligations.
- A change intent may be `constrained_by` an engineering decision.
- A commit `claims` a change intent.
- A verdict `evaluates` requirement obligations for a commit and cites evidence events.

The existing untyped `cited_decisions` field remains historical compatibility data. It must not be presented as business intent.

### 4. Obligations have stable identities

Each requirement revision contains one or more obligation IDs with testable statements. IDs are append-only within the requirement lineage. Changed meaning gets a new obligation ID; a retired obligation leaves a tombstone. Rewording that demonstrably preserves meaning may retain the ID but still produces a new artifact revision.

### 5. Requirement satisfaction is tri-state, per obligation

Closed P5 verdict outcomes:

- `SATISFIED`
- `NOT_SATISFIED`
- `INCONCLUSIVE`

Absence of a verdict or evidence is `UNVERIFIED`, derived by the projection and never stored as a verdict. No aggregate may render green while an applicable obligation is unverified, inconclusive, contradicted, bound to another revision, or evaluated for another commit.

### 6. Evidence does not grade itself

A check witness may state which obligation it was designed to exercise, but that association is an authored claim. The independent verdict names the evidence event IDs it considered and owns the obligation outcome. Builder-produced evidence cannot directly set `SATISFIED`.

### 7. Applicability is not guessed in P5

P5 does not infer business intent from filenames, diff contents, commit subjects, or an AI summary. It validates explicit `Intent:` commit trailers. The first enforcement gate checks the integrity of declared intent; it does not require every repository commit to carry business intent. A later phase may add a separately ratified applicability policy for behavior-changing work.

### 8. Specs are reviewed like code (founder-ratified 2026-09-09)

A requirement revision may receive a **requirement-review verdict** from a reviewer who did not author it. The review judges **verifiability only** — each obligation gets one closed outcome: `VERIFIABLE` (observable, unambiguous, at least one conceivable evidence route, non-contradictory with its siblings), `AMBIGUOUS`, `UNTESTABLE`, or `CONTRADICTORY`. It never judges business merit — whether the requirement is wise stays with the declared business authority, per rule 1.

The gate is SOFT and derived, never stored: a chain whose targeted obligation lacks a review verdict marking it `VERIFIABLE` renders a `spec: unreviewed` (or `spec: ambiguous` / `untestable` / `contradictory`) component and is capped below green — `SATISFIED` and `DEPLOYED_VERIFIED` are unreachable, while `IMPLEMENTED_UNVERIFIED` remains reachable. Authoring and targeting are never blocked; the gap is visible from the moment a change intent is declared, which is the cheapest moment to fix an unverifiable spec. The rule is provider-independent — it applies to the canonical model, so a requirement authored in any tool meets the same bar, closing the dodge-by-tooling route.

Honesty limit, same boundary as rule 1: reviewer independence is DECLARED. Equal declared author and reviewer identities fail closed; authenticated independence waits for the approval feed. The hard form — review verdicts required, not merely capping — applies only to Proofbound's own P5 requirements (Task 8 self-hosting). Rubber-stamping has its own falsifier (Measurements, falsifier 7).

## Proposed artifact contracts

The family SPEC and each provider SPEC must pin the exact formats before implementation. The recommended bootstrap format is UTF-8 Markdown with one strict JSON metadata block, parsed with the Go standard library. This avoids a new YAML dependency and keeps explanatory prose adjacent to structured meaning. Unknown, duplicate, null, missing, out-of-order where required, or malformed fields fail closed. These sections define BOTH the native provider's on-disk file format AND the canonical payload minimums every provider emits: a foreign mapper produces the same canonical fields — `schema` names the canonical payload version, while `artifact_path` and `artifact_sha256` bind the SOURCE file it mapped from. Nothing is fabricated to complete a record: a segment the source does not author stays absent, and the chain renders the gap.

### Business decision v1

Minimum structured fields:

- `schema: proofbound.business-decision.v1`
- `decision_id`
- `status`: `proposed`, `accepted`, `superseded`, or `withdrawn`
- `outcome`
- `declared_owner`
- `declared_approver`
- `decided_at`
- `supersedes`: optional exact BD revision references
- `artifact_path`
- `artifact_sha256`

The body explains context, alternatives, consequences, success measures, and revisit triggers.

### Requirement v1

Minimum structured fields:

- `schema: proofbound.requirement.v1`
- `requirement_id`
- `status`: `proposed`, `active`, `superseded`, or `retired`
- one or more exact `authorized_by` BD revision references
- `declared_owner`
- `obligations`: stable ID, statement, and state (`active` or `retired`)
- `supersedes`: optional exact BR revision references
- `artifact_path`
- `artifact_sha256`

The body may contain examples and rationale, but P5 does not compile prose into tests or temporal logic. Behavior locks, example ledgers, ambiguity measurement, and risk envelopes remain later Intent Fabric work.

### Change intent v1

Minimum structured fields:

- `schema: proofbound.change-intent.v1`
- `intent_id`
- `status`: `proposed`, `accepted`, `superseded`, or `withdrawn`
- `declared_sponsor`
- typed exact-revision requirement targets
- obligation IDs for each target
- optional exact VD constraints
- `artifact_path`
- `artifact_sha256`

One CI may target several obligations and several commits may claim the same CI. A CI must not name the commit that first introduces it; avoiding that circular reference is deliberate.

### Commit claim v1

The explicit Git trailer is:

```text
Intent: CI-<slug>-<suffix>
```

At ingestion, the Git adapter resolves the named CI from that commit's own tree, validates it, and records its exact artifact digest. An arbitrary ID-shaped mention is not a claim. Multiple trailers are sorted and de-duplicated. A trailer naming a missing, malformed, withdrawn, or ambiguous CI fails closed.

### Obligation verdict v2

Extend, do not reinterpret, `vera.verdict.v1`. The v2 artifact adds:

- exact reviewed commit
- exact CI revision
- exact BR revisions
- one outcome per targeted obligation
- evidence event IDs considered for each outcome
- optional finding IDs and defect commit as today

An `ACCEPTABLE` aggregate verdict is invalid unless every targeted obligation is `SATISFIED`. `NEEDS_WORK` may contain any mixture but must not hide individual outcomes. Cited evidence events must exist at a ledger sequence at or before the verdict's own observation — no future or floating evidence.

### Requirement review v1

Minimum structured fields:
- `schema: proofbound.requirement-review.v1`
- `requirement_id` plus the exact `artifact_sha256` of the reviewed requirement revision
- `declared_reviewer` — must differ from the reviewed revision's declared owner/author; equal declared identities fail closed
- one closed outcome per obligation id (`VERIFIABLE`, `AMBIGUOUS`, `UNTESTABLE`, `CONTRADICTORY`)
- optional findings prose per obligation
- `artifact_path` and its own `artifact_sha256`

Committed on receipt like every verdict. Outcomes bind to the exact reviewed revision and never upgrade a successor revision.

## Proposed kernel architecture

### Core registry

Add registered kinds and sources only after their SPEC amendments:

- kinds: `business_decision.recorded`, `requirement.recorded`, `change_intent.recorded`, `requirement.reviewed`
- sources: `intent.<provider>` — one per configured intent provider (P5: `intent.records`, `intent.specdir`)

Do not add `approval.observed` without an authenticated approval feed. Existing event-envelope, canonicalization, idempotency, and revision semantics remain unchanged.

### The intent-provider connector family (internal/connector/intent/...)

One connector family owns intent ingestion; each provider implements the Intent Provider contract and its conformance suite. The reference provider is repo-native records (below, unchanged in substance); the second is the file-native specdir mapper, whose SPEC names the exact directory convention(s) it parses [SPEC DECISION on the convention variant].

#### Provider: repo-native records

Owns strict parsing and event creation for committed BD, BR, and CI artifacts. Its injected reader must expose committed revisions, not arbitrary working-tree bytes.

Recovery requirement: a fresh clone must be able to reconstruct every artifact revision referenced by an ingested commit. Recovery design, RULED in round 1: accepted revisions remain distinct committed files, each linking its predecessor (append-only revision artifacts) — current-tree scanning suffices. Historical Git blob traversal was REJECTED: more machinery, and it would not generalize to foreign providers. The general guarantee is the replay-sufficiency invariant (Sources-of-intent section): meaning reconstructs from the ledger alone for every provider. The implementation must not allow in-place mutation to erase a referenced revision — the conformance suite carries a fixture proving that fails closed.

#### Provider: specdir (file-native foreign mapper)

Committed spec-directory files map to BR-equivalents (requirements plus acceptance criteria → obligations with stable ids) and optionally CI-equivalents. No BD is fabricated — the chain renders `authorization: undeclared`. Mapping failures fail closed. A mapping that is impossible without a canonical-schema change is a STOP condition (F9 falsifier), never a schema bend.

### Git connector revision

Preserve `cited_decisions` for backward compatibility, and add a versioned `intent_refs` payload field containing CI ID and artifact digest. Bump the connector wire version. Pinned payload vectors, rewrite tests, projection validation, and mutation calibration must be updated.

Trailer grammar gains an optional provider qualifier (F6): unqualified `Intent: CI-<slug>-<suffix>` resolves ONLY against the commit's own tree (the ruled tree-resolution semantics, unchanged); qualified `Intent: <provider>:<record-id>` resolves per that provider's point-in-time rule, and the RESOLVED artifact digest is recorded in the event — the resolution is the observation. Unknown provider prefix, ambiguous resolution, or digest mismatch fails closed.

`sync all` ordering must ingest record revisions from ALL configured providers before commits that reference them, or apply both in one transaction. A dangling reference is an error, never a partially trusted projection.

### Reviews connector revision

Parse v1 artifacts unchanged. Add a separate v2 parser and event mapping. Never reinterpret stored v1 payloads using v2 semantics. Validate that all referenced event IDs, commits, record revisions, and obligation IDs exist during projection; malformed or dangling proof chains fail closed. The connector also parses proofbound.requirement-review.v1 artifacts into requirement.reviewed events: a dangling requirement-revision reference, an outcome for an obligation not present in the bound revision, or equal declared author and reviewer identities fails closed.

### Projections

Add rebuildable, proof-bearing projections using ordinary relational tables:

- `business_decisions_view`
- `requirements_view`
- `requirement_obligations_view`
- `change_intents_view`
- `intent_targets_view`
- `commit_intents_view`
- `obligation_verdicts_view`
- `requirement_reviews_view`

Every row retains originating event ID and ledger sequence. Revision selection is ledger-ordered and must never rewrite historical relations. A separate derived report assembles the chain; the tables are not a hand-maintained graph.

### CLI and report

Add:

```text
proofbound sync intent [<provider>|all]
proofbound report intent <intent-id>
proofbound report requirement <requirement-id>
proofbound intent check --commit <sha>
```

Reports render exact record revisions, declared authority, commit proof, evidence event IDs, obligation outcomes, deployment proof, freshness, and gaps. `proofbound intent check` initially validates only commits that declare `Intent:` trailers.

Derived chain states are:

- `DECLARED`: valid CI exists, no commit claim observed
- `IMPLEMENTED_UNVERIFIED`: commit claims the CI, no complete verdict
- `SATISFIED`: every target obligation is satisfied for the exact commit and revisions, and every targeted obligation's requirement revision carries a non-author review verdict marking it VERIFIABLE
- `NOT_SATISFIED`: at least one target obligation is not satisfied
- `INCONCLUSIVE`: no failure, but at least one conclusion is inconclusive
- `DEPLOYED_UNVERIFIED`: deployment observed without complete satisfaction
- `DEPLOYED_VERIFIED`: deployment observed for a satisfied commit
- `SUPERSEDED`: relevant business record, requirement, intent, or commit was superseded

Reports must show component states rather than collapse contradictory deployments or multiple commits into one misleading status.

Reports additionally render a per-requirement spec-review component — verifiable / ambiguous / untestable / contradictory / unreviewed — derived, never stored. Any non-verifiable component caps the chain below green while blocking nothing (semantic rule 8).

### Gates

Introduce gates in this order:

1. `intent-reference-integrity`: every declared `Intent:` resolves to an exact committed CI revision and every CI target resolves to exact active BR obligations.
2. `intent-verdict-integrity`: every obligation outcome references existing evidence and matches the exact commit and record revisions.
3. `intent-delivery-readiness`: for explicitly scoped commits, all targeted obligations are satisfied before a delivery command proceeds.

Each begins in canary against historical events. The first two can graduate within P5 after bad fixtures prove they block. The third graduates ONLY where a delivery boundary Proofbound actually controls calls it — in P5 that is its own `make delivery-enforce`, via self-hosting (F7); for any external consumer it remains canary until that consumer's pipeline calls the gate. Observing a deployment after the fact is never prevention.

Gate 1's integrity scope includes provider conformance: a configured provider failing its conformance suite is a blocking integrity failure, not a warning.

## Task sequence and mechanical acceptance

No task begins until the round-1 findings are dispositioned (done in this v2), the semantic VD is accepted by the human, and the roadmap is explicitly amended. (Task 0 and Task 0.5 are exempt — they produce that ratification; see Authorization boundary.)

| # | Task | Mechanical Definition of Done |
|---|---|---|
| 0 | Ratify semantics and threat model | VD resolves the round-1 adjudication findings (F1–F10); names claim/observation/verdict boundaries, authority limit, record types, revision rule, relation vocabulary, obligation identity rule, compatibility policy, and falsifiers; `ROADMAP.md` gains the accepted phase DoD; the same change fixes the stale README remaining-work list or records why not (F10) |
| 0.5 | Identity migration to Proofbound | The DoD in the Identity-migration section: classified zero-unclassified vera inventory, byte-exact pinned vectors, green gate, dated alias advisory |
| 1 | Freeze artifact schemas, the Intent Provider contract, and fixtures | `internal/connector/intent/SPEC.md` (family) + per-provider SPECs precede code; exact valid vectors and hostile fixtures cover malformed UTF-8, duplicate/unknown fields, digest mismatch, traversal, ID/path mismatch, invalid lifecycle transitions, missing revisions, dangling relations, and obligation renumbering; plus the provider-conformance fixture classes: cross-provider id collision, provider-prefix spoofing, canonicalization instability (one record, two digests), upstream-mutation digest mismatch, cross-provider sync-ordering, cross-provider supersedes, and an intentionally unmappable foreign fixture that must STOP rather than bend the schema; artifact-digest and payload-digest checks exercised independently (removing either comparison is a killed mutant); requirement-review v1 schema frozen with fixtures: author-equals-reviewer fails closed, wrong-revision digest binding fails closed, unknown outcome value fails closed, outcome-for-absent-obligation fails closed |
| 2 | Implement committed records connector | `proofbound sync intent records` ingests valid BD/BR/CI revisions, re-ingest appends zero, changed revisions append exactly one, uncommitted bytes are ignored, and a fresh clone/rebuild recovers every referenced revision; the replay-sufficiency probe passes: drop projections, fresh clone, and every relation and obligation statement reconstructs from the ledger alone (F4) |
| 2b | Implement the specdir foreign mapper | The same conformance suite passes against specdir fixtures; BR-equivalents map with stable obligation ids; no BD is fabricated (chain renders authorization: undeclared); the unmappable fixture STOPs; zero canonical-schema changes were required to complete the mapping |
| 3 | Bind commits to exact CI revisions | Only explicit `Intent:` trailers create claims; arbitrary mentions do not; missing/malformed/withdrawn refs fail closed; amend/rebase/branch-switch behavior remains correct; old `cited_decisions` reports remain readable; provider-qualified trailers resolve per provider point-in-time rules and record the resolved digest; unknown-provider and spoofed-prefix trailers fail closed |
| 4 | Project and report provenance chain | Incremental and from-genesis row sets match; every row/report segment carries event proof; reports distinguish missing, contradictory, superseded, and unverified states; missing proof fails closed; the spec-review component derives correctly — unreviewed and non-verifiable outcomes cap the chain below green and block nothing |
| 5 | Add obligation verdict v2 + requirement review v1 | v1 remains byte-semantically compatible; v2 validates exact commit/CI/BR/evidence refs; aggregate acceptance cannot coexist with a non-satisfied obligation; builder-authored evidence alone cannot produce satisfaction; cited evidence must precede the verdict in ledger sequence; requirement.reviewed events parse and bind exact requirement revisions; equal declared author/reviewer fails closed |
| 6 | Join observed deployments | Intent report joins GitHub deployment events by exact commit; multiple environments and revisions remain distinct; stale/missing deployment data is explicit; `DEPLOYED_UNVERIFIED` fixture is visible and non-green |
| 7 | Canary and promote integrity gates | Historical canary runs recorded; intentionally bad intent and verdict chains produce proof-bearing BLOCKED results; good chains PASS; integrity gates promote only after independent acceptance |
| 8 | Self-host and independently verify | Proofbound's own P5 change has a BD/BR/CI chain, exact commit binding, evidence, v2 independent verdict, and report; package mutation sweeps are calibrated and green; a non-author independent verdict is committed; bare `make check` and `proofbound verify` pass; `intent-delivery-readiness` is wired into `make delivery-enforce` and demonstrably BLOCKS an intentionally bad self-hosted chain there (external-consumer posture remains canary); every requirement in Proofbound's own P5 chain carries a non-author review verdict with all targeted obligations VERIFIABLE (the hard form applies to self-hosting only) |

## Required test classes

Beyond happy paths, the package SPECs must derive tests for:

- Same ID, different artifact bytes; old commits retain the old meaning.
- Requirement text changes without an obligation-ID change.
- Obligation ID is removed or reused.
- CI refers to an inactive, missing, or wrong-digest requirement.
- Commit body mentions a CI without an `Intent:` trailer.
- Commit trailer resolves in the working tree but not in that commit's tree.
- Historical commit references a revision no longer at the current path.
- Verdict cites evidence for another commit or another requirement revision.
- Verdict says aggregate `ACCEPTABLE` while one obligation is inconclusive.
- Evidence event is absent, malformed, builder-authored, or not admitted by policy.
- Deployment points to a superseded or unverified commit.
- Same CI is implemented by multiple commits and deployed to multiple environments.
- Rebuild after projection deletion produces canonical row-set equality.
- Mutation operators that remove digest, commit, revision, obligation, and event-ID comparisons are killed by discriminating tests.
- Cross-provider record-id collision is rejected by provider-scoped namespacing.
- A trailer naming an unconfigured or spoofed provider prefix fails closed.
- The same upstream record canonicalizes to one digest across runs (canonicalization stability, pinned by vector).
- An upstream record mutated between resolution and ingest fails closed on digest mismatch.
- `supersedes` across providers preserves both lineages.
- The intentionally unmappable foreign fixture stops the phase rather than bending the canonical schema.
- Drop-projections plus fresh-clone replay reconstructs every relation and obligation statement from the ledger alone.
- A verdict citing an evidence event with a later ledger sequence than the verdict's observation fails closed.
- Records from one provider referenced by commits before that provider's sync fail closed until ordering is satisfied (cross-provider sync-ordering).
- A review verdict bound to a superseded requirement revision remains valid for that exact revision and never upgrades its successor.
- Equal declared author and reviewer identities on a requirement review fail closed.
- A chain green in every other respect renders capped, not green, when one targeted obligation is unreviewed.
- A review outcome naming an obligation absent from the bound revision fails closed.

## Migration and compatibility

- Ledger migration remains append-only; existing events are never rewritten.
- Existing `commit.recorded` payloads without `intent_refs` remain valid historical v1 events and render `intent=undeclared`, not malformed.
- Existing `vera.verdict.v1` artifacts and events retain their current meaning. They do not acquire obligation satisfaction retroactively.
- Existing `VD-*` citations remain engineering-decision references only.
- Projection schema version increments; rebuild, rather than data patching, creates new derived rows.
- No retroactive claim that historical commits had business intent. A baseline may link surviving requirements prospectively and mark earlier behavior `UNVERIFIED`.

## Non-goals

- No automatic extraction of requirements from tickets, meetings, commit messages, or prose.
- No claim that a declared approver field proves identity or authorization.
- No live API connector to a mutable upstream in P5 — file-native providers only; the provider contract pins how snapshot-ingestion providers will work, and building one waits for a real feed.
- No vendor-named coupling in kernel code, schemas, or sources — provider identifiers are generic convention names bound to real systems only in configuration.
- No natural-language-to-test compiler, temporal-logic compiler, behavior lock, example ledger, or ambiguity tournament in P5.
- No new graph database or mutable relationship store.
- No cryptographic signing unless a separately decided identity and key feed exists.
- No universal “every commit requires intent” gate without a ratified applicability mechanism.
- No production-blocking claim until the gate is called by the actual deployment boundary.
- No inference that passing checks means requirements are satisfied.
- No merit judgment in requirement reviews — verifiability only; business wisdom stays with the declared authority.
- No authenticated reviewer identity in P5 — independence is declared, failing closed only on equal declared identities.

## Risks and mitigations

| Risk | Required mitigation |
|---|---|
| Approval theater | Label repo metadata as declared authority; reserve verified approval for authenticated feeds |
| Second-artifact drift | Exact content digests on every relation; append-only revisions; derive every report from ledger events |
| Requirement prose becomes worse code | P5 keeps obligations human-readable and verifier-evaluated; compilation waits for behavior-lock design |
| Process tax causes bypass | Intent requirement is scoped, canary measured, and not universal in P5; measure adoption and false-block rate |
| Self-classification becomes an escape hatch | Do not enforce a universal gate from author-selected class; applicability requires a later independent policy |
| Current-tree revision loss | Append-only revision files (historical traversal rejected in round 1); replay-sufficient payloads make meaning-recovery provider-independent; fresh-clone recovery is acceptance-critical |
| Evidence laundering | Verdict owns outcomes and names evidence IDs; evidence cannot assert its own satisfaction |
| Aggregate green hides gaps | Per-obligation closed outcomes; missing is derived UNVERIFIED; fail-closed aggregation |
| Revision joins create false matches | Every relation includes kind, ID, and digest; mutation tests remove each comparison independently |
| Observability is marketed as prevention | Deployment reports say “observed”; blocking is claimed only when wired into the delivery boundary |
| A foreign format warps the canonical model | The conformance suite plus the STOP falsifier: a mapping that needs a canonical-schema change mid-phase halts the phase for re-ratification; the model is never bent to a format |
| Mutable upstream rewrites its own history | API-backed providers deferred; when built, snapshot ingestion only — the ledger's observed sequence is proof, upstream revision metadata is a claim |
| Spec review becomes rubber-stamp theater | Closed per-obligation outcomes; findings-per-review measured; falsifier 7 redesigns or retires the gate |

## Measurements and falsifiers

P5 is not successful merely because the schema works. Record these during the self-hosting trial:

- Human minutes to author and approve BD, BR, and CI artifacts.
- Percentage of fields changed during independent review.
- Number of ambiguous obligations found before implementation.
- Number of false or missing commit-to-intent links.
- Number of verdict/evidence mismatches caught mechanically.
- False-block and false-pass counts during gate canary.
- Report reconstruction time from an empty projection store.
- Per provider: records mapped, records rejected by conformance, false-block rate.
- Review findings per requirement review (a sustained zero across many reviews, while downstream INCONCLUSIVE verdicts still occur, is the rubber-stamp signal).

Revisit or simplify the design if any of these occur:

1. The three-artifact chain takes longer to maintain than the verification evidence it replaces for three consecutive changes.
2. Independent reviewers cannot distinguish BD, BR, and CI authorship or repeatedly duplicate the same datum across them.
3. Obligation outcomes remain subjective labels without discriminating evidence in three accepted changes.
4. Fresh-clone reconstruction cannot recover the exact meaning attached to an old commit.
5. Canary cannot identify a useful applicability boundary without trusting the builder's own label.
6. Completing a foreign-provider mapping required changing the canonical schema — the canonical model is wrong; stop and re-ratify rather than bend it.
7. Requirement reviews rubber-stamp — near-zero findings across many reviews while downstream INCONCLUSIVE outcomes persist: the review is theater; redesign or retire the gate.

## Round-1 adjudication — dispositioned

The ten questions were adjudicated in round 1 — full answers in p5-adjudication-round1.md; one-line dispositions:

1. BD → BR → CI is minimal and correct as the canonical ledger model, but not as an authoring mandate: providers supply what they author and partial chains render as gaps.
2. Immutable append-only revision files for the native provider; historical blob traversal rejected because it is more machinery and would not generalize to foreign providers.
3. The CI artifact stays in P5 — sponsorship, multi-commit scope, revision-targeted obligations, and now the cross-source join point.
4. Declared vs verified is the right boundary; externally-sourced fields are equally DECLARED even when the upstream authenticates them.
5. Verdict v2 is honest with three SPEC invariants: the verdict owns outcomes, builder-authored evidence can never set `SATISFIED`, and cited evidence must exist at or before the verdict's ledger sequence.
6. The tri-state is sufficient; `UNVERIFIED` stays derived-never-stored and no fourth authored state is added.
7. Backward compatibility preserves historical meaning, with the frozen/live split made explicit and legacy commits rendering `intent=undeclared`.
8. The only consuming delivery boundary in P5 is Proofbound's own `make delivery-enforce`; external consumption stays canary until a real pipeline calls the gate.
9. The missing test family is provider conformance — id collision, prefix spoofing, canonicalization instability, upstream-mutation mismatch, cross-provider ordering and supersedes, the unmappable fixture, and the replay-sufficiency probe.
10. Smallest defensible step with exactly two providers and API-backed providers deferred; one provider leaves the abstraction unproven, three or a live API provider is premature.

| Finding | Severity | Where folded in this v2 |
|---|---|---|
| F1 — intent is multi-source | HIGH | Sources-of-intent section; Tasks 1 / 2 / 2b / 3; required test classes; risks; non-goals |
| F2 — identity migration not operationalized | HIGH | Identity-migration section; Task 0.5 |
| F3 — two digests, one rule | MED | Two-digest rule in Sources-of-intent; Task 1 fixtures |
| F4 — replay self-sufficiency across providers | MED | Replay self-sufficiency in Sources-of-intent; Task 2 DoD |
| F5 — source naming | MED | Core registry sources (`intent.<provider>`) |
| F6 — trailer resolution across providers | MED | Git-connector trailer grammar; Task 3 |
| F7 — gate 3's real boundary | MED | Gates section (graduation rule) |
| F8 — verifier naming | LOW | Verifier-neutral wording throughout |
| F9 — provider falsifier | LOW | Measurements and falsifiers; unmappable-fixture STOP condition |
| F10 — baseline hygiene | LOW | Task 0 DoD |

A founder-ratification round (2026-09-09) followed round 1 and added semantic rule 8 (requirement review, soft-gate form) plus the four approved tradeoffs — provenance in p5-founder-ratification.md; rule 8 is a founder decision, not an adjudication finding.

## Authorization boundary

Independent-verifier validation may change any proposed semantic or execution detail. After its verdict:

1. Round-1 verdict preserved (p5-adjudication-round1.md) and its findings dispositioned in this plan — DONE.
2. Founder ratification recorded 2026-09-09 (p5-founder-ratification.md): two real providers; identity migration as Task 0.5; partial chains first-class; requirement review in soft-gate form — DONE.
3. Commit both artifacts on receipt, then mint the semantic VD citing them.
4. Amend `ROADMAP.md` with the accepted P5 DoD.
5. Begin Task 0.5, then the task sequence, under spec-first and independent-verification rules.
