---
schema: vera.verdict.v1
verdict_id: p7-b1-unit3-spec-review-round6-20261104
status: ACCEPTABLE
reviewed_commit: 9393af4e70711cbd5071e47373d9aeaa2c66404a
findings:
  - finding_id: p7-b1-unit3-spec-review-round6-NEW1
    severity: LOW
    summary: §5.1 "closes the N8 gap by construction" slightly overstates; the forgery is reduced to a contract violation caught by §9.3a code inspection / Law-9 review, which the same sentence states — cosmetic wording only
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round6-NEW2
    severity: LOW
    summary: §5.1 / §0 "the same call site the unit test calibrates" is imprecise shorthand for "the same function/implementation" (test and fixture are two call sites of one sole-producer function) — cosmetic wording only
    verdict: CONFIRMED
artifact_path: docs/verification/verdicts/p7-b1-unit3-spec-review-round6-20261104.md
artifact_sha: 731df279a3495132520f585fc1805f44f8af5635359a54fc0f0bae119df3ff34
---

# P7 B1 unit-3 runtime-fixture spec (round 6) — non-author re-verify — ACCEPTABLE

## Provenance

- **Under review:** the round-6 revision of the DRAFT spec
  `docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md` (authored this session, commit
  `9393af4`; spec-first per Build Law 6 — no code under review). This re-verifies round-5 findings
  N8 (MED) and N9 (LOW).
- **HEAD at dispatch and at receipt:** `9393af4e70711cbd5071e47373d9aeaa2c66404a` — unchanged
  between dispatch and receipt (spec-review only; the tree was not modified during the review). Two
  cosmetic wording nits (NEW-1/NEW-2 below) are applied in a follow-up commit immediately after this
  verdict; they change no substance and, per the reviewer, need no re-review — this ACCEPTABLE
  verdict is about `9393af4`, which carries the loose phrasings as noted.
- **Driving mechanism:** `/proofbound-review` (CLAUDE.md pre-merge adversarial review), round 6.
  Tier: connector/peripheral — a single adversarial review of a design document.
- **Reviewer identity:** `dispatched-non-author-reviewer-round6-20261104` — an independent reviewer
  dispatched via the Agent tool; it did NOT author the spec or any revision. Adjudicated by this
  session, which authored the round-6 revision; the **acceptance capacity comes from the non-author
  reviewer** (this session is not a valid acceptor of its own revision). **This is an ACCEPTANCE of
  the spec document** (Law 9 satisfied: a non-author verdict).
- **Scope of this acceptance:** it accepts the unit-3 runtime-fixture **spec** (the contract,
  invariants TWIN-INV-18..22, controls, and §8 decisions) as sound and implementation-ready. It does
  **not** accept any produced fixture, witness, proof, `case_status`, or `DEF-P7-001` movement — each
  remains separately gated behind the governed implementation, `make mutants` + a non-author
  ACCEPTABLE package verdict (ROADMAP P1 bar, Law 9), and the §9.3a live acceptance run. The change
  intent `CI-p7-runtime-fixture-1a2b3c` is still `proposed`; the **owner** flips it to `accepted` to
  authorize the governed `kernel/*` work.
- **Adjudicator note:** the no-false-PASS reasoning was re-checked against the earlier rounds'
  code findings and holds — for the shipped black-box configuration no accidental divergent-binary
  PASS exists (all on-disk checks and the `/proc` witness bind the same `PostgresBinarySHA256`, and
  the staged file is exec'd unchanged), and the one residual deliberate-forgery path (a fixture
  copying the digest instead of reading `/proc`) is reduced to a contract violation caught by the
  §9.3a code inspection and the Law-9 governed-kernel review of the fixture source — which the spec
  now states honestly rather than over-claiming against. No "test that cannot fail" remains in the
  control set.
- **Cleanroom redaction:** none required; the report below contains no absolute paths or usernames.

## Findings asymmetry notice

Per `VD-verification-asymmetry-2dyjnd`, the two LOW wording remedies below are the reviewer's
proposals; they are applied verbatim as cosmetic edits in the follow-up commit and introduce no new
claim, so they do not themselves require a fresh non-author verdict.

## Verdict (verbatim reviewer report)

The following is the dispatched reviewer's report, reproduced verbatim.

---

ROUND-6 NON-AUTHOR RE-VERIFY — P7 B1 unit-3 runtime-fixture spec (HEAD 9393af4, file docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md). I did not author this spec or any revision. Spec-first per Law 6; no code under review.

## Per-finding verdicts

N8 — RESOLVED. All three required elements are present and mutually consistent:
- Over-claim withdrawn: §5.1 "Single call site" paragraph (lines 329–330) states resolved_sha256's value "is non-discriminating on the positive path ... so inspecting the value cannot by itself prove the /proc read happened. The spec does not claim otherwise," and explicitly labels the earlier "cannot populate resolved_sha256" wording "a false over-claim ... withdrawn." No residual sentence anywhere asserts the observable alone proves the read (checked §1.1, §5 table, §8.3, §9.3a).
- Single sole-producer, bound to the proof_binding path: §5.1 (lines 332–334) and the TWIN-INV-18 row (line 271) both require witnessLaunchedProcess to be the single implementation and sole producer of resolved_sha256, with the fixture's production proof_binding emission obtaining it "by calling that one function ... no second path, no literal"; a direct literal assignment is called out as a contract violation.
- §9.3a code inspection: §9 item 3a (lines 464–468) requires the non-author to "confirm by code inspection that resolved_sha256 on the fixture's proof_binding path is produced by the single calibrated witnessLaunchedProcess call reading /proc/<pid>/exe, not a literal or a copy," and names an acceptance that skips this as not having calibrated the witness.
Taken together these close the declaration-forgery path to the extent a spec can: the design provides no sanctioned second producer, and the residual deliberate-forgery path is correctly reduced to "caught by §9.3a code inspection / Law-9 governed-kernel source review" (stated at lines 338–340). No internal contradiction with §1/§4's "the fixture is the gate" framing (line 340 explicitly ties the two together) or with the frozen DeriveCaseStatus/CASE-INV contracts (§5 still routes TWIN-INV-18+22 to the single frozen proof_binding key and does not add a committed-artifact re-hash gate).

N9 — RESOLVED. The §0 Round-4 paragraph (lines 52–58) no longer presents any "§9.3a requires the swap control" sentence; it carries an inline "(Superseded in round 5: the swap control is now optional and the required calibration is a direct witness-function unit test ...)" note that is consistent with the live §9.3a (witness-function unit test + resolved_sha256 inspection + the new code-inspection step). No stale swap-control requirement is presented as current.

## No-false-PASS assessment
Confirmed: for the shipped black-box config there remains no accidental divergent-binary PASS (ArchiveSHA256 / inventory walk / staging re-hash / PostgresBinarySHA256 all bind the same digest and the staged file is exec'd unchanged, so a divergent binary dies at step 2 before /proc). The only residual PASS-not-entailing-a-real-/proc-read path is the deliberate literal-copy forgery, and round 6 reduces it exactly to "contract violation caught by code inspection / Law-9 review" — it does not claim more (the "guarantee is the single bound, calibrated, source-reviewed call site," line 340). No "test that cannot fail" in the control set: absent→Skip, tampered-archive→step-2 reject, witness-function negative (distinct error)/positive (match), TWIN-INV-20 injected mutation MUST FAIL, TWIN-INV-21 suppressed removal MUST FAIL — each can fail.

## New findings (non-blocking)
- NEW-1 (LOW, wording). §5.1 line 338 says the round-6 construction "closes the N8 declaration-forgery gap by construction." Read in isolation "by construction" slightly overstates — the forgery is not structurally impossible; a fixture can still violate the MUST-call contract in source, and the actual enforcement is the §9.3a code-inspection / Law-9 review named in the very same sentence and at line 340. The sentence self-corrects ("*and* is re-checked at acceptance ... the guarantee is the ... source-reviewed call site"), so the paragraph as a whole is accurate; only the bare phrase is loose. Not blocking, but if a round-7 touch happens, "closes by construction" would read more precisely as "reduces to a contract violation caught by code inspection / Law-9 review."
- NEW-2 (LOW, wording). §5.1 line 334 and §0 line 38 call the fixture's production invocation "the same call site the §5.1 unit test calibrates." Strictly the unit test and the fixture share the same sole-producer function/implementation, invoked from two different call sites (test code vs. fixture). The surrounding text ("single implementation," "sole producer," "no second path, no literal") makes the intent unambiguous, so this will not mislead an implementer, but "same call site" is imprecise shorthand for "same function."
Neither finding introduces a new contradiction or over-claim of substance; both are cosmetic and carried from the accepted round-5 reviewer phrasing.

## Cross-reference check
Round-number and section pointers are consistent: §9 item 3a references N4/N5/N7/N8; §0 round-6 bullet, TWIN-INV-18 row, §5.1 and §9.3a all agree on the single-producer + code-inspection requirement; the §0 Round-4 superseded note matches the live §5.1/§9.3a. No stale pointer that would mislead an implementer.

## Overall recommendation
ACCEPTABLE for implementation. N8 and N9 are both RESOLVED, the no-false-PASS reasoning is sound and the spec no longer asserts any protection it lacks. The two LOW items are cosmetic wording nits that do not need another review round — the owner can flip the change intent (CI-p7-runtime-fixture-1a2b3c) to accepted, optionally tightening the "by construction" / "same call site" phrasing in passing. This does not require a round 7.
(Note: I am a reviewer only — I made no file edits and recorded no verdict artifact; per Law 9 the accepting non-author/owner adjudicates and records the verdict.)

---

## Adjudication summary

- **Status: ACCEPTABLE** — the unit-3 runtime-fixture spec is accepted by a non-author verdict as
  sound and implementation-ready. Six adversarial rounds drove it from "closure rests on `LOCKED`
  (pure digest arithmetic)" to a fixture-gated design whose `case_status = PASS` entails a real
  from-source `postgres` having been the launched process, observed via a single calibrated,
  source-reviewed `/proc/<pid>/exe` witness, with the no-fetch mechanism owner-settled
  ([[VD-p7-nofetch-witness-2026-11-04]]).
- **Two LOW cosmetic wording nits** (NEW-1 "by construction"; NEW-2 "same call site") are applied as
  the reviewer worded them in the immediately-following commit; they change no substance and need no
  re-review.
- **Next (owner):** flip `CI-p7-runtime-fixture-1a2b3c` from `proposed` to `accepted` to authorize the
  governed `kernel/*` implementation. Implementation then proceeds spec-first per §9 (subset prep +
  compatibility proof closing F5 → fixture + TWIN-INV-18..22 → behavior-lock/case-manifest over the
  real proof → `make mutants` + non-author ACCEPTABLE package verdict + the §9.3a live acceptance
  run). Only on that package acceptance can `DEF-P7-001` move toward RESOLVED.
