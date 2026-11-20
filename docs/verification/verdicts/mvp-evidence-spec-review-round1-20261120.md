---
schema: vera.verdict.v1
verdict_id: mvp-evidence-spec-review-round1-20261120
status: NEEDS_WORK
reviewed_commit: 98b33a7826c0d41044cec495e69c793a46186512
findings:
  - finding_id: mvp-evidence-spec-review-round1-20261120-F1
    severity: MED
  - finding_id: mvp-evidence-spec-review-round1-20261120-F2
    severity: MED
  - finding_id: mvp-evidence-spec-review-round1-20261120-F3
    severity: MED
  - finding_id: mvp-evidence-spec-review-round1-20261120-F4
    severity: MED
  - finding_id: mvp-evidence-spec-review-round1-20261120-F5
    severity: LOW
  - finding_id: mvp-evidence-spec-review-round1-20261120-F6
    severity: LOW
  - finding_id: mvp-evidence-spec-review-round1-20261120-F7
    severity: LOW
artifact_path: docs/verification/verdicts/mvp-evidence-spec-review-round1-20261120.md
artifact_sha: c971d4fb2ecef4c630a975c1007f1517fead710c9481c2ce7692d47d0585bcc4
---
## Provenance
Non-author reviewer agent, spec round 1 of 2 (Law 11), on docs/plans/mvp-evidence-portability-spec.md (draft 1, untracked) at HEAD 98b33a7. Read, read-only: reviews/v2.go (ParseObligationVerdict, decodeExactOrdered, topLevelKeys), reviews/SPEC.md R-INV-10..13, projections.go reduceObligationVerdict, cli/starter.go (draft, seal, decoder), cli/cli.go sync-all ordering, core Source/Kind/native-id rules, store dedupe, cli/SPEC.md CLI-INV list, sdd cli_test.go runCLI. Ran only grep/sed; no tests, no make check.

## Status
NEEDS_WORK: four MED findings. The core design (additive optional field, resolve by source/kind/native_id, legacy ids unchanged) is sound and sync ordering is fine (see "Checked, no defect"). The defects are enforceability and one trust hole.

## Findings
- **F1 (MED). "Exactly these fields, `evidence_refs` last" is not enforceable as written.** `decodeExactOrdered` checks key order only at the top level (`topLevelKeys`); nested `ObligationOutcome` objects get only `DisallowUnknownFields`, so field order inside an outcome is not checked today and the spec's "last" cannot be a failing test. Also the event payload is the re-marshalled struct, so serialized order is struct order regardless of source order. Change: either drop "last"/exact-order from §2.1 and R-INV-14, or add an explicit nested key-order check. Also state the mechanics: `EvidenceRefs []string` needs `json:"evidence_refs,omitempty"` (else every existing verdict's payload gains `"evidence_refs":null`, changes content_sha and re-appends events on existing ledgers, the very back-compat break §2.1 denies), and "non-empty when present" needs an explicit `!= nil && len==0` rejection because `[]` and absent are otherwise indistinguishable after decode.
- **F2 (MED). Nothing binds a reference to the reviewed commit, so any event can satisfy evidence.** The parser/projection check only that a ref is well-formed and resolves at/before the verdict. A SATISFIED outcome may cite any event (e.g. an unrelated session or a self-authored record), which defeats R-INV-11's evidence-bounded intent (the existing "authored evidence cannot grade itself" text has no enforcing code at all; grep finds none, and the spec amends R-INV-11 without addressing it). Since reviewed_commit existence is already checked against commits_view, the ref adds no extra existence guarantee. Change: in projection require every ref to be `git:commit.recorded:<v.ReviewedCommit>` (the only evidence kind in the Non-goals), or state plainly that evidence stays unrestricted and drop the "self-grading" implication.
- **F3 (MED). Forward-compatibility is unstated.** A v0.3.0 binary hitting a verdict with `evidence_refs` fails `DisallowUnknownFields` and the reviews sync stops at once (malformed), so a repo whose Action pins v0.3.0 turns red with an obscure decode error after anyone seals with v0.3.1. Spec only covers old verdicts on a new binary. Change: add the pin/upgrade requirement to §2.4 and release notes, and name the failure mode.
- **F4 (MED). Proving tests are underspecified and two cannot fail for the claimed reason.** (a) CLI-INV-20 cites `TestReviewSealWritesNothingOnAnyFailure`, a failure-path test; nothing proves the success path writes `evidence_refs:[git:commit.recorded:<sha>]` and `evidence_event_ids:[]`. Add a success assertion on the sealed file. (b) CLI-INV-31's `TestPRCheckOnAFreshLedger` must be pinned: runCLI uses `DATABASE_URL=""` and cwd, so a second `git clone` dir gets its own embedded ledger (good), but the spec must say the clone copies the committed verdict, readiness gate and sig, runs `sync all`, and asserts PASS; plus a negative control that a v0.3.0-style event-id verdict on the same clone yields `dangling`, otherwise the test could pass without the fix. Put it in a non-integration-tagged file (make check omits that tag). (c) R-INV-11's existing proving test is in reviews_test (`TestObligationVerdictEvidenceValidation`); the spec swaps in the projections test only, dropping the parse-level coverage, and R-INV-14 needs cases for unsorted, duplicate, malformed, empty-present and SATISFIED-with-neither-list.
- **F5 (LOW). Resolution query not pinned.** Several events can share `(source,native_id)` with different `content_sha` (store dedupes only on the triple; `git/2` re-ingest). "Earliest" is right but write the query: `WHERE source=$1 AND kind=$2 AND native_id=$3 ORDER BY seq LIMIT 1`, so a same-native-id event of another kind cannot satisfy it, and the "future" check compares that minimum seq.
- **F6 (LOW). Dedupe/union view semantics.** The view's `evidence_event_ids` becomes "sorted union of legacy and resolved ids"; with both lists present a duplicated id (legacy id equal to a resolved id) should be deduped, and the view upsert (`ON CONFLICT ... seq <`) is unchanged, so say so. Also the error text for a bad ref should not embed an unbounded native id.
- **F7 (LOW). Numbering.** CLI-INV-31 is free (highest is 30) and R-INV-14 is free (R-INV-13 highest); no collision. Note R-INV-11's SPEC row proof is `reviews_test`, so keep the Law-2 single home when amending.

Checked, no defect: sync all orders intent, git, checks, sessions, reviews last (cli.go), so on a fresh ledger the git commit event always has a lower seq than the verdict; R-INV-13's hold needs CommitClaimed, which needs the git event, so a verdict cannot be appended before its reviewed commit; a later "reviews" commit adds its own git event but does not affect the reviewed commit's ref. Existing verdicts parse and project unchanged provided F1's `omitempty` is applied.
