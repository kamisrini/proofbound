# Proofbound session commands

The `/proofbound-*` commands are the Claude Code slash commands that make this repository's
[session protocol](../CLAUDE.md) operational. This repo is built under its own laws
([`CLAUDE.md`](../CLAUDE.md)): the build process must leave the same kind of durable, checkable
record the product itself is about — decisions, package contracts, independent verdicts, and a live
resume note — rather than knowledge that lives only in a session that later disappears. Each command
compiles one recurring discipline into a single entry point so it is run the same way every time
instead of improvised.

This page is a **catalog**. The full why/what/how for each command lives in its command file under
[`.claude/commands/`](../.claude/commands/) — that is its one home, and this page links rather than
restating it (Law 2, one home per datum).

## The commands

| Command | Definition | What it does | When to run it |
| --- | --- | --- | --- |
| `/proofbound-next` | [proofbound-next.md](../.claude/commands/proofbound-next.md) | Orient from `notes/state.md`, git, and the gate's honest state; propose the next *authorized* task. Ends with a recommendation, not an edit. | At session start, or any time you need to re-orient. |
| `/proofbound-spec` | [proofbound-spec.md](../.claude/commands/proofbound-spec.md) | Write a kernel package `SPEC.md` **before** its code (Law 6): interface, numbered `INV-*` invariants with their proving tests, non-goals, and test derivations. | Before implementing any kernel package. |
| `/proofbound-decide` | [proofbound-decide.md](../.claude/commands/proofbound-decide.md) | Record a decision as a durable `docs/decisions/VD-*.md` and regenerate the decision index. Decisions have one home; everything else cites the ID. | The moment a real decision is made during work — not at session end. |
| `/proofbound-review` | [proofbound-review.md](../.claude/commands/proofbound-review.md) | Run the adversarial pass against the Build Laws and the package SPEC, and commit a durable verdict. Law 9: a review of your own work is a non-accepting pre-check — acceptance needs a non-author. | Before merging a meaningful chunk. |
| `/proofbound-wrap` | [proofbound-wrap.md](../.claude/commands/proofbound-wrap.md) | Run the canonical end-of-day durability closeout in [`eod-prompt.md`](eod-prompt.md): journal, rewrite the resume note, run `make check`, commit/push coherently, and verify the backup bundle. | At session end. |

## How to use them

Start a Claude Code session from the repository root — commands and hooks load only from this
directory ([`CLAUDE.md`](../CLAUDE.md)) — and type the slash command, for example `/proofbound-next`.
A typical change runs through them in order:

1. `/proofbound-next` — pick up the next authorized task.
2. `/proofbound-spec` — before writing a new kernel package.
3. `/proofbound-decide` — whenever a real decision is made during the work.
4. `/proofbound-review` — before merging a meaningful chunk; obtain a non-author verdict.
5. `/proofbound-wrap` — close the session durably.

These commands assume this repository's checkout and tooling. They are a build-workflow aid for this
project, not part of the shipped product.

## Related

- [`CLAUDE.md`](../CLAUDE.md) — the Build Constitution and the full session protocol.
- [`docs/eod-prompt.md`](eod-prompt.md) — the canonical closeout that `/proofbound-wrap` runs.
