---
schema: vera.verdict.v1
verdict_id: r2-m4-drift-recheck-20261119
status: ACCEPTABLE
reviewed_commit: 6ed9231e35036ed883acefc118eaf2464d237ef7
findings: []
artifact_path: docs/verification/verdicts/r2-m4-drift-recheck-20261119.md
artifact_sha: 4f21f1a096db34013078ba2c1af152b1d34ce6dbaf451cfa562659436447cc93
---
# r2-m4 drift re-check (F1 fix)

## Provenance
Non-author reviewer (Law 9), scoped re-check under Law 11. Read `git show 6ed9231` (drift.go, build.go, drift_test.go, state_revision_test.go, BR 0004 vs 0005 diff). Ran in a detached worktree at <scratch>/recheck-m4 (GOCACHE/TMPDIR under <scratch>), since removed: `go test ./internal/cli/` ok; `go test -run 'TestStateRevision|TestResolve' ./internal/connector/intent/sdd/` ok. No sdd TestDrift*/M3*/StarterKit* run. Main checkout untouched.

## Answers
1. F1 closed. "Current" revision now comes from Provider.StateRevision(HEAD, name), read from the HEAD tree, never ledger seq. Revert, branch switch, archive/unarchive all resolve from HEAD's own tree. Verified B then reverted to A: HEAD=A is not covered, so STALE (unit case plus e2e asserting appended=0). Renamed or deleted change gives REMOVED; a HEAD revision not ledgered gives UNSYNCED; neither can read CURRENT. Name derivation: change-intent artifact_path is the change directory (build.go:106); changeName takes path.Base and strips a YYYY-MM-DD- prefix with the same sddDatedRE as cli.go; TestChangeNameStripsTheArchiveDate covers active, archived and bare-date. Verified both A and B with HEAD on A: warranties scanned newest first, first with promise equal to HEAD's wins, so the A warranty (OV-A, CURRENT); asserted by the "verified both, head back on A" case.
2. Resolve refactor is behaviour-preserving. changeAt keeps the same checks in the same order (inDomain, Read error, empty state, mapState, archiveBase, stateSkip, missing change, cr.skip) with identical errors; Resolve then does the singleArchive parent scan and absent check as before. TestStateRevisionAgreesWithResolveExceptArchivedEarlier compares both over all of expected-resolve.json and requires at least one archived-earlier row.
3. Tests discriminate. Mutant: HEAD promise lookup uses the newest warranty's digest instead of head.digest (the ledger-order-style bug); TestDriftComparesPromises fails it ("edited promise" read CURRENT). A second mutant removing the UNSYNCED branch did not compile (unused variable), so it was not tested. BR 0005 self-digest verified (zero last artifact_sha256, sha256 = 4311072f...de73, matches). It supersedes 0004's 83d98dbe...; only O-13 retired and O-14 added, other obligations byte-identical, header prose differs.

LOW: a change whose active directory name itself starts with YYYY-MM-DD- would have its prefix stripped; this fails safe (REMOVED/UNKNOWN), never CURRENT.

## Verdict
ACCEPTABLE. F1 closed; no HIGH or MED findings.
