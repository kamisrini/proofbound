---
schema: vera.verdict.v1
verdict_id: p7-flake-ephemeral-port-kernel-round2-20261109
status: ACCEPTABLE
reviewed_commit: caffe711f58752e68676ff9b695f2a683be8e81e
findings: []
artifact_path: docs/verification/verdicts/p7-flake-ephemeral-port-kernel-round2-20261109.md
artifact_sha: 7eb7c72c917521493c1c414dc18e887ea9811ac114fea16836b5c5a3e30e4a1a
---

# DEF-KERNELCHECK-FLAKE-001 ephemeral-port kernel fix — non-author verdict (round 2)

**ACCEPTABLE.** The round-1 BLOCKING finding F1 (stale `@generated` route matrix leaving `make check`
RED) is RESOLVED. An independent non-author reviewer — the same reviewer that raised F1, re-engaged
after the author's fix; it authored neither `057bc07` nor `8700760` — re-verified the fix at HEAD
`caffe711f58752e68676ff9b695f2a683be8e81e`. Together with the round-1 verdict
(`p7-flake-ephemeral-port-kernel-round1-20261109.md`), the ephemeral-port fix for
`DEF-KERNELCHECK-FLAKE-001` is non-author accepted under change intent `CI-flake-ephemeral-port-7d5f10`
(BR-gate-integrity-4b2e08 O-1). The owner remains the final acceptor.

## Basis (what the reviewer independently ran)

1. `git show 8700760`: touches ONLY `docs/verification/p6-event-route-matrix.md`, 1 insertion / 1
   deletion — the "Code commit" pin `9d07b5f…` → `057bc07…`; route table byte-identical; no test or
   kernel logic. The fix is exactly the prescribed mechanical regeneration, no scope creep.
2. `scripts/p6-route-matrix.sh --check`: RC=0, clean (no "artifact stale").
3. Pin currency: the artifact's "Code commit" = `057bc07…` = `git log -1 --format=%H --
   kernel/internal/{projections,cli,twin,gates}`; later commits (`0d36784`, `caffe71`) touch none of
   those paths, so the derived pin stays current at HEAD.
4. `make check` run alone from repo root: exit 0, final "0 issues"; the full gate ran to completion
   (round 1 halted at `hooks-test`). The "index stale" line is the index-check self-test's own
   negative control, not a failure.

The round-1 verification of the underlying fix stands: test-only, in-scope, `go vet` clean (default +
`-tags integration`), `TestEmbeddedPort` port contract intact, `migration -count=3` green with no
55432 contention.

## Residuals (carried from round 1; noted, not blocking)

- F2 — `freeTestPort` listen-`:0`→close→rebind TOCTOU window; inherent to the pattern, mirrors
  `twin.temporaryPort`; accepted by precedent.
- F3 — other fixed ports outside this intent's scope: `twin/replay_test.go:339` (55440, untagged, no
  current cross-package contention), `store/integration_test.go` (55441, integration-tagged),
  `cmd/p6dbprobe/main.go` (55433, untracked). Candidates for a follow-up/deferral.

## Reviewer identity

Independent non-author reviewer (fresh-context subagent dispatched by the owning session, 2026-11-09;
same reviewer as round 1, re-engaged). Authored neither the kernel change nor the route-matrix fix.

## Findings

None (F1 resolved; F2/F3 are accepted residuals, not defects of this change).
