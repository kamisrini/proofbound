# R2 M3 — binding tenant commits to `intent.sdd` change intents (SPEC draft 7)

**Status:** draft 7, 2026-11-19. It answers non-author round 6,
[`r2-m3-spec-review-round6-20261119`](../verification/verdicts/r2-m3-spec-review-round6-20261119.md)
(NEEDS_WORK: 2 MED, 2 LOW), over [draft 6](R2-sdd-m3-git-binding-spec-draft6.md).

**How to read it.** Draft 7 restates only what changes. Everything else stands as in draft 6, which in
turn incorporates draft 5; the union is the contract.
- Round 5's answers are in draft 6 §0d.
- Round 4's answers are in draft 5 §0c.
- Rounds 1-3's answers are in draft 4 §0-§0b.

**Owner decisions, 2026-11-19:** D1-D5, plus **D6** (§3), all take the recommended option.

**Before code lands:** it still needs a non-author re-check (Law 9), and the owner's approval of the
change-intent wording (draft 5 §8, Law 6).

## 0e. Round 6, answered

| Finding | Answer in draft 7 |
|---|---|
| F1 MED — appending revisions unconditionally re-admits the ordering wedge; holding them quietly is fail-open | **D6 (owner, 2026-11-19): a revision whose order-only prerequisites are missing is refused loudly** (§4.6). It is not appended, so the ledger cannot wedge. The reviews sync processes the remaining artifacts and then **fails** (non-zero exit), naming each refused revision. So every pipeline that syncs before it enforces (`sync all`, `verify`, `delivery-enforce`) goes red, and the withdrawn outcome cannot pass silently. The recovery is the same as for claims: `sync intent sdd`, then `sync git`. M3-INV-5's derivation gains the round-6 sequence as §7.3 |
| F2 MED — `migration.Parse` rejects a one-record archive | §7.2 and row 12 copy the repository's **full five-record** `docs/verification/p6-historical-evidence.jsonl` into the tenant verbatim (round 6 executed: it imports into an empty ledger and rebuilds). `E` is the pinned `checks` event `01M28TPW9C8R7ND19MNDCJ9GDG` (seq 1670). Later appends start above seq 1834, so `E` precedes every verdict |
| F3 LOW — the identity `(kind, native id)` conflates v1 and v2; the snapshot misses same-sync appends | The identity is now **(payload schema, native id)**: v1 and v2 verdicts are distinct even though they share a kind. The identity set is **updated on every append** within the sync, so `verify`'s two passes classify each file the same way. Two files with one identity are, by construction of the views (which key on `verdict_id` / `review_id`), revisions of one artifact; files processed in path order become a first revision and then a revision. Row 11 has cases for each |
| F4 LOW — residual artifact under-specification | §7.1 step 4 now states that every verdict reference carries **empty `obligation_ids`**. The `change_intent` reference has `record_kind` `change_intent` and relation `evaluates`; it is **not** copied from sdd's change-intent targets, which carry `implements` and obligation ids. Obligation outcomes are **sorted and unique** by `(source, requirement_id, artifact_sha256, obligation_id)`. §7.2 writes `gates/x1.yaml` |
| Observation — undecodable git payload in the `LedgerView` pass | A ledgered `git` `commit.recorded` payload that does not decode as `connectorgit.Commit` **fails the reviews sync**. Projections could not reduce it either, so the ledger is already unprojectable, and failing is fail-closed |
| Observation — draft 6 §5 dropped draft 5's R-INV-5 note | Restored in §5: R-INV-5's "stops on malformed input" is unchanged, and it gains "a refused revision (R-INV-13) does not stop the sync; it fails it after the remaining artifacts" |

## 1. Evidence

As draft 6 §1, plus these facts:
- **`migration.Parse` pins the archive.** It requires exactly five records, each matching the pinned
  table (seq, event id, source, native id, kind, connector version, timestamps, line SHA-256).
  `Import` requires an empty ledger, and `openHistoricalEvidenceStore` enables it (round 6, executed).
- **v1 and v2 verdicts share kind `review.verdict`.** Only the payload's `schema` field distinguishes
  them (`reviews.go` `Sync`).
- **`verify` runs the reviews sync twice** and requires the second pass to append nothing
  (`requireNoSecondSyncAppends`). Any error from either pass fails `verify`.

## 3. Owner decisions

D1-D5 as in draft 6 §3 (decided 2026-11-19). **D6 — a revision whose prerequisites are missing**
(decided 2026-11-19). The options were:
- held quietly: fail-open, rejected in round 5;
- appended: a permanent wedge, rejected in round 6;
- **refused loudly** (chosen): not appended, and the sync fails after processing the rest.

The accepted cost: on a long-lived ledger, `sync all` and `delivery-enforce` stay red until
`sync intent sdd` and `sync git` have run. Meanwhile a canary or enforce run **without** a prior sync
still reads the old revision's outcome. Enforcement in this system always syncs first
(`delivery-enforce.sh`: `sync all`, then `gates enforce`), and that run is red.

## 4.6 The reviews hold and refusal (D5, D6)

The rule replaces draft 6 §4.6's rule list. It applies after parsing and before append.
- **Identity.** An artifact's identity is `(schema, native id)`, where `native id` is `verdict_id` or
  `review_id`.
  - The CLI's view keeps the set of identities of ledgered `reviews` events, reading the schema from
    each payload's `schema` field.
  - The connector adds an identity to the set **on every append** in the same sync.
- **Never checked.** A `vera.verdict.v1` is appended as today.
- **Prerequisites** (unchanged from draft 6):
  - for a v2 verdict: `CommitClaimed` for its exact claim, and `RequirementLedgered` for every listed
    requirement;
  - for a requirement review: `RequirementLedgered` for its requirement.

  Evidence is not consulted.
- **All prerequisites present:** the artifact is appended, whether it is a first revision or a later
  one.
- **Prerequisites missing, first revision** (identity not in the set): it is **held**. It is not
  appended or counted as malformed, and it is reported as `held <cause> <path>` (D5).
- **Prerequisites missing, later revision** (identity in the set): it is **refused**. It is not
  appended, and it is recorded in `Result.Refused []Hold{Path, Cause}`. The sync continues with the
  remaining artifacts.
  - After the last artifact, `Sync` returns an error wrapping `ErrRevisionRefused`, which lists every
    refused path with its cause. The sync run is finished with that error, and the CLI exits non-zero.
  - Before the error, the CLI prints the counts line, the held lines and one
    `refused <cause> <path>` line per refusal (D6).
- **Causes:** `claim-not-ledgered`, `requirement-not-ledgered`.
- **View errors.** These fail the sync, as any reader error does:
  - a ledgered `git` `commit.recorded` payload that does not decode as `connectorgit.Commit`;
  - a `reviews` payload without a `schema` string.
- **Why nothing wedges.** No review artifact is ever appended before its order-only prerequisites,
  whatever its revision (draft 6's sufficiency argument now covers both kinds).
- **Why nothing passes silently.**
  - A held first revision adds no warrant.
  - A refused revision leaves the old outcome in the views, but every sync that sees it fails. So
    every enforcement path that syncs first is red until the prerequisites are ledgered and the
    revision appends.
- **Tenant conflicts.** Two files with one identity whose prerequisites differ surface as a
  refusal, because the second one in path order is a revision. That is loud, and it is the honest
  outcome for two artifacts that claim to be one verdict.

## 5. Sites

As draft 6 §5, with the reviews rows reading:

| Site | Change |
|---|---|
| `connector/reviews/reviews.go` `Deps`, `New`, `Sync`, `Result` | §4.6: `LedgerView` (`Ledgered(schema, nativeID)`, `CommitClaimed`, `RequirementLedgered`), `Result.Held`, `Result.Refused`, `ErrRevisionRefused` |
| `connector/reviews/SPEC.md` | draft 6's amendments, with R-INV-13 covering hold **and** refusal. R-INV-5: "stops on malformed input" unchanged; a refusal does not stop the sync but fails it after the remaining artifacts. Interface lists `Refused` and `ErrRevisionRefused` |
| `cli/cli.go` `syncReviewsOnStore`, `storeLedgerView`, `sync reviews` / `sync all` output | §4.6: held and refused lines; a non-zero exit on refusal |
| `cli/SPEC.md` CLI-INV-7 / CLI-INV-8 | the refused lines and the exit status |

## 6. Invariants

As draft 6 §6, with these changes:

5. **M3-INV-5.** Scope as in draft 6. Its proof now rests on M3-INV-11 and M3-INV-12 together: no
   review artifact is ledgered ahead of its order-only prerequisites, whatever its revision.
11. **M3-INV-11** — No **first** revision of a v2 verdict or requirement review is ledgered before its
    order-only prerequisites. Such a revision is held, reported, and appended by the first later sync
    that finds them. A hold never fails a sync.
12. **M3-INV-12** (new, D6) — A **later** revision, meaning its `(schema, native id)` is already
    ledgered or was appended earlier in the same sync, is likewise never ledgered before its
    prerequisites.
    - It is refused: the remaining artifacts are still processed, and then the sync fails, naming it.
    - With prerequisites present, a later revision is appended as before M3.
    - v1 verdicts are never checked.

## 7. Test derivations

As draft 6 §7, with these rows replaced:

| INV | Test (where) | Kills | Controls |
|---|---|---|---|
| 5, 12 | cli package, embedded store, untagged: §7.2 and §7.3 | skipping the ledger check; disabling the hold; appending a refused revision; failing before processing the remaining artifacts | §7.2 PASS; §7.3: rebuild succeeds after the refused sync, the run exits non-zero, and after recovery the revision appends and the sync exits 0 |
| 11, 12 | reviews unit, fake `LedgerView`. Cases: draft 6's held cases; a later revision with its claim missing, refused, with a later valid artifact still appended and the error listing the path; a later revision with prerequisites present, appended; **a v2 with a ledgered v1's id**, a first revision and so held, not refused; **two files with one identity in one sync**, the first appended and the second (claim missing) refused, so the snapshot is updated on append; a v1 never consulting the view; evidence never consulted | dropping any prerequisite check; refusing a first revision; holding a later revision quietly; keying identity on kind instead of schema; a snapshot not updated on append; stopping at the first refusal | neutral: artifacts with prerequisites present append exactly as before |
| 12 (view) | draft 6's row 12, with the full five-record archive (F2) and a `reviews` v1 and v2 sharing an id (identity by schema). Also: a ledgered git payload that does not decode fails the view | as draft 6 row 12, plus identity by kind; skipping undecodable payloads | as draft 6 row 12 |

Every new test is calibrated before it is believed: a hand mutant per "kills" cell must make it fail,
and a neutral edit must leave it passing.

### 7.1 Row 8 procedure: step 4, restated

Each artifact is written with its self-digest and its exact-ordered JSON front matter.
- **Review `RR-m3-x`:**
  - reviewer `reviewer-b`;
  - a requirement reference with relation `reviews` and empty `obligation_ids`;
  - one VERIFIABLE outcome per obligation id, sorted and unique.
- **Verdict `OV-m3-x1`:**
  - ACCEPTABLE, `declared_reviewer` `reviewer-b`, `reviewed_commit` X1, `findings: []`;
  - `change_intent` with source `intent.sdd`, `record_kind` `change_intent`, the CI's record id and
    digest, relation `evaluates`, and empty `obligation_ids`. It is built from the read-back id and
    digest, **not** copied from sdd's targets;
  - `requirements`: one reference with `record_kind` `requirement`, relation `evaluates`, and empty
    `obligation_ids`;
  - one SATISFIED outcome per obligation, sorted and unique by
    `(source, requirement_id, artifact_sha256, obligation_id)`, with evidence `[X1's git event id]`.
- **Verdict `OV-m3-x3`:** the same, with `reviewed_commit` X3 and evidence `[X3's git event id]`.

### 7.2 Row 5 procedure: changes

- **Step 1.** The tenant carries the repository's five-record
  `docs/verification/p6-historical-evidence.jsonl`, copied verbatim. `E` is
  `01M28TPW9C8R7ND19MNDCJ9GDG`.
- **Step 2** also writes `gates/x1.yaml` (as in §7.1 step 5, `scope_commit` X1).

Everything else stands as in draft 6 §7.2.

### 7.3 The round-6 sequence (refusal, M3-INV-5 and M3-INV-12)

Same harness, continuing on interleaving A's ledger after it reads PASS.
1. **Edit and claim.** Edit `add-auth`'s spec so its requirement gets a new revision `R'@d2`, and
   commit X5 with `Intent: sdd:add-auth` (no `migrate` needed; the archive is already imported).
2. **Revise in place.** Revise `RR-m3-x` to review `R'@d2` (its new obligation ids), and `OV-m3-x1`
   in place to `reviewed_commit` X5, with the new references and evidence `[E]`. Commit both as X6.
3. **Gate.** Add `gates/x5.yaml`.
4. **`sync all`.** It exits **non-zero** and prints:
   - `claim-skip not-ledgered sdd:add-auth at X5`;
   - `refused requirement-not-ledgered <RR file>`;
   - `refused claim-not-ledgered <OV file>`.

   `rebuild` **succeeds**.
5. **Recover.** `sync intent sdd`, then `sync git`, then `sync reviews`. It exits 0, and both revisions
   are appended. `rebuild` succeeds, and `gates canary` reads x5 `state=PASS`.

Without D6's refusal, step 4 appends both revisions ahead of `R'@d2` and X5's binding, and step 5's
rebuild fails: that is round 6's F1.

## 9. Non-goals

As draft 6 §9. Evidence portability stays with the owner, unfiled.
