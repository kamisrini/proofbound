NEEDS_WORK

## Calibration and evidence

Reviewed frozen commit `3470e60a67caf1bda414d7baf08bfb57b76ffd8b` as a non-author verifier. The repository remained unchanged.

- `make check` passed.
- `go test -race ./internal/connector/git/... -count=1` passed.
- The author’s 77/77 mutation result is accepted as calibrated author evidence, not semantic proof. I did not rerun the mutation sweep.
- I copied the module to `/tmp` and added adversarial tests there only. Existing tests stayed green; the hostile fixtures below failed deterministically.
- `git fsck --no-dangling` passed on the real repository, so the current checkout does not itself contain the missing-object corruption described below.
- Real repository data was checked before suggesting remedies:
  - all decision files throughout current reachable history are directly under `docs/decisions/`;
  - HEAD paths tested are valid UTF-8;
  - real merge `1bd805119df942b10ee1b43ed12e21c54475877d` has a first-parent change to `LICENSE`, while the connector records no touched files for it.

## Closure against the stated HARM/routes

The central harm is not closed: a green connector can still append incomplete commit facts, fabricate a decision citation, and report a broken repository as a successful sync. Because identity includes payload content and the ledger is append-only, the affected payload cannot be repaired in place.

1. Shallow, grafted, and replacement-rewritten history: the tested forms are closed. Missing-object refs remain an open partial-history route.
2. Date watermark and ref coverage: correctness is closed by older-date, branch, tag, and detached-HEAD tests. The documented ordering contract has a low-severity mismatch.
3. Stash/notes/replace exclusions and prefix boundaries: closed for the exercised ref classes.
4. Citation fabrication: open. A decision-looking file in a nested directory is accepted as though it were the exact decision artifact.
5. Hostile content, filenames, and scalar mapping: framing and valid UTF-8 hostile paths are closed. Invalid UTF-8 path bytes collapse in the final JSON, and merge commits receive empty file attribution.
6. Broken repository versus empty repository: open. A ref containing a valid object-id spelling for a missing object is silently ignored.
7. Append errors and count partitioning: closed by direct inspection and discriminating tests.

## Findings

### HIGH-1 — A ref to a missing object is silently accepted, reproducing the recorded broken-as-empty harm route

`validateRefs` verifies only that a ref resolves syntactically, not that its object exists ([gitcmd.go:207](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:207), especially lines 228–243). `rev-list --all` then silently ignores this form ([gitcmd.go:57](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:57)), while `Tips` explicitly swallows peel failures ([gitcmd.go:94](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:94)).

Reproduction:

1. Create a normal repository and commit.
2. Construct `gitcmd.Repo`.
3. Write `1111111111111111111111111111111111111111` to `.git/refs/heads/broken` without creating that object.
4. Both `repo.Commits(ctx)` and `repo.Tips(ctx)` return nil errors. The broken ref and its history are omitted.

This violates G-INV-6 and the protocol’s “failure from Git is always an error” rule ([gitcmd/SPEC.md:33](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:33)). It is materially different from the existing corrupt-ref test, which exercises malformed ref text.

Verified finding; proposed remedy is not yet verified. The implementation needs existence validation for every admitted ref, including packed refs, while still handling valid non-commit refs according to an explicit rule. `git fsck` proves the current repository is clean but is not by itself a selected product remedy.

### HIGH-2 — Distinct legal Git path bytes collapse to the same payload spelling

The `-z` path parser correctly obtains raw bytes, but immediately converts each path to a Go string ([gitcmd.go:272](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:272)). `json.Marshal` later replaces invalid UTF-8 with U+FFFD ([connector.go:99](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/connector.go:99)).

A commit touching two files named by byte sequences `FE` and `FF` produced:

```json
"files_touched":["�","�"]
```

The two distinct paths are therefore neither significant nor recoverable, contrary to G-INV-12 ([gitcmd/SPEC.md:54](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:54)). This can permanently append incomplete or ambiguous commit facts.

Verified finding; proposed remedy is not yet verified. The contract must choose either fail-closed refusal of non-UTF-8 paths or a reversible wire encoding. Refusal would accept current HEAD’s paths, but that compatibility check does not prove the remedy.

### HIGH-3 — Nested files fabricate exact decision citations

Citation resolution recursively lists `docs/decisions`, discards the directory with `path.Base`, and treats any `VD-*.md` basename as a decision ([gitcmd.go:158](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:158)).

A commit containing only:

```text
docs/decisions/archive/VD-nested-aaaaaa.md
```

and mentioning `VD-nested-aaaaaa` yielded that ID in `CitedDecisions`. This violates G-INV-13’s exact-file requirement ([gitcmd/SPEC.md:56](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:56)) and directly reproduces the stated citation-fabrication harm.

Verified finding; proposed remedy is not yet verified. Requiring the exact direct path is compatible with all decision paths in current reachable repository history, but a discriminating nested-path test must establish the remedy.

### HIGH-4 — Every merge commit receives empty `files_touched`

The file query uses `git diff-tree` without a merge diff mode ([gitcmd.go:150](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:150)). Git therefore emits no paths for merge commits, even when the merge tree differs from a parent.

A hostile merge fixture that introduced `from-feature` returned `FilesTouched == nil`. This is also live against repository data: merge `1bd8051` changed `LICENSE` relative to its first parent, but the current command emits zero paths.

This permanently appends incomplete commit facts. The SPEC repeats the exact command ([gitcmd/SPEC.md:28](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:28)) but never defines merge semantics, so simply adding a flag is not a verified fix. First define whether “files touched” means first-parent change, union across parents, combined conflict-resolution change, or another rule; then test the chosen harm-level behavior against a real merge.

### MED

None.

### LOW-1 — `Commits` does not satisfy its documented committer-date ordering

The interface promises committer-date order, oldest first ([git/SPEC.md:92](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/SPEC.md:92)), but `--reverse --date-order` preserves ancestry constraints rather than strictly sorting timestamps ([gitcmd.go:57](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:57)).

A parent dated 2030 followed by its child dated 2020 was returned 2030 then 2020. The SPEC explicitly says ordering is not a correctness property, so this is low severity, but the interface claim is false and should be corrected or implemented.

### LOW-2 — `New` accepts a typed-nil `Repo`, contrary to INV-27

The constructor checks only `d.Repo == nil` ([connector.go:62](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/connector.go:62)). A `(*fakeRepo)(nil)` stored in the interface is accepted and produces a non-nil connector. An implementation whose method dereferences its receiver can then panic during sync.

This is a verified contract mismatch with “every dependency is required,” but ordinary wiring through successful `gitcmd.New` does not produce this shape, so severity is low. Any remedy needs a direct typed-nil test rather than relying on the existing nil-interface case.

## Acceptance rationale

Task 4 is not acceptable under Law 9. The mechanical DoD, race run, and mutation sweep are useful evidence, but four untested routes still reach the finding’s own harm: missing history can be reported as successful, path identity can be destroyed, citations can be fabricated, and merge facts can be incomplete. Acceptance requires fixing the invariants—not only these reproductions—then rerunning the package checks and mutation sweep and obtaining a new non-author verdict.
