---
schema: vera.verdict.v1
verdict_id: r2-speckit-review-20261119
status: ACCEPTABLE
reviewed_commit: 6b58186d964079d7acd81d04e3420a2426bfceb2
findings: []
artifact_path: docs/verification/verdicts/r2-speckit-review-20261119.md
artifact_sha: 71ee7532f305f001275a5af02cbc0c8c9f6bf7a5d66b746f53679dba7b128335
---

# R2 Spec Kit (6b58186): connector-tier review

## Provenance
- Reviewer: independent non-author (Law 9), one connector-tier review (Law 11); HIGH/MED only, no mutation sweep.
- Commit: 6b58186d964079d7acd81d04e3420a2426bfceb2. I wrote none of it.
- Read: the full diff of sdd/{speckit,profile,ids,build,provider,reader}.go, SPEC rows INV-58/64/68..71 and their retirements, the testdata/expected-skips and fixture-profile changes, the cli starter diff (cli.go, starter.go, starterkit/*) and starter_test.go.
- Ran, in a detached worktree under <scratch>/review-sk (Go 1.27.4, GOCACHE and TMPDIR under <scratch>):
  - the targeted sdd tests: pass;
  - the whole sdd package with `-p 1` (284.9s, one process on port 55432): pass.
  - I did not run the cli package tests separately.
- Probe: a throwaway test (deleted) committed one synthetic Spec Kit tree with 15 feature shapes and printed the mapped state. Results are below.
- The worktree is removed and pruned. The main checkout is untouched; `git status` shows only the two pre-existing untracked paths.

## Questions asked
- **OpenSpec ids, digests, skips and Resolve unchanged: yes.**
  - `openspecObligations` is the old obligation loop extracted verbatim.
  - The call site sets `r.skip` and `continue`s exactly as the old `return nil` did.
  - The owner check, digest and sort still come after it, so the order of skip causes is the same.
  - `suffix` takes the kit, and OpenSpec passes `Kit="openspec"`: same bytes as before. The pinned id vectors pass, and the test also pins `CI-install-skills-globally-57r52f`.
  - The reader's archive skip is guarded by `Archive != ""`, which an OpenSpec profile always satisfies (`literalPath` rejects "").
  - The OpenSpec profile branch keeps the same mode, archive and path checks. An OpenSpec `sibling` key was already rejected as an unknown key and is now rejected explicitly.
  - Expected-skips, expected-resolve and TestResolveIsTotal all pass.
- **State key: still distinguishes every Spec Kit state.**
  - The key is profile entry, root entry, "-".
  - `Read` lists the whole tree under root, so any in-feature change alters the root oid.
  - A Spec Kit and an OpenSpec profile can never share a key, because the profile oid differs.
- **Profile strictness: no mixed or ambiguous profile is accepted for the wrong kit.**
  - Spec Kit requires mode `feature-sibling`, no `archive` key, a sibling that is a single segment, and exactly `[root/*/spec.md]`.
  - OpenSpec requires `delta-sections`, an archive and no sibling.
  - Any other kit is `profile-invalid`.
  - `sibling` is whitelisted case-sensitively in `profileKeys`, and the `Mode`, `Status`, `Root` and `Archive` fields are read through pointers, so a missing key is distinguished from an empty one.
- **Spec Kit mapping probe (all as intended, no panic):**

| Feature shape | Result |
|---|---|
| Symlinked feature dir | `non-regular-file` skip |
| `tasks.md` is a directory or a symlink | requirement kept, change skipped `non-regular-file` |
| Symlinked `spec.md` | requirement skip `non-regular-file`, change `cascade` |
| Blank statement | `statement-blank`, change `cascade` |
| Duplicate FR id | `obligation-collision`, change `cascade` |
| `spec.md` with no obligation lines, or only fenced ones | no record and no skip, change absent |
| CRLF line endings | trailing whitespace trimmed |
| Name `a` | `slug-too-short` |
| `tasks.md` but no `spec.md` | change absent |
| Regular file under `specs/` | ignored |
| Spec without `tasks.md` | requirement only |

## Findings
No HIGH or MED findings. LOW (at most 3, one line each):
1. LOW: the obligation regex needs exactly `- **FR-001**:`. Spec Kit-style variants such as `- **FR-001:**`, `* **FR-001**: ...` and indented bullets are dropped silently, with no record and no skip. A spec mixing forms binds only the matched obligations. This is the documented INV-58 retirement, but a spec with zero matched lines is invisible rather than reported.
2. LOW: `changes.sibling` accepts any single segment, including `spec.md`, which makes every spec an instant change intent. It is the tenant's own profile, so there is no cross-tenant harm.
3. LOW: `TestInitSpecKitWritesASpecKitProfile` only checks JSON fields, despite its comment. The provider's acceptance of the starter profile is covered only by the e2e/starter-kit tests (the cli package tests were not run in this review).

## Verdict
ACCEPTABLE. Existing OpenSpec behaviour, ids and skip order are unchanged. The Spec Kit mapping handles the hostile shapes I tried without a wrong record or a panic. The profile is strict per kit. The state key remains complete without an archive line.
