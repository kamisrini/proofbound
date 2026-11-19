---
schema: vera.verdict.v1
verdict_id: r2-speckit-specdir-review-20261119
status: ACCEPTABLE
reviewed_commit: 16b103a81cd6c88fcb59d2effb161e56dc83c92e
findings: []
artifact_path: docs/verification/verdicts/r2-speckit-specdir-review-20261119.md
artifact_sha: ef773e7972b3933191bd37557568fc602aa54669f5f6af4b1e3624b984f7da10
---
## Provenance
Non-author reviewer (Law 9), Law 11 scoped review of 16b103a. I read the cli.go diff, `intentProviders`, `committedIntentReader`, `gitOutput`, `intentsdd` profile validation, the CLI-INV-22 and SDD-INV-56 rows, and `specdir_speckit_test.go`. I ran `go test ./internal/cli/ -run 'TestSpecKitProfileRetires|TestIntent|TestVerify'` in a detached worktree under <scratch>, with GOCACHE and TMPDIR under <scratch>. It passed. I ran one mutation (replaced the guard with `false && specKitOwnsSpecs(root)`). The `speckit on specs` case failed with "2 providers, want 1". I did not run the e2e test. The worktree is removed and the main checkout is untouched.

## Answers
1. **Silent drops: none that matter.**
   - Real specdir-format files under specs/ alongside a speckit profile are no longer ingested. Before the change that repo failed loudly, because specdir returns an error on every malformed artifact. Nobody can have relied on it, and the commit and CLI-INV-22 document the switch-off. LOW: the failure mode is now a quiet omission. A one-line note in the release notes is enough.
   - A profile committed only on a branch is invisible from main. The provider set is then unchanged, so this is the old behaviour and not a regression.
   - Empty repo or no HEAD: `gitOutput` returns "", so the function returns false and the set is unchanged.
   - `verify`: both passes call `syncIntentOnStore(…, "all")` against the same HEAD, so the provider set is identical across pass 1 and pass 2.
   - History-native sync: specdir reads each commit's tree, but the switch reads only HEAD. If HEAD declares speckit, specdir is off for all history. Records already in the append-only ledger stay. A fresh ledger would not ingest old specdir-era records. This is narrow, and HEAD is the right choice for the current-contract question.
   - LOW: `sync intent specdir` by name now errors with `unknown provider "specdir"` on a Spec Kit project. That is loud and arguably correct.
2. **SDD-INV-56 holds.** `selection == "sdd"` is still the only branch that builds the sdd provider or a GitReader. `specKitOwnsSpecs` does one `git show HEAD:<profile>` and a bare JSON decode, and never constructs or walks sdd. The `all` set can only shrink to `records`, so it never gains `sdd`. The existing TestIntentAllExcludesSDD-family tests passed.
3. **Leniency is acceptable, not fail-open.**
   - An invalid-but-speckit profile (strict validation fails) still drops specdir. `sync intent sdd` then rejects the profile loudly. `sync all` never reads specs by design, and specdir could not ingest Spec Kit files anyway. Nothing that was ingestible is lost.
   - Valid profiles need a literal root, so overlap forms such as "./specs" cannot reach the check.
   - Malformed JSON, a missing HEAD and non-speckit kits all fall to the old behaviour, which is the safe direction.
   - LOW: a profile that is invalid in an unrelated field disables specdir while sdd would refuse it. Acceptable, and arguably the profile should be linted at `init`.
4. **The tests discriminate.** `TestSpecKitProfileRetiresSpecdirFromTheIntentSet` covers five cases: none, openspec, speckit on specs, speckit elsewhere, and invalid JSON. The mutation killed it, and the control cases pin the unchanged side. The e2e test at speckit_e2e_test.go:26-36 now runs `sync intent sdd && sync all` and asserts there is no claim-skip.

## Verdict
ACCEPTABLE. No HIGH or MED findings.
