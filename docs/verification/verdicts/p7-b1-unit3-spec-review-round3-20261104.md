---
schema: vera.verdict.v1
verdict_id: p7-b1-unit3-spec-review-round3-20261104
status: NEEDS_CHANGES
reviewed_commit: db2a9a13407fbeeb12c703bf6da46fd170cc2040
findings:
  - finding_id: p7-b1-unit3-spec-review-round3-N4
    severity: MED
    summary: the §5.1 "swapped binary" negative control trips at the step-2 archive-hash gate before launch, so it never exercises the TWIN-INV-18 /proc/<pid>/exe witness that N3 made load-bearing for proof_binding=PASS — the strongest new witness is uncalibrated
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round3-N1
    severity: HIGH
    summary: round-2 N1 RESOLVED — witness verification relocated onto the fixture consistently; no residual self-verify claim
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round3-N2
    severity: MED
    summary: round-2 N2 RESOLVED — false integrity-script precedent removed; no-automated-gate residual disclosed deliberately
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round3-N3
    severity: LOW
    summary: round-2 N3 RESOLVED — proof_binding=PASS gated on launched-process witness, determinism necessary-not-sufficient
    verdict: CONFIRMED
artifact_path: docs/verification/verdicts/p7-b1-unit3-spec-review-round3-20261104.md
artifact_sha: 53c9d137d784ac6baf45163d7368284d5a10b9c7377288d2e4f686de2c30b48b
---

# P7 B1 unit-3 runtime-fixture spec (round 3) — non-author re-verify

## Provenance

- **Under review:** the round-3 revision of the DRAFT spec
  `docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md` (authored this session, commit
  `db2a9a1`; spec-first per Build Law 6 — no code under review). This re-verifies the round-2
  findings N1 (HIGH), N2 (MED), N3 (LOW).
- **HEAD at dispatch and at receipt:** `db2a9a13407fbeeb12c703bf6da46fd170cc2040` — unchanged
  between dispatch and receipt (spec-review only; the tree was not modified during the review).
- **Driving mechanism:** `/proofbound-review` (CLAUDE.md pre-merge adversarial review), round 3.
  Tier: connector/peripheral — a single adversarial review of a design document (no mutation
  evidence; no code yet).
- **Reviewer identity:** `dispatched-non-author-reviewer-round3-20261104` — an independent reviewer
  dispatched via the Agent tool; it did NOT author the spec or any revision. Adjudicated by this
  session, which **authored the round-3 revision** — so this session is NOT a valid acceptor of it;
  the acceptance capacity comes from the non-author reviewer. The verdict returns NEEDS_CHANGES
  (one open MED), so acceptance is not at issue. Law 9 satisfied.
- **Adjudicator independent verification:** the three resolved findings were confirmed against the
  round-3 text (every PASS/witness/verification mention now locates verification in the fixture, not
  `DeriveCaseStatus`; §6.1 correction matches `scripts/p6-artifact-integrity.sh:41`'s verdicts-only
  walk; the N3 `proof_binding` clause is present). The **new finding N4 is confirmed from the spec's
  own structure**: §4 orders the fixture as step 2 (verify `ArchiveSHA256` against the committed
  manifest) → step 4 (launch) → step 6 (witnesses). A negative control that swaps/alters the archived
  binary changes `ArchiveSHA256` and so fails at step 2, **before any postmaster launches**, so the
  TWIN-INV-18 `/proc/<pid>/exe` comparison is never reached — the control cannot distinguish a real
  `/proc` witness from an archive-hash rubber-stamp. N4 is not a false-PASS soundness hole (no path
  yields PASS without a matching binary), but it leaves the load-bearing witness uncalibrated.
- **Cleanroom redaction:** none required; the report below contains no absolute paths or usernames.

## Findings asymmetry notice

Per `VD-verification-asymmetry-2dyjnd`, a verified finding transfers zero trust to its suggested
remedy. The reviewer's sketch (swap the launched binary without tripping the archive-hash gate) is a
proposal, not an accepted fix; the author decides the remedy and a non-author re-verifies.

## Verdict (verbatim reviewer report)

The following is the dispatched reviewer's report, reproduced verbatim.

---

Round-3 non-author re-verify — P7 B1 unit-3 runtime-fixture spec (HEAD db2a9a1). I did not author this spec or any revision. Checked the round-3 spec in full against the round-2 verdict, casemanifest.go:195-425, scripts/p6-artifact-integrity.sh, and CLAUDE.md Laws 1/9.

PER-FINDING

N1 (HIGH) — RESOLVED. The witness-verification claim is now consistently and honestly relocated onto the FIXTURE in every place that mentions PASS/witnesses/verification. §0, §1 (both bullets — the case-manifest bullet now states plainly "the committed manifest is a declaration, not self-verifying … neither re-hashes a witness nor launches anything (casemanifest.go:197-222, 398-425) … forgeable by declaration"), §1 "What actually establishes the runtime fact", §1.1 soundness chain ("anchored on the fixture, not on the committed artifact"; "committed case_status=PASS is … a record … not itself the proof (N1: DeriveCaseStatus re-verifies nothing)"), §4 step 6 ("the fixture — not DeriveCaseStatus — is the gate"), §5 header, §5.1, §6, §9 item 3 and §9.3a all agree. A grep for DeriveCaseStatus/re-verify/self-verify/forgeable turns up NO residual sentence claiming the committed artifact self-verifies; the old round-2 wording ("DeriveCaseStatus verifies each witness digest before folding", "re-verified by DeriveCaseStatus") is gone. This matches the code: DeriveCaseStatus is a pure worst-wins fold and validateOutcome only checks evidence-list well-formedness. The §1.1 chain is now honestly anchored on fixture + §9.3a live run + §5.1 negative control, and the claim-relocation is logically sufficient: acceptance (not the committed file) establishes the runtime fact, exactly as every other evidence artifact in this repo draws trust from the non-author who ran it. Spec also explicitly notes that a committed-artifact cross-check gate would change the frozen DeriveCaseStatus contract and is out of scope — correct scoping.

N2 (MED) — RESOLVED (with a disclosed residual). §6.1 now carries an explicit "Honest correction (round-2 N2)": these artifacts are NOT covered by p6-artifact-integrity.sh — confirmed, that script does `find "$verdict_dir" -maxdepth 1` over docs/verification/verdicts/ only (line 41), self-digest-validates only front-matter verdicts there plus two hard-coded P5/P6 exhibits. The false precedent is removed; the "no automated integrity gate" residual is stated plainly as a deliberate choice, not hidden, with optional future hardening noted as out of scope. The no-@generated/documentary conclusion is legitimate under Law 1: the generating input (offline archive) is intentionally absent from make check, so the artifacts are one-time captured evidence, not in-CI-regenerable derived state — the same basis on which units 1-2 build/verify-evidence are already accepted unmarked. Residual (now disclosed, not a defect): the committed case-manifest JSON has no integrity hook AND DeriveCaseStatus re-verifies nothing, so a committed case_status=PASS could be hand-edited post-acceptance undetected by make check. Spec acknowledges this; it is an owner acceptance call, not an over-claim.

N3 (LOW) — RESOLVED. §5 adds an explicit clause: "proof_binding = PASS … only when the launched-process witness (TWIN-INV-18) verifies; the determinism check (TWIN-INV-22 / EvaluateLock==LOCKED) is necessary but not sufficient … A run that reproduces the proof digests (LOCKED) but whose /proc/<pid>/exe does not equal PostgresBinarySHA256 yields proof_binding = FAIL." Consistent with the frozen nine-key CASE-INV-3: no new key added; both witnesses still map to proof_binding, gated at fixture-authoring level.

NEW FINDING

N4 (MED) — The §5.1 negative control, as specified, does not actually exercise the launched-process witness (TWIN-INV-18) — the very witness N3 just made load-bearing for proof_binding=PASS. §4 orders the fixture: step 2 verifies ArchiveSHA256 against the committed manifest, step 4 launches, step 6 runs the witnesses. A "binary swapped / bytes altered" archive (§5.1, §8 item 3) changes ArchiveSHA256 (and the inventory's PostgresBinarySHA256), so it fails at the step-2 archive-hash gate BEFORE any postmaster launches — the TWIN-INV-18 /proc/<pid>/exe comparison is never reached. Failure scenario: a fixture that rubber-stamps proof_binding on "ArchiveSHA256 == committed" and never genuinely reads /proc/<pid>/exe would still pass the specified negative control (swapped archive → no PASS, for the wrong reason), giving the §9.3a non-author false confidence that the launched-process witness has teeth. This is precisely the Law 9 corollary concern (a positive control that must die must kill the right thing). To calibrate TWIN-INV-18 the control must swap the launched binary at the launch layer while keeping the committed archive/manifest consistent (so step 2 passes and only step 6 fails) — the spec does not specify how to do this, and its black-box §9.3a run plus a step-2-tripping control do not distinguish a real /proc check from an archive-hash rubber-stamp. Not a false-PASS soundness hole (no path yields PASS without a matching binary), but it undercuts the calibration of the strongest new witness.

CONSISTENCY CHECKS THAT PASS: frozen vera.replay.v1 proof shape (§7 non-goals) preserved — witnesses live in the case-manifest, not the proof. TWIN-INV-18..22 append cleanly (TWIN-INV-5 tombstone). No new CASE-INV key. Law 6 honored (spec-first, no code, DEF-P7-001 unmoved, CI-p7-runtime-fixture-1a2b3c stays proposed). §8.4 no-fetch owner decision correctly left open. No path to acceptance (as opposed to the forgeable committed artifact in isolation) exists without a real from-source postgres run, PROVIDED the non-author's §9.3a run and negative control genuinely exercise TWIN-INV-18 — which is exactly what N4 flags as not yet guaranteed.

OVERALL RECOMMENDATION: The three round-2 findings (N1/N2/N3) are honestly and correctly resolved; the LOCKED→determinism down-scoping and the fixture-as-gate relocation are sound. I recommend a small round 4 (or a targeted spec clause) before the change intent leaves `proposed`, to fix N4 — specify a negative control that swaps the launched binary without tripping the step-2 archive-hash gate, so the launched-process witness that now gates proof_binding=PASS is actually calibrated. N4 does not create a false-PASS path, so if the owner prefers, it could instead be pinned as an explicit implementation-gate requirement rather than a full round 4 — but because it concerns the exact witness N3 hinges PASS on, it should be nailed down in text before implementation, not discovered during it. §8.4 owner decision remains correctly pending and is not a blocker.

---

## Adjudication summary

- **Status: NEEDS_CHANGES** (one open MED). The three round-2 findings are RESOLVED; the
  fixture-as-gate relocation (N1) and the Law-1 correction (N2) are sound and honest. The spec is
  now structurally correct — the one remaining item is a **control-calibration** gap, not a
  soundness hole.
- **N4 is the Law-9-corollary bar applied to the strongest new witness:** the TWIN-INV-18 negative
  control must be specified to swap the **launched** binary after the step-2 archive-hash gate has
  passed (keeping the committed archive/manifest internally consistent), so the ONLY check that can
  reject the control is the `/proc/<pid>/exe` witness — proving that witness is load-bearing and not
  rubber-stamped by the archive hash. The round-4 revision adds this to §5.1 (and the §9 gate), and a
  non-author re-verifies before `CI-p7-runtime-fixture-1a2b3c` leaves `proposed`.
- **§8.4 (no-fetch mechanism) remains the one open owner decision** and is not a blocker to the N4
  correction.
