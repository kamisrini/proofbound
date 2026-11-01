---
schema: vera.verdict.v1
verdict_id: p7-postgres-source-verify-hardening-round2-20261101
status: ACCEPTABLE
reviewed_commit: 95dfb478a73ad7f4752c31b51c4e7ea01e581991
findings:
  - id: 1
    severity: LOW
    law: 2
    summary: test `mkman` helper still emits `download_sha256` in its throwaway synthetic manifest; cosmetic, not a drift seed, no false-accept
artifact_path: docs/verification/verdicts/p7-postgres-source-verify-hardening-round2-20261101.md
artifact_sha: 4056fab0701685cd4afbf6c5e9ab56a6b91b937021d2bb0f57c6abbae2254568
---

# P7 B1 unit-1 hardening (round 2) — non-author verdict

## Provenance

- **Under review:** the round-2 fix `10b4204` (rewrite the self-test symlink positive control to
  isolate the `! -L` guard), together with the underlying hardening `d9590da`, reviewed at HEAD
  `95dfb478a73ad7f4752c31b51c4e7ea01e581991`. This closes the NEEDS_WORK finding of
  `p7-postgres-source-verify-hardening-round1-20261101`.
- **Driving mechanism:** `/proofbound-review` (CLAUDE.md pre-merge review), connector/peripheral tier.
- **Reviewer identity:** an independent reviewer dispatched via the Agent tool by the authoring
  session, which did NOT write the code. Law 9 is satisfied: this is a non-author acceptance.
- **Code author/committer:** `VERA Maintainer` (authoring session).
- **Cleanroom redaction:** none required; the report below contains no absolute paths or usernames.

## Verdict (verbatim reviewer report)

status: ACCEPTABLE
reviewed_commit: 95dfb478a73ad7f4752c31b51c4e7ea01e581991
what I ran:
- `git rev-parse HEAD` -> 95dfb478a73ad7f4752c31b51c4e7ea01e581991 (confirmed).
- `git show d9590da` and `git show 10b4204` -> read both diffs. d9590da adds the `! -L` guard + drops download_sha256 from the pinned manifest; 10b4204 rewrites ONLY the self-test symlink control (harness-only, verifier byte-unchanged between the two).
- Read all three files at HEAD in full (verifier scripts/p7-postgres-source-verify.sh, self-test scripts/tests/p7-postgres-source-verify.test.sh, manifest docs/verification/p7-replay-lock-fixtures/postgres-source-16.4.json).
- Real self-test `bash scripts/tests/p7-postgres-source-verify.test.sh` -> "p7-postgres-source-verify.test: ok", exit 0.
- CENTRAL CHECK: made a stripped copy of the verifier in a scratch dir with ONLY the `[[ ! -L $tarball ]] || die ...` line removed (diff confirmed exactly one line gone; the real repo file never touched, `git diff --stat` empty). Repointed a copy of the test's `verify=` at the stripped copy and ran it -> it printed "symlink passing both hash and size (guard isolation) was accepted" and exited NON-ZERO (STRIPPED_SELFTEST_EXIT=1). The control is genuinely load-bearing.
- STEP 6 hand-reconstruction: built the isolation fixture by hand. payload=40 content bytes; symlink target path = 40 'a' chars so `stat -c %s` on the symlink = 40 = pinned size; `sha256sum` (deref) = pin. Against the stripped verifier the fixture was ACCEPTED (exit 0, "OK ... size=40"); against the real verifier it was REJECTED with "tarball must be a regular file, not a symlink" (exit 1). Mechanism confirmed: both the hash check and the size check pass, so only the `! -L` guard can reject it.
- False-accept battery against the REAL verifier (all 11 PASS): pristine accepted; tampered byte rejected; wrong size rejected; manifest!=checksum-file divergence rejected; bytes-vs-pin rejected; missing tarball rejected; uppercase-hex pin rejected; short-hex pin rejected; symlink-to-wrong-bytes rejected; dangling symlink rejected; empty PATH (jq/sha256sum/stat missing) fails closed.
- `grep -rn download_sha256 .` and `grep @generated` on the three files; `grep hooks-test` in Makefile (the self-test is auto-discovered via `scripts/tests/*.test.sh`).

findings:
1. LOW / cosmetic — the test's `mkman` helper still emits `download_sha256:"$1"` in its synthetic throwaway manifest, even though d9590da removed that field from the real pinned manifest. The verifier never reads it, the test manifest is regenerated per run and is not an authoritative home, so this is not a Law-2 drift seed and cannot cause a false-accept. Purely a leftover inconsistency with the manifest cleanup intent. Not blocking.
No false-accept defects found.

scenarios I tried and could not break: stripping the guard (control fires), hand-built guard-isolation symlink (hash+size both pass, guard rejects), symlink to wrong bytes, dangling symlink, tampered byte, +1 size, manifest/checksum divergence, bytes-vs-pin with agreeing-but-wrong pin, missing tarball, uppercase and short hex pins, and jq/sha256sum/stat unavailable (fails closed).

build-laws check:
- Law 1: no `@generated` markers on any of the three hand-authored files; nothing claims to be derived. OK.
- Law 2: the duplicate `download_sha256` drift seed is removed from the pinned manifest; the verifier reads `published_sha256` only. Remaining grep hits are legitimate (round-1 verdict + journal/state narrative, the plan's schema-field list, setup-sources.json where download_sha256 legitimately differs from consumed_sha256, and the test's throwaway synthetic manifest — finding 1, cosmetic). OK.
- Law 3: the positive control now fails when the enforcement is removed — proven by the stripped-guard self-test exiting non-zero. This closes the round-1 "test that could not fail" defect. The prior controls (tamper/size/divergence/bytes-vs-pin/missing) are all still present and independently reject. OK.
- Law 9: this is the owed non-author re-verify of 10b4204; delivered as a non-author verdict. OK.

reviewer summary: The round-2 fix is correct and the harness defect from round 1 is genuinely closed. The symlink control is now load-bearing: deleting the `! -L` guard makes the self-test fail with the intended message, and I reproduced the isolation fixture by hand to confirm both the hash and size checks pass so only the guard can reject it. No false-accept regressions across an 11-case battery; the verifier fails closed on missing tooling and on every tamper/divergence/symlink variant. The one finding is a cosmetic leftover `download_sha256` in a throwaway test manifest with no provenance impact. High confidence: ACCEPTABLE.

## Disposition (author note)

Accepted at the reviewed commit `95dfb47`. The round-1 Law-3 defect is closed: the reviewer
independently proved the control load-bearing (stripped-guard self-test exits non-zero) and
reconstructed the isolation fixture by hand. Finding 1 is LOW/cosmetic — the `mkman` helper's
synthetic throwaway manifest still carries `download_sha256`, but it is regenerated per run, never
read by the verifier, and is not an authoritative home, so it is not a Law-2 drift seed and cannot
affect any outcome. Following the standing precedent for a non-false-accept cosmetic LOW (not worth
moving the tree off an accepted commit and re-opening the non-author re-verify cycle), it is recorded
here as optional future cleanup rather than fixed under this acceptance. Unit-1 source provenance and
its hardening are accepted; `DEF-P7-001` stays OPEN/BLOCKED pending unit 2 (build + license/notice
inventory) and unit 3 (runtime fixture).
