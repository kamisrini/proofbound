# P6 Consolidation Plan — Deep Completion of P1–P5 (Proofbound) — draft 2

**Status:** DRAFT 2 — round-1 build-machine adjudication folded; awaiting founder ratification of
the four inputs in the Authorization boundary. No implementation task is authorized yet.

**Provenance:** preserves draft 1 at SHA-256
`4564e0f0aafd81d88c0d5fc0322bcb5640e61508b07db918c911911c38f3a764` and folds
[`p6-consolidation-plan-round1.md`](../verification/verdicts/p6-consolidation-plan-round1.md),
committed with the reviewed draft in `ceeae16`.

**Doctrine (founder direction, 2026-09-12):** depth before width. P6 completes, hardens,
re-verifies, and measures the promises already made in P1–P5. Behavior locks, regeneration, a
production-shaped twin, governor, portable warranties, marketplace, and end-user UI are P7+.

## Goal and closure meaning

P6 may call P1–P5 “deep-complete” only when every discovered live promise is either:

1. satisfied by cited mechanical evidence; or
2. explicitly superseded, narrowed, or retired by a decision record that updates every live
   authority in the same change.

Merely labeling an unsatisfied promise `defer-to-P7` or `wontfix` does not complete it. The decision
must change the promise; otherwise the row remains open.

## Authorization preflight — before Task 0

These are governance steps, not implementation tasks, and occur in this order:

1. Preserve draft 1 and commit the build-machine round-one adjudication on receipt — DONE in
   `ceeae16`.
2. Fold all adjudication findings into this draft 2 — DONE when this file is committed.
3. Record the founder's four ratification inputs in
   `docs/verification/verdicts/p6-founder-ratification.md`, committed verbatim on receipt.
4. Mint the P6 semantic VD citing the ratification and adjudication; amend `ROADMAP.md` with the
   ratified P6 section in the same commit.
5. Only then begin Task 0.

Task 0 does not amend the roadmap; that contradictory draft-1 requirement is removed.

## Task 0 — closed debt-census contract

Task 0 first commits `docs/plans/p6-census-SPEC.md`, then the census mechanism and its proving
fixtures, then its first result. The mechanism is `scripts/p6-census.sh`; its canonical output is
`docs/plans/p6-census.md`.

### Machine-readable row schema

The output contains exactly one table with these columns:

```text
row_id | category | subject | observation | probe | state | disposition | decision | evidence
```

- `row_id`: stable `C<number>-<zero-padded ordinal>`; never reused.
- `category`: one of C1–C8 below.
- `probe`: an executable command or named test, not prose such as “inspect manually.”
- `state`: `open` or `closed`.
- `disposition`: `close-in-P6`, `defer-to-P7`, or `wontfix`.
- `decision`: `—` only for `close-in-P6`; otherwise one exact accepted VD path.
- `evidence`: `—` while open; when closed, one or more exact artifact paths, test names, event IDs,
  commit IDs, or command-result records.

The script fails on an unknown column value, duplicate row ID, unclassified row, a closed row with
no evidence, a defer/wontfix row with no accepted decision, or a decision that does not update the
live promise it supersedes. `scripts/tests/p6-census.test.sh` proves each failure and a valid pass.

### Closed scan universe

The census does not pretend natural-language understanding is mechanical. It scans these finite
surfaces:

- Live authorities: `CLAUDE.md`, `README.md`, `ROADMAP.md`, `notes/state.md`, `docs/gates.md`, the
  accepted phase plans, accepted VDs, and current package SPECs.
- Historical/frozen exhibits, adjudications, journals, and superseded draft plans are inventoried
  but excluded from current-truth comparison. A live authority quoting a historical baseline must
  delimit or label that quote; otherwise it is a doc-truth row.
- Recognized mechanism tokens: exact backtick-quoted `make <target>` invocations; paths under
  `scripts/`, `kernel/scripts/`, `.claude/hooks/`, `.github/workflows/`, `gates/`, and
  `kernel/internal/`; gate IDs from `gates/*.yaml`; and machine-readable `file_test.go::TestName`
  citations.
- Actual surfaces: targets from `make -qp`; tracked paths from `git ls-files`; gate definitions
  from `gates/*.yaml`; packages from `go list ./...`; registered kinds/sources from a core registry
  test helper; and acceptance artifacts from `docs/verification/`.

False prose matches are rows until classified with a durable exclusion in the census SPEC. Silent
regex exceptions are forbidden.

### Census categories

#### C1 — Live documentation truth

For every recognized mechanism token in the live authorities, prove the target/path/gate/test
exists. For every public Make target and gate definition, prove it is documented in its required
home (`README.md` or `CLAUDE.md` for user/operator commands; `docs/gates.md` for gates/checks).
Stale phase-status claims and unlabeled historical wording are rows.

#### C2 — Gate estate and build-law enforcement

For every row in `docs/gates.md`, prove its enforced path or Make target exists, its self-test
exists, and the current Make DAG invokes it at the claimed boundary. Enumerate all advisory rows and
compare expiry dates to the run date. P6 requires zero expired advisories and explicit disposition
of every advisory due at P6 start; owned, dated, still-future advisories may remain.

Known probes must confirm or refute the current missing surfaces: `make backup`, `make meta-tax`,
`make wrap-verify`, `make laws-lock`, `make state`, a useful `make short`, the `.claude` hook claims,
commit cadence, cleanroom, lesson recurrence, figure provenance, skip lint, prescription lint, state
freshness, and full invariant-citation resolution. The Proofbound legacy aliases are due at P6 start
and therefore receive an explicit row.

#### C3 — Current package acceptance

The package set is every package under `go list ./internal/...` containing at least one non-test Go
file. For each package, record:

- the package tree object from `git rev-parse <frozen-commit>:kernel/<package-path>`;
- calibrated mutation result (`neutral=survived`, `invalid=invalid`, `lethal=killed`), candidate
  count, killed count, invalid count, and declared-survivor count; and
- a non-author verdict naming the same frozen implementation commit and package tree object.

Test-only packages and command wrappers are separately enumerated and receive an explicit
`not-in-production-package-set` result rather than disappearing. C3 is measured at Task 0 but is
closed only in Task 8, after the final code-changing task.

#### C4 — Connector reality

Enumerate every production connector package and classify its acceptance as `live`, `synthetic`, or
`none`, with the exact code commit and evidence artifact. Enumerate each narrowing and non-goal from
its accepted VD/SPEC; every item is either retained by an exact decision or changed with tests.

Sessions live acceptance requires at least one genuine, quiescent harness JSONL file, at least one
accepted `session.observed` event, a second sync with zero appends, and a report of parsed versus
skipped line counts. Payload inspection must confirm metadata-only capture. If no lawful corpus is
available, a decision must rescope, replace, or retire the connector and update roadmap/README/SPEC
in the same change.

For GitHub, rerun or explicitly retain the owner/repository allowlist, 200-record v1 bound,
workflow/deployment-only scope, exact-commit joining, and missing-data behavior. Where fixtures
configure multiple repositories/providers, a proving test must retain tenant/source identity.

P6 does not claim ambient witness capture. Plain `make check` remains independent of Proofbound.
`make delivery-enforce` remains the one explicit controlled boundary and its self-test must prove
that all promoted witnesses are refreshed and ingested before enforcement.

#### C5 — Full event-universe route proof

Generate a matrix over every registered `(source, kind)` pair and each P1–P4 mechanism:
projection apply/rebuild, `proofbound verify`, twin replay, and gate evaluation. Every matrix cell is
exactly `consume`, `ignore`, or `reject` and cites one proving test. An unclassified pair is a row.

The matrix includes the P5 kinds `business_decision.recorded`, `requirement.recorded`,
`change_intent.recorded`, and `requirement.reviewed`, plus `intent.records` and `intent.specdir`.
Projection incremental/rebuild snapshots must match for every consumed pair. Twin fixtures must
exercise each current pair through the unchanged `vera.replay.v1` proof. Existing frozen vectors
remain byte-exact; additional vectors are distinct files and never edit or rename a pinned vector.

#### C6 — Proofbound intent self-coverage

The historical coverage interval begins immediately after acceptance commit `f426ca8` and ends at
the census run commit. For every commit in that interval, apply the founder-ratified path predicate
below and report: changed paths, applicable yes/no, valid `Intent:` claim yes/no, and exact CI digest
when present. Coverage is valid applicable claims divided by applicable commits; zero applicable
commits renders `not-applicable`, never 100%.

Enumerate every current ACTIVE requirement revision from all configured providers, every active
obligation, its latest exact-revision non-author review outcome, and every Proofbound delivery chain
state. Missing review or chain proof is a row.

#### C7 — P5 measurements and falsifiers

The result artifact contains one row for each of the nine P5 measurements and seven P5 falsifiers.
Measurement result is exactly `measured`, `not-recoverable`, or `not-applicable`; falsifier result is
exactly `false`, `fired`, or `indeterminate`. Every non-measured/indeterminate row states why and the
decision or action it triggers.

Allowed provenance is typed per datum: ledger event ID for ledger facts; commit/artifact plus command
for repository facts; dated session record for human-time facts. Event IDs are required only where
the ledger could have observed the datum. “Evaluated honestly” is replaced by complete row counts
and closed values.

#### C8 — Reproducibility, platform claims, and artifact integrity

From a fresh local clone of the frozen implementation commit with no `.proofbound` state, record:

- bare `make check`;
- first and second `proofbound sync all` append counts;
- `proofbound verify` from an empty projection store;
- exact rebuild row-set equality; and
- intent/requirement reports for every self-hosted chain.

Enumerate every claimed operating-system path. Each has a real run on that OS or an accepted
decision narrowing support. Verify every committed verdict/ratification self-digest where its schema
defines one, every referenced commit exists, and every acceptance artifact path resolves. Verify the
P6 motivating vision-progress assessment is preserved under `docs/verification/` or explicitly
classified as non-author analysis.

### Task 0 DoD

Commands/artifacts:

```text
docs/plans/p6-census-SPEC.md exists
scripts/tests/p6-census.test.sh passes
scripts/p6-census.sh --check exits 0
docs/plans/p6-census.md exists and is committed
```

Countable conditions: C1–C8 all appear; every emitted row has a unique ID and disposition; zero
unclassified rows; every defer/wontfix row has a decision that changes the live promise.

## Founder input (b) candidate — closed intent-applicability rule

This exact candidate is submitted for ratification; implementation may not silently alter it.

A commit is behavior-changing when its changed-path set intersects any of:

```text
Makefile
check-windows.ps1
setup-windows.ps1
CLAUDE.md
.github/workflows/**
gates/**
scripts/**
tools/**
kernel/**
docs/gates.md
docs/allowed-skips.txt
docs/allowed-survivors.txt
docs/decisions/INDEX.md
docs/invariants.lock
docs/laws.lock
```

The only exclusions inside included directory classes are `kernel/**/SPEC.md` and
`kernel/**/testdata/**`; `_test.go` files remain included because they alter admitted evidence.
All other `docs/**`, `notes/**`, vision files, `README.md`, `ROADMAP.md`, `.gitignore`, and `LICENSE`
are non-applicable unless also named explicitly above.

Changed-path semantics are path-only and content-independent: a root commit uses all paths; an
ordinary commit diffs its first parent; a merge uses the union of diffs against every parent;
renames/copies test both old and new paths; deletions test the deleted path. Any match makes the
commit applicable. An applicable commit passes only when the Git connector resolves at least one
valid exact `Intent:` claim from that commit.

Canary falsifier submitted as part of input (b): any known false negative blocks ratification; more
than 10% false positives across the complete post-`f426ca8` history fires redesign before enforce.

## Task sequence after authorization

| # | Task | Mechanical Definition of Done |
|---|---|---|
| 0 | Closed debt census | Task 0 DoD above; first C1–C8 census committed |
| 1 | Documentation, gate, and P0/P1 mechanism truth | Every C1/C2 `close-in-P6` row closed; named mechanisms exist and are wired with self-tests or the live promise is decision-amended; zero expired advisories; legacy alias disposition complete; bare `make check` exits 0 |
| 2 | Connector reality and controlled witness boundary | Every C4 `close-in-P6` row closed; sessions meets the numeric live-corpus DoD or is decision-rescoped; GitHub narrowness inventory complete; `make delivery-enforce` self-test proves witnessed ordering; plain `make check` remains product-independent |
| 3 | Event-universe and reproducibility re-proof | Every C5/C8 `close-in-P6` row closed; source-kind route matrix has zero unclassified cells; fresh-clone command record committed; frozen vectors unchanged; `proofbound verify` exits 0 |
| 4 | Intent applicability canary→enforce | Ratified path matcher and hostile tests committed; complete post-`f426ca8` canary report committed; falsifier threshold not fired; every applicable post-enforcement commit carries a valid exact intent; delivery boundary blocks one seeded missing-intent commit |
| 5 | Requirement-review completion | Every C6 review row closed; each active obligation has an exact-revision non-author review; non-verifiable obligations produce a finding and revision/retirement rather than a silent cap |
| 6 | Measurements and falsifier evaluation | Every C7 row has one allowed closed result and typed provenance; every fired falsifier has its required action/decision; no missing P5 measurement or falsifier row |
| 7 | Conditional snapshot-provider decision | Default: skipped by founder decision and contract remains pinned for P7+. If the founder designates a named real feed, the ratification explicitly amends the no-width doctrine; only then may a separate spec-first provider task and acceptance bar be inserted before Task 8 |
| 8 | Final package acceptance sweep | Freeze one implementation commit after Tasks 1–7; C3 contains every production package tree; calibrated mutation sweeps complete; one or more non-author verdicts bind that commit and all tree objects; zero package rows below the bar |
| 9 | Consolidation round and close | Full non-author round-C verdict over P1–P5 at Task-8 commit is ACCEPTABLE and committed on receipt; census rerun has zero open `close-in-P6` rows and no live promise hidden by defer/wontfix; roadmap P1–P5 deep-complete annotations cite evidence; state/journal current; bare `make check` and `proofbound verify` exit 0 |

## Non-goals

- No behavior locks, witnessed regeneration, or natural-language-to-test machinery.
- No production-shaped twin: no traffic replay, chaos, surrogates, or confidence-scored futures.
- No cryptographic signing, verifier identities, portable warranties, or marketplace.
- No governor, policy envelopes, credential control, or autonomy ratchet.
- No end-user UI and no external customer delivery boundaries.
- No new event kinds.
- No new provider unless founder input (c) explicitly amends the no-width doctrine; default is skip.
- No QA-ladder level jump; P6 measures and hardens the existing L1–L2 foundations.
- No claim that an explicit `make delivery-enforce` command is ambient or unskippable outside the
  boundary it actually controls.

## Measurements and falsifiers

Measurements:

- census rows opened/closed by category and sweep;
- intent self-coverage numerator, denominator, and percentage;
- witnessed-run coverage numerator, denominator, and percentage at the controlled boundary;
- expired, due-at-P6-start, and valid-future advisory counts;
- production packages below/at the current-code acceptance bar; and
- active requirements/obligations lacking exact-revision review.

Falsifiers:

1. If a category gains new rows on its third identical full-universe sweep, its scan is inert or
   scope-blind; reopen the scan before closing more rows.
2. If more than half of closed rows are C1 doc-truth-only and founder input (d)'s execution ceiling
   is reached, stop P6 early, record remaining rows without a deep-complete claim, and begin P7
   planning only after an explicit founder decision.
3. If requirement reviews show zero findings while any downstream self-hosted verdict is
   `INCONCLUSIVE`, treat the review mechanism as indeterminate and re-ratify it.
4. Apply the exact founder-ratified false-negative/false-positive threshold carried in input (b);
   if fired, do not enforce the applicability gate.

## ROADMAP amendment text — apply after ratification

> ## P6 — Consolidation: deep completion of P1–P5 (no new capability)
>
> Depth before width (founder direction 2026-09-12). P6 closes a mechanically generated C1–C8
> census covering live documentation, gate/build-law estate, current package acceptance, connector
> reality, the full source-kind event universe, Proofbound intent self-coverage, P5 measurements,
> and clean-clone/platform/artifact integrity. Every live P1–P5 promise is either evidenced or
> explicitly superseded, narrowed, or retired; deferral alone is not completion.
>
> **DoD:** committed census with zero unclassified rows and zero open `close-in-P6` rows; no live
> promise hidden by defer/wontfix; all production packages mutation-green and non-author accepted on
> one frozen final implementation commit; zero expired advisories; full event-route and clean-clone
> proof; round-C non-author verdict ACCEPTABLE and committed; bare `make check` and
> `proofbound verify` green. P7+ capability planning begins only after P6 closes.

If input (c) designates a provider, the founder ratification and roadmap text must explicitly state
that this is a narrow exception to “no new capability.” Otherwise Task 7 is recorded as skipped.

## Authorization boundary — four founder inputs

The founder must ratify or amend exactly these inputs before the semantic VD:

1. **Phase:** P6 is consolidation/deep completion; wider vision capability work is P7+.
2. **Applicability:** the exact path predicate, Git diff semantics, `f426ca8` anchor, and
   zero-false-negative/10%-false-positive canary threshold proposed above.
3. **Snapshot provider:** name a concrete lawful export feed and explicitly permit the width
   exception, or skip Task 7 by decision (recommended under depth-before-width).
4. **Stop-early effort ceiling:** choose a countable maximum of active P6 execution hours after the
   Task 0 census commit, excluding founder/verifier waiting time. Recommended: 40 hours.

After ratification, mint the semantic VD and amend `ROADMAP.md`; then execute Task 0 through Task 9
in order under spec-first, commit-is-durability, calibrated mutation acceptance, and non-author
verdicts committed on receipt.
