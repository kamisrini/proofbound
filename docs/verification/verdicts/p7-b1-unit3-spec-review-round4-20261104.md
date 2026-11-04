---
schema: vera.verdict.v1
verdict_id: p7-b1-unit3-spec-review-round4-20261104
status: NEEDS_CHANGES
reviewed_commit: 7bb163fbcbdc3bc09bc25fdacae8f24ba58c0621
findings:
  - finding_id: p7-b1-unit3-spec-review-round4-N5
    severity: MED
    summary: the §5.1 launched-binary swap control is not implementable from the black-box fixture — no seam exists between staging and launch, so it tacitly requires a new governed kernel test-seam the spec neither names nor scopes
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round4-N6
    severity: LOW
    summary: §9.3a asks the non-author to confirm the control fails "only at /proc, not step-2", but the spec mandates no distinct per-layer failure signal, so the layer cannot actually be told apart
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round4-N7
    severity: LOW
    summary: the positive companion ("unmodified run must reach+pass /proc") specifies no observable that the /proc witness actually executed, so it cannot independently prove the witness ran rather than PASS emerging from the archive hash
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round4-N4
    severity: MED
    summary: round-3 N4 attribution/labeling half RESOLVED (tampered-archive -> step-2 gate; swap -> /proc witness, both labeled; §9.3a requires the swap specifically); mechanism half rolls into N5
    verdict: CONFIRMED
artifact_path: docs/verification/verdicts/p7-b1-unit3-spec-review-round4-20261104.md
artifact_sha: 708ee5e1e8c6ed0dfc62781fb1ce808d899ca89a8851ab43db4441652bde12db
---

# P7 B1 unit-3 runtime-fixture spec (round 4) — non-author re-verify

## Provenance

- **Under review:** the round-4 revision of the DRAFT spec
  `docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md` (authored this session, commit
  `7bb163f`; spec-first per Build Law 6 — no code under review). This re-verifies the round-3 finding
  N4 (MED).
- **HEAD at dispatch and at receipt:** `7bb163f` — unchanged between dispatch and receipt
  (spec-review only; the tree was not modified during the review). Note: a decision record
  `VD-p7-nofetch-witness-2026-11-04.md` (§8.4 owner decision) was authored after dispatch and is
  orthogonal to the N4/N5 content under review.
- **Driving mechanism:** `/proofbound-review` (CLAUDE.md pre-merge adversarial review), round 4.
  Tier: connector/peripheral — a single adversarial review of a design document.
- **Reviewer identity:** `dispatched-non-author-reviewer-round4-20261104` — an independent reviewer
  dispatched via the Agent tool; it did NOT author the spec or any revision. Adjudicated by this
  session, which **authored the round-4 revision** — so the acceptance capacity comes from the
  non-author reviewer; this verdict returns NEEDS_CHANGES, so acceptance is not at issue. Law 9
  satisfied.
- **Adjudicator independent verification:** N5 is confirmed against the cited launch flow —
  `replayIsolated` runs `prepare(root)` then `store.Open(..., BinariesDir)` as consecutive
  statements in an un-exported function, `binariesDir` is an internal temp path the fixture never
  learns, and `ReplayIsolatedWithVerifiedPostgres` is a single black-box call, so no seam exists for
  a black-box fixture to swap the launched executable between staging verification and launch; the
  staging re-hash (`verified_archive.go:267-273`) binds the same `PostgresBinarySHA256` as the /proc
  comparison, so a byte-divergent binary present at stage time dies at staging, not at /proc.
  Implementing the swap control therefore requires adding a governed test-only seam to `kernel/*`,
  which the round-4 text does not name or scope. N6/N7 confirmed from the spec text. No false-PASS
  path is introduced.
- **Cleanroom redaction:** none required; the report below contains no absolute paths or usernames.

## Findings asymmetry notice

Per `VD-verification-asymmetry-2dyjnd`, a verified finding transfers zero trust to its suggested
remedy. The reviewer's fix sketches (name+scope a seam; per-layer error signal; a /proc-ran
observable) are proposals; the author decides the remedy and a non-author re-verifies.

## Verdict (verbatim reviewer report)

The following is the dispatched reviewer's report, reproduced verbatim.

---

ROUND-4 NON-AUTHOR RE-VERIFY — P7 B1 unit-3 runtime-fixture spec (HEAD 7bb163f). I did not author this spec or any revision.

## N4: PARTIALLY RESOLVED

The **attribution/labeling** half of N4 is resolved; the **mechanism** half is not, and it collides with the actual code in the way the caller suspected.

Resolved in text (§0, §1.1, §5 TWIN-INV-18 row, §5.1, §8 item 3, §9 item 3, §9.3a): the controls are now correctly layered and attributed — tampered-archive is explicitly labeled as calibrating only the step-2 archive-hash gate and "must NOT be relied on to show the launched-process witness has teeth"; the launched-binary swap is named as the TWIN-INV-18 control; a positive companion is added (swapped build passes every other check; unmodified run must reach+pass /proc); and §9.3a now requires the accepting non-author to run the swap control specifically and confirm it fails only at the /proc layer. That directly answers the round-3 attribution gap.

Not resolved: the swap control as specified is **not implementable against the current black-box API**, and the spec neither says how nor flags the governed kernel change it would require.

## The mechanism collision (the caller's core question)

Launch flow (replay.go:127-188 `replayIsolated`): `prepare(root)` → `stageVerifiedPostgres` extracts to `root/pgbin` and re-hashes bin/postgres once (verified_archive.go:267-273) → returns `binariesDir` → `store.Open(..., BinariesDir: binariesDir)` → embedded-postgres launches with `BinariesPath=binariesDir`, exec'ing `binariesDir/bin/postgres` directly (store.go:108). The file persists on disk from staging through launch, and the staging re-hash does NOT re-run before launch.

Two consequences:
1. The on-disk re-hash (267-273) does NOT re-fire before launch, so a disk swap performed strictly after `prepare` returns would reach /proc uncaught by staging — so there is no wrong-layer collision *with the re-hash itself*, PROVIDED the swap is timed in that window and no re-staging intervenes (the spec never pins this down).
2. But there is **no exposed seam** for a black-box fixture to act in that window. `prepare(root)` (replay.go:155) and `store.Open` (replay.go:160) are consecutive statements inside the un-exported `replayIsolated`; `binariesDir` is an internal temp path (`makeTempTwinRoot`/`root/pgbin`) the fixture never learns; `ReplayIsolatedWithVerifiedPostgres` is a single black-box call. The spec's two concrete examples both fail against this: "replace binariesDir/bin/postgres after staging verification but before pg_ctl start" (no hook between replay.go:158 and :160; path not exposed) and "redirect the launcher" (the launcher is store.Open→embeddedpostgres, not injectable from the fixture). You also cannot do it by crafting `VerifiedPostgresInput`: `PostgresBinarySHA256` binds the archive digest (110-113), the per-file walk, the 267-273 re-hash AND the /proc comparison to the same value — any byte-divergent binary on disk pre-stage dies at staging; one that matches staging also matches /proc.

So implementing this control requires ADDING a test-only injection seam to `kernel/*` (a post-stage/pre-launch callback, exposing binariesDir, or a swap hook). That is itself governed kernel code (§8 item 5, needs its own CI-*/spec/acceptance), and the spec does not acknowledge it. §5.1's "Only TWIN-INV-18's /proc check can now reject it" is stated as if trivially doable — a mild implementability over-claim given no such seam exists.

## Attribution verdict
Correct now: tampered-archive → step-2 gate; swap → /proc witness. Both labeled, and §9.3a requires exercising the swap specifically.

## New findings
- **N5 (MED) — swap-control mechanism under-specified / needs an unscoped governed kernel seam.** As exemplified it is not implementable from the fixture; the window exists on disk but no API seam reaches it. This is the exact witness N3 hinges PASS on, so it should be nailed in text (name the seam, scope it as governed kernel code) before CI leaves `proposed`, not discovered in implementation.
- **N6 (LOW) — no per-layer distinguishable failure signal.** §9.3a asks the non-author to confirm the swap fails "only at the /proc check, not step-2," but the spec mandates no distinct inspectable error identifier per layer. The staging mismatch already emits "twin: launched postgres executable digest mismatch" (verified_archive.go:271); the new /proc witness needs a clearly distinct, assertable signal, or the non-author cannot actually tell which layer rejected.
- **N7 (LOW) — positive companion leans circularly on the swap control.** "the unmodified run must reach and pass the /proc check" has no specified observable that the /proc witness actually executed (vs. PASS emerging from the archive hash). As worded it cannot independently prove the witness ran; it relies on the (under-specified) swap control to do that. A cheap fix: require the fixture to surface that the /proc witness was computed and matched.

## No-false-PASS still holds
No path to `case_status=PASS` or clean §9.3a acceptance exists without the real pinned postgres having been the launched process: every on-disk check and /proc bind to the same `PostgresBinarySHA256`. No "test that cannot fail" is introduced; §9.3a's "fails only at /proc" requirement is the right guard (modulo N6). Frozen `vera.replay.v1` shape, nine-key CASE-INV-3, Law 6 (no code, DEF-P7-001 unmoved, CI stays proposed) all still honored. §8.4 no-fetch owner decision correctly left open.

## Overall recommendation
NEEDS a small round 5 (or, at minimum, a hardened §8 implementation-gate clause before `CI-p7-runtime-fixture-1a2b3c` leaves `proposed`). The attribution fix is sound and the layering + positive companion are the right shape — this is close — but the swap control's mechanism is the one thing N4 was really about, and it is under-specified and collides with the black-box API: it tacitly requires a new governed kernel test-seam the spec neither names nor scopes. Fix N5 (name+scope the seam), N6 (per-layer failure signal), and N7 (observable that /proc ran) in text, then non-author re-verify. Not a false-PASS soundness hole.

---

## Adjudication summary

- **Status: NEEDS_CHANGES** (one open MED + two LOW). The N4 attribution half is RESOLVED; no
  false-PASS path exists and no "test that cannot fail" is introduced. The remaining work is
  spec-completeness about **how the TWIN-INV-18 witness is calibrated**, not soundness.
- **The cleanest round-5 remedy** (author's call, to be re-verified) is likely to calibrate the
  witness's discriminating power by a **direct unit test of the witness function** against a live
  process running a *different* executable (no production swap seam needed), reserve any full-path
  swap control as an explicitly-scoped governed kernel test-seam if end-to-end calibration is also
  wanted (N5), require a **distinct assertable failure identifier** for the /proc witness vs. the
  staging re-hash (N6), and require the fixture to **surface that the /proc witness was computed and
  matched** on the positive path (N7).
- **§8.4 is now settled by the owner** (2026-11-04, [[VD-p7-nofetch-witness-2026-11-04]]: network
  namespace primary + labeled declaration fallback); the round-5 revision folds that into the spec.
