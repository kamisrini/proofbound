# P6 Task 3 event-universe and reproducibility SPEC

## Scope

Task 3 re-proves the current 13 registered `(source, kind)` pairs across projection, verify, twin,
and the configured gate estate, then closes the C8 reproducibility/integrity rows. It adds no event
kind, projection, gate, provider, platform promise, or product capability. Frozen v1 wire identities
and pinned vectors are never edited.

## Dual-platform execution contract

The founder ratified retention of both Linux and native Windows for the existing execution path in
`docs/verification/verdicts/p6-dual-platform-ratification.md`, with semantic decision
`docs/decisions/VD-p6-dual-platform-2026-09-13.md`. This preserves an existing tracked platform
claim; it does not add product capability.

Linux requires Git, Bash, Go >=1.26, `jq`, GNU Make, `golangci-lint`, and the standard checked-script
tools. Native Windows requires Git for Windows or MSYS2 with Bash, Git, Go >=1.26, `jq`, GNU Make,
`golangci-lint`, and the standard checked-script tools. The PowerShell entry point may install or
inspect prerequisites, but GNU Make must execute the repository's Bash recipes. WSL is not native
Windows evidence.

For each OS, the acceptance record must include the tool versions, the tracked platform entry point,
bare `make check`, witnessed check/ledger prerequisites, complete two-pass `sync all`,
`proofbound verify`, rebuild equality, and self-hosted intent/requirement reports. A missing
prerequisite or failed command leaves that OS red and C8-002 open; no partial corpus check is a green
platform result.

## Closed route matrix

The source-kind universe is exactly:

1. `git / commit.recorded`
2. `checks / check.run`
3. `sessions / session.observed`
4. `reviews / review.verdict`
5. `reviews / requirement.reviewed`
6. `github / github.workflow_run`
7. `github / github.deployment`
8. `intent.records / business_decision.recorded`
9. `intent.records / requirement.recorded`
10. `intent.records / change_intent.recorded`
11. `intent.specdir / business_decision.recorded`
12. `intent.specdir / requirement.recorded`
13. `intent.specdir / change_intent.recorded`

Every pair is `consume` for projection, verify, and twin: the projector has a closed route for it;
verify applies/rebuilds the same projector over the complete ledger; and twin binds and projects the
same event bytes under the frozen `vera.replay.v1` proof. For the current configured gate estate,
Git commits, check runs, both review kinds, and all six intent-provider pairs are `consume`;
sessions and both GitHub kinds are `ignore`. `ignore` means no current gate definition or semantic
gate rule selects the event; it does not mean the generic gate evaluator could never be configured
for that pair. No registered pair is `reject`; mismatched or unregistered pairs reject at the core
or projector boundary and are outside the 13-pair registered universe.

| ID | Assertion | Proving test |
|---|---|---|
| T3-INV-1 | The matrix is exactly 13 × 4 with no duplicate or missing cell and only consume/ignore/reject values | `scripts/tests/p6-route-matrix.test.sh` |
| T3-INV-2 | Projection's supported route table contains every registered pair and rejects representative mismatches | `kernel/internal/projections/projection_test.go::TestSupportedEventMatrix` |
| T3-INV-3 | Verify applies, snapshots, rebuilds, and compares the complete projector | `kernel/internal/cli/cli_test.go::TestVerifyChecksEveryStageError` and `sync_integration_test.go::TestSyncChecksRebuildAndVerify` |
| T3-INV-4 | Twin's frozen proof binds every registered pair's payload bytes and preserves pair identity | `kernel/internal/twin/replay_test.go::TestReplayProofCoversEveryRegisteredEventPair` |
| T3-INV-5 | The current semantic/data-gate selectors consume exactly their named pair set and ignore sessions/GitHub | `kernel/internal/gates/gates_integration_test.go::TestCurrentGateEstateEventRoutes` |

## Reproducibility and integrity

- A fresh detached Linux clone at a named commit runs bare `make check`, performs two complete
  `sync all` passes with zero second-pass appends, runs `proofbound verify`, rebuilds without snapshot
  drift, and renders the self-hosted intent and requirement reports. The clone path and database are
  disposable and never become evidence themselves; command output and exact commit do.
- The tracked PowerShell entry point is run by real Windows PowerShell. A green run is Windows-path
  evidence. If Windows lacks a prerequisite or the Bash/Go suite is not portable, the result is red
  and requires a founder narrowing decision; WSL alone is not called native Windows proof.
- The artifact-integrity checker validates every schema-bearing verdict's declared path and
  normalized self-digest, validates its reviewed commit when present, and verifies the P5/P6
  adjudication exhibit digests. Documentary verdicts are enumerated rather than mistaken for wire
  artifacts.
- The ignored founder-requested vision assessment is copied verbatim into
  `docs/verification/p6-vision-progress-assessment.md`; its source remains ignored status, while the
  committed copy is the durable P6 evidence.
- The snapshot provider remains decision-skipped under the P6 semantic VD.

## Acceptance artifacts

- `docs/verification/p6-event-route-matrix.md`
- `docs/verification/p6-fresh-clone-linux.md`
- `docs/verification/p6-windows-platform.md`
- `docs/verification/p6-artifact-integrity.md`
- `docs/verification/p6-vision-progress-assessment.md`
- `docs/verification/p6-historical-evidence.jsonl` and the explicit
  `migrate historical-evidence` command record

Task 3 closes only when the 52-cell matrix checker, artifact-integrity checker, frozen-vector tests,
fresh-clone record, native PowerShell result (or accepted narrowing), `proofbound verify`, and bare
`make check` are green, and every C5/C8 `close-in-P6` row is evidence-bound.

## Non-goals

- No change to existing event schemas, routes, frozen proof bytes, or pinned vectors.
- No claim that current ignored gate pairs are permanently unsupported.
- No CI service, platform expansion, provider, snapshot feed, or external deployment boundary.
