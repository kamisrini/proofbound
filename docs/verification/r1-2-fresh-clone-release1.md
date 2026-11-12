# R1-2 — Release 1 fresh-clone proof (Linux)

**Date:** 2026-11-12

**Frozen commit:** `bab17a7` (pushed `origin/main` HEAD)

**Clone mode:** local `git clone --no-local` into a disposable dir, no copied `.proofbound` state,
**fresh embedded-Postgres data dir** and **fresh Go build + golangci-lint caches** (a faithful
cold "clean machine" build). The clone path and database are not evidence; the frozen commit and the
command exit codes below are.

**Toolchain (documented prereqs — this is a documented-prereqs clean clone, not a bare box):**
Go `1.27.4`, `golangci-lint 2.15.0` (built with go1.27.3) on `PATH` via `~/go/bin`. The `proofbound`
binary builds pure-Go (`CGO_ENABLED=0`). On first run the embedded Postgres binaries download into
`.proofbound/pgbin`; no system Postgres, gcc, Java, or Docker is required for the Flight Recorder path.

**Result:** **PASS for the R1-2 Definition of Done** (fresh clone builds + `make check` green). The
live Release-1 surface (sync / report / rebuild) also passes end-to-end when bootstrapped in the
correct order; the one failing command, `verify`, fails only on the already-tracked
`DEF-DELIVERY-ENFORCE-VERDICTS` deferral. Two findings are recorded below.

## R1-2 core — build + gate

| Command | Exit | Result |
|---|---:|---|
| `git clone --no-local <repo> clone` | 0 | Clone at `bab17a7`; worktree clean (the two permanent untracked paths are untracked in the source, so they are correctly absent from the clone). |
| `CGO_ENABLED=0 go build -o proofbound ./cmd/proofbound` | 0 | Pure-Go binary (18 MB). |
| `PATH=~/go/bin:$PATH GOCACHE=<fresh> GOLANGCI_LINT_CACHE=<fresh> make check` | 0 | Cold-cache `go build ./...` + `go test ./...` (no `-tags integration`) + `golangci-lint run ./...` → `0 issues.`; all hook self-tests `ok`. Expected negative-fixture diagnostics emitted separately by the harness. |

## Live Release-1 surface — correct bootstrap order (migrate first, on an empty ledger)

| Command | Exit | Result |
|---|---:|---|
| `proofbound migrate historical-evidence` | 0 | `imported=5` (the P5 intent prerequisite closure + cited evidence envelopes). **Must run first, on an empty ledger** — see Finding A. |
| `proofbound sync intent all` | 0 | `listed=32 appended=29 existing=3`. |
| `proofbound sync git` | 0 | `listed=601 appended=600 existing=1`. |
| `proofbound sync git` (re-run) | 0 | `listed=601 appended=0 existing=601` — idempotent. |
| `proofbound sync intent all` (re-run) | 0 | `listed=32 appended=0 existing=32` — idempotent. |
| `proofbound report week` | 0 | Rendered the week `2026-11-05 → 2026-11-12`, `commits count=64`, each with files / cited decisions / proof event IDs. |
| `proofbound report intent CI-implement-p5-0a1b2c` | 0 | Self-hosted intent chain `state=SATISFIED`; both `BR-intent-chain-d4e5f6` obligations `VERIFIABLE/SATISFIED`; exact revision proofs resolved. |
| `proofbound report requirement BR-self-hosting-correctness-4607a3` | 0 | Rendered `status=proposed` with declared authorization proof. |
| `proofbound intent check --commit bab17a7` | 0 | `commit=bab17a7 intent=none` (HEAD is an ungoverned `notes:` commit — correct). |
| `proofbound rebuild` | 0 | Explicit projection rebuild completed with no error. |
| `proofbound verify` | **1** | **Fails** — `initial reviews sync: reviews connector: docs/verification/verdicts/p7-2b-detector-power-spec-review-round1-20261105.md: unknown finding field "summary"`. This is the known `DEF-DELIVERY-ENFORCE-VERDICTS` (Finding B). |

## Finding A — bootstrap order is load-bearing and undocumented (usability / doc gap)

`migrate historical-evidence` must run **first, on an empty ledger**, before any `sync`. If `sync`
runs first the ledger is no longer empty, `migrate` refuses (`historical evidence migration: ledger
is not empty`), the P5 historical intent closure (`CI-implement-p5-0a1b2c@f90b216…`) is never imported,
and the projection apply then **fails closed** on `commit seq 139: dangling intent revision …` — so
`report week`, `rebuild`, and `verify` all break. This is **not a product-correctness bug**: the
cited digest `f90b216…` exactly matches the record's current `artifact_sha256`, and with the correct
order every command above passes. It is a usability finding: the required bootstrap sequence is
undocumented. It reinforces the open R1 wrap-up item "write a real end-user quickstart", which must
state the order **`migrate historical-evidence` → `sync intent all` → `sync git` → `report`/`verify`**.

## Finding B — `verify` is red on a fresh clone (DEF-DELIVERY-ENFORCE-VERDICTS)

`proofbound verify` runs an initial reviews sync, and the reviews connector rejects pre-existing
committed verdicts that are not yet `vera.verdict.v1`-conformant (here, a `summary` finding field the
connector forbids). This is the already-tracked `DEF-DELIVERY-ENFORCE-VERDICTS` (currently OPEN,
labeled R1-4 "owner-optional" because `make delivery-enforce` is not a `make check` gate). This proof
shows the impact is larger than a gate: **`verify` — a headline Release-1 verb ("records *and*
verifies") — does not run on a fresh clone until R1-4 is done.** Recommend the owner consider
reclassifying R1-4 from owner-optional to required for R1. (Decision left to the owner; not changed here.)

## Scope note

The R1-2 Definition of Done (fresh clone builds + `make check` green) is **met**. Findings A and B
concern the live product surface beyond that DoD and are recorded here so the clean-clone proof is
honest about what a new user actually encounters; neither is a regression introduced by this work.
