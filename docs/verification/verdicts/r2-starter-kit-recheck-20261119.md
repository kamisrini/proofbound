---
schema: vera.verdict.v1
verdict_id: r2-starter-kit-recheck-20261119
status: ACCEPTABLE
reviewed_commit: 747b9e6623dc70ee0ca48ba1cc2d154b52e1d791
findings: []
artifact_path: docs/verification/verdicts/r2-starter-kit-recheck-20261119.md
artifact_sha: cc46a2a97688ee0dcac805476c0cdaaa3a59983656201e5e925bc894bc999fa5
---

## Provenance
Reviewer: independent non-author (Law 9), scoped re-check of F1 only (Law 11 budget). Commit 747b9e6 (full SHA above). I read the full `git show` of cli.go and starter.go, the current ledgerDraft, reviewSeal and mergeDecisions, and projections.go:558 (the owner/reviewer check). I ran `go test ./internal/cli/` in a detached worktree at <scratch>/recheck-kit, with GOCACHE and TMPDIR under <scratch>. I did not run the connector/intent/sdd package. The worktree is removed and the main checkout is untouched.

## Answers
1. F1 is closed. Seal now builds a fresh draft from the ledger with ledgerDraft. It copies only the outcome and finding (plus reviewer name) from draft.json, through mergeDecisions, and the draft's own reference, digest, owner, statement and evidence fields are dropped. By case:
   - Edited change_intent, requirement reference or digest: the decision lookup key no longer matches, so seal refuses with "does not match the ledger".
   - Edited statements, declared_owner or evidence ids: ignored, because the ledger's values are used.
   - Duplicated obligations: the lookup map collapses them, so the output is unchanged and no duplicate reaches the artifact.
   - Missing or extra obligations: refused. The `used != len(...)` check catches extras; a missing key is refused directly.
   - Edited commit field: the commit is looked up in the ledger and everything is rebuilt from that commit's own ledger data. The result is a valid verdict for another genuinely ledgered commit, which the projection accepts. A reviewer can seal a different commit than the draft's directory implies, but cannot forge anything. I rate this LOW.
   - Reviewer whitespace: the name is trimmed once, compared with the ledger owner, and written trimmed. That is the same string the projection compares at projections.go:558 (exact equality), so the two stay consistent.
   - A draft edit cannot produce a PASS that a genuine non-owner decision would not. The owner comes from the ledger and evidence is always the ledgered git event. The only free inputs are the reviewer's own outcomes and name, which is what the reviewer is meant to supply.
   - LOW: the owner check is exact-match, so "Alice" and "alice" count as different people. This was already the case before the fix and is not part of F1.
2. Nothing broke. The reviewDraft body moved into ledgerDraft with its logic unchanged, so `review draft` produces the same draft. The full `go test ./internal/cli/` passes (ok, 125s), including TestReviewSealWritesAcceptedArtifacts and the end-to-end starter-kit flow. The O_EXCL writes with removal on error look correct.
3. The tests discriminate. I replaced mergeDecisions with `draft := edited` in the worktree. TestReviewSealTakesEvidenceFromTheLedger failed, as did four TestReviewSealWritesNothingOnAnyFailure subtests (review_digest_edited, obligation_removed, obligation_added, owner_edited_in_draft) plus artifact_rejected. That is six failures, matching the commit message. The control cases (blank_reviewer, owner_reviews_own_work, undecided_outcome and others) correctly survived.

## Verdict
ACCEPTABLE. F1 is closed and no HIGH or MED findings remain.
