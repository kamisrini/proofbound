---
schema: vera.verdict.v1
verdict_id: task4-current-round4
status: ACCEPTABLE
reviewed_commit: d794ff7c691c3dc788505672ae817b33039b0ffc
findings: []
artifact_path: docs/verification/verdicts/task4-current-round4.md
artifact_sha: ccd4440b1580ae89b6d0884e4e714b45df0fc90e0417d7a4d970e1a3e4811083
---

ACCEPTABLE

## Round 4 scope and calibration

Independent adversarial review of frozen commit `d794ff7c691c3dc788505672ae817b33039b0ffc`.

The repository remained frozen and unchanged throughout review:

- `git rev-parse HEAD` remained exactly the tested commit.
- `git status --short` was empty.
- `git diff --exit-code` passed.
- `git fsck --no-dangling` passed.
- All hostile additions ran only in a copied module under `/tmp`; no repository file or note was modified.

Verification results:

- Bare `make check` passed. The `index stale; run make index` output was the index self-test’s expected positive control.
- `go test -race ./internal/connector/git/... -count=1` passed using a writable temporary Go cache. The initial bare race invocation reached only the sandbox’s read-only Go-cache restriction, not a product/test failure.
- The invariant-numbering and invariant-table production checks passed independently.
- Both mechanism self-tests passed independently.
- All five relevant scripts remain committed executable as mode `100755`.
- The real historical merge `1bd805119df942b10ee1b43ed12e21c54475877d` still resolves through the actual adapter to `FilesTouched:["LICENSE"]`.
- The author’s final gitcmd result of 77/77 killed and combined parent/child result of 95/95 killed are accepted as calibrated author evidence, not semantic proof. Mutation was not rerun independently.

## Round 3 remedy adjudication

1. **Repository-selection environment: CLOSED.** Every child command now removes inherited environment entries whose names case-insensitively begin `GIT_` ([gitcmd.go:334](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:334), [gitcmd.go:349](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:349)). Independent hostile repositories exercised:

   - `GIT_DIR` and `GIT_WORK_TREE`
   - `GIT_COMMON_DIR`
   - `GIT_OBJECT_DIRECTORY`
   - `GIT_ALTERNATE_OBJECT_DIRECTORIES`
   - `GIT_NAMESPACE`
   - `GIT_INDEX_FILE`
   - `GIT_CONFIG_COUNT`, `GIT_CONFIG_KEY_*`, and `GIT_CONFIG_VALUE_*`
   - `GIT_CONFIG_GLOBAL` and `GIT_CONFIG_SYSTEM`
   - discovery, ceiling, and replacement-related `GIT_*` inputs
   - mixed/lowercase key removal through the helper’s real case-insensitive path

   In every case, `New(A)` and `Commits` returned repository A’s SHA and path, never repository B’s data.

2. **Normal and linked worktrees: CLOSED.** A normal repository and a linked worktree continued to resolve their own Git metadata under simultaneous hostile Git-dir, common-dir, and object-directory variables. A legitimate shared clone using a persisted `.git/objects/info/alternates` file also remained readable, showing that sanitization does not break root-described alternate storage.

3. **Gitlink configuration: CLOSED.** Both tree-delta routes explicitly request `--ignore-submodules=none` ([gitcmd.go:159](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:159)). Independent root and non-root gitlink changes remained present under:

   - repository-local `diff.ignoreSubmodules=all` and `none`;
   - repository-local `submodule.<name>.ignore=all`;
   - HOME global configuration;
   - XDG global configuration;
   - nested gitlink paths with a real `.gitmodules` mapping.

   The root commit consistently reported `[".gitmodules","nested/sub"]`; the changed gitlink consistently reported `["nested/sub"]`.

4. **Detached HEAD immediate-object semantics: CLOSED.** Validation resolves `HEAD^{object}`, inspects that immediate object with `cat-file -t`, and requires type `commit` ([gitcmd.go:284](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:284)). Missing objects, blobs, trees, annotated tags, and nested annotated tags were independently refused by both `Commits` and `Tips`; a directly detached commit remained accepted.

5. **Other configuration and path routes: NO NEW HARM FOUND.** Repository attributes, an external diff driver, a global attributes file, HOME/XDG configuration, and a hostile global `core.worktree` setting did not produce another repository’s history or a changed successful payload. Attribute/diff-driver settings did not alter `FilesTouched`.

## Closure against HARM/routes

The Round 3 HARM was successful ingestion of the wrong repository, contradictory `FilesTouched` for one SHA, or acceptance of a detached non-commit object—facts that cannot be repaired in the append-only ledger.

Every enumerated route is closed:

1. `GIT_DIR`/`GIT_WORK_TREE` no longer override the requested root.
2. The entire case-insensitive `GIT_*` class is removed, including common-dir, object-store, namespace, index, discovery, and config-injection routes.
3. Repository and user `diff.ignoreSubmodules` settings cannot suppress gitlinks.
4. Root and non-root tree-delta commands independently force the same gitlink semantics.
5. Detached HEAD checks immediate type rather than peelability.
6. Normal repositories, linked worktrees, and persisted alternates remain operational after sanitization.

No tested route successfully produced wrong-source or contradictory payload data.

## Invariant mechanism audit

G-INV-19 through G-INV-21 are genuine appended identities in the specification ([gitcmd/SPEC.md:80](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:80)) with machine-readable proving-test rows ([gitcmd/SPEC.md:121](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:121)).

The lock diff from `ce6937003a1bb3d655425e993f439aee113ca2ee` is exactly three appended rows:

- G-INV-19 at [invariants.lock:49](/mnt/d/Users/thamm/Desktop/Projects/Vera/docs/invariants.lock:49)
- G-INV-20 at [invariants.lock:50](/mnt/d/Users/thamm/Desktop/Projects/Vera/docs/invariants.lock:50)
- G-INV-21 at [invariants.lock:51](/mnt/d/Users/thamm/Desktop/Projects/Vera/docs/invariants.lock:51)

No prior lock row changed or moved. Both blocking checks remain wired into `make check` ([Makefile:3](/mnt/d/Users/thamm/Desktop/Projects/Vera/Makefile:3)).

## Findings

### HIGH

None.

### MED

None.

### LOW

None.

## Residual limits

- The adapter deliberately refuses repositories whose required object store exists only through an inherited `GIT_ALTERNATE_OBJECT_DIRECTORIES`; accepting that input would reopen G-INV-19’s object-store redirection route. Persisted, root-described alternates were verified.
- The review did not attempt an exhaustive proof over every Git version, platform environment implementation, system configuration key, or concurrent repository mutation.
- The adapter still relies on the `git` executable selected through the process `PATH`; G-INV-19 governs repository selection by Git state, not executable provenance.
- Mutation evidence remains limited by its operator set and exclusion of `_test.go`; its green result does not establish that assertions express the right contract.

## Acceptance rationale

Task 4 is ACCEPTABLE under Law 9 at commit `d794ff7c691c3dc788505672ae817b33039b0ffc`.

The three Round 3 findings are closed at their full stated route classes, not merely at the original examples. Independent hostile repositories could not redirect the requested root, configuration could not change root or non-root gitlink payloads, and every immediate detached non-commit type was refused. The remediation preserved legitimate linked-worktree and persisted-alternate execution, the append-only invariant mechanism is exact and self-tested, the full bare gate and race suite are green, and no new correctness route reaching the stated append-only harm was found.
