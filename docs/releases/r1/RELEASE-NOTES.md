# Proofbound — Release 1 Release Notes

**Release 1 — "the Trust Record"**
Suggested tag: `v0.1.0` · Line decided in [`VD-release-line-kwdh4v`](../../decisions/VD-release-line-kwdh4v.md) · Scope tracked in [`plans/RELEASE-PLAN.md`](../../plans/RELEASE-PLAN.md)

---

## What Release 1 is

Release 1 is the **records-and-verifies** cut of Proofbound. It ingests a project's real
delivery record — the commits that happened, the decisions that were cited, the checks that
ran, the verdicts that were issued — into an append-only event ledger, and lets you read it
back as durable, machine-readable proof bound to the exact commit it describes.

It answers questions that ordinary build logs and green CI badges do not preserve:

- What was this change supposed to do, and which decision authorized it?
- Which exact commit was recorded, and what is its permanent proof ID?
- What evidence and verdicts are bound to it?
- What is *still* unverified?

New to Proofbound? Start with the [**User Guide**](USER-GUIDE.md) (step-by-step, with a sample
app you build from scratch) or the [**5-minute demo**](DEMO.md).

## What's included

| Capability | Phase | What it gives you |
|---|---|---|
| **Flight Recorder** | P1 | Append-only event ledger; git, check-runner, and agent-session connectors; projections; `proofbound verify` (rebuild + snapshot equality) |
| **Gates-as-data** | P2 | Gate definitions evaluated against the ledger, with a canary → enforce promotion workflow |
| **GitHub connector** | P3 | Read-only ingestion of GitHub Actions runs and deployments |
| **Intent provenance** | P5 | A business-decision → requirement → change-intent (BD/BR/CI) chain, bound to the commit that satisfies it |
| **Behavior locks + runtime replay** | P7 B1 | Behavior locks over the twin-replay path, replayed against a real PostgreSQL |

Every phase above is deep-complete and **independently accepted** under the repository's
non-author verification rule.

## Highlights for adopters

- **Pure-Go binary.** Builds with `CGO_ENABLED=0`. No C toolchain, no Java, no Docker needed
  to build the binary or run the Flight Recorder.
- **Bring your own Postgres, or use embedded dev-mode.** Point `DATABASE_URL` at your own
  PostgreSQL, or leave it unset and Proofbound runs a local embedded PostgreSQL under
  `./.proofbound`.
- **Works on any git repository.** Proofbound records whichever repo you run it inside — not
  just its own checkout.
- **Idempotent ingestion.** Re-running `sync` appends only new events.
- **Proven on a clean clone.** A fresh clone builds and passes the full gate on a clean
  machine — see [`verification/r1-2-fresh-clone-release1.md`](../../verification/r1-2-fresh-clone-release1.md).

## What Release 1 does *not* do yet (this is Release 2)

Release 1 ships a system that **records and verifies**. It does **not**:

- regenerate failing code (witnessed regeneration — the full loop),
- predict outcomes or publish calibration (the Living Twin as a product),
- govern actions with policy envelopes or a human decision inbox (Governor),
- issue portable, signed warranties (the Trust Engine), or
- redistribute a bundled database.

Those are planned for **Release 2** and are not built. The full R1-vs-R2 boundary is in
[`plans/RELEASE-PLAN.md`](../../plans/RELEASE-PLAN.md).

## Known limitations and honest caveats

- **`proofbound verify` is Proofbound's own delivery boundary.** On a plain application repo it
  reports `no make check witness found in spool` and exits non-zero — that is expected. `verify`
  rebuilds projections and checks snapshot equality against a *witnessed* `make check` run, which
  only Proofbound's self-hosting workflow produces. For your own project, the trust surface is
  `sync` + `report` + `rebuild`; `verify` is not part of the basic flow. See the
  [User Guide](USER-GUIDE.md#what-verify-is-and-isnt).
- **Embedded dev-mode is for local use, not redistribution.** First run downloads an embedded
  PostgreSQL (~tens of MB) into `./.proofbound/pgbin`, which needs outbound internet once. In a
  locked-down environment, use `DATABASE_URL` against your own Postgres instead.
- **Embedded PostgreSQL prints startup/shutdown logs to stdout.** They are interleaved with
  Proofbound's output in dev-mode. Using your own `DATABASE_URL` removes this noise. See the
  [User Guide](USER-GUIDE.md#a-note-on-the-postgresql-log-noise).
- **`report week` shows a trailing 7-day window.** Commits older than that are in the ledger but
  do not appear in the weekly view.
- **GitHub connector is read-only** and needs `PROOFBOUND_GITHUB_OWNER`, `PROOFBOUND_GITHUB_REPOS`,
  and a `GITHUB_TOKEN`.

## Install

```bash
git clone <this-repo-url> proofbound && cd proofbound/kernel
CGO_ENABLED=0 go build -o "$HOME/bin/proofbound" ./cmd/proofbound   # put it on your PATH
```

Prerequisites: Git and Go 1.26+ (the maintained build uses Go 1.27). `golangci-lint` and GNU
Make are only needed to run Proofbound's *own* `make check`, not to build or use the binary.

## Tagging this release

Release 1 is a documentation-and-acceptance milestone over the existing accepted phases. To tag it:

```bash
git tag -a v0.1.0 -m "Release 1 — the Trust Record"
git push origin v0.1.0
```

---

*Next: [User Guide](USER-GUIDE.md) · [5-minute demo](DEMO.md)*
