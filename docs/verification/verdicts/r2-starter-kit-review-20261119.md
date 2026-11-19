---
schema: vera.verdict.v1
verdict_id: r2-starter-kit-review-20261119
status: NEEDS_WORK
reviewed_commit: 69c9d8343faecca53f0d54c0de9c225cd7e94509
findings:
  - finding_id: r2-starter-kit-review-20261119-F1
    severity: MED
artifact_path: docs/verification/verdicts/r2-starter-kit-review-20261119.md
artifact_sha: 3378d559e67bc46011ccc1d9fc6de70bb7e7f738c3acf90b865c679ce0344db0
---

## Provenance
- Reviewer: independent non-author (Law 9), one connector-tier review under the Law 11 budget. No mutation sweep.
- Commit 69c9d83. Read: docs/plans/R2-starter-kit-spec.md, the CLI SPEC rows CLI-INV-18..20, kernel/internal/cli/starter.go, the starterkit/* files, the cli.go routing, starter_test.go and the sdd e2e test. I also read the reviews parsers (reviews/v2.go) and the projection reducers and readiness-gate SQL (projections.go, gates.go).
- Ran: one scratch variant of `TestStarterKitReviewFlow` in a detached worktree at <scratch>/review-kit, with GOCACHE and TMPDIR under <scratch>. The worktree is removed. The main checkout is unchanged: `git status` shows only the two untracked paths that were there before.

## Findings

### F1 (MED): `review seal` validates only offline, so one edited draft field can permanently wedge the tenant's projection
Seal never consults the ledger. It trusts everything in the editable `draft.json`: `evidence_event_ids`, the requirement and change-intent digests, `declared_owner`, and the obligation set. The reviews parser only checks the shape of event ids (valid ULID, sorted). It does not check that they exist. The projection reducer (`reduceObligationVerdict` and `reduceRequirementReview`) rejects dangling evidence, dangling revisions, a non-exact obligation set, and reviewer == ledger owner.

Concrete scenario, reproduced:
1. The reviewer agent replaces the pre-filled evidence id with a well-formed but nonexistent one (`01ARZ3NDEKTSV4RRFFQ69G5FAV`).
2. `seal` succeeds and the artifacts are committed.
3. `sync reviews` reports `appended=2` and exits 0. The events are now in the append-only ledger.
4. `gates canary` prints `ready-<c12> state=BLOCKED reason="projection failed: v2 verdict evidence ... is dangling"`.
5. `Projector.Apply` replays from its checkpoint, so every later gate evaluation hits the same bad event. The event cannot be removed from the ledger.

The effect is fail-closed. I found no way to get a false PASS, because the projection and the gate SQL enforce owner, coverage and exact targets. But the starter kit's primary path (an LLM agent editing JSON by hand) can brick the tenant's intent gates, not merely block one commit. CLI-INV-20 and the spec text ("validates everything", "re-parses ... before writing") do not warn about this.

Suggested fix: have `seal` open the ledger read-only and check that:
- each evidence id exists;
- each requirement and change-intent digest is a ledgered revision;
- the obligation set equals the ledgered targets;
- the reviewer differs from the ledger's owner, not the draft's `declared_owner`.

Alternatively, rebuild those fields from the ledger as the code comment claims ("seal rebuilds everything else") and ignore the draft's copies. Add a seal test with a well-formed nonexistent evidence id. `TestReviewSealWritesNothingOnAnyFailure` uses `"not-an-event"`, which only proves the parser's shape check.

## LOW (one line each)
- The final write loop in `reviewSeal` has no rollback, so a mid-loop I/O failure (permissions, disk full) leaves partial artifacts. This contradicts "any failure writes nothing", which holds only for validation failures. `os.WriteFile` also follows symlinks and truncates a file created after the `Lstat` check.
- The reviewer-vs-owner check is an exact string compare (the reviewer is trimmed, the owner is not, and case matters), both in seal and in the projection. `Alice` and `alice` pass as different people. This is consistent with the stated declared-not-authenticated limit, but the spec and tests only exercise the exact match.
- The `commit-msg` hook accepts `Intent: ` anywhere in the message, not only as a trailer. If `commit-msg` already exists, init prints `exists` and the kit's trailer enforcement is silently absent. init also installs into whatever `core.hooksPath` points to, for example a shared `.husky`.

## Questions answered, no defect found
- **init, overwriting:** `writeNew` uses `Lstat`, so it never overwrites an existing file or symlink. `.gitignore` and `AGENTS.md` are append-only, guarded by an exact-line check and the begin marker. The owner name goes through `json.Marshal`. A second run printed `exists` or `unchanged` for everything, and `TestInitOpenSpecNeverOverwritesAndIsIdempotent` covers it.
- **draft event and claims:** it uses the latest git event for the sha. Two sdd claims give two verdicts, and the review dedup key is the requirement revision. A change intent or requirement missing from the ledger is a clear error. A commit with no sdd ref is refused, and it refuses to replace an existing draft.
- **seal, other tampering:** removing the `reviews` entries, or editing the requirement review, yields BLOCKED (the readiness SQL needs a VERIFIABLE review per targeted obligation, enforced by the projection). The projection checks the ledger owner, so editing `declared_owner` cannot create a PASS. It wedges the ledger, which is the F1 class.
- **dir argument:** a relative dir resolves against the working directory. `draft.Commit` is length-checked only, so a draft for a non-ledgered commit gives a dangling-commit projection error, again F1 class.
- **SPEC rows and tests:** CLI-INV-18..20 match the code. `TestStarterKitReviewFlow` is a real discriminator: SATISFIED gives PASS with the exact counts, INCONCLUSIVE gives BLOCKED, and it runs through the real ledger.

## Verdict
NEEDS_WORK on F1 only. Everything else is ACCEPTABLE, with the 3 LOW items recorded and not blocking.
