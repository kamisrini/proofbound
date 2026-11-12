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

**Release 1 — "the Trust Record" — is feature-complete and independently accepted.** This is the
*records-and-verifies* cut of Proofbound: it ingests a project's real delivery record and verifies it,
run against a Postgres you supply (or an embedded dev-mode database). The release line is decided in
[`VD-release-line-kwdh4v`](docs/decisions/VD-release-line-kwdh4v.md); the maintained scope view is
[`docs/plans/RELEASE-PLAN.md`](docs/plans/RELEASE-PLAN.md).

Release 1 is built from phases that are each deep-complete and accepted under the repository's
independent (non-author) verification rule:

- **Flight Recorder (P1)** — an append-only event ledger with git, check-runner, and agent-session
  connectors, projections, and `proofbound verify` (projection rebuild + snapshot equality).
- **Gates-as-data (P2)** — gate definitions evaluated against the ledger, with a canary→enforce
  promotion workflow.
- **First external connector (P3)** — read-only GitHub Actions / deployments ingestion.
- **Intent provenance (P5)** — a business-decision → requirement → change-intent (BD/BR/CI) chain,
  with governed change intents bound to the exact commit that satisfies them.
- **Behavior locks + runtime replay (P7 B1)** — behavior locks over the twin-replay path, replayed
  against a real PostgreSQL, with provenance and independent acceptance.

The `proofbound` binary builds as **pure Go** (`CGO_ENABLED=0`): the Flight Recorder path needs no
C toolchain, Java, or Docker. Proofbound's first tenant is Proofbound — the project ingests and
verifies **its own** commits, checks, and intent records as its standing acceptance test. A reproduced
clean-clone proof of this release is recorded in
[`docs/verification/r1-2-fresh-clone-release1.md`](docs/verification/r1-2-fresh-clone-release1.md).

**What Release 1 does not do yet (this is Release 2):** it does not regenerate failing code, predict
outcomes, govern actions with policy envelopes, issue portable signed warranties, or redistribute a
bundled database. Those capabilities are planned and not built.

**Release 1 wrap-up still in progress:** a small number of hardening items remain before Release 1 is
declared shipped — see [`docs/plans/RELEASE-PLAN.md`](docs/plans/RELEASE-PLAN.md). Notably,
`proofbound verify` on a brand-new clone currently requires the verdict-schema cleanup tracked as
R1-4; until that lands, `verify` fails on pre-existing non-conformant verdicts.

## Quick start

### Prerequisites

- Git, Bash, and GNU Make
- Go 1.26 or newer (the maintained build uses Go 1.27)
- `golangci-lint` on your `PATH` (required by `make check`; install to `$(go env GOPATH)/bin`)
- **No Docker or C toolchain is required** to build the binary or run the Flight Recorder. Docker is
  only needed for the PostgreSQL integration tests and mutation sweeps, which are not part of the
  default `make check`.

For native Windows, use Git for Windows or MSYS2 so Bash and GNU Make can run the repository's
recipes; WSL is not native Windows verification. The tracked `setup-windows.ps1` and
`check-windows.ps1` scripts are the entry points.

### Verify the checkout

From the repository root:

```bash
make check
```

The gate builds the Go packages, runs the unit tests (without the `integration` build tag), runs
`golangci-lint`, and runs the repository's own hook, link, and freshness checks. For a faster inner
loop use `make short`.

### Run Proofbound on a project

Build the command-line tool:

```bash
cd kernel
CGO_ENABLED=0 go build -o proofbound ./cmd/proofbound
```

Choose a database. Either point Proofbound at your own PostgreSQL:

```bash
export DATABASE_URL='postgres://user:pass@host:5432/dbname?sslmode=disable'
```

…or leave `DATABASE_URL` unset to use the **embedded dev-mode** database. On first run Proofbound
downloads an embedded PostgreSQL into `./.proofbound/pgbin` and keeps its data under `./.proofbound`.
Dev-mode is for local use, not redistribution.

Then ingest and read the record. **Order matters** — if this repository carries a one-shot historical
portability archive (Proofbound's own checkout does), the migration must run **first, on an empty
ledger, before any `sync`**:

```bash
# 1. Import the historical archive FIRST, on an empty ledger (only for repos that carry one,
#    such as this one). Running sync first makes the ledger non-empty and this step will refuse,
#    which later causes report/verify to fail closed on a dangling intent reference.
proofbound migrate historical-evidence

# 2. Ingest the intent records, then the git history.
proofbound sync intent all
proofbound sync git

# 3. Read and verify.
proofbound report week                 # weekly proof view: commits, cited decisions, proof IDs
proofbound report intent   <CI-id>     # render one change-intent chain and its verdicts
proofbound report requirement <BR-id>  # render one requirement and its authorization
proofbound verify                      # projection rebuild + snapshot equality (see R1-4 note above)
```

`sync` is idempotent: re-running it appends only new events. The full command surface is:

```
proofbound sync {git|checks|sessions|reviews|github|all}
proofbound sync intent {records|specdir|all}
proofbound migrate historical-evidence
proofbound rebuild | verify
proofbound report {week|github|intent <id>|requirement <id>}
proofbound intent check --commit <sha>
proofbound gates {canary|enforce}
proofbound twin replay [--through <seq>] [--candidate <file>]
```

## Repository layout

| Path | Purpose |
| --- | --- |
| [`kernel/`](kernel/) | Go implementation and package tests |
| [`docs/`](docs/) | Specifications, decisions, and verification documentation |
| [`scripts/`](scripts/) | Repository checks and development tooling |
| [`ROADMAP.md`](ROADMAP.md) | Project phases and completion criteria |
| [`docs/plans/RELEASE-PLAN.md`](docs/plans/RELEASE-PLAN.md) | Release 1 vs Release 2 scope |

## Contributing

The project is being developed specification-first. Before changing implementation behavior, read
the relevant package `SPEC.md` and run the checks that cover it. Please open an issue describing the
behavior, evidence, or workflow problem before proposing a larger change.

Maintainers and automated build agents can find the internal operating rules in
[`CLAUDE.md`](CLAUDE.md), the current resume state in [`notes/state.md`](notes/state.md), and the
long-form product context in [`vision-plain-english.md`](vision-plain-english.md). Those documents
are intentionally separate from this end-user overview. Build agents drive the repository through the
`/proofbound-*` session commands — see the catalog in
[`docs/session-commands.md`](docs/session-commands.md).

## License

See [`LICENSE`](LICENSE).
