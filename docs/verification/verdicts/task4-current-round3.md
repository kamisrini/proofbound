NEEDS_WORK

## Round 3 scope and calibration

Independent review of frozen commit `ce6937003a1bb3d655425e993f439aee113ca2ee`.

The repository remained unchanged: HEAD stayed at the tested commit, `git status --short` was empty, and `git diff --exit-code` passed.

- Bare `make check` passed. The `index stale; run make index` line was the index self-test’s expected positive-control output.
- `go test -race ./internal/connector/git/... -count=1` passed.
- The author’s gitcmd 74/74 and combined 92/92 mutation results are accepted as calibrated author evidence, not semantic proof. I did not rerun mutation.
- Hostile additions ran only in a copied module under `/tmp`; repository files were not modified.
- `git fsck --no-dangling` passed on the real repository.
- The real historical merge `1bd805119df942b10ee1b43ed12e21c54475877d` still resolves through the actual adapter to `FilesTouched:["LICENSE"]`.

## Round 2 remedy adjudication

1. **Rename configuration: CLOSED for `diff.renames`.** Toggling both repository-local and user/global `diff.renames` around the same SHA produced the required two-path result.
2. **Invalid UTF-8 scalar bytes: CLOSED for actual reachable scalar routes.** Independent raw commit objects with invalid bytes in author name, author email, committer name, committer email, subject, and body were all refused. The helper also covers SHA/date field positions that Git normally synthesizes.
3. **Annotated-tag missing referents: CLOSED for direct and nested chains.** Direct and tag-to-tag missing referents failed in both `Commits` and `Tips`. Valid nested tags to commits peeled correctly; valid direct blob refs and annotated tags to blobs remained non-history without becoming errors.
4. **Actual path refusal: CLOSED.** Root and non-root commits with invalid UTF-8 filenames were refused through the adapter; valid hostile UTF-8 paths survived.
5. **Invariant mechanism and numbering: CLOSED for the Round 2 defect.**
   - All five new scripts/tests are committed executable as mode `100755`.
   - Standalone numbering and table-shape self-tests passed.
   - Standalone production lints passed.
   - `make check` invokes both blocking checks.
   - The lock diff is exactly G-INV-15 through G-INV-18 appended after G-INV-14 within the gitcmd namespace; no prior lock row changed.
   - The proving-test cells now conform to the one-token machine-readable shape.
   - The documented residual limit is honest: table shape is enforced, historical file/test resolution is not.
6. **P1 ordering prose and parent dependency list: CLOSED.** The plan now describes reverse Git date-order, and the parent SPEC lists `reflect`.
7. **Merge semantics: CLOSED for synthetic and real data.** Root/non-root behavior, a hostile merge, and the real historical merge all matched first-parent semantics.
8. **Detached HEAD routes: PARTIAL.** Missing and blob-valued detached HEADs are refused. A detached HEAD containing an annotated-tag object is still accepted; see MED-1.

Additional stable controls passed:

- User-level `diff.renames` cannot override the command-level setting.
- `i18n.logOutputEncoding=ISO-8859-1` did not alter or reject a UTF-8 payload.
- `GIT_NAMESPACE` did not hide local history in the exercised repository.
- Valid nested annotated tags and valid non-commit tag/ref routes retained their intended behavior.

## Closure against HARM/routes

The Round 2 reproductions are substantially closed, but the central append-only harm remains reachable through two new ordinary Git state inputs:

- repository-selection environment can silently redirect the adapter to a different repository while `New` is passed the intended root;
- `diff.ignoreSubmodules` can change `FilesTouched` for one SHA.

Both routes report success and can permanently append wrong or contradictory facts. A malformed detached HEAD containing an annotated tag object also remains distinguishable from the explicitly claimed fail-closed behavior only in prose, not in product behavior.

## Findings

### HIGH-1 — `GIT_DIR` and `GIT_WORK_TREE` can silently redirect the connector to another repository

Every Git process inherits the caller environment unchanged ([gitcmd.go:326](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:326)). `-C r.root` changes the working directory but does not override repository-selection environment variables.

Reproduction:

1. Create repository A with one commit `A` and repository B with a different commit `B`.
2. Set `GIT_DIR=<B>/.git` and `GIT_WORK_TREE=<A>`.
3. Call `gitcmd.New(<A>)`, then `Commits`.
4. Construction and listing both succeed, but the result contains B’s SHA, subject, and touched path—not A’s.

The hostile run returned B’s `repo-b` commit and `FilesTouched:["b"]` while the constructor had been given A’s root. This is not merely a failure mode: it is a successful ingestion of the wrong source, directly reaching the stated fabricated/incomplete-facts harm.

This also falsifies the protocol claim that repository/user configuration cannot alter payload semantics ([gitcmd/SPEC.md:25](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:25)) and the `New(root)` repository identity contract.

The finding is verified; a remedy is not. Repository-selecting Git environment must be explicitly controlled and tested without accidentally removing legitimate execution environment. A fix should prove A remains A under hostile `GIT_DIR`, `GIT_WORK_TREE`, and any other selected environment variables, while normal repositories and linked-worktree behavior still work.

### HIGH-2 — `diff.ignoreSubmodules` still changes one SHA’s `FilesTouched`

The non-root file path uses porcelain `git diff --name-only` ([gitcmd.go:159](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:159)). The command pins `diff.renames=false` but not submodule-ignore semantics ([gitcmd.go:326](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:326)).

Reproduction:

1. Commit a gitlink at path `sub`.
2. Commit a change from one gitlink commit ID to another.
3. Read the same tip with `diff.ignoreSubmodules=none`: `FilesTouched == ["sub"]`.
4. Read the same tip with `diff.ignoreSubmodules=all`: `FilesTouched == []`.

Both reads succeeded against the identical SHA. This directly violates the same-commit stability guarantee and causes different `content_sha` values for one native ID. G-INV-16 correctly closes rename configuration only; it does not close this independent Git configuration route.

The finding is verified; a remedy is not. The first-parent tree-delta contract needs explicit, configuration-independent gitlink semantics, discriminated for both root and non-root commits and for repository and user-level configuration.

### MED-1 — A detached HEAD containing an annotated-tag object is accepted despite the remediation’s non-commit-HEAD claim

Detached HEAD validation uses `HEAD^{commit}` ([gitcmd.go:284](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:284)), which accepts any object peelable to a commit rather than requiring the immediate HEAD object to be a commit.

Reproduction:

1. Create a normal commit and an annotated tag pointing to it.
2. Write the annotated tag object’s SHA directly into `.git/HEAD`.
3. Call `Commits` and `Tips`.
4. Both succeed. `Commits` returns the peeled commit; `Tips` publishes peeled commit IDs for `HEAD`, the branch, and the tag.

Ordinary `git checkout <annotated-tag>` writes the peeled commit into detached HEAD; an immediate tag object in HEAD is therefore a hand-corrupt/nonstandard state. The remediation journal explicitly claims detached non-commit HEADs are broken ([journal:169](/mnt/d/Users/thamm/Desktop/Projects/Vera/notes/journal/2026-08-24.md:169)), but the committed test covers only a blob.

This does not corrupt the commit payload in the reproduced case, so severity is MED rather than HIGH. It does, however, leave a claimed broken-repository route reporting success. The contract must explicitly choose immediate-object or peelable-to-commit semantics; either a product rejection test or a corrected narrower claim is required. The finding does not pre-verify either remedy.

## LOW

None.

## Acceptance rationale

Task 4 remains unaccepted under Law 9 at commit `ce6937003a1bb3d655425e993f439aee113ca2ee`.

Round 2’s specific HIGH findings and mechanism defects are closed, and the restored numbering/table mechanisms are sound within their documented scope. Acceptance still fails because two new HIGH routes reach the exact append-only harm: one can ingest an entirely different repository under the requested root, and one can re-key the same commit through ordinary submodule diff configuration. The detached annotated-tag HEAD also leaves the stated fail-closed boundary unresolved.

Fix the invariant classes rather than only these fixtures, add appended invariant identities where needed, regenerate the lock in the same commit, test the remedies against linked worktrees and root/non-root gitlinks, rerun bare `make check`, race tests, and mutation sweeps, then request another non-author verdict.
