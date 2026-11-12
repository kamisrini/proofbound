---
schema: vera.verdict.v1
verdict_id: p7-b1-unit3-spec-review-round2-20261104
status: NEEDS_WORK
reviewed_commit: b64a46eaff77c44038963fa451bb66cd89282c66
findings:
  - finding_id: p7-b1-unit3-spec-review-round2-N1
    severity: HIGH
  - finding_id: p7-b1-unit3-spec-review-round2-N2
    severity: MED
  - finding_id: p7-b1-unit3-spec-review-round2-N3
    severity: LOW
  - finding_id: p7-b1-unit3-spec-review-round2-F1
    severity: HIGH
  - finding_id: p7-b1-unit3-spec-review-round2-F7
    severity: MED
artifact_path: docs/verification/verdicts/p7-b1-unit3-spec-review-round2-20261104.md
artifact_sha: 947f02f7d0ea370abcbc8dfab84278a3f46ad5688c6b493dcf9bbd7f0ad16d9f
---

# P7 B1 unit-3 runtime-fixture spec (round 2) — non-author re-verify

## Provenance

- **Under review:** the round-2 revision of the DRAFT spec
  `docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md` (authored this session, commit
  `b64a46e`; spec-first per Build Law 6 — no code under review). This re-verifies the fixes claimed
  against the round-1 verdict `p7-b1-unit3-spec-review-round1-20261104.md` (findings F1–F11).
- **HEAD at dispatch and at receipt:** `b64a46eaff77c44038963fa451bb66cd89282c66` — unchanged
  between dispatch and receipt (spec-review only; the tree was not modified during the review).
- **Driving mechanism:** `/proofbound-review` (CLAUDE.md pre-merge adversarial review), round 2.
  Tier: connector/peripheral — a single adversarial review of a design document (no mutation
  evidence; no code yet).
- **Reviewer identity:** `dispatched-non-author-reviewer-round2-20261104` — an independent reviewer
  dispatched via the Agent tool; it did NOT author the spec or its round-2 revision. Adjudicated by
  this session, which **authored the round-2 revision** — so this session is NOT a valid acceptor of
  it; this verdict's acceptance capacity comes from the non-author reviewer. The verdict returns
  NEEDS_CHANGES, so acceptance is not at issue regardless. Law 9 satisfied.
- **Adjudicator independent verification** (the adjudicator discards findings lacking a reproducible
  proof; the two confirmed structural findings were checked against the code before acceptance):
  - **N1 CONFIRMED** — `kernel/internal/twin/manifest/casemanifest.go:197-222`: `DeriveCaseStatus`
    is a pure worst-wins fold over author-declared `outcome.Status` values; it re-hashes nothing and
    compares no witness `sha256`. `validateOutcome` (`casemanifest.go:398-425`) requires a PASS
    outcome only to cite a non-empty evidence list naming an existing artifact role; it never checks
    a witness digest against real bytes. The round-2 spec's claims (§1 "PASS is reachable only when
    the bound runtime witnesses … verify"; §5 "DeriveCaseStatus verifies each witness digest before
    folding"; §6 "re-verified by DeriveCaseStatus") are therefore false against the frozen,
    accepted contract. `case_status=PASS` remains a self-declared state — forgeable by committing a
    witness record whose `sha256` is copied from `PostgresBinarySHA256`, with no postgres launched.
  - **N2 CONFIRMED** — `scripts/p6-artifact-integrity.sh:41` walks only
    `docs/verification/verdicts` at `-maxdepth 1`. Build-evidence / verify-evidence under
    `docs/verification/p7-replay-lock-prerequisites/` are not validated by it, so the §6.1 precedent
    ("treated exactly as … verdicts, build-evidence, verify-evidence [validated by
    p6-artifact-integrity]") is factually wrong: the §6 artifacts would get no automated integrity
    check. The documentary/no-`@generated` conclusion may still be defensible, but the stated
    justification is incorrect and must be corrected or paired with a real validation gate.
  - N3 (LOW) accepted as a reasonable spec-clarity gap at the reviewer's severity.
  - F1 and F7 re-graded PARTIALLY (both roll up into N1 and N2 respectively).
- **Cleanroom redaction:** none required; the report below contains no absolute paths or usernames.

## Findings asymmetry notice

Per `VD-verification-asymmetry-2dyjnd`, a verified finding transfers zero trust to any remedy it
suggests. The reviewer's "the fix is a claim-relocation" framing is a proposal, not an accepted fix.
The spec author decides the remedy; a non-author must re-verify the revised spec (round 3). Findings
are stated as broken invariants / over-claims, not as instructions.

## Verdict (verbatim reviewer report)

The following is the dispatched reviewer's report, reproduced verbatim.

---

# Round-2 non-author adversarial re-verify — P7 B1 unit-3 runtime-fixture spec

**HEAD b64a46e. I did not author this spec.** I re-checked each round-1 finding against the revised spec, the accepted twin/manifest SPECs, and the actual code (`replay.go`, `verified_archive.go`, `casemanifest.go`, `scripts/p6-artifact-integrity.sh`).

## Per round-1 finding

| ID | Verdict | Why (one line) |
|---|---|---|
| F1 (HIGH) | **PARTIALLY** | LOCKED is now correctly scoped to determinism-only (honest, correct), but closure was moved onto `case_status=PASS`, which is ANOTHER pure fold that does not witness the run — see N1. |
| F2 (HIGH) | **RESOLVED** | TWIN-INV-18 is now `/proc/<pid>/exe` of the postmaster PID with `(deleted)` rejection + a REQUIRED negative control; on-disk re-hash and `version()` explicitly rejected as non-witnessing; §5.2 amendment stays honest. PID resolution via `postmaster.pid` is correct (postmaster is the long-running process; `pg_ctl start` blocks until up, so no start/read race). |
| F3 (MED) | **RESOLVED** | Ambient dynamic deps (system `libz`, `libc`) are an explicit residual; `ldd` set is informational, not a witnessed claim; §7 adds the matching non-goal. |
| F4 (MED) | **RESOLVED** | Primary witness is a network namespace (`unshare -rn`); the fetch-strategy assertion is labeled a declaration, not a witness; the namespace-vs-declaration choice is correctly left OPEN as the one §8.4 owner decision. |
| F5 (MED) | **RESOLVED** | §8.1 downgraded to "RESOLVED FOR THE FULL TREE; SUBSET UNPROVEN"; proving the `bin/`+`share/` subset runs is an explicit implementation step (§9.2). |
| F6 (MED) | **RESOLVED** | Settled as drop-all, no dereference, assert symlink-free before hashing; prep-time dereference attack surface removed and the reasoning is stated. |
| F7 (MED) | **PARTIALLY** | The @generated/documentary dichotomy is now explicitly addressed (good), but the justification rests on a precedent that does not exist — see N2. |
| F8 (MED) | **RESOLVED** | Toolchain identity is now a citation (evidence path + revision), explicitly "not a copy of its values". |
| F9 (MED) | **RESOLVED** | TWIN-INV-20 now carries a required positive control (injected mutation MUST make the witness FAIL); TWIN-INV-21 witness "stats the path and asserts absence" with a required positive control (suppressed removal MUST FAIL). |
| F10 (MED) | **RESOLVED** | Ephemeral port + short socket + serial isolation specified against DEF-KERNELCHECK-FLAKE-001; §3a makes a live (non-skip) run an acceptance precondition. |
| F11 (LOW) | **RESOLVED (at control level)** | §5.1 specifies absent→UNVERIFIED / swapped→FAIL / real→PASS and warns a permissive `RequirementResolver` stub may not flip the real-run witnesses. Teeth depend on the fixture, not the manifest — see N1. |

## NEW findings

### N1 (HIGH) — The round-2 central mechanism contradicts the accepted case-manifest contract; `case_status=PASS` is still forgeable by declaration.
This is the answer to the caller's most important question, and it is the unresolved core of F1.

The spec repeatedly attributes witness verification to the committed-artifact pipeline:
- §1: "PASS is reachable **only** when the bound runtime witnesses ... are present and verify."
- §1.1 soundness chain: "`case_status = PASS` ⟹ ... the launched-process witness ... verified."
- §5 header: "**`DeriveCaseStatus` verifies each witness digest before folding the outcome to PASS**."
- §6: witness records "each digest bound into the case manifest and **re-verified by `DeriveCaseStatus`**."

All four are false against the accepted, frozen contract and the code:
- `DeriveCaseStatus` (`kernel/internal/twin/manifest/casemanifest.go:197-222`) is a **pure worst-wins fold over author-declared `outcome.Status` values**. It re-hashes nothing, launches nothing, and never compares a witness `sha256` to `PostgresBinarySHA256`.
- `ValidateCaseManifest` / `validateOutcome` (`casemanifest.go:398-425`) require only that a PASS outcome cite a **non-empty evidence list naming an existing artifact role**; `validateArtifacts`/`validateWitnesses` require only a well-formed path and a syntactically valid 64-hex digest — never that the digest matches real file bytes. This is exactly CASE-INV-4/5 as accepted ("`PASS` require[s] evidence naming an existing artifact role").

Consequence: `case_status=PASS` is precisely a **self-declared state**, the very thing §5 says it is not. The runtime-fact claim was moved from one pure fold (LOCKED = digest arithmetic, correctly down-scoped) onto a second pure fold (case_status = declaration arithmetic). The §1.1 soundness chain is **unsound for the committed artifact in isolation**.

**Forgery scenario (the caller's circularity hypothesis, confirmed):** an author writes a `launched-process` witness record `{kind, path, sha256}` with `sha256` set to the committed `PostgresBinarySHA256` (a copied value), adds an artifact with that role, declares `proof_binding=PASS` (plus the other outcomes), and commits a well-formed proof/lock. `ValidateCaseManifest` passes, `DeriveCaseStatus` folds to PASS, `EvaluateLock`→LOCKED — **with no PostgreSQL ever launched.** The author produces both the witness record and the manifest that cites it, and nothing in the committed pipeline breaks the circle.

The only non-circular guarantee actually lives in (a) the **fixture proving-test** that must compare `/proc/<pid>/exe` to `PostgresBinarySHA256` at run time and refuse to emit `proof_binding=PASS` otherwise, (b) the §3a **live non-author run**, and (c) the §5.1 **negative control**. The spec's §4-step-6 ("asserts ... `DeriveCaseStatus == PASS` with all runtime witnesses verifying") and §9.3 ("with every witness verifying") read correctly only if understood as *fixture* behavior — but §1/§5/§6 mis-locate that verification in `DeriveCaseStatus`. If implemented literally this would require modifying the accepted `DeriveCaseStatus` contract (scope creep needing its own CI + acceptance); if not, it is silently false and leaves PASS forgeable. The spec must stop claiming the case manifest/`DeriveCaseStatus` re-verifies witnesses, and explicitly rest soundness on the fixture + §3a live run + §5.1 negative control (or specify a new gate that genuinely cross-checks each witness `sha256` against its source, e.g. launched-process `sha256` == `PostgresBinarySHA256`).

### N2 (MED) — §6.1's Law-1 exemption cites a precedent that does not cover these artifacts.
§6.1 justifies "no `@generated`, validated structurally (self-digest + path resolution via `scripts/p6-artifact-integrity.sh`)" by claiming the new artifacts are "treated exactly as ... verdicts, build-evidence, verify-evidence." But `scripts/p6-artifact-integrity.sh` (read in full) walks **only** `docs/verification/verdicts/*.md` (maxdepth 1), self-digest-validates only schema-bearing front-matter files there (plain documentary ones are merely counted), plus two hard-coded P5/P6 exhibits. Build-evidence and verify-evidence live in `docs/verification/p7-replay-lock-prerequisites/` (confirmed: plain `#`-heading markdown, no front-matter, no `@generated`) and are **not validated by that script at all**. The §6 artifacts this unit commits (prepared-archive manifest JSON, proof bytes, behavior-lock + case-manifest JSON, witness records, command logs) would likewise fall outside the verdicts dir and get **no automated integrity check**. So the exemption is granted on a precedent the repo does not have: a load-bearing case-manifest JSON could be hand-edited post-hoc and `make check` would not catch it. The only §6 artifact actually covered is the round-2 verdict itself (it lands in `verdicts/`). F7's conclusion (documentary, no `@generated`) may be defensible, but its stated justification is factually wrong and should either be corrected or paired with a real validation gate for these artifacts.

### N3 (LOW) — TWIN-INV-18 and TWIN-INV-22 both fold into the single `proof_binding` outcome key.
The nine case-manifest outcome keys are fixed (CASE-INV-3). The §5 table maps BOTH the launched-process provenance witness (TWIN-INV-18) AND the determinism check (TWIN-INV-22) to `proof_binding`. There is no dedicated `launched_process` slot, so the strongest new witness shares a key with the weak determinism re-eval. The spec does not explicitly require the fixture to gate `proof_binding=PASS` on BOTH; as written, a case declaring `proof_binding=PASS` on determinism grounds alone would still fold to PASS. Worth an explicit "proof_binding=PASS requires the launched-process witness to verify, not merely determinism" clause.

## Consistency checks that still pass
- Frozen `vera.replay.v1` proof shape (§7) is preserved; witnesses live in the case-manifest, not the proof — no contradiction with TWIN-INV-1..17 or the authored/observed boundary.
- TWIN-INV-18..22 append correctly (TWIN-INV-5 is a pre-existing tombstone).
- Law 6 (spec-first, no code, no DEF-P7-001 movement, change-intent kept `proposed`) honored; Law 2 toolchain citation (F8) correct; Law 9 negative/positive controls now present at the spec level (F2/F9/F11).

## Overall recommendation
**Needs a targeted round 3 before implementation.** The revision resolves F2–F6, F8–F11 honestly and well, and the LOCKED→determinism down-scoping is the right move. But F1 is only PARTIALLY resolved: N1 is a HIGH structural over-claim — the spec asserts `DeriveCaseStatus`/the case manifest re-verifies witness digests when the accepted contract and code show it is a pure fold over author-declared statuses, leaving `case_status=PASS` forgeable and the §1.1 soundness chain unsound for the committed artifact. The fix is largely a claim-relocation (rest soundness on the fixture proving-test + §3a live run + §5.1 negative control, and/or specify a gate that actually cross-checks each witness digest), not a redesign — but because it is a HIGH over-claim about a frozen contract, plus the N2 precedent defect, it must be corrected and re-verified by a non-author before the change intent leaves `proposed`. The §8.4 no-fetch owner decision remains correctly open and is not a blocker to that correction.

---

## Adjudication summary

- **Status: NEEDS_CHANGES.** Nine of eleven round-1 findings (F2–F6, F8–F11) are honestly RESOLVED;
  the LOCKED→determinism down-scoping is correct. But N1 (HIGH, confirmed) shows the round-2 fix
  relocated the over-claim rather than removing it: it rested closure on `case_status=PASS`, which is
  a second pure fold that re-verifies nothing, so PASS is still forgeable by declaration and the
  §1.1 soundness chain is unsound for the committed artifact in isolation. N2 (MED, confirmed)
  invalidates the §6.1 Law-1 precedent.
- **This is not an acceptance.** A round-3 revision must (i) relocate the witness-verification claim
  out of `DeriveCaseStatus`/the committed manifest and onto the **fixture proving-test + §3a live
  non-author run + §5.1 negative control** (and/or specify a genuine new cross-check gate as its own
  scoped change) so the soundness chain holds against the real artifact; (ii) correct the §6.1 Law-1
  justification (and either accept honestly that these artifacts get no automated integrity check,
  as build/verify-evidence already don't, or add a real validation gate); and (iii) gate
  `proof_binding=PASS` explicitly on the launched-process witness, not determinism alone (N3). A
  non-author must re-verify round 3 before `CI-p7-runtime-fixture-1a2b3c` leaves `proposed`.
- **§8.4 (no-fetch mechanism) remains the one open owner decision** and is not a blocker to the
  round-3 correction.


## Findings — original pre-schema block (R1-4 provenance)

> The structured `findings:` in the frontmatter were normalized to `vera.verdict.v1`
> (`finding_id` + `severity` only). The **original** block, including each finding's full
> `summary`/`verdict` text and its original severity word, is preserved verbatim below so no
> reviewer detail is lost. `info`→`LOW` and `low-medium`→`MED` were the only severity remaps.
> See `docs/deferrals.tsv` → `DEF-DELIVERY-ENFORCE-VERDICTS`.

```yaml
findings:
  - finding_id: p7-b1-unit3-spec-review-round2-N1
    severity: HIGH
    summary: case_status=PASS is still forgeable by declaration; DeriveCaseStatus is a pure fold and re-verifies no witness digest, so §1/§5/§6 over-claim
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round2-N2
    severity: MED
    summary: §6.1 Law-1 exemption cites p6-artifact-integrity as covering build/verify-evidence, but that script only walks docs/verification/verdicts; the §6 artifacts get no automated integrity check
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round2-N3
    severity: LOW
    summary: TWIN-INV-18 (launched-process) and TWIN-INV-22 (determinism) both fold into the single proof_binding key; proof_binding=PASS not explicitly gated on the launched-process witness
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round2-F1
    severity: HIGH
    summary: F1 only PARTIALLY resolved — LOCKED correctly down-scoped, but closure moved onto case_status=PASS which is another pure fold (see N1)
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round2-F7
    severity: MED
    summary: F7 only PARTIALLY resolved — dichotomy addressed, but justification rests on a precedent that does not exist (see N2)
    verdict: CONFIRMED
```