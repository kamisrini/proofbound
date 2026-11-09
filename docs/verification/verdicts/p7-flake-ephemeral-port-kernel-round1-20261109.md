---
schema: vera.verdict.v1
verdict_id: p7-flake-ephemeral-port-kernel-round1-20261109
status: NEEDS_WORK
reviewed_commit: 057bc077b24f3b831b7074be83e5dccd7a3a7c5a
findings:
  - finding_id: F1
    severity: HIGH
  - finding_id: F2
    severity: LOW
  - finding_id: F3
    severity: LOW
artifact_path: docs/verification/verdicts/p7-flake-ephemeral-port-kernel-round1-20261109.md
artifact_sha: 476c0b1409d4e0cb85d7628d0442f1f2a1371f1469b6db0f8addfd25d75afca1
---

# DEF-KERNELCHECK-FLAKE-001 ephemeral-port kernel fix — non-author verdict (round 1)

**NEEDS_WORK.** An independent non-author reviewer verified commit `057bc07` (the ephemeral-port
test fix) against change intent `CI-flake-ephemeral-port-7d5f10` and the Build Laws. The test change
itself is correct, minimal, test-only, and in-scope — but the commit shipped a stale `@generated`
artifact that left `make check` RED at HEAD, violating the intent's acceptance boundary
("`make check` green AND deterministic"). Findings transfer zero trust to the fix; the blocking one
(F1) has since been fixed by the author (`8700760`) and a round-2 non-author re-verification is
dispatched.

## Basis (what the reviewer ran)

- `git show 057bc07`: TEST-ONLY — 8 `_test.go` files (4 new `freeport_test.go` helpers + 4 one-line
  `Port: freeTestPort(t)` call-site edits). Product `embeddedPort(0)==55432` (`store.go:177`) and the
  `surface_test` port contract untouched. Confirmed.
- `go vet ./...` and `go vet -tags integration ./internal/projections/ ./internal/store/
  ./internal/gates/`: both exit 0 — compiles both ways.
- `go test ./internal/store/ -run TestEmbeddedPort -count=1`: PASS — port contract holds.
- `go test ./internal/migration/ -count=3`: ok, no "already listening on port 55432" — determinism
  fix confirmed.
- `make check`: FAILED (exit 2) at `hooks-test` → `p6-route-matrix` self-test.
- Scope: within the 4 named packages; violates no exclusion. Confirmed.

## Findings

- **F1 (HIGH, BLOCKING; FIXED in `8700760`) — stale `@generated` route matrix (Law 1).**
  `docs/verification/p6-event-route-matrix.md` derives its "Code commit" pin from
  `git log -1 -- kernel/internal/{projections,cli,twin,gates}` (`scripts/p6-route-matrix.sh:57`).
  `057bc07` modified `projections` and `gates`, advancing the derived commit to `057bc07…`, but the
  artifact still pinned the prior `9d07b5f…`, so `--check` fails and `make check` is RED. The author's
  pre-commit `make check` passed only because `git log` cannot see uncommitted changes — committing
  invalidated the artifact (an artifact-staleness trap), falsifying the commit message's "make check
  green". Mechanical fix: `scripts/p6-route-matrix.sh --write` + commit (only the pin line changes).
  *Author note:* done in `8700760`; `route-matrix --check` now clean. Independently verified by the
  owning session before the fix (pinned `9d07b5f…` vs derived `057bc07…`).
- **F2 (LOW, residual, accepted) — TOCTOU in `freeTestPort`.** listen `127.0.0.1:0` → read port →
  close → embedded-postgres binds later: a window exists where another process could grab the freed
  port. Inherent to the pattern and identical to the established `twin.temporaryPort` (`replay.go:363`)
  the change deliberately mirrors. Low risk, accepted by precedent; residual, not eliminated.
- **F3 (LOW, out-of-scope, candidate for follow-up) — other fixed ports remain in tests.**
  `twin/replay_test.go:339` `Port: 55440` (untagged, runs under make check; distinct from 55432 and
  unique to twin, so no current cross-package contention, but the same fixed-port smell);
  `store/integration_test.go:393/404` `Port: 55441` (integration-tagged only);
  `kernel/cmd/p6dbprobe/main.go:13` default `55433` (untracked dir). None contend with the 55432 set;
  all outside `CI-flake-ephemeral-port-7d5f10`'s named 4-package scope.

## Reviewer identity

Independent non-author reviewer (fresh-context subagent dispatched by the owning session, 2026-11-09).
Did not author the kernel change.
