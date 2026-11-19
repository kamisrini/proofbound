# R2 M3 — binding tenant commits to `intent.sdd` change intents (SPEC draft 6)

**Status:** draft 6, 2026-11-19. It answers non-author round 5,
[`r2-m3-spec-review-round5-20261119`](../verification/verdicts/r2-m3-spec-review-round5-20261119.md)
(NEEDS_WORK: 3 MED, 3 LOW), over [draft 5](R2-sdd-m3-git-binding-spec-draft5.md). Round 4 and its
answers are in draft 5 §0c; rounds 1-3 are in draft 4 §0-§0b. **Owner decisions, 2026-11-19:** D1-D5
take the recommended option (§3). Round 5 narrows how D5 is implemented, not whether: §3 says how.
This draft still needs a non-author re-check (Law 9) and the owner's approval of the change-intent
wording (§8) before any code lands (Law 6). **Parent contract:**
`kernel/internal/connector/intent/sdd/SPEC.md` (R2 M1, accepted). **Requirement served:**
`BR-sdd-warranty-i9miv7` O-8 and O-10 ("fails only a claim on that state").

## 0d. Round 5, answered

| Finding | Answer in draft 6 |
|---|---|
| F1 MED — evidence ids are ledger-local, not order-only, so §7.2's PASS control is red | **Evidence is removed from the hold** (§4.6). A missing evidence id is not something a later sync can supply on a ledger other than the one that minted it, so holding for it would hold forever. A verdict whose evidence is absent is appended and fails Apply, **exactly as before M3** (pre-existing; §9). §7.2 now makes its evidence portable the way this repository does: a one-record archive imported by `migrate historical-evidence` into the empty ledger, so the id exists on every ledger the test builds. §1's "order-only" list is corrected to items 1-3 |
| F2 MED — holding an in-place revision keeps a withdrawn SATISFIED outcome warranting a commit | **The hold applies only to an artifact's first revision** (§4.6). If any `reviews` event with the same kind and native id (`verdict_id` / `review_id`) is already ledgered, the new revision is appended unconditionally, as before M3, so it either overwrites the outcome or fails Apply loudly. Either way, a withdrawn SATISFIED outcome can no longer keep warranting a commit. A first revision that is held adds no warrant, so trust stays fail-closed. M3-INV-11 states both halves; row 11 has a revision case |
| F3 MED — the real `LedgerView` is untested outside `sdd`; the "0 files" premise is false | **The premise was wrong.** The authoring session's grep searched for a YAML `schema:` line, but v2 and requirement-review front matter is one JSON line. Three such artifacts are committed (§1). So the hold runs on every sync of this repository, and §1 now says so. New **row 12** tests the real `LedgerView` on the embedded store, untagged, over a mixed ledger: `records`, `specdir` and `sdd` requirements; git events carrying refs of each provider; and historically imported events. It kills the three mutants round 5 names (sdd-only refs, `intent.sdd`-only requirements, a source-filtered read) |
| F4 LOW — §7.1 under-specifies artifacts | §7.1 now names every field it reads back: requirement id and digest, obligation ids, change-intent id and digest. It names the relations (`evaluates` for the verdict, `reviews` for the review), distinct `verdict_id`s for X1 and X3, X2 and X3 touching no file under `add-auth`, and the gate's required `description` and `expires` |
| F5 LOW — M3-INV-5 overreaches; event-id semantics unstated | M3-INV-5 is scoped to `sdd` claims and to "content-valid" artifacts, which are defined. The `records` / `specdir` carve-out (`DEF-INTENT-GIT-ORDERING-001`) is now in the invariant itself. `EventLedgered` is gone (F1), so its comparison semantics no longer arise. §7.2 adds a second interleaving and states that the derivation samples interleavings rather than enumerating them |
| F6 LOW — reviews SPEC text not listed; a second decoder implied; the lazy read untested | §5 lists the reviews SPEC Purpose sentence, R-INV-11 and the requirement-review contract sentence as amended. Git refs are decoded with `connectorgit.Commit`, the exported type the git connector itself marshals (`connector.go:28`), not with projections' unexported `commitPayload`, so there is no second decoder. Row 12 counts the ledger reads: 0 over a ledger with only v1 verdicts, 1 once a v2 verdict is listed |

**Raised to the owner, not filed** (§9): before M3, a v2 verdict whose evidence id exists only on the
ledger that minted it already fails Apply on every other ledger. That includes a fresh CI ledger built
by `sync all`, unless the evidence was historically imported. M3 neither causes nor fixes it.

## 1. Evidence

As draft 5 §1, corrected:
- **Order-only prerequisites of the v2 verdict reducer** are items 1-3: the reviewed commit in
  `commits_view`, the exact claim in `commit_intents_view`, and every listed requirement revision in
  `requirements_view`. Item 4 (`intent_targets_view`) is present whenever item 2 is (round 5
  confirmed). **Item 5 (evidence) is not order-only.** Event ids are random ULIDs (`newIDs`,
  `crypto/rand`), so an id minted on one ledger never appears on another, except through
  `migrate historical-evidence`, which imports ids verbatim (round-5 F1).
- **Order-only prerequisite of the requirement-review reducer:** the requirement revision in
  `requirements_view` (and `requirement_obligations_view`, filled in the same `reduceIntent` call).
- **This repository does hold v2 / review artifacts** (round-5 F3), each with a single-line JSON
  front matter:
  - `p5-obligation-verdict-round1.md` (`intent.records`, evidence imported historically);
  - `p5-requirement-review-round1.md`;
  - `p7-replay-requirement-review-round1-20261014.md`.

  All three parse and sort ahead of the first malformed file, so every `sync reviews`, `sync all` and
  `verify` here would build the `LedgerView`.
- **Revisions upsert.** `obligation_verdicts_view` is upserted on
  `(verdict_id, requirement…, obligation_id)` and `requirement_reviews_view` on `(review_id, requirement…, obligation_id)`, so a later
  ledgered revision overwrites the earlier one's outcome (round-5 F2).
- **The git payload's producer type is exported:** `connectorgit.Commit` with
  `IntentRefs []IntentRef` (`connector/git/connector.go:28-42`). Projections decode the same bytes
  with their unexported `commitPayload`.

## 2. Purpose, boundary, entry conditions

As draft 5 §2. The reviews connector's change is the hold, as narrowed in §4.6.

## 3. Owner decisions (decided 2026-11-19: the recommended option)

D1-D4 as in draft 5 §3. **D5** is the reviews hold. Round 5 narrows it in two ways, and neither
changes the option chosen:
1. the hold waits only for prerequisites a later sync can supply on **this** ledger (the claim and
   the requirements, not evidence);
2. it applies only to an artifact's first revision.

Both narrowings make the hold strictly less able to suppress an artifact, so neither can widen what
the owner approved.

## 4. Contract

§4.1-§4.5 as in draft 5 (trailer values, resolution and anti-spoofing, the ledger check `L`, claim
outcomes and skips, projections and the readiness gate).

### 4.6 The reviews hold (D5)

- **New dependency.** `reviews.Deps` gains `Ledger LedgerView`, **required** (a nil or typed-nil value
  fails `New`, R-INV-6):

  ```go
  type LedgerView interface {
      // Ledgered reports whether a reviews event of this kind and native id is ledgered (any revision).
      Ledgered(ctx context.Context, kind core.Kind, nativeID string) (bool, error)
      // CommitClaimed reports whether a ledgered git commit event for sha carries exactly this ref.
      CommitClaimed(ctx context.Context, sha, provider, recordID, digest string) (bool, error)
      // RequirementLedgered reports whether an intent requirement.recorded event of this source, id and digest is ledgered.
      RequirementLedgered(ctx context.Context, source core.Source, recordID, digest string) (bool, error)
  }
  ```
- **The hold rule.** It applies after an artifact parses (the parse rules and their failures are
  unchanged) and before it is appended. A `vera.verdict.v1` is never held.
  1. **A revision is never held.** If `Ledgered(kind, nativeID)` holds, the artifact is appended
     unconditionally, as before M3.
  2. **A first revision** of a `proofbound.obligation-verdict.v2` is appended only if both of these
     hold:
     - `CommitClaimed(reviewed_commit, trimPrefix(change_intent.source, "intent."),
       change_intent.record_id, change_intent.artifact_sha256)`;
     - `RequirementLedgered` for every listed requirement.
  3. **A first revision** of a `proofbound.requirement-review.v1` is appended only if
     `RequirementLedgered` holds for its requirement.
  4. **Evidence is not consulted.** A missing evidence id fails Apply, as before M3.

  An artifact that is not appended is **held**: not counted as malformed, and the sync continues. The
  checks run in the order listed, and the first one that fails names the cause:
  `claim-not-ledgered` or `requirement-not-ledgered`.
- **Result and output.** As draft 5:
  - `Result.Held []Hold{Path, Cause}`;
  - a held path is not in the cursor;
  - `sync reviews` and `sync all` print `held <cause> <path>` after their counts line;
  - `verify` prints none;
  - a hold never changes an exit status.
- **The CLI's `LedgerView`.** It is built over the same store, **lazily**, on the first v2 verdict or
  requirement review. Then one `ReadEvents` pass with an **empty** filter fills three sets:
  - `(kind, native id)` of every `reviews` event;
  - `(sha, provider, record id, digest)` of every ref in every `git` `commit.recorded` payload,
    decoded as `connectorgit.Commit`;
  - `(source, requirement id, digest)` of every `requirement.recorded` event whose source is an
    `intent.*` source, the digest decoded from the payload's `artifact_sha256`.

  The pass costs O(ledger), once per sync that lists a v2 or requirement-review artifact. In this
  repository that is every reviews sync (§1).
- **Why the hold is sufficient.** It is made against the ledger as it stands before the append, and
  the ledger is append-only. So, when the reducer reaches the artifact's event:
  - the claim row exists, because the git event carrying it was projected first, and for `sdd` it
    passes the dangling check (M3-INV-5);
  - `commits_view` holds the sha;
  - the requirement rows exist.

  The remaining reasons Apply can fail are the artifact's **content** and its **evidence** (both
  pre-existing, §9), and, for `records` / `specdir` claims, `DEF-INTENT-GIT-ORDERING-001`.
- **Why the hold is fail-closed on trust.**
  - A held first revision is in no projection, so it adds no warrant.
  - A revision is never held, so no earlier outcome survives a later committed revision because of
    the hold (round-5 F2).
- **The accepted cost.** A first revision whose claim or requirement never becomes ledgered (a
  digest typo, a claim on a commit whose `sdd` claim is skipped) used to make Apply fail loudly. Now
  it prints `held …` on every sync, and the commit stays unwarranted.

## 5. Sites (complete)

As draft 5 §5, with these rows replaced or added:

| Site | Change |
|---|---|
| `connector/reviews/reviews.go` `Deps`, `New`, `Sync`, `Result` | §4.6 |
| `connector/reviews/SPEC.md` | new R-INV-13 (the hold, both halves); R-INV-6 lists `Ledger`; Interface lists `LedgerView`, `Hold`, `Result.Held`. **Amended text** (round-5 F6): the Purpose sentence "each valid schema-bearing artifact emits one … event" gains "unless it is a held first revision (R-INV-13)"; R-INV-11's "dangling or future evidence fails projection" stays true (evidence is not held) and gains a cross-reference saying so; the requirement-review contract's "projection validation rejects a dangling … revision digest" gains "a first revision whose requirement is not yet ledgered is held by the connector instead (R-INV-13)" |
| `connector/reviews/reviews_test.go` | `Deps` literals at its four construction sites gain a `LedgerView` (a fake that admits everything keeps their behaviour) |
| `cli/cli.go` `syncReviewsOnStore` + a `storeLedgerView` type | §4.6; decodes git payloads as `connectorgit.Commit` (no second decoder) |
| `cli` package tests (untagged) | rows 5, 8 and 12 |

## 6. Invariants (M3-INV-n)

M3-INV-1 to M3-INV-4 and M3-INV-6 to M3-INV-10 as in draft 5 §6. These two are restated:

5. **M3-INV-5** — No ledgered commit event names an `sdd` revision that is not itself ledgered at a
   lower sequence number.
   - **Scope:** a tenant whose `Intent:` values are all `sdd:` claims (for `records` / `specdir`
     claims, `sync git` before their intent sync can still dangle: `DEF-INTENT-GIT-ORDERING-001`),
     and whose committed review artifacts are **content-valid**.
   - **What it promises:** for such a tenant, a projection rebuild succeeds after any interleaving of
     `sync git`, `sync intent sdd`, `sync reviews` and `sync all`.
   - **Content-valid** means the artifact projects on a ledger that holds every prerequisite it
     cites, including its evidence ids.

11. **M3-INV-11** — The hold has two halves:
    - **First revisions wait.** No **first** revision of a v2 verdict or requirement review is
      ledgered before its order-only prerequisites (claim, requirement revisions). Such a revision is
      held, reported, and appended by the first later `sync reviews` that finds them ledgered.
    - **Later revisions never wait.** A revision of an already-ledgered artifact, and every v1
      verdict, is appended as before M3.

    A hold never fails a sync or changes an exit status.

## 7. Test derivations

Rows 1-4, 6, 7, 9 and 10 as in draft 5 §7. Replaced or new:

| INV | Test (where) | Kills | Controls |
|---|---|---|---|
| 5 | cli package, embedded store, untagged (§7.2) | skipping the ledger check; disabling the hold | the bound commit's row and the verdict both project after recovery, in both interleavings |
| 8 | cli package, embedded store, untagged (§7.1) | the gate ignoring `sdd` rows (X1 must PASS); keeping the bindable ref of a two-claim commit (X3 must stay BLOCKED); disabling the hold (X3's verdict appends, Apply fails, X1 BLOCKED) | X1 `state=PASS`, X2 and X3 `state=BLOCKED`, on one ledger; X3's verdict reported `held claim-not-ledgered` |
| 11 | reviews unit with a fake `LedgerView`. Cases: each prerequisite missing (the claim, each requirement), each held with its cause and absent from the appender and cursor; a requirement review with its requirement missing; **a revision whose native id is ledgered, with its claim missing, appended** (round-5 F2); a v1 verdict never consulting the view; the same held artifact appended on a later sync once the fake admits it; a verdict citing an absent evidence id, appended (evidence not consulted) | dropping any one prerequisite check; holding a revision; consulting evidence; counting a hold as malformed; failing the sync on a hold; holding a v1 | neutral: an artifact whose prerequisites are all present appends exactly as before |
| 12 | cli package, embedded store, untagged: the real `storeLedgerView` over a ledger holding `intent.records`, `intent.specdir` and `intent.sdd` requirement revisions; git events whose refs name each provider; a historically imported event (`migrate historical-evidence` on the empty ledger first); and one `reviews` event. Ask every method about each. Also a read counter: 0 ledger reads for a reviews sync listing only v1 verdicts, 1 once a v2 verdict is listed | a view collecting only `sdd` refs; only `intent.sdd` requirements; reading with a source or kind filter; ignoring imported events; an eager ledger read | negatives: a ref with a wrong digest, a requirement with a wrong source, an unknown native id, all false |

Every new test is calibrated before it is believed: a hand mutant per "kills" cell must make it fail,
and a neutral edit must leave it passing.

### 7.1 Row 8 procedure

An internal test of package `cli` (no build tag). It calls `runCommand`, and between commands it opens
the same store with `openStore(ctx, root, "")` to read events.
1. **Tenant.** `git init` a repository in a temp dir with an sdd profile (OpenSpec, owner `owner-a`) and
   an active change `add-auth` with one requirement. Commit:
   - `X1` with `Intent: sdd:add-auth`;
   - `X2` with `Intent: sdd:nope`;
   - `X3` with both `Intent: sdd:add-auth` and `Intent: sdd:nope`.

   X2 and X3 change only a file outside `openspec/` (for example `src/x`), so `add-auth`'s revision is
   the same at X1, X2 and X3. X3 is HEAD.
2. **Ledger, part 1.** `sync intent sdd`, then `sync git`. Expect `claim-skip unknown-change sdd:nope`
   at X2 and X3, and X1 bound.
3. **Read back.** From the `intent.sdd` events for `add-auth`, read:
   - the change intent's record id and `artifact_sha256`;
   - the requirement's record id and `artifact_sha256` (sdd's requirement id embeds the digest);
   - every obligation id the requirement payload lists.

   From the `git` events, read X1's and X3's event ids.
4. **Artifacts.** Write each with its self-digest (`ArtifactSHA256`'s zeroing rule) and the v2 /
   review exact-ordered JSON front matter:
   - one `proofbound.requirement-review.v1`, id `RR-m3-x`, reviewer `reviewer-b` (≠ `owner-a`), its
     requirement reference with relation `reviews`, one VERIFIABLE outcome per obligation id;
   - a `proofbound.obligation-verdict.v2`, id `OV-m3-x1`, ACCEPTABLE, `declared_reviewer`
     `reviewer-b`, `reviewed_commit` X1, `change_intent` the sdd CI reference, `requirements` the
     requirement reference with relation `evaluates`, one SATISFIED outcome per obligation with
     evidence `[X1's git event id]`, `findings: []`;
   - the same for X3 with a **distinct** id `OV-m3-x3`, `reviewed_commit` X3, evidence `[X3's git
     event id]`.

   Commit them under `docs/verification/verdicts/` as `X4` (no trailer). X4 is HEAD.
5. **Gates.** Write `gates/x1.yaml`, `gates/x2.yaml` and `gates/x3.yaml` in the working tree. Each is a
   **JSON** `proofbound.gate.v1` definition with these fields:
   - a distinct `id`;
   - a `description`;
   - `expires` `2099-01-01`;
   - `mode` `canary`;
   - `rule` `intent-delivery-readiness`;
   - `scope_commit` that commit's full sha.
6. **Ledger, part 2.** `sync git`, then `sync reviews`. Expect the review and `OV-m3-x1` appended, and
   `held claim-not-ledgered docs/verification/verdicts/<OV-m3-x3 file>`.
7. **Judge.** `gates canary`. Assert `state=PASS` for x1 and `state=BLOCKED` for x2 and x3.

BLOCKED is meaningful only beside the PASS on the same ledger. A failed Apply makes every readiness
gate BLOCKED, so the x1 PASS is what proves the ledger projects.

### 7.2 Row 5 procedure

Same package and harness. The evidence is made portable, so one committed verdict is valid on every
ledger the test builds.
1. **Tenant.** As §7.1 step 1, X1 only. The tenant also carries
   `docs/verification/p6-historical-evidence.jsonl` with **one** record: a valid ledger record with a
   fixed event id `E` (for example a `checks` event; the shape is that of the repository's own
   archive).
2. **Artifacts, before any sync.** Compute the change-intent and requirement references and the
   obligation ids **without a ledger**, by running the `sdd` provider's `Revisions` over the tenant.
   They are content-derived, so they equal what `sync intent sdd` will ledger. Write the review and
   `OV-m3-x1` as in §7.1 step 4, with evidence `[E]`. Commit them as X4.
3. **Interleaving A, on a fresh state root.**
   1. `migrate historical-evidence`, which needs an empty ledger.
   2. `sync all`, which never syncs `sdd`. Expect X1 `claim-skip not-ledgered`, the verdict
      `held claim-not-ledgered` and the review `held requirement-not-ledgered`. `rebuild` succeeds.
   3. `sync intent sdd`, `sync git` (X1 now bound), then `sync reviews` (both appended). `rebuild`
      succeeds, and `gates canary` on x1 reads `state=PASS`.
4. **Interleaving B, on another fresh root.** `migrate historical-evidence`, `sync reviews` (both
   held), `sync intent sdd`, `sync reviews` (the review appends, the verdict is still held: the claim
   is not ledgered), `sync git`, `sync reviews` (the verdict appends). `rebuild` succeeds after each
   step, and x1 reads `state=PASS` at the end.

Without the hold, step 3.2 appends the verdict ahead of the binding, and step 3.3's rebuild fails.
**These two interleavings sample M3-INV-5; they do not enumerate it.** The argument in §4.3 and §4.6
covers the rest.

## 8. Proposed change intent (owner approves the wording)

As draft 5 §8.

## 9. Non-goals

As draft 5 §9, plus:
- **Evidence portability (pre-existing, raised to the owner).** A v2 verdict whose evidence id exists
  only on the ledger that minted it fails Apply on every other ledger, before and after M3. The hold
  deliberately does not wait for it, because waiting would be forever (round-5 F1).
