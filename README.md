# Proofbound

Proofbound is a warranty layer for software delivery.

It helps teams answer questions that ordinary build logs and test dashboards do not preserve well:

- What was this software change supposed to do?
- What evidence shows that it does it?
- Which exact commit was verified?
- Who performed or owns the verification?
- What remains unverified if something fails later?

The result is durable, machine-readable proof bound to the software it describes—not just a
temporary “tests passed” message.

## Why Proofbound?

Software is increasingly produced and changed by automated systems. That makes it important to
preserve more than source code and a green pipeline. Proofbound is designed to make intent,
verification evidence, verdicts, provenance, and accountability travel with each meaningful change.

Proofbound complements CI/CD and test runners. Those systems execute builds and checks; Proofbound
records what was claimed, what was checked, what the evidence supports, and where the responsibility
for the result lies.

## How it works

At a high level, a delivery workflow will:

1. State the behavior or invariant a change is expected to preserve.
2. Run the relevant tests, checks, and verification procedures.
3. Bind the evidence and verdict to the exact commit and project context.
4. Preserve the result as an auditable artifact.
5. Make proven, failed, and unverified behavior distinguishable.

Proofbound can be introduced alongside a new project or used to assess an existing product. An
existing product starts with a baseline: it can document what is already tested and verified, but
it cannot retroactively prove behavior for which no evidence exists.

## Project status

This repository is an early-stage working scaffold, not yet a finished end-user product.

As of 2026-08-26, P1 Tasks 0–9 and P2 gates-as-data are implemented and accepted under the
repository’s independent verification rule. P2 is closed; the next planned phase is P3.

Current project status: P2 gates-as-data complete; P3 external connector is next.

Currently implemented and verified (P3 slice is under verification):

- Go kernel and core event primitives
- Durable store foundations with migrations and append/read operations
- Transaction handling and embedded/external database configuration
- PostgreSQL-backed integration tests
- Git, checks, and sessions connectors
- Projection rebuild and `vera verify`
- `vera report week`, including event-ID proof and `[superseded]` commit marking
- Review-verdict ingestion, review projection rows, and ledger-ordered red-verdict chains
- Machine-enforced spec-first coverage for every kernel package
- P2 gate canary evaluation for the latest `make check` witness
- Explicit canary-to-enforce gate promotion workflow; current seven-gate set promoted to enforce
- Generated-files freshness gate witness (`make index-check-witnessed`) with compound matching
- Law citation lint gate witness (`make law-citation-witnessed`)
- SPEC invariant numbering gate witness (`make spec-numbering-witnessed`)
- Invariant table gate witness (`make invariant-table-witnessed`)
- Link lint gate witness (`make link-witnessed`)
- Kernel build/test/lint gate witness (`make kernel-check-witnessed`)
- Explicit delivery enforcement workflow (`make delivery-enforce`): fresh witnesses, ingestion, then fail-closed gate enforcement
- P2 canary-to-enforce bad-witness acceptance evidence
- P3 GitHub connector decision and initial workflow/deployment ingestion slice (under verification)
- P3 live fixture-selection evidence for `github/docs`
- Verification and mutation-testing infrastructure

Remaining work:

- P3: complete the GitHub deployed-where / tested-what projection and live freshness acceptance

The mutation harness’s report-package integration calibration currently exceeds its 30-second
ceiling after cache initialization; this limitation is documented in the Task 8 evidence. The
standard `make check` gate is green.

## Quick start

### Prerequisites

- Git
- Go 1.26 or newer
- Bash
- `golangci-lint` (required by the full check)
- Docker, if running PostgreSQL integration or mutation tests

### Build and test

From the repository root:

```bash
make check
```

The full gate runs repository checks, link and documentation checks, builds the Go packages, runs
the unit tests, and runs linting. For a faster inner loop:

```bash
make short
```

To run the promoted data gates as a delivery boundary, use `make delivery-enforce`. It serializes
the workflow, emits fresh witnesses for every promoted target, ingests them, and then runs
`vera gates enforce`.

The current command-line entry point is a scaffold while the product workflow is being built. The
working implementation and tests live under [`kernel/`](kernel/).

## Repository layout

| Path | Purpose |
| --- | --- |
| [`kernel/`](kernel/) | Go implementation and package tests |
| [`docs/`](docs/) | Specifications, decisions, and verification documentation |
| [`scripts/`](scripts/) | Repository checks and development tooling |
| [`ROADMAP.md`](ROADMAP.md) | Project phases and completion criteria |

## Contributing

The project is being developed specification-first. Before changing implementation behavior, read
the relevant package `SPEC.md` and run the checks that cover it. Please open an issue describing the
behavior, evidence, or workflow problem before proposing a larger change.

Maintainers and automated build agents can find the internal operating rules in
[`CLAUDE.md`](CLAUDE.md), the current resume state in [`notes/state.md`](notes/state.md), and the
long-form product context in [`vision-plain-english.md`](vision-plain-english.md). Those documents
are intentionally separate from this end-user overview.

## License

See [`LICENSE`](LICENSE).
