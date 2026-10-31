---
schema: vera.verdict.v1
verdict_id: p7-twin-replay-cli-round1-20261031
status: ACCEPTABLE-WITH-NOTES
reviewed_commit: a83438365f5a81730987a95d3d853dbd2320f12e
findings:
  - id: F1
    severity: low-medium
    status: fixed
  - id: F2
    severity: info
    status: hardened
artifact_path: docs/verification/verdicts/p7-twin-replay-cli-round1-20261031.md
artifact_sha: c1b5635ba5c00f7195df81f74780cc767b201ce96cc853e33d14a61760c2b94d
---

# `proofbound twin replay` CLI — non-author review (round 1)

**ACCEPTABLE-WITH-NOTES.** Independent non-author reviewer (Law 9; dispatched reviewer that did not
author the code) reviewed commit `a83438365f5a81730987a95d3d853dbd2320f12e`
(`feat(cli): add proofbound twin replay`), connector/peripheral tier — one adversarial review per
[`VD-review-ceremony-tiers-v2b6kp`](../../decisions/VD-review-ceremony-tiers-v2b6kp.md).

## Reviewer basis

Read the full commit diff, `kernel/internal/twin/replay.go` (`ReplayIsolated`), and the projection
`Apply`/`Snapshot` paths. Ran `go vet ./internal/cli/` (clean) and
`go test ./internal/cli/ -run 'TwinReplay|ParseTwinReplay' -count=1` → PASS (embedded postgres).

## Confirmed

- **CLI-INV-11 (no source-ledger mutation): CONFIRMED.** The command only reads the source store
  (project/snapshot/ReadEvents/ReplayIsolated); no append/write path reaches the events table, and
  `ReplayIsolated` does all candidate work in a throwaway store it creates and removes. The proving
  tests snapshot the event seqs before/after and assert equality, so a mutation **would** be caught —
  the tests discriminate, not vacuous.
- **CLI-INV-12 (counterfactual verdict + proof): SOUND** on the default (through=latest) path; the
  tests assert `changed` + `candidate_events=1` + distinct baseline/candidate snapshot digests, so a
  hard-coded verdict or a non-reflected candidate would fail.
- Build laws: Law 6 (SPEC updated same commit), Law 8 (only `strconv` + `internal/twin`, no new
  dependency), cli SPEC non-goals (no reinterpretation of `vera.replay.v1` bytes). All OK. No
  escalation warranted.

## Findings and resolution

- **F1 (low-medium) — FIXED in `b9f7b7f`.** The baseline was always the full committed projection
  (`projector.Snapshot`) regardless of `--through`, so `twin replay --through <seq>` below the latest
  seq compared a prefix projection against the full projection and reported a misleading
  `verdict=changed` with no candidate (conflating dropped events with the candidate's effect).
  Reproduction: seed 2 events, run `--through 1` with no candidate → `changed`, `candidate_events=0`.
  The author fix derives the baseline from the same `--through` prefix (an isolated
  `ReplayIsolated(through, [])` run); regression test `TestTwinReplayThroughBelowLatestIsCoherent`
  pins `--through` below latest as `preserved`. (Remedy adopted after independent re-derivation of the
  defect, per [`VD-verification-asymmetry-2dyjnd`](../../decisions/VD-verification-asymmetry-2dyjnd.md) —
  findings and remedies carry separate trust.)
- **F2 (info) — HARDENED in `b9f7b7f`.** `loadTwinCandidate` did not reject trailing data after the
  JSON object; it now errors on `decoder.More()`. Duplicate flags taking the last value are harmless
  and left as-is.

The reviewed change does not mutate the source ledger, both stated invariants hold with
discriminating tests, and the one correctness finding is fixed with a locking test.
