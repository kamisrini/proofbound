# P5 Proposed Execution Plan — Intent Provenance
**Status:** DRAFT FOR FABLE VALIDATION — NOT AUTHORIZED FOR IMPLEMENTATION  

**Proposed phase:** P5, after the accepted P4 twin spike  

**Decision required before execution:** a ratified VD must adopt or amend the semantic model in

Task 0; this draft does not amend `ROADMAP.md` and is not itself a product decision.  

**Audience:** an independent product/architecture verifier first, then a fresh implementation

session after explicit human approval.

## Fable handoff — repository state before P5

This section is the minimum self-contained context for review on a machine with no prior session

history. It describes the accepted implementation through P4; everything later in this document is

a P5 proposal, not existing behavior. The snapshot used to prepare this handoff is `main` at

`786d6e1` (2026-08-27). Verify the actual checkout with `git status --short --branch` and

`git log -1 --oneline` because a transferred checkout may be newer or contain local work.

### Product and trust model already established

Proofbound (called VERA in the older internal documents and CLI) is a warranty layer for software

delivery. Its kernel records immutable observations in an append-only PostgreSQL event ledger,

derives disposable/rebuildable projections, and renders proof-bearing reports. The governing

boundary is intentionally strict:

- connector observations, authored claims, and independent verdicts are different kinds of fact;

- derived tables are never a second writable truth and must rebuild to the same row set;

- ingestion is idempotent on source-native identity plus content identity;

- evidence and verdicts bind to exact commits/artifact bytes rather than mutable names or positions;

- new primitives require a real feed, and accepted packages require both mechanical evidence and an

  independent non-author verdict.

The load-bearing product decisions are indexed in `docs/decisions/INDEX.md`. Start with

`VD-north-star-6io56h`, `VD-stack-go-fid9mi`, `VD-commit-is-durability-6p29qf`,

`VD-verdicts-are-artifacts-rl0rab`, `VD-verification-asymmetry-2dyjnd`, and

`VD-no-graph-2aa4vz`. `CLAUDE.md` contains repository operating laws; `ROADMAP.md` is phase status;

`notes/state.md` is the canonical resume note. If those sources disagree, do not silently reconcile

them in code—identify the contradiction in the review.

### Accepted implementation through P4

| Phase | What exists now | Primary acceptance evidence |

|---|---|---|

| P0 | Repository laws, decision records, generated indexes/locks, check scripts, hooks self-test, and the `make check` gate. | `ROADMAP.md`; `docs/gates.md` |

| P1 | Go kernel; PostgreSQL migrations and append/read/transaction store; Git, check-witness, session, and review-verdict connectors; rebuildable projections; weekly proof report; ledger-ordered red-verdict/change/next-verdict chain; blocking package-SPEC coverage. | `docs/plans/P1-flight-recorder-plan.md`; `docs/verification/task6-final-evidence.md` through `task9-final-evidence.md`; accepted verdicts under `docs/verification/verdicts/` |

| P2 | Seven YAML gate definitions; ledger-backed canary evaluation; explicit fail-closed enforcement; expiry checks; serialized `make delivery-enforce` workflow that refreshes witnesses, ingests them, then enforces. | `docs/verification/p2-gate-evidence.md` |

| P3 | Bounded read-only GitHub Actions/deployments connector for an owner plus repository allowlist; exact-commit joined tested-what/deployed-where projection; freshness and event proof in the report. Live acceptance against public `github/docs` synced 200 records in 7 seconds, rendered 106 groups, and replayed with 0 new events in 4 seconds. | `docs/decisions/VD-p3-github-connector-2026-08-26.md`; `docs/verification/p3-github-live-acceptance.md` |

| P4 | `internal/twin` bounded replay; candidate-sequence validation; disposable embedded-PostgreSQL projection; deterministic `vera.replay.v1` proof including payload bytes; source-ledger preservation and cleanup on failure; pure in-memory forecast calibration. The accepted mutation sweep killed 31/31 candidates. | `docs/decisions/VD-p4-twin-replay-calibration-2026-08-26.md`; `docs/verification/verdicts/p4-current-round1.md` |

P4 does **not** persist prediction events. That remains deferred until a real prediction feed has a

separate decision and acceptance evidence. Nothing currently implements business decisions (`BD`),

behavioral requirements (`BR`), change intents (`CI`), commit claims, or obligation-level verdicts;

those are the new semantic objects proposed below.

### Current code and runtime surfaces

- `kernel/internal/core`: event kinds, identities, validation, and clocks.

- `kernel/internal/store`: PostgreSQL ledger, migration `migrations/001_ledger.sql`, append/read,

  locking, and transaction boundaries. Tests can use embedded PostgreSQL or an external

  `DATABASE_URL`.

- `kernel/internal/connector/{git,checks,sessions,reviews,github}`: existing observation adapters.

- `kernel/internal/projections`: incremental application, destructive rebuild from the ledger,

  snapshots, weekly/review chains, and GitHub delivery reporting.

- `kernel/internal/gates`: YAML gate loading, canary/enforce evaluation, expiry, and ledger proof.

- `kernel/internal/twin`: isolated replay and calibration spike accepted in P4.

- `kernel/cmd/vera`: CLI entry point. Current surface is

  `vera sync {git|checks|sessions|reviews|github|all}`, `vera rebuild`, `vera verify`,

  `vera report {week|github}`, and `vera gates {canary|enforce}`.

Local runtime state lives under `.vera/` and is not authoritative source code. Check witnesses are

strict `vera.witness.v1` JSON spool artifacts; review files use `vera.verdict.v1`; replay proof uses

`vera.replay.v1`. Preserve backward compatibility for all three when evaluating the P5 schemas.

### Reproducing the baseline on another machine

Prerequisites are Git, Bash, Go 1.26 or newer, `golangci-lint`, and PostgreSQL. Docker with

`postgres:16-alpine` has been used for external-database integration and mutation runs. From the

repository root:

```bash

git status --short --branch

git log -1 --oneline

make check

```

`make check` is the full blocking repository gate. Run it bare so its exit status is not hidden by a

shell pipeline. `make short` currently runs only the hooks self-test and is not evidence that Go

tests passed. `make mutants` is the slower package-acceptance mechanism and is intentionally outside

`make check`.

For a ledger-backed verification run, provide a disposable PostgreSQL database through

`DATABASE_URL`, create a fresh witnessed check with `make check-witnessed`, then run `make verify`.

For the promoted delivery boundary use `make delivery-enforce`. GitHub sync additionally requires

`VERA_GITHUB_OWNER` and `VERA_GITHUB_REPOS`; it is network-backed and intentionally limited to the

configured repositories and v1 collection bound.

### Known baseline limitations and review cautions

- This is an early-stage CLI/kernel, not a finished end-user product.

- `README.md` still lists P4 work under “Remaining work” even though its next paragraph,

  `ROADMAP.md`, `notes/state.md`, and the committed independent verdict record P4 as accepted. Treat

  that list as stale documentation, not as the phase authority.

- Some imported prose describes checks that are absent from this checkout. Only visible scripts,

  their self-tests, kernel build/test/lint, YAML gates, and committed evidence count as mechanisms.

- `make backup` is absent; `make short` is incomplete; state-freshness, skip-lint, prescription-lint,

  and full historical invariant-citation resolution remain mechanism debt.

- Session ingestion was accepted with synthetic fixtures because no real session JSONL corpus was

  available. Do not upgrade that evidence into a claim of live-source acceptance.

- The GitHub connector is read-only and deliberately narrow; unmatched workflow/deployment records

  render as missing instead of being inferred across commit SHAs.

- Projection tables may be dropped and rebuilt. P5 must extend the ledger-first model rather than

  make a projection or graph database authoritative.

- Acceptance means the cited frozen revision and evidence passed the repository's rules. It does

  not make hand-authored authority authentic, prove business intent, or prove more than the tested

  routes and mutation operators cover.

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

P5 does **not** prove that a business decision was wise, that prose perfectly captured what the

business meant, or that a hand-authored approver field is authentic. It records claims as claims,

observations as observations, and verdicts as verdicts.

## Why separate record types are required

The current `VD-*` records answer engineering questions: why a stack, boundary, or build rule was

chosen. A business decision, a behavioral requirement, and a change's implementation claim have

different authors, lifetimes, revision rules, and proof semantics. A generic `type` label on the

current VD template would still leave those meanings conflated.

P5 introduces separate semantic artifacts while reusing one ingestion and ledger mechanism:

| Artifact | Meaning | Typical authority | Changes at |

|---|---|---|---|

| Engineering decision (`VD`) | Why an engineering constraint was chosen | Technical owner | Architecture-policy cadence |

| Business decision (`BD`) | Why a business outcome or constraint was authorized | Named business authority | Business-decision cadence |

| Requirement (`BR`) | What observable obligations must hold | Domain owner delegated by a BD | Behavior cadence |

| Change intent (`CI`) | What one delivery claims to do to exact BR obligations | Change sponsor/builder | Delivery cadence |

| Verdict | Whether evidence supports each obligation for an exact commit | Independent verifier | Verification cadence |

These remain files during bootstrap, consistent with `VD-no-graph-2aa4vz`. The ledger stores

immutable observations of committed artifact revisions; projections and joins are derived. No graph

database and no second writable truth are introduced.

## Semantic rules to ratify before code

### 1. A record is not automatically a fact

The connector can observe that committed bytes exist and who Git says authored the commit. It

cannot infer that the record is correct or that a named person approved it. Event names and reports

must preserve this boundary:

- `business_decision.recorded` means an artifact revision was observed.

- `requirement.recorded` means obligations were declared in an artifact revision.

- `change_intent.recorded` means a delivery claim was declared.

- `approval.observed` is reserved for an authenticated approval source added later.

- `review.verdict` remains a verifier conclusion, not an observed law of nature.

Until an authenticated approval connector exists, repo-native `owner` and `approved_by` fields are

displayed as **declared authority**, never “verified approval.” Git signature verification is not in

P5 unless separately decided and fed by a real identity source.

### 2. Every relation binds an exact revision

An ID alone is insufficient because record contents can change. Every relation carries:

```json

{

  "record_kind": "requirement",

  "record_id": "BR-<slug>-<suffix>",

  "artifact_sha256": "<64 lowercase hex>",

  "relation": "implements",

  "obligation_ids": ["O-1", "O-2"]

}

```

`artifact_sha256` identifies the exact canonical artifact bytes. A later revision produces a new

ledger event and cannot reinterpret an old commit, verdict, or deployment.

### 3. Relations are typed

Closed relation values for P5:

- A requirement is `authorized_by` a business decision.

- A change intent `implements`, `modifies`, `repairs`, or `retires` requirement obligations.

- A change intent may be `constrained_by` an engineering decision.

- A commit `claims` a change intent.

- A verdict `evaluates` requirement obligations for a commit and cites evidence events.

The existing untyped `cited_decisions` field remains historical compatibility data. It must not be

presented as business intent.

### 4. Obligations have stable identities

Each requirement revision contains one or more obligation IDs with testable statements. IDs are

append-only within the requirement lineage. Changed meaning gets a new obligation ID; a retired

obligation leaves a tombstone. Rewording that demonstrably preserves meaning may retain the ID but

still produces a new artifact revision.

### 5. Requirement satisfaction is tri-state, per obligation

Closed P5 verdict outcomes:

- `SATISFIED`

- `NOT_SATISFIED`

- `INCONCLUSIVE`

Absence of a verdict or evidence is `UNVERIFIED`, derived by the projection and never stored as a

verdict. No aggregate may render green while an applicable obligation is unverified,

inconclusive, contradicted, bound to another revision, or evaluated for another commit.

### 6. Evidence does not grade itself

A check witness may state which obligation it was designed to exercise, but that association is an

authored claim. The independent verdict names the evidence event IDs it considered and owns the

obligation outcome. Builder-produced evidence cannot directly set `SATISFIED`.

### 7. Applicability is not guessed in P5

P5 does not infer business intent from filenames, diff contents, commit subjects, or an AI summary.

It validates explicit `Intent:` commit trailers. The first enforcement gate checks the integrity of

declared intent; it does not require every repository commit to carry business intent. A later phase

may add a separately ratified applicability policy for behavior-changing work.

## Proposed artifact contracts

The package SPEC must pin the exact wire format before implementation. The recommended bootstrap

format is UTF-8 Markdown with one strict JSON metadata block, parsed with the Go standard library.

This avoids a new YAML dependency and keeps explanatory prose adjacent to structured meaning.

Unknown, duplicate, null, missing, out-of-order where required, or malformed fields fail closed.

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

The body may contain examples and rationale, but P5 does not compile prose into tests or temporal

logic. Behavior locks, example ledgers, ambiguity measurement, and risk envelopes remain later

Intent Fabric work.

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

One CI may target several obligations and several commits may claim the same CI. A CI must not name

the commit that first introduces it; avoiding that circular reference is deliberate.

### Commit claim v1

The explicit Git trailer is:

```text

Intent: CI-<slug>-<suffix>

```

At ingestion, the Git adapter resolves the named CI from that commit's own tree, validates it, and

records its exact artifact digest. An arbitrary ID-shaped mention is not a claim. Multiple trailers

are sorted and de-duplicated. A trailer naming a missing, malformed, withdrawn, or ambiguous CI

fails closed.

### Obligation verdict v2

Extend, do not reinterpret, `vera.verdict.v1`. The v2 artifact adds:

- exact reviewed commit

- exact CI revision

- exact BR revisions

- one outcome per targeted obligation

- evidence event IDs considered for each outcome

- optional finding IDs and defect commit as today

An `ACCEPTABLE` aggregate verdict is invalid unless every targeted obligation is `SATISFIED`.

`NEEDS_WORK` may contain any mixture but must not hide individual outcomes.

## Proposed kernel architecture

### Core registry

Add registered kinds and sources only after their SPEC amendments:

- kinds: `business_decision.recorded`, `requirement.recorded`, `change_intent.recorded`

- source: `records` for repo-native records

Do not add `approval.observed` without an authenticated approval feed. Existing event-envelope,

canonicalization, idempotency, and revision semantics remain unchanged.

### `internal/connector/records`

Owns strict parsing and event creation for committed BD, BR, and CI artifacts. Its injected reader

must expose committed revisions, not arbitrary working-tree bytes.

Recovery requirement: a fresh clone must be able to reconstruct every artifact revision referenced

by an ingested commit. The SPEC must choose and prove one of these designs before code:

1. **Append-only revision artifacts (recommended for P5):** accepted revisions remain as distinct

   committed files, and a new revision links to its predecessor. Current-tree scanning is sufficient.

2. Historical Git traversal: the reader enumerates and de-duplicates every referenced blob across

   repository history.

The implementation must not silently assume current-tree scanning is sufficient while allowing

in-place mutation to erase a referenced revision.

### Git connector revision

Preserve `cited_decisions` for backward compatibility, and add a versioned `intent_refs` payload

field containing CI ID and artifact digest. Bump the connector wire version. Pinned payload vectors,

rewrite tests, projection validation, and mutation calibration must be updated.

`sync all` ordering must ingest record revisions before commits that reference them, or apply both in

one transaction. A dangling reference is an error, never a partially trusted projection.

### Reviews connector revision

Parse v1 artifacts unchanged. Add a separate v2 parser and event mapping. Never reinterpret stored

v1 payloads using v2 semantics. Validate that all referenced event IDs, commits, record revisions,

and obligation IDs exist during projection; malformed or dangling proof chains fail closed.

### Projections

Add rebuildable, proof-bearing projections using ordinary relational tables:

- `business_decisions_view`

- `requirements_view`

- `requirement_obligations_view`

- `change_intents_view`

- `intent_targets_view`

- `commit_intents_view`

- `obligation_verdicts_view`

Every row retains originating event ID and ledger sequence. Revision selection is ledger-ordered and

must never rewrite historical relations. A separate derived report assembles the chain; the tables

are not a hand-maintained graph.

### CLI and report

Add:

```text

vera sync records

vera report intent <intent-id>

vera report requirement <requirement-id>

vera intent check --commit <sha>

```

Reports render exact record revisions, declared authority, commit proof, evidence event IDs,

obligation outcomes, deployment proof, freshness, and gaps. `vera intent check` initially validates

only commits that declare `Intent:` trailers.

Derived chain states are:

- `DECLARED`: valid CI exists, no commit claim observed

- `IMPLEMENTED_UNVERIFIED`: commit claims the CI, no complete verdict

- `SATISFIED`: every target obligation is satisfied for the exact commit and revisions

- `NOT_SATISFIED`: at least one target obligation is not satisfied

- `INCONCLUSIVE`: no failure, but at least one conclusion is inconclusive

- `DEPLOYED_UNVERIFIED`: deployment observed without complete satisfaction

- `DEPLOYED_VERIFIED`: deployment observed for a satisfied commit

- `SUPERSEDED`: relevant business record, requirement, intent, or commit was superseded

Reports must show component states rather than collapse contradictory deployments or multiple

commits into one misleading status.

### Gates

Introduce gates in this order:

1. `intent-reference-integrity`: every declared `Intent:` resolves to an exact committed CI revision

   and every CI target resolves to exact active BR obligations.

2. `intent-verdict-integrity`: every obligation outcome references existing evidence and matches the

   exact commit and record revisions.

3. `intent-delivery-readiness`: for explicitly scoped commits, all targeted obligations are

   satisfied before a delivery command proceeds.

Each begins in canary against historical events. The first two can graduate within P5 after bad

fixtures prove they block. The third must remain canary until a real delivery boundary consumes it;

observing a GitHub deployment after it happened is not deployment prevention.

## Task sequence and mechanical acceptance

No task begins until Fable's verdict is incorporated, the semantic VD is accepted, and the roadmap

is explicitly amended.

| # | Task | Mechanical Definition of Done |

|---|---|---|

| 0 | Ratify semantics and threat model | VD resolves the Fable findings; names claim/observation/verdict boundaries, authority limit, record types, revision rule, relation vocabulary, obligation identity rule, compatibility policy, and falsifiers; `ROADMAP.md` gains the accepted phase DoD |

| 1 | Freeze artifact schemas and fixtures | `internal/connector/records/SPEC.md` precedes code; exact valid vectors and hostile fixtures cover malformed UTF-8, duplicate/unknown fields, digest mismatch, traversal, ID/path mismatch, invalid lifecycle transitions, missing revisions, dangling relations, and obligation renumbering |

| 2 | Implement committed records connector | `vera sync records` ingests valid BD/BR/CI revisions, re-ingest appends zero, changed revisions append exactly one, uncommitted bytes are ignored, and a fresh clone/rebuild recovers every referenced revision |

| 3 | Bind commits to exact CI revisions | Only explicit `Intent:` trailers create claims; arbitrary mentions do not; missing/malformed/withdrawn refs fail closed; amend/rebase/branch-switch behavior remains correct; old `cited_decisions` reports remain readable |

| 4 | Project and report provenance chain | Incremental and from-genesis row sets match; every row/report segment carries event proof; reports distinguish missing, contradictory, superseded, and unverified states; missing proof fails closed |

| 5 | Add obligation verdict v2 | v1 remains byte-semantically compatible; v2 validates exact commit/CI/BR/evidence refs; aggregate acceptance cannot coexist with a non-satisfied obligation; builder-authored evidence alone cannot produce satisfaction |

| 6 | Join observed deployments | Intent report joins GitHub deployment events by exact commit; multiple environments and revisions remain distinct; stale/missing deployment data is explicit; `DEPLOYED_UNVERIFIED` fixture is visible and non-green |

| 7 | Canary and promote integrity gates | Historical canary runs recorded; intentionally bad intent and verdict chains produce proof-bearing BLOCKED results; good chains PASS; integrity gates promote only after independent acceptance |

| 8 | Self-host and independently verify | Proofbound's own P5 change has a BD/BR/CI chain, exact commit binding, evidence, v2 independent verdict, and report; package mutation sweeps are calibrated and green; non-author Fable/adversarial verdict is committed; bare `make check` and `vera verify` pass |

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

- Mutation operators that remove digest, commit, revision, obligation, and event-ID comparisons are

  killed by discriminating tests.

## Migration and compatibility

- Ledger migration remains append-only; existing events are never rewritten.

- Existing `commit.recorded` payloads without `intent_refs` remain valid historical v1 events and

  render `intent=undeclared`, not malformed.

- Existing `vera.verdict.v1` artifacts and events retain their current meaning. They do not acquire

  obligation satisfaction retroactively.

- Existing `VD-*` citations remain engineering-decision references only.

- Projection schema version increments; rebuild, rather than data patching, creates new derived rows.

- No retroactive claim that historical commits had business intent. A baseline may link surviving

  requirements prospectively and mark earlier behavior `UNVERIFIED`.

## Non-goals

- No automatic extraction of requirements from tickets, meetings, commit messages, or prose.

- No claim that a declared approver field proves identity or authorization.

- No Jira/Productboard/CRM connector in the first slice; source-neutral interfaces must permit one.

- No natural-language-to-test compiler, temporal-logic compiler, behavior lock, example ledger, or

  ambiguity tournament in P5.

- No new graph database or mutable relationship store.

- No cryptographic signing unless a separately decided identity and key feed exists.

- No universal “every commit requires intent” gate without a ratified applicability mechanism.

- No production-blocking claim until the gate is called by the actual deployment boundary.

- No inference that passing checks means requirements are satisfied.

## Risks and mitigations

| Risk | Required mitigation |

|---|---|

| Approval theater | Label repo metadata as declared authority; reserve verified approval for authenticated feeds |

| Second-artifact drift | Exact content digests on every relation; append-only revisions; derive every report from ledger events |

| Requirement prose becomes worse code | P5 keeps obligations human-readable and verifier-evaluated; compilation waits for behavior-lock design |

| Process tax causes bypass | Intent requirement is scoped, canary measured, and not universal in P5; measure adoption and false-block rate |

| Self-classification becomes an escape hatch | Do not enforce a universal gate from author-selected class; applicability requires a later independent policy |

| Current-tree revision loss | Append-only revision files or proven historical traversal; fresh-clone recovery is acceptance-critical |

| Evidence laundering | Verdict owns outcomes and names evidence IDs; evidence cannot assert its own satisfaction |

| Aggregate green hides gaps | Per-obligation closed outcomes; missing is derived UNVERIFIED; fail-closed aggregation |

| Revision joins create false matches | Every relation includes kind, ID, and digest; mutation tests remove each comparison independently |

| Observability is marketed as prevention | Deployment reports say “observed”; blocking is claimed only when wired into the delivery boundary |

## Measurements and falsifiers

P5 is not successful merely because the schema works. Record these during the self-hosting trial:

- Human minutes to author and approve BD, BR, and CI artifacts.

- Percentage of fields changed during independent review.

- Number of ambiguous obligations found before implementation.

- Number of false or missing commit-to-intent links.

- Number of verdict/evidence mismatches caught mechanically.

- False-block and false-pass counts during gate canary.

- Report reconstruction time from an empty projection store.

Revisit or simplify the design if any of these occur:

1. The three-artifact chain takes longer to maintain than the verification evidence it replaces for

   three consecutive changes.

2. Independent reviewers cannot distinguish BD, BR, and CI authorship or repeatedly duplicate the

   same datum across them.

3. Obligation outcomes remain subjective labels without discriminating evidence in three accepted

   changes.

4. Fresh-clone reconstruction cannot recover the exact meaning attached to an old commit.

5. Canary cannot identify a useful applicability boundary without trusting the builder's own label.

## Questions for Fable to adjudicate

Fable should return `ACCEPTABLE` or `NEEDS_WORK` and address these explicitly:

1. Is BD -> BR -> CI the minimum semantic separation, or does it create a second-artifact/process

system that will drift or be bypassed?

2. Should accepted revisions be immutable files, or should the connector reconstruct historical

blobs? Which choice best satisfies fresh-clone replay and one-home-per-datum?

3. Is a CI artifact necessary in P5, or can an explicit commit/PR claim bind directly to BR

obligations without losing sponsorship, multi-commit scope, or revision history?

4. Does the proposed authority language avoid presenting committed metadata as authenticated

approval?

5. Can verdict v2 honestly establish obligation satisfaction, or does it need an additional

independent claim/evidence relationship?

6. Are `SATISFIED`, `NOT_SATISFIED`, and `INCONCLUSIVE` sufficient without collapsing missing into a

stored state?

7. Does backward compatibility preserve historical meaning without licensing silent gaps?

8. What concrete delivery boundary can consume `intent-delivery-readiness`, and what must remain

canary until that boundary exists?

9. Which invariant classes or attack routes are absent from the required tests?

10. Is this the smallest defensible step toward the vision's behavior lock, or is any component

    premature?

## Authorization boundary

Fable validation may change any proposed semantic or execution detail. After its verdict:

1. Preserve the verdict as an artifact bound to this plan's reviewed commit.

2. Resolve every `NEEDS_WORK` finding in the plan before code.

3. Ask the human to approve the resulting value and scope tradeoffs.

4. Mint the semantic VD and amend `ROADMAP.md`.

5. Only then begin Task 1 under spec-first and independent-verification rules.
