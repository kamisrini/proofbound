# Proofbound — Release 2 Release Notes

**Release 2 — "the SDD Warranty Connector"**
Suggested tag: `v0.2.0` · Line decided in [`VD-release-line-r2-sdd-t5j34p`](../../decisions/VD-release-line-r2-sdd-t5j34p.md) · Scope tracked in [`plans/RELEASE-PLAN.md`](../../plans/RELEASE-PLAN.md) · Plan: [`sdd-warranty-connector-plan.md`](../../plans/sdd-warranty-connector-plan.md)

**Status: awaiting owner acceptance.** Every milestone below passed one independent (non-author)
review under the Law 11 review budget. Release 2 is accepted when the owner says so.

---

## What Release 2 is

Release 2 turns Proofbound into a **warranty for spec-driven development**. Teams that write specs
with [OpenSpec](https://github.com/Fission-AI/OpenSpec) or [Spec Kit](https://github.com/github/spec-kit),
and build with any agent, get a record that answers:

- Which spec does this commit claim to implement, at exactly which revision?
- Has anyone other than the author verified every obligation of that spec, and what did they decide?
- Has the spec changed since it was verified, and if so, which obligations?

Proofbound does not replace your agent or your spec tool. It reads what they leave in git, so it
works beside any harness. It carries all of Release 1.

New here? Start with the [**Release 2 demo**](DEMO.md): one command runs Proofbound on OpenSpec's own
repository.

## What's included

| Capability | What it gives you | Run it |
|---|---|---|
| **SDD connector** (OpenSpec, Spec Kit) | Every change or feature, and every requirement in it, recorded from the project's real git history with exact content digests | `proofbound sync intent sdd` |
| **Commit binding** | A commit ending `Intent: sdd:<name>` binds to the exact revision of that change; a claim on a change that does not exist is reported, never silently accepted | `proofbound sync git` |
| **Readiness gate, with reasons** | A commit is warranted only when every obligation it claims is SATISFIED and every requirement VERIFIABLE, by a reviewer who is not the spec's owner; BLOCKED says why, with counts | `proofbound gates canary` |
| **Starter kit** | Profile, `.gitignore` lines, agent instructions (AGENTS.md), a Claude Code reviewer subagent, and a commit hook, in one command | `proofbound init {openspec\|speckit}` |
| **Reviewer workflow** | Draft every obligation a commit claims, decide each one, seal: writes the review and verdict artifacts and the commit's own gate; everything but the decisions comes from the ledger | `proofbound review draft <commit>` · `review seal <dir>` |
| **Drift** | Flags every warranty whose spec now promises something else, obligation by obligation, with the old and new wording; ignores archive moves and non-promise edits | `proofbound report drift` |
| **Lookup by name** | Report a change by its folder name instead of an internal id | `proofbound report intent sdd:<name>` |
| **Clean output** | The embedded database's server log goes to `.proofbound/db.log`, not your terminal | — |

## Highlights for adopters

- **Proven on real data.** OpenSpec's own repository: 1,059 commits and 648 spec revisions, read
  in about a minute. A claimed commit went from BLOCKED (0 of 4 verified) to PASS after review. One
  reworded obligation showed up in `report drift` as exactly that obligation.
- **Drift compares promises, not files.** Measured on OpenSpec's history: archiving a change
  changes every digest, yet 60 of 73 archived changes promise the same thing. A digest rule would
  raise a false alarm on every one of them; Proofbound flags only the real rewrites.
- **Fail-closed by construction.** A commit claiming an unknown change is reported, not bound. A
  review can't be sealed by the spec's owner, with an undecided outcome, or with references that
  don't match the ledger. Drift reads the repository at HEAD, so a revert to a never-verified spec
  is STALE.

## Known limitations and honest caveats

- **Independence is declared, not proven.** A reviewer states their name; on a one-person project
  the record says who *claimed* to verify, not that they are independent. Seal refuses the spec's
  own owner as reviewer.
- **Verification is a human or agent decision.** Proofbound records and checks verdicts; it does
  not decide whether code satisfies a requirement. Until someone reviews, gates say BLOCKED.
- **Spec Kit was proven on fixtures and end-to-end tests, not on a real public project.** Spec
  Kit's own repository does not use its layout, and its spec template matches the mapping.
- **Spec Kit lines must use the template's exact form** (`- **FR-001**: ...`). Near-miss forms are
  read as prose, silently. Status lines (`**Status**: Draft`) are not mapped.
- **OpenSpec main specs, `design.md`, `tasks.md` and Spec Kit's `plan.md` are not parsed.** Only
  change deltas and feature specs carry obligations.
- **Drift is information, not a gate.** `report drift` exits 0. The readiness gate still judges a
  commit against the exact revision it claimed.
- **The commit hook is a convenience.** `git commit --no-verify` skips it.
- **Low-severity review findings are recorded, not fixed** (Law 11). Each is in its milestone's
  verdict under `docs/verification/verdicts/r2-*`.

## Install

As Release 1:

```bash
git clone <this-repo-url> proofbound && cd proofbound/kernel
CGO_ENABLED=0 go build -o "$HOME/bin/proofbound" ./cmd/proofbound   # put it on your PATH
```

Then, in your OpenSpec or Spec Kit project: `proofbound init openspec` (or `init speckit`). See
[the demo page](DEMO.md#use-it-on-your-own-openspec-or-spec-kit-project).

## Tagging this release

After the owner accepts Release 2:

```bash
git tag -a v0.2.0 -m "Release 2 — the SDD Warranty Connector"
git push origin v0.2.0
```

---

*Next: [Release 2 demo](DEMO.md) · [Release 1 notes](../r1/RELEASE-NOTES.md)*
