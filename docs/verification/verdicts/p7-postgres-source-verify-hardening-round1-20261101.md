---
schema: vera.verdict.v1
verdict_id: p7-postgres-source-verify-hardening-round1-20261101
status: NEEDS_WORK
reviewed_commit: 0a1ab68d3be9b9f84f5ee69b8ddc27c50a2182d8
findings:
  - finding_id: 1
    severity: MED
artifact_path: docs/verification/verdicts/p7-postgres-source-verify-hardening-round1-20261101.md
artifact_sha: 184fd0ced204c126729b22152b5c9199a1b3d8687d4392737f230e3076330e73
---

# P7 B1 unit-1 hardening (round 1) — non-author verdict

## Provenance

- **Under review:** the unit-1 source-provenance hardening at commit `d9590da` (the `! -L` guard in
  `scripts/p7-postgres-source-verify.sh`; the symlink positive control added to
  `scripts/tests/p7-postgres-source-verify.test.sh`; dropping the unused duplicate `download_sha256`
  from `docs/verification/p7-replay-lock-fixtures/postgres-source-16.4.json`), reviewed at HEAD
  `0a1ab68d3be9b9f84f5ee69b8ddc27c50a2182d8`. The hardening applies the two LOW findings of the
  prior accepted verdict `p7-postgres-source-verify-round1-20261101`.
- **Driving mechanism:** `/proofbound-review` (CLAUDE.md pre-merge review), connector/peripheral tier.
- **Reviewer identity:** an independent reviewer dispatched via the Agent tool by the authoring
  session, which did NOT write the code. Law 9 is satisfied: this is a non-author adjudication.
- **Code author/committer:** `VERA Maintainer` (authoring session).
- **Cleanroom redaction:** none required; the report below contains no absolute paths or usernames.

## Verdict (verbatim reviewer report)

status: NEEDS-CHANGES
reviewed_commit: 0a1ab68d3be9b9f84f5ee69b8ddc27c50a2182d8
what I ran:
- `git rev-parse HEAD` -> 0a1ab68d3be9b9f84f5ee69b8ddc27c50a2182d8 (confirmed)
- `git show d9590da` -> read full diff: +1 guard line in verifier (`[[ ! -L $tarball ]] || die ...`), +8 lines symlink positive control in self-test, -1 line (`download_sha256`) in manifest.
- Read all three files at HEAD in full (verifier 67 lines, self-test 74 lines, manifest 16 lines).
- `bash scripts/tests/p7-postgres-source-verify.test.sh` -> `p7-postgres-source-verify.test: ok`, exit=0.
- Gate wiring: Makefile `hooks-test:` loops `for f in scripts/tests/*.test.sh; do bash "$f" || exit 1; done`, and `check: hooks-test ...` — self-test is in the gate.
- `grep -rn download_sha256 .` -> remaining hits are: the plan-amendment draft (schema prose), the round1 verdict (evidence), setup-sources.json (sibling manifest, legitimately carries it and legitimately differs there), the self-test's inline fixture manifest (harmless unused field), and notes/journal (narrative). No consumer reads `download_sha256` from THIS manifest; the verifier reads `published_sha256` only.
- Own battery in a scratch dir with my own fixture bytes: symlink->correct-content bytes REJECTED; symlink->wrong bytes REJECTED; regular file ACCEPTED; hardlink ACCEPTED (expected — indistinguishable from regular, fine); symlink->directory REJECTED ("tarball not found"); dangling symlink REJECTED ("tarball not found"); symlinked .sha256 checksum file (regular tarball) ACCEPTED but cannot bypass the pin — tampering the symlink target's digest -> REJECTED ("manifest published_sha256 != published checksum file").
- Ran actual verifier: pristine regular-file fixture -> clean OK accept; symlink fixture -> `die: tarball must be a regular file, not a symlink`.
- Regression battery: uppercase hex REJECTED ("not 64 lowercase hex"); 63-char hex REJECTED; empty PATH -> `jq is required` (fail closed); jq-present/sha256sum-absent -> `sha256sum is required` (fail closed); tamper/wrong-size/divergence/bytes-vs-pin/missing all die (self-test green).
- Law 3 isolation test: stripped the `! -L` guard from a COPY of the verifier (0 guard lines), ran the REAL self-test against that copy from repo root -> `p7-postgres-source-verify.test: ok`, exit=0 (self-test STILL PASSES with the guard removed).
- Mechanism: on the test's symlink fixture, `stat -c %s` on the symlink returns 46 (the target-path string length), `stat -Lc %s` returns 32 (real content), `sha256sum` follows the link so the hash MATCHES the pin. Guard-stripped verifier rejects with `tarball size mismatch: got 46, pinned 32` — the SIZE check, not the guard, rejects it.
- Law 1: `grep -l '@generated'` on the three files -> none (correct; hand-authored). Manifest: `grep -c download_sha256` -> 0; `jq empty` -> valid JSON.

findings:
1. MED (Law 3 / harness, broken invariant) — the added symlink positive control does NOT isolate the behaviour this commit added. Concrete: delete the `[[ ! -L $tarball ]]` guard and the full self-test still exits 0 — the symlink fixture is already rejected by the size check, because `stat -c %s` on a symlink reports the target-path string length (46 in the fixture), not the dereferenced content size (32). The commit message and the test comment both assert "only the new -L check can reject this, since hash and size would match (Law 3)"; that is false — the hash matches but the size does NOT, so the guard could be deleted or silently rot and the control stays green. Under Build Law 3 ("Enforcement proves it fires") and the Law 9 corollary ("a verdict is only as good as its harness"), a positive control that passes with the enforcement removed does not prove the enforcement fires. This is NOT a false-accept — it is a false "proof-of-firing": in production a symlinked tarball is always rejected anyway (by the size check and additionally by the guard), so the security property holds robustly as defense-in-depth. The defect is that the harness does not defend the guard it was written to defend, which is precisely the failure class this repo legislates against.
2. LOW (informational, not a defect) — a symlinked `.sha256` checksum file is accepted (the commit only guards the tarball). Not a gap: the authoritative pin is the manifest `published_sha256`, the tarball bytes must hash to it, and the checksum file is only cross-checked against the manifest; a symlinked checksum whose target digest diverges is still rejected. It cannot cause a false-accept.

scenarios I tried and could not break (no false-accept found): symlink tarball -> correct bytes; -> wrong bytes; -> directory; dangling; symlinked checksum with divergent target digest; regular file; hardlink; uppercase/63-char hex; missing jq/sha256sum (fail closed); tampered byte; wrong size; manifest/checksum divergence; bytes-vs-pin; missing tarball. No construction found where a symlinked tarball is ACCEPTED.

build-laws check:
- Law 1: PASS — no `@generated` markers; the three files are hand-authored source.
- Law 2: PASS — the same-file `download_sha256`==`published_sha256` drift seed is removed from this manifest; the verifier reads `published_sha256` only; sibling setup-sources.json legitimately retains its own fields.
- Law 3: FAIL (finding 1) — the new self-test control does not fire on the guarded behaviour; the self-test stays green with the `-L` guard removed, so it does not prove the guard fires, and the commit's Law-3 rationale is factually wrong.
- Law 9: SATISFIED by this delivery — the hardening was author-green only and the non-author re-verify was owed; this review is that verdict, and it is NEEDS-CHANGES.

reviewer summary: The hardening's runtime behaviour is correct and robust — every symlinked-tarball construction I tried is rejected, no false-accept exists, fail-closed and hex/size/divergence checks all hold, and the Law-2 drift seed is cleanly removed. The defect is in the harness: the symlink positive control added "to prove the -L guard fires (Law 3)" does not isolate the guard — the self-test passes unchanged with the guard deleted, because the size check rejects the fixture first, and the commit's own claim that "hash and size would match" is false. Given this repo's foundational stance on Law 3 and the "test that could not fail" lesson, a non-isolating positive control is a real defect, so I return NEEDS-CHANGES.

## Disposition (author note)

Finding 1 accepted as a real Law-3 defect — this is exactly the "test that could not fail" class
(Law 9 corollary). The runtime behaviour was correct (no false-accept), but the positive control did
not defend the guard: with the `! -L` guard deleted the self-test stayed green because `stat -c %s`
on a symlink reports the target-path length, not the dereferenced size, so the size check rejected
the fixture first. Fixed in `10b4204` (harness-only): the control now makes the symlink's
target-path length equal the content size, so both the hash and size checks pass and only the guard
can reject it; proven load-bearing by stripping the guard from a copy (fixture then ACCEPTED). The
misleading "hash and size would match" claim in `d9590da` is corrected in the `10b4204` message and
the test comment. Finding 2 is informational (no false-accept). The round-2 re-verify of `10b4204`
is `p7-postgres-source-verify-hardening-round2-20261101`.


## Findings — original pre-schema block (R1-4 provenance)

> The structured `findings:` in the frontmatter were normalized to `vera.verdict.v1`
> (`finding_id` + `severity` only). The **original** block, including each finding's full
> `summary`/`verdict` text and its original severity word, is preserved verbatim below so no
> reviewer detail is lost. `info`→`LOW` and `low-medium`→`MED` were the only severity remaps.
> See `docs/deferrals.tsv` → `DEF-DELIVERY-ENFORCE-VERDICTS`.

```yaml
findings:
  - id: 1
    severity: MED
    law: 3
    summary: symlink positive control does not isolate the `! -L` guard (test stays green with the guard removed); not a false-accept
```