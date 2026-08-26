---
schema: vera.verdict.v1
verdict_id: task4-current-round2
status: NEEDS_WORK
reviewed_commit: 132e18716ff4890137718e0aaf89491a7d6fdd3a
findings:
  - finding_id: task4-current-round2-finding
    severity: MED
artifact_path: docs/verification/verdicts/task4-current-round2.md
artifact_sha: 4ecbeebcca73a150f382607160ec7b906b668653b7b3d5fe206d3cc459e38f74
---

NEEDS_WORK

## Round 2 scope and calibration

Independent review of frozen commit `132e18716ff4890137718e0aaf89491a7d6fdd3a`.

The repository remained unchanged: HEAD stayed at `132e187`, `git status --short` was empty, and `git diff --exit-code` passed.

- `make check` passed. Its `index stale; run make index` line was the expected positive-control output from the index self-test; the gate exited zero.
- `go test -race ./internal/connector/git/... -count=1` passed.
- The author’s gitcmd 70/70 and combined 88/88 mutation results are accepted as calibrated author evidence, not proof of semantic correctness. I did not rerun mutation.
- Hostile tests were added only to a copied module under `/tmp`. Existing tests were the neutral control; new fixtures were required to discriminate the routes below.
- `git fsck --no-dangling` passed on the real repository.
- Real merge `1bd805119df942b10ee1b43ed12e21c54475877d` now resolves through the actual adapter to `FilesTouched:["LICENSE"]`.

## Round 1 remedy adjudication

1. **Loose, packed, and detached-HEAD missing object targets: CLOSED for the stated direct-target forms.** Independent runs of `TestRepo_MissingObjectRefIsAnError` and `TestRepo_MissingDetachedHEADObjectIsAnError` passed for both `Commits` and `Tips`.
2. **Non-UTF-8 paths: PRODUCT REMEDY CLOSED, PROVING MECHANISM INCOMPLETE.** Independent real-adapter fixtures with byte `FE` filenames in both root and non-root commits were refused. The committed test exercises only the parsing helper, discussed under MED-2.
3. **Direct versus nested citations: CLOSED.** A nested-only `docs/decisions/archive/VD-*.md` fixture no longer yielded a citation; exact direct files still did.
4. **First-parent merge semantics: CLOSED for the authored fixture and current real data.** The synthetic merge returned the feature path, and the historical repository merge returned `LICENSE`. A new configuration-dependent route still makes the same SHA unstable; see HIGH-1.
5. **Ordering contract: CLOSED in the package SPEC.** It now accurately promises reverse Git date-order with parents before children, not strict timestamp sorting. The P1 plan retains the old statement; see LOW-1.
6. **Typed-nil pointer repository: CLOSED.** The constructor rejects a `(*fakeRepo)(nil)`.

The original six reproductions were repaired. Acceptance still fails because new routes reach the same append-only harm.

## Closure against HARM/routes

The central harm remains open: the same commit SHA can produce different payloads under ordinary repository configuration, and legal Git scalar bytes can be irreversibly changed while the connector reports success. A broken annotated tag is also still silently omitted by `Tips`.

- Partial-history defenses close shallow, graft, replacement, and direct missing-ref targets. A transitive missing annotated-tag target remains open in `Tips`.
- Watermark and history-scope routes remain closed.
- Stash, notes, replace exclusions, and legitimate prefix boundaries remain closed.
- Exact-tree citation resolution remains closed.
- NUL framing, valid UTF-8 hostile paths, scalar placeholder selection, and first-parent merge selection remain closed. Configuration-dependent rename detection and invalid UTF-8 scalar bytes remain open.
- Append counts, append failures, and cursor non-input behavior remain closed.

## Findings

### HIGH-1 — Local `diff.renames` configuration changes one SHA’s payload

Non-root file discovery invokes porcelain `git diff --name-only` without pinning rename behavior ([gitcmd.go:151](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:151), especially line 161). The common command wrapper pins only replacement and quote-path behavior ([gitcmd.go:299](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:299)).

Reproduction:

1. Commit `old-name`.
2. Rename it to `new-name` without changing content and commit again.
3. Read the same tip with `diff.renames=false`: `FilesTouched == ["new-name","old-name"]`.
4. Read the same tip with `diff.renames=true`: `FilesTouched == ["new-name"]`.

Both calls succeeded against the same repository and SHA. This directly violates INV-5’s same-commit stability guarantee ([git/SPEC.md:180](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/SPEC.md:180)). Since `FilesTouched` feeds `content_sha`, changing local config causes a second contradictory append for one native SHA.

This is a verified finding. A likely remedy is to define and force rename semantics independently of repository/user configuration, but no proposed flag is considered verified until a test toggles the real Git config around one SHA and proves identical payloads.

### HIGH-2 — Invalid UTF-8 scalar content is silently rewritten in the event payload

The remediation validates path streams only. Scalar fields are converted directly from Git bytes to Go strings ([gitcmd.go:107](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd.go:107), especially lines 129–138), then marshalled without UTF-8 validation ([connector.go:109](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/connector.go:109)).

I created a valid commit object whose subject bytes were:

```text
62 61 64 2d ff
```

The actual adapter returned the subject bytes unchanged as `6261642dff`, but successful JSON marshalling emitted:

```json
"subject":"bad-\ufffd"
```

The connector therefore permanently records content different from the observed commit while reporting success. The same route applies to author and committer names/emails, which follow the same byte-to-string conversion. This reaches the stated incomplete-facts harm even though path refusal itself works.

This is a verified finding. The contract must define scalar encoding behavior—such as fail-closed UTF-8 validation or a verified reversible/decoding policy—before a remedy can be accepted. Extending only the path helper does not close this route.

### MED-1 — An annotated tag object with a missing referent is silently omitted by `Tips`

Ref validation proves only that the ref’s immediate object exists ([gitcmd.go:259](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:259)). `Tips` then swallows failure to peel that object to a commit ([gitcmd.go:95](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd.go:95)).

Reproduction:

1. Create an annotated tag object that exists locally but names missing commit `1111111111111111111111111111111111111111`.
2. Point `refs/tags/broken` at that tag object.
3. `Commits` correctly errors with `fatal: bad object 111...`.
4. `Tips` returns success containing only `HEAD` and the branch, silently omitting the broken tag.

This violates the adapter protocol’s “a failure from Git is always an error” rule ([gitcmd/SPEC.md:35](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:35)) and G-INV-6’s broken-repository rule ([gitcmd/SPEC.md:48](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:48)). In a steady connector run `Commits` catches this first, so severity is below HIGH; direct `Tips` use or mutation between the two calls can still fabricate an incomplete successful cursor.

The finding is verified. Validation must cover the referent closure required to peel admitted refs, while retaining an explicit rule for valid non-commit refs. Merely checking the immediate tag object is insufficient.

### MED-2 — New remedy invariants were folded into G-INV-12 and the invalid-path citation does not prove adapter behavior

G-INV-12 remains permanently identified in the lock as “Path bytes are significant; nothing is trimmed,” but remediation also placed first-parent merge semantics into that existing invariant ([gitcmd/SPEC.md:57](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:57)). Merge semantics are a new independent correctness claim and should have received an appended invariant ID, preserving the reviewable lock diff.

The table rows for G-INV-6 and G-INV-12 now contain comma-separated test names after one `file::test` citation ([gitcmd/SPEC.md:85](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:85), [gitcmd/SPEC.md:91](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/SPEC.md:91)). That does not conform to the pinned machine-readable one-citation format in [core/SPEC.md:425](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/core/SPEC.md:425).

More importantly, `TestNULUTF8Strings_RejectsNonUTF8PathIdentity` calls only `nulUTF8Strings` ([gitcmd_test.go:117](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/gitcmd/gitcmd_test.go:117)). If the actual root or non-root file-reading path stopped calling that helper, the claimed proving test would remain green. My independent adapter fixtures prove the product currently works, but the committed mechanism cannot see future recurrence.

This is a verified mechanism defect. The appropriate structure is not prescribed here; any remedy must preserve permanent numbering, produce a reviewable lock diff for genuinely new invariants, conform to the pinned citation grammar, and include an actual-adapter invalid-path test.

### LOW-1 — The P1 plan retains the false strict ordering claim

The package contract was corrected, but [P1-flight-recorder-plan.md:69](/mnt/d/Users/thamm/Desktop/Projects/Vera/docs/plans/P1-flight-recorder-plan.md:69) still says `git rev-list --all` is “committer-date ordered.” The implementation and package SPEC correctly describe reverse Git date-order with ancestry constraints. This does not affect correctness because ordering is explicitly presentational, but Task 4’s design sources now contradict each other.

### LOW-2 — The parent SPEC dependency list omitted the new `reflect` dependency

The typed-nil remedy imports `reflect` ([connector.go:9](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/connector.go:9)), while the dependency declaration still lists the previous standard-library set without it ([git/SPEC.md:448](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/connector/git/SPEC.md:448)). The adjacent prose explicitly says the list was corrected by reading the import block because “a dependency list written from memory is a claim like any other,” making this a direct recurrence of that low-severity drift.

## Acceptance rationale

Task 4 remains unaccepted under Law 9 at commit `132e187`. Round 1’s concrete reproductions are fixed, but two HIGH routes still reach the append-only harm: repository configuration can re-key one SHA into contradictory events, and scalar byte identity can be silently destroyed. The broken-tag and proving-mechanism findings also prevent the remediation from supporting its broader claims.

Fix the invariants rather than only these fixtures, append new invariant IDs instead of broadening an old identity, regenerate the invariant lock in the same commit, add actual-adapter discrimination, rerun `make check`, race tests, and the mutation sweeps, then request another non-author verdict.
