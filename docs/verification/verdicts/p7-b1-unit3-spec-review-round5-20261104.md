---
schema: vera.verdict.v1
verdict_id: p7-b1-unit3-spec-review-round5-20261104
status: NEEDS_WORK
reviewed_commit: 5287ab41f75e3e7e7b030f1f13793a30b9850d98
findings:
  - finding_id: p7-b1-unit3-spec-review-round5-N8
    severity: MED
  - finding_id: p7-b1-unit3-spec-review-round5-N9
    severity: LOW
  - finding_id: p7-b1-unit3-spec-review-round5-N5
    severity: MED
  - finding_id: p7-b1-unit3-spec-review-round5-N6
    severity: LOW
  - finding_id: p7-b1-unit3-spec-review-round5-N7
    severity: LOW
artifact_path: docs/verification/verdicts/p7-b1-unit3-spec-review-round5-20261104.md
artifact_sha: dd08c27b3a8875b44d3c0bbc21ff8a05adcb424913ddcb06b15c695de53e8f00
---

# P7 B1 unit-3 runtime-fixture spec (round 5) — non-author re-verify

## Provenance

- **Under review:** the round-5 revision of the DRAFT spec
  `docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md` (authored this session, commit
  `5287ab4`; spec-first per Build Law 6 — no code under review). This re-verifies round-4 findings
  N5 (MED), N6 (LOW), N7 (LOW), and the §8.4 fold-in of [[VD-p7-nofetch-witness-2026-11-04]].
- **HEAD at dispatch and at receipt:** `5287ab41f75e3e7e7b030f1f13793a30b9850d98` — unchanged
  between dispatch and receipt (spec-review only; the tree was not modified during the review).
- **Driving mechanism:** `/proofbound-review` (CLAUDE.md pre-merge adversarial review), round 5.
  Tier: connector/peripheral — a single adversarial review of a design document.
- **Reviewer identity:** `dispatched-non-author-reviewer-round5-20261104` — an independent reviewer
  dispatched via the Agent tool; it did NOT author the spec or any revision. Adjudicated by this
  session, which **authored the round-5 revision** — so the acceptance capacity comes from the
  non-author reviewer; this verdict returns NEEDS_CHANGES, so acceptance is not at issue. Law 9
  satisfied.
- **Adjudicator independent verification:** N8 is confirmed by reading the round-5 §5.1 text, which
  asserts "A fixture that never reads /proc/<pid>/exe cannot populate resolved_sha256 and so cannot
  produce a valid PASS witness record." That is an over-claim: `resolved_sha256` is an authored hex
  field a fixture can set to `input.PostgresBinarySHA256` without calling `readlink` on
  `/proc/<pid>/exe`, and on the positive path a genuine read yields exactly that same value, so
  inspecting it cannot distinguish a real read from a copied digest. The spec does not require the
  fixture's production `proof_binding` emission to call the single calibrated `witnessLaunchedProcess`
  function, so that function can be an orphan the fixture bypasses — a declaration-forgery path of the
  same class as round-2 N1. N9 confirmed from the §0 text. For the shipped black-box configuration
  there remains no *accidental* divergent-binary false-PASS (all on-disk checks and the /proc check
  bind to the same `PostgresBinarySHA256`, and the staged file is exec'd unchanged); N8 is the only
  residual PASS-not-entailing-a-real-/proc-read path and it is a deliberate forgery, caught by the
  Law-9 kernel review of the governed fixture source at implementation — but the spec must not assert
  a protection it does not provide.
- **Cleanroom redaction:** none required; the report below contains no absolute paths or usernames.

## Findings asymmetry notice

Per `VD-verification-asymmetry-2dyjnd`, a verified finding transfers zero trust to its suggested
remedy. The reviewer's fix sketch (single call site + §9.3a code inspection + withdraw the
over-claim) is a proposal; the author decides the remedy and a non-author re-verifies.

## Verdict (verbatim reviewer report)

The following is the dispatched reviewer's report, reproduced verbatim.

---

ROUND-5 NON-AUTHOR RE-VERIFY — P7 B1 unit-3 runtime-fixture spec (HEAD 5287ab4). I did not author this spec or any revision. Spec-first per Law 6; no code under review.

## Per-finding verdicts

**N5 — RESOLVED.** The implementability/scoping defect is fixed. `witnessLaunchedProcess(pid, expectedSHA256) (resolvedSHA256, error)` is a standalone testable function; a direct unit test can spawn a live process running a byte-divergent executable and a process running the pinned binary and assert (distinct error / no-error + match) — genuinely implementable and it proves the FUNCTION discriminates, with no production swap seam. The optional end-to-end swap seam is now correctly named as a test-only post-stage/pre-launch hook scoped as governed `kernel/*` code carrying its own CI-*/SPEC/acceptance (§5.1 item, §8 item 5), not an implicit requirement. This answers round-4 N5 as stated.

**N6 — RESOLVED.** §5.1/TWIN-INV-18/§8 item 3 require a distinct typed error (e.g. `ErrLaunchedProcessDigestMismatch`), assertable (errors.Is) and explicitly separate from the staging re-hash message at verified_archive.go:271 (`twin: launched postgres executable digest mismatch`, a plain unwrapped fmt.Errorf). Concrete and clearly separable. Confirmed against code: the two error sites are different functions, so a per-layer signal is achievable.

**N7 — PARTIALLY RESOLVED.** Recording `{method:"proc-exe", resolved_pid, resolved_sha256}` and requiring §9.3a inspection is a real improvement over round-4 (round-4 surfaced nothing) and it blocks the most naive case (a record with no resolved_sha256 at all). BUT the closure is incomplete and the spec over-claims it: see new finding N8. The resolved_sha256 VALUE is non-discriminating on the positive path — a genuine /proc read of the pinned postgres yields exactly `PostgresBinarySHA256`, which is also exactly what a forged copy yields, so inspecting the value cannot by itself distinguish "read /proc" from "copied the expected digest."

**§8.4 — RESOLVED (faithful to the VD).** §8 item 4 and the TWIN-INV-19 row match VD-p7-nofetch-witness-2026-11-04 exactly: no-egress network namespace (`unshare -rn`/equivalent) as primary witness; labeled configuration-declaration fallback (fetch strategy not constructed + local BinariesPath) recorded honestly as a declaration; each run records which it used; namespace-only (B) and declaration-only (C) rejected. No drift from the recorded decision.

## New findings

**N8 (MED) — the spec does not bind the fixture's run-time witness to the calibrated function, and §5.1 contains a factual over-claim.** §5.1 asserts: "A fixture that never reads /proc/<pid>/exe cannot populate resolved_sha256 and so cannot produce a valid PASS witness record." This is false. resolved_sha256 is just a hex field in an authored record; a fixture can set `resolved_sha256 = input.PostgresBinarySHA256` and `resolved_pid = <real postmaster pid>` without ever calling readlink on /proc/<pid>/exe. The record is well-formed, proof_binding=PASS is emitted (resolved_sha256 == PostgresBinarySHA256 holds by construction), and the §9.3a non-author inspecting resolved_sha256 sees it equal the pinned digest — which is what ANY genuine pass must also show, so inspection does not catch it. The direct-unit-test calibration (N5) does not save this: it proves a function `witnessLaunchedProcess` discriminates, but nothing in the spec requires the fixture's production proof_binding path to CALL that same function (single implementation / single call site), so the calibrated function can be an orphan the fixture bypasses. By cleanup time (TWIN-INV-21 removes the temp root and the server is stopped) resolved_pid is no longer independently checkable, so §9.3a cannot re-derive the digest from the live process.
Concrete failure scenario: fixture copies the pinned digest into the witness record, authors all outcomes PASS → case_status=PASS; the witness-function unit test still passes (it tests the orphaned helper); §9.3a sees resolved_sha256 == PostgresBinarySHA256 and accepts. This is a PASS reachable by declaration — the same forgeable-by-declaration class as round-2 N1 and the Law-9 corollary ("a test that cannot fail"), which this spec's fixture-is-the-gate architecture exists to prevent.
Cheap textual fix (author's call, to be re-verified): (a) require a SINGLE `witnessLaunchedProcess` implementation that is the only producer of resolved_sha256 and is invoked on the fixture's proof_binding path (the unit test and the fixture share that one call site); (b) require §9.3a to confirm by code inspection that the production emission reads /proc through that function (not a literal); (c) withdraw/correct the §5.1 "cannot populate resolved_sha256" over-claim. Note this residual is a deliberate-forgery path, not an accidental divergent-binary false-PASS — see below — so a non-author who reads the governed fixture source (which Law 9 kernel review entails) would catch it; but the spec should not assert a protection it does not provide.

**N9 (LOW, doc-consistency nit).** §0's historical "Round 4" paragraph still says "§9.3a now requires the accepting non-author to run the launched-binary-swap control specifically." The current §9.3a (and §9 item 3a) no longer requires the swap control — it requires the witness-function unit test + resolved_sha256 inspection, and the swap is now explicitly optional. A reader of §0 hits a stale requirement that contradicts the live §9.3a. Harmonize the §0 narrative with round-5's §9.3a.

## No-false-PASS assessment (the caller's core question)
For the SHIPPED black-box configuration (no optional test seam), there is no path to case_status=PASS with a *divergent on-disk binary*: ArchiveSHA256, the per-file inventory walk, the staging re-hash (verified_archive.go:267-273), and PostgresBinarySHA256 all bind to the same digest, the staged file persists unchanged through launch, and embedded-postgres execs that staged file — so any byte-divergent binary dies at staging before /proc. The ONLY residual path to a PASS not entailing a real /proc observation of the pinned process is N8 (fixture copies the digest rather than reading /proc). That is a declaration forgery, not an accidental hole, but the spec claims it is impossible, which is why it must be corrected. The explicitly-specified calibration controls (unit-test negative=distinct error / positive=match; TWIN-INV-20 injected-mutation must FAIL; TWIN-INV-21 suppressed-removal must FAIL) each state a control that can fail — no "test that cannot fail" in the control set itself; the gap is the fixture-to-witness binding (N8).

## Overall recommendation
NOT YET ACCEPTABLE — needs a small round 6. N5, N6 and §8.4 are RESOLVED and the design is very close, but N7 is only partially closed and §5.1 carries a false over-claim (N8, MED) that, uncorrected, would let an implementer treat resolved_sha256 as self-proving when it is value-non-discriminating on the positive path. The fix is purely textual (bind the fixture's proof_binding emission to the single calibrated witnessLaunchedProcess call site + §9.3a code-inspection of that path + withdraw the over-claim), plus the N9 §0 consistency nit. Re-verify by a non-author after those edits.

---

## Adjudication summary

- **Status: NEEDS_CHANGES** (one MED + one LOW). N5, N6, §8.4 RESOLVED; the witness-function
  calibration and the owner-settled no-fetch mechanism are sound and faithful. The one substantive
  item, N8, is an **over-claim correction + a binding requirement**: the spec must (a) require a
  single `witnessLaunchedProcess` implementation that is the sole producer of `resolved_sha256` and
  is invoked on the fixture's `proof_binding` path (unit test and fixture share that one call site);
  (b) require §9.3a to confirm by **code inspection** that the production emission reads `/proc`
  through that function, not a literal; and (c) withdraw the §5.1 "cannot populate resolved_sha256"
  over-claim. N9 harmonizes the §0 Round-4 narrative with the live §9.3a.
- **No accidental false-PASS path exists in the shipped configuration;** N8 is a deliberate-forgery
  path that the Law-9 kernel review of the governed fixture will catch at implementation — the
  round-6 edit ensures the spec does not *claim* a protection it lacks and that the implementer binds
  the witness correctly.
- The fix is textual; round 6 is expected to converge.


## Findings — original pre-schema block (R1-4 provenance)

> The structured `findings:` in the frontmatter were normalized to `vera.verdict.v1`
> (`finding_id` + `severity` only). The **original** block, including each finding's full
> `summary`/`verdict` text and its original severity word, is preserved verbatim below so no
> reviewer detail is lost. `info`→`LOW` and `low-medium`→`MED` were the only severity remaps.
> See `docs/deferrals.tsv` → `DEF-DELIVERY-ENFORCE-VERDICTS`.

```yaml
findings:
  - finding_id: p7-b1-unit3-spec-review-round5-N8
    severity: MED
    summary: §5.1 over-claims that a fixture that never reads /proc cannot populate resolved_sha256; the spec does not bind the fixture's proof_binding emission to the single calibrated witnessLaunchedProcess call site, so PASS stays forgeable by copying the pinned digest
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round5-N9
    severity: LOW
    summary: §0's historical Round-4 paragraph still states §9.3a requires the launched-binary-swap control, contradicting the live round-5 §9.3a (witness-function unit test + resolved_sha256 inspection; swap now optional)
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round5-N5
    severity: MED
    summary: round-4 N5 RESOLVED — witness calibrated by a direct unit test of a standalone witnessLaunchedProcess; optional end-to-end seam scoped as governed kernel code
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round5-N6
    severity: LOW
    summary: round-4 N6 RESOLVED — distinct typed /proc mismatch error, assertable and separate from verified_archive.go:271
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round5-N7
    severity: LOW
    summary: round-4 N7 PARTIALLY — the resolved_sha256 observable blocks the naive empty case but is value-non-discriminating on the positive path; closure completed by fixing N8
    verdict: CONFIRMED
```