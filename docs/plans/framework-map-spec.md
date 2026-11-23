# Spec: what each Audit room control maps to, in four frameworks

**Status:** draft 1, 2026-11-23, owner said "lets next one" (item 2 of the agent-first list);
reviewed with the code by one non-author (§6). **Implements:** `CI-framework-map-w7k2qd` →
`BR-framework-map-n4c8te` O-1 (requirement `proposed`). **Contract row:** CLI-INV-65.

## 0. Why

An auditor or compliance lead does not ask "did C10 hold?". They ask "show me your change-management
evidence for SOC 2 CC8.1", or "what supports ISO 27001 A.8.32 for code your agents wrote?". Today the
Audit room answers ten questions in its own words. After this change, each control also names the
framework criteria it gives evidence toward, both on the page and in the transcript. One generated
document lists the whole mapping with its caveats.

## 1. The frameworks and versions

| Framework | Version cited |
|---|---|
| AICPA Trust Services Criteria (SOC 2) | 2017 criteria, 2022 revised points of focus |
| ISO/IEC 27001 | 2022, Annex A |
| NIST SP 800-218, Secure Software Development Framework (SSDF) | v1.1 (February 2022) |
| EU AI Act, Regulation (EU) 2024/1689 | as published |

SSDF is added to the three planned: it is the framework US buyers ask about for software they
acquire (secure software attestation). A warranty is partial evidence for its PS.2: it lets an
acquirer verify the review evidence for a source revision, not the integrity of built artifacts.

## 2. The words (the honesty rules)

- Each entry says **"maps to"**: the control gives evidence toward the criterion. It never says
  "complies with", "satisfies" or "certifies". Every surface says that whether a criterion is met is
  the auditor's judgment, and that a criterion needs more than this evidence.
- **EU AI Act.** Its obligations fall on providers and deployers of *high-risk AI systems*.
  - A coding agent used to build software is not, by that use, a high-risk AI system. Art. 12
    (record-keeping) and Art. 14 (human oversight) are about the high-risk system's own logs and
    oversight, so Proofbound does not map to them, and the document says why.
  - The articles that apply are those for a provider of a high-risk AI system whose code agents help
    write:
    - **Art. 17(1)(a)**: the procedures for managing modifications;
    - **Art. 17(1)(b)–(d)**: design control; development and quality control; examination, test and
      validation;
    - **Art. 17(1)(k)**: record-keeping;
    - **Art. 11 with Annex IV points 2(g) and 6**: documented validation, and the relevant changes
      made through the lifecycle.
- An empty cell means no claim. It does not mean the framework is silent.

## 3. The mapping (the single home is the Go table; the document is generated from it)

| Control | SOC 2 | ISO 27001:2022 | SSDF v1.1 | EU AI Act (high-risk provider) |
|---|---|---|---|---|
| C1 claimed changes passed review | CC8.1 | A.8.32 | PW.7 | Art. 17(1)(a), 17(1)(d), Annex IV(2)(g) |
| C2 no code without a spec (signed waivers for emergencies) | CC8.1 | A.8.32 | — | Art. 17(1)(a), 17(1)(c) |
| C3 warranties verify | CC8.1 | A.8.32 | PS.2 (partial), PS.3 | Art. 11 and Annex IV(6), 17(1)(k) |
| C4 drift disclosed | CC8.1 | A.8.32 | — | Art. 11 and Annex IV(6) |
| C5 every commit accounted for | CC8.1 | A.8.32 | — | Art. 17(1)(k) |
| C6 changes to who may sign, waivers and the check | CC6.1, CC8.1 | A.5.18, A.8.9 | PO.3, PS.1 | — |
| C7 spec before code | CC8.1 | A.8.32, A.8.25 | — | Art. 17(1)(b) |
| C8 no self-review | CC5.1, CC8.1 | A.5.3 | — | Art. 17(1)(d) |
| C9 commits signed by a declared key | CC6.1 | A.5.16, A.8.24 | PS.1 | — |
| C10 agent changes inside the declared area | CC8.1 | A.8.32 | PS.1 | Art. 17(1)(a) |

Every reason is written as "evidence toward …", never as the criterion's outcome; a test refuses
outcome words ("proves", "ensures", "protects", "complete", …) in reasons.

Each cell is also listed with one line of reason in the generated document. 

## 4. Where it shows

- **Code.** `kernel/internal/cli/frameworks.go` holds `frameworkMap`: control ID → entries
  `{framework, criterion, reason}`, plus the four framework names and versions.
- **Audit room.** Each control card gets a "Maps to" line (for example "SOC 2 CC8.1 · ISO 27001
  A.8.32 …"), each entry with its reason as a tooltip. Under the controls, one muted line carries the
  §2 caveat and links to the document on GitHub.
- **Transcript.** `auditControl` gains `maps_to: [{framework, criterion}]`. The canonical transcript
  of a range therefore differs from one produced by a build without this field. Attestations sign
  the transcript digest at the time; nothing re-derives an old one, so nothing breaks.
- **Document.** `docs/compliance/FRAMEWORK-MAP.md` is **generated** (`@generated`, Law 1) by
  `cd kernel && go test ./internal/cli -run TestFrameworkMapDoc -update`. Without `-update`, the same
  test fails when the file differs, so `make check` catches drift.

## 5. Non-goals and tests

- **Non-goals:**
  - no new control;
  - no per-criterion pass/fail;
  - no other frameworks (HIPAA, PCI DSS, SLSA, NIST 800-53), recorded for later;
  - no change to `check`'s output.
- **Tests:**
  - `TestFrameworkMap`: every control C1..C10 has at least one entry; every entry's framework is one
    of the four; every criterion matches its framework's ID format; no entry for EU AI Act Art. 12 or
    Art. 14 (§2); the Audit room page and the transcript carry each control's entries.
  - `TestFrameworkMapDoc`: the golden document.

**Acceptance:** the Audit room on the agent demo's `shop` repository shows "Maps to" on every control.

## 6. Review

`framework-map-review-20261123`: NEEDS_WORK (F1..F3 MED, fixed; F4..F7 LOW, recorded per Law 11).

**F1 (MED), fixed.** Reasons claimed outcomes the controls do not produce: C9 "proves which person or
agent", C6 "protect" and "under change control", C5 "is complete", C10 "checked". Every reason now
states the evidence ("evidence toward …: …"). C9 says that who held the key is outside the evidence.
A test refuses outcome words in reasons, and the card label reads "Maps to (evidence toward, not
compliance)".

**F2 (MED), fixed.** Stretched mappings were dropped or qualified:
- SOC 2 CC3.4, an entity-level risk-assessment criterion: dropped.
- ISO A.8.29, which is *security* testing: dropped.
- SSDF PS.2: marked partial.
- SSDF PW.7: reworded (it does not require reviewer independence) and removed from C8.

**F3 (MED), fixed.** C2 cited Art. 17(1)(b) for development control, which is (c). Added Art. 17(1)(a)
(managing modifications) to C1, C2 and C10, 17(1)(k) (record-keeping) to C3 and C5, and Annex IV
point 2(g) to C1.

**LOW findings, recorded:**
- **F4:** better or omitted criteria, for example CC6.2 and CC6.3 for C6, ISO A.8.30, A.8.15 and
  A.5.33, SSDF PO.3.3 and PO.4.
- **F5:** the ID format test accepts IDs that do not exist; there is no allowlist of official titles.
- **F6:** the document link points to `main` and not the release; one caveat line instead of one
  per card; "ISO/IEC" in the label; no "as of" date for the EU AI Act; SSDF v1.2 status.
- **F7:** the generic `-update` flag name.
