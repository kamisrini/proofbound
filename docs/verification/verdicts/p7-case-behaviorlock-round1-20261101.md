---
schema: vera.verdict.v1
verdict_id: p7-case-behaviorlock-round1-20261101
status: ACCEPTABLE
reviewed_commit: 6dc93724a782414d936376ac2b450834dad7693b
findings: []
artifact_path: docs/verification/verdicts/p7-case-behaviorlock-round1-20261101.md
artifact_sha: 057b0bce341f0cd73563d80fb2c58d33291707c5711efd3f93bf3715cdc403da
---

# P7 case-manifest + ReplayIsolated behavior-lock — non-author verdict

## Provenance

- **Under review:** commit `6dc93724a782414d936376ac2b450834dad7693b` (the two non-runtime
  contracts `proofbound.p7.case-manifest.v1` and `proofbound.p7.behavior-lock.v1` in
  `kernel/internal/twin/manifest`, their tests, the package `SPEC.md`, `doc.go`, and the two
  change-intent records).
- **HEAD at dispatch and at receipt:** `6dc93724a782414d936376ac2b450834dad7693b` — unchanged
  between dispatch and receipt of this verdict.
- **Driving mechanism:** `/proofbound-review` (CLAUDE.md pre-merge review); the reviewer
  independently ran `scripts/mutants.sh internal/twin/manifest`, `go test`, `go vet`, and
  `scripts/p6-artifact-integrity.sh --check`.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261101` — an independent reviewer
  dispatched via the Agent tool by the authoring session, which did NOT write the code. Law 9 is
  satisfied: this is an acceptance, not a self pre-check.
- **Code author/committer:** `VERA Maintainer` (authoring session).
- **Cleanroom redaction:** none required; the report below contains no absolute paths or usernames.

## Verdict (verbatim reviewer report)

status: ACCEPTABLE
reviewed_commit: 6dc93724a782414d936376ac2b450834dad7693b

Mutation calibration + summary (verbatim from scripts/mutants.sh internal/twin/manifest):
calibration neutral=survived invalid=invalid lethal=killed
summary candidates=335 killed=330 invalid=0 survived=5

All five survivors are pre-declared equivalents in docs/allowed-survivors.txt, and the allowlist's line#ord keys match the sweep exactly (declaration.go:189#30, 192#31, 216#41, 234#37; manifest.go:382#59). ZERO survivors in casemanifest.go or behaviorlock.go — the two files under review. Calibration controls all behaved (neutral survived, invalid was invalid, lethal was killed), so the counts are believable.

Independent checks run:
- cd kernel && go test -count=1 ./internal/twin/manifest/ → ok (0.064s). go vet ./internal/twin/manifest/ → clean.
- bash scripts/p6-artifact-integrity.sh --check → EXIT=0 (both change-intent record self-digests validate).
- grep for os/net/os-exec/database-sql/http in the two new .go files → none; only the documentary JSON field name "network_policy" appears. EvaluateLock executes no replay, fetches nothing, runs no binary/PostgreSQL. Non-runtime confirmed.
- Both change-intent records (CI-p7-case-manifest-1a2b3c, CI-p7-behavior-lock-1a2b3c) explicitly exclude runtime execution, DEF-P7-001 activation, and any PASS/satisfaction claim. Commit message states "Author-green only; non-author verdict pending" — correct Law 9 posture (I am that non-author verifier). SPEC.md exists and its Interface block matches the shipped exports; CASE-INV-1..6 and LOCK-INV-1..6 are documented with named proving tests (Law 6 satisfied).

findings:
I found no BLOCKING or HIGH defects. Two LOW observations:

- LOW — behaviorlock.go:405-412 (candidateDigest), and by extension the *Digest helpers. Harm: the digest FORMULAS are never pinned to a known-answer value. Every test derives the expected digest by calling the same production function (behaviorlock_test.go:56, :218 both call candidateDigest(...); the round-trip/digest tests hash the bytes the code just produced), so a silent formula regression — e.g. changing the newline join separator in candidateDigest to "" — would be self-consistent on both sides of ValidateBehaviorLock and invisible to the suite. The "proof candidate_digest off" test (behaviorlock_test.go:169) sets a constant cb1 and would still reject against any deterministic formula, so it does not discriminate the formula either. This is not a sweep survivor: the mutation engine (tools/mutants/main.go) only swaps the four binary operators ==/!=/&&/||, so line 410 is uncovered by construction, not left undefended. Real-world bound: candidate_digest must agree with the external vera.replay.v1 proof producer; a drift there would make EvaluateLock derive UNVERIFIED for genuine proofs (fails closed — never a false LOCKED or false PASS), so this is a cross-component spec-drift hardening gap, not a soundness hole. A known-answer assertion (e.g. candidateDigest(lockCandidates()) == a hardcoded hex) would close it.

- LOW/informational — behaviorlock.go:420-432 (EvaluateLock). Harm: the LockEvaluation fields Identity, SourceCandidate, and Baseline are set to VERIFIED unconditionally; only Proof and Intent feed Overall (Overall=LOCKED iff Proof==VERIFIED && Intent==VERIFIED). This is sound because ValidateBehaviorLock already enforces identity/source/candidate/baseline consistency and EvaluateLock returns an error (not a derived result) for any invalid lock, so no false LOCKED is reachable and SPEC LOCK-INV-5/6 hold. But the three reported criteria are cosmetic — they always read VERIFIED for any lock that validates — so a consumer should not treat them as independent per-criterion checks. No failing scenario exists (it fails closed); noted only for report honesty.

Scenarios I tried to break and could not: PASS-by-absence (rejected — PASS needs evidence naming an existing artifact role, and all-nine-PASS is the only fold to PASS); case_status mismatch vs worst-wins fold (rejected); case_status forced to UNDECLARED (impossible — fold never yields it); unknown/missing/null/duplicate/extra fields in case-manifest and behavior-lock, including nested command/witness/artifact and the intent_binding {state} vs tuple disjunction (all rejected via exactObject + core.Canonicalize duplicate-key/non-canonical rejection); mismatched/empty/unknown-field proof bytes, UNDECLARED intent with a valid proof, nil resolver on a declared intent, resolver error (all derive UNVERIFIED, never LOCKED); baseline through_seq/source_digest divergence from source_prefix (rejected); sequence_map not bijective with candidate ids or non-ascending (rejected); a 40-hex (git-rev) parent where a 64-hex lock digest is required (rejected). evaluateProof's full-struct == over ReplayProof compares schema + all five members, so no partial/zero-value proof matches a validated lock.

Reviewer summary: I independently read the SPEC, both .go files, both test files, the shared helpers (manifest.go validateRevision/validateSHA256/validateRelativePath/exactObject, core.Canonicalize), and the two change-intent records; ran the package tests, vet, the artifact-integrity check, and a full calibrated mutation sweep (335/330/0/5, controls green, no new survivors in the reviewed files); and tried a battery of wrong-accept/wrong-reject constructions against every CASE-INV and LOCK-INV. Each invariant is true of the code and its named proving test genuinely discriminates it, with the single qualification that digest-formula identity (candidateDigest in particular) is pinned only for change-sensitivity, not for known-answer value — a LOW hardening gap that fails closed. The package is non-runtime, claims no runtime PASS, and does not touch DEF-P7-001. I recommend ACCEPTABLE. Confidence: high.

## Disposition (author note)

Accepted at the reviewed commit. The two LOW observations fail closed (they can only ever produce a
conservative `UNVERIFIED`, never a false `LOCKED`/`PASS`) and are non-blocking. The known-answer
digest pin (first LOW) is recorded as an optional future hardening; applying it would move the tree
off this reviewed commit and is not worth re-opening an ACCEPTABLE acceptance for a fail-closed LOW.
