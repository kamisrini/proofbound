# P5 Task 0.5 evidence — Proofbound identity migration

**Captured:** 2026-09-10
**Status:** mechanically green; durability does not imply independent acceptance.

## Live surfaces

- Primary command composition moved to `kernel/internal/cli`, with the live executable at
  `kernel/cmd/proofbound` and a thin deprecated alias at `kernel/cmd/vera`.
- The Go module was already `github.com/kamisrini/proofbound/kernel` at the received baseline; all
  live imports remain under that module.
- Runtime paths, caches, temporary replay roots, scripts, tests, current docs, and gate definitions
  use Proofbound identity. Gate schema is `proofbound.gate.v1`.
- Live environment names use `PROOFBOUND_*`; the shipped `VERA_*` forms are fallback-only aliases,
  with live values taking precedence and a dated warning on legacy consumption.
- The local 105 MB runtime tree was moved intact from `.vera/` to `.proofbound/` after confirming
  the recorded lock owner was not alive. `proofbound verify` subsequently opened, replayed, and
  closed that moved embedded ledger successfully (exit 0).

The migration and conflict procedure is documented in
[`proofbound-identity-migration.md`](../proofbound-identity-migration.md). Alias removal is registered
in `docs/gates.md` for 2026-12-31 or P6 start, whichever comes first.

## Frozen compatibility

The strings and semantics `vera.witness.v1`, `vera.verdict.v1`, and `vera.replay.v1` remain unchanged
in parsers, projections, emitters, replay proof, SPECs, and tests. The focused package tests for
checks, reviews, projections, CLI integration, and twin replay passed; the aggregate full gate also
passed, so the existing byte/canonicalization assertions continued to reproduce.

The received P5 exhibit remains byte-exact at SHA-256
`9d203c96243a73cad2cc119662a9b58ea04dc7e9b17d8e5ed3b1a418118d5ffa`.

## Classified inventory

`scripts/identity-inventory.sh` enumerates every case-insensitive standalone legacy identity token
in content plus legacy path components. It emits one category per hit and fails on any fifth state.
Its own negative control injects an unclassified live-product use and proves the check goes red.
`make check` now includes this blocking inventory.

Final inventory after this evidence file was added:

```text
summary  frozen-wire=34 frozen-history=239 deprecated-alias=47 baseline-quote=39 unclassified=0
```

Allowed categories are `frozen-wire`, `frozen-history`, `deprecated-alias`, and `baseline-quote`.

## Commands and results

```text
go test ./internal/cli ./internal/gates ./internal/store ./internal/twin ./internal/connector/checks ./cmd/proofbound ./cmd/vera -count=1
PASS

go test ./internal/connector/reviews ./internal/cli -count=1
PASS

make check
PASS (exit 0; all packages; golangci-lint: 0 issues)

go run ./cmd/proofbound verify
PASS (exit 0 against moved embedded ledger)
```

The `make check` command was bare and unpiped; only the process `PATH` was extended to the existing
`/home/thamm/go/bin` installation so the host could locate `golangci-lint`.
