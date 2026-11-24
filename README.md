# Proofbound

**Proofbound is a warranty layer for code written by AI agents.**

People write code at a pace other people can review. Agents do not: they produce many changes a day,
each larger than a person would write, and the commit log rarely says which agent did what, what it
was asked to do, or whether it stayed inside that job. Proofbound keeps that record, and makes it
something you can check.

For every change an agent makes, Proofbound answers:

- **What was the agent asked to do?** The spec it claims, which must be committed before the code.
- **Did it stay inside that?** Every changed line outside the spec's declared area is listed, and a
  reviewer must accept each one with a reason or reject it.
- **What evidence shows it did it?** Each promise is checked by a reviewer who is not the agent, the
  commit's author or the spec owner, and Proofbound undoes each part of the change to show whether the
  tests notice.
- **Which exact commit, and which agent?** Agents are named from authorship and co-author lines, each
  committed line is traced to the Claude Code session that wrote it (or flagged as written by no
  recorded agent edit), and commits can be signed by keys the repository declares.
- **What remains unverified?** Code with no spec, promises that changed after review, and every
  exception, with who accepted it and why.

The result is a signed warranty bound to the exact commit, which anyone can verify without your
infrastructure, not a green check that disappears with the CI log.

## You don't have to trust the agent

Proofbound does not make a model safe, and does not try to. It makes what an agent did **visible and
checkable** to the people who answer for the software: the engineering lead, the reviewer, the
auditor. Its checks are built to hold even when the agent is wrong, careless or working against you:

| Holds even if the agent misbehaves | Why |
|---|---|
| Lines changed outside the spec's area are named (C10) | read from git, not from anything the agent reports |
| A test that cannot fail is caught (C11) | Proofbound undoes each part of the change and runs the tests itself |
| No one grades their own work (C8) | a review by the change's author, agent or spec owner is refused |
| Signed commits (C9) and signed reviews | each signature is checked against keys the repository declares |
| The warranty | bound to the exact commit; anyone can verify it from a clone |

Two features help while the agent works, but the agent runs on the same machine and could get around
them, so Proofbound never relies on them alone:
- **The write-time perimeter:** a Claude Code hook refuses an edit outside the spec's area as it is
  made. A shell command bypasses it; C10 still judges every committed line.
- **Session provenance:** a Claude Code hook logs which session wrote which lines. The agent could
  write to its own log; the reviewer signs what the logs showed, and lines no recorded edit wrote are
  named.

Each page and record says what it does not prove.

## Why agents change the problem

- **Volume.** No one can read everything an agent writes. Review has to aim at the promises and at
  what fell outside them.
- **Identity.** A git author name is whatever was typed. Proofbound names agents, and can require
  that commits be signed by declared keys.
- **Self-grading.** An agent that writes the code and the tests can pass its own work. Proofbound
  never accepts a verdict from the change's author, agent or spec owner, and shows when the tests do
  not depend on the change at all.
- **The perimeter.** Compliance asks whether the agent acted only inside what was authorized. A commit
  log cannot answer that; Proofbound's Audit room answers it control by control.

**Start here:** install and the agent demo, in [docs/releases/r3/README.md](docs/releases/r3/README.md).
The [agent demo](docs/releases/r3/DEMO-AGENT.md) shows an agent going off-script, step by step.

The same record works for changes people make. Proofbound complements CI/CD and test runners: they run
the checks; Proofbound records what was promised, what was checked, what the evidence supports, and
who is accountable for the result.

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

**What Release 1 does not do yet:** it does not regenerate failing code, predict outcomes, govern
actions with policy envelopes, issue portable signed warranties, or redistribute a bundled database —
those are **Release 3**. Its first slice, the portable signed warranty, is accepted and published as
v0.3.0 to v0.3.2 ([`docs/releases/r3/DEMO.md`](docs/releases/r3/DEMO.md)). **Release 2**, the SDD Warranty Connector, is
built and accepted (2026-11-19): it warrants the output of spec-driven-development tools (OpenSpec and
Spec Kit) by binding each commit to the exact spec revision it claims, recording independent
verification, and flagging drift when a verified spec changes.

**Release 1 wrap-up is complete.** The hardening items that make Release 1 usable by someone else
(R1-1 through R1-4) are all done — the clean-clone proof, staged-aware pin freshness, and the
verdict-schema cleanup that lets `proofbound verify` run on a fresh clone. The maintained scope and
status view is [`docs/plans/RELEASE-PLAN.md`](docs/plans/RELEASE-PLAN.md).

**New to Proofbound? Start here:** [`docs/releases/r3/README.md`](docs/releases/r3/README.md), one page
with install, the quickstart, every command and demo, and links to the earlier releases.

Earlier releases' pages, for history:

- [`docs/releases/r1/USER-GUIDE.md`](docs/releases/r1/USER-GUIDE.md) — step-by-step, with a sample
  app you build from scratch, and how to run Proofbound on your own project (including a single-dev
  corporate setup).
- [`docs/releases/r1/DEMO.md`](docs/releases/r1/DEMO.md) — a copy-paste 5-minute demo.
- [`docs/releases/r1/RELEASE-NOTES.md`](docs/releases/r1/RELEASE-NOTES.md) — what Release 1
  includes, its limitations, and what is deferred to later releases.
- [`docs/releases/r2/DEMO.md`](docs/releases/r2/DEMO.md) — Release 2 (accepted 2026-11-19): the whole
  warranty loop on OpenSpec or Spec Kit in about two minutes, OpenSpec's own repository at scale, and
  how to use it on your project. What it includes and its caveats:
  [`RELEASE-NOTES.md`](docs/releases/r2/RELEASE-NOTES.md).
- [`docs/releases/r3/DEMO.md`](docs/releases/r3/DEMO.md) — Release 3, slice 1 (accepted 2026-11-20):
  export a warranted commit as one signed file that anyone can verify against a clone, with no
  database.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/kamisrini/proofbound/main/install.sh | sh
```

Then follow the [quickstart](docs/releases/mvp/QUICKSTART.md) to your first warranty. Current release:
[v0.3.2](https://github.com/kamisrini/proofbound/releases/tag/v0.3.2).

## Quick start (from source)

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

### GitHub connector (optional)

To ingest GitHub Actions runs and deployments for your repositories, configure the connector via
environment variables before `proofbound sync github`:

```bash
export PROOFBOUND_GITHUB_OWNER='your-org-or-user'
export PROOFBOUND_GITHUB_REPOS='repo-a,repo-b'        # comma-separated
export GITHUB_TOKEN='ghp_...'                         # a read-only token
export PROOFBOUND_GITHUB_API_BASE_URL='https://api.github.com'   # optional; set for GitHub Enterprise
proofbound sync github
proofbound report github
```

Without `PROOFBOUND_GITHUB_OWNER`/`PROOFBOUND_GITHUB_REPOS` set, `sync github` fails with
`github connector: API, Owner, IDs, and Logger are required`.

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

### Maintainer / operator targets

Beyond `make check` and `make short`, the repository exposes operator targets used by maintainers and
automated build agents:

- `make verify` — projection rebuild + snapshot-equality verification against a configured ledger.
- `make kernel-check` — build, test (no `integration` tag), and lint the Go kernel only.
- `make delivery-enforce` — serialize fresh witnesses, ingest them, then run `proofbound gates enforce`
  as a delivery boundary.
- `make gates-canary` / `make gates-enforce` — evaluate the promoted data gates in canary or enforcing mode.
- `make check-witnessed` — run the gate and publish a witness of the result to the local spool.
- `make index` / `make index-check` — regenerate or verify the freshness of generated indexes.
- `make invariants-lock` — refresh the invariant lock files.
- `make identity-inventory` — audit identity markers for stray live aliases.
- `make mutants PKG=internal/<package>` — run a package mutation sweep (package-acceptance evidence;
  intentionally outside `make check`).

Witnessed and delivery targets require a configured local ledger. For a fresh clone carrying a
historical portability archive, run `proofbound migrate historical-evidence` before verification.

## Repository layout

| Path | Purpose |
| --- | --- |
| [`kernel/`](kernel/) | Go implementation and package tests |
| [`docs/`](docs/) | Specifications, decisions, and verification documentation |
| [`scripts/`](scripts/) | Repository checks and development tooling |
| [`ROADMAP.md`](ROADMAP.md) | Project phases and completion criteria |
| [`docs/plans/RELEASE-PLAN.md`](docs/plans/RELEASE-PLAN.md) | Release 1 / 2 / 3 scope |

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
