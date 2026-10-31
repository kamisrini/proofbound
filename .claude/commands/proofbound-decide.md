---
description: Record a decision as a durable VD-* record, immediately (not at session end)
---

Record a decision as a durable decision record. Do this the moment a decision is made during work,
not at session end (CLAUDE.md § Session protocol).

First, check this is actually a decision worth a record: a choice between real alternatives, with
consequences we are committing to. A task note, a plan step, or a fact goes in the journal or
`notes/state.md`, not here. "Why we chose X" has exactly one home — `docs/decisions/VD-*.md` — and
everything else cites the ID (Law 2). If an existing VD already covers this, amend that file in place
(add a dated `**Amended YYYY-MM-DD:**` paragraph as the existing records do) rather than minting a
competing one.

Then:

1. Mint the ID. Copy the shape of `docs/decisions/000-template.md`. The slug is short and stable.
   The suffix follows existing naming: a foundational/standing decision uses a random 6-char token
   — `LC_ALL=C tr -dc 'a-km-z2-9' < /dev/urandom | head -c 6`, re-mint if the filename already
   exists — while a phase decision uses a `YYYY-MM-DD` date suffix (see the `VD-p5-…`, `VD-p7-…`
   files). Match whichever the decision is.
2. Write `docs/decisions/VD-<slug>-<suffix>.md` with the template's fields, all filled:
   - Title line exactly `# VD-<slug>-<suffix>: <Title>` (the index generator parses this).
   - `**Status:**` (`Accepted`, or `Superseded by VD-…`), `**Date:**` today (absolute ISO).
   - `**Context:**` the situation and the forcing question. `**Decision:**` concrete, testable
     phrasing — not an aspiration. `**Consequences:**` what it makes easy, what it makes hard, what
     we gave up (state the accepted cost plainly). `**Revisit when:**` a concrete trigger, never
     "later".
3. Regenerate the index: `make index` (INDEX.md is `@generated` — never hand-edit it, Law 1).
4. Confirm cheaply: `scripts/link-lint.sh` (the new VD must resolve in INDEX.md) and
   `scripts/law-citation-lint.sh` if the record cites any Build Law.
5. If any OTHER file should cite this decision, add the bare `VD-<slug>-<suffix>` reference there —
   link, never restate the rationale (Law 2).

**Governed-path note:** the `VD-*.md` file itself is not a governed path, but `make index` rewrites
`docs/decisions/INDEX.md`, which IS governed (`scripts/intent-applicability.sh`). So committing a
decision is a governed change and expects an owner-approved change intent. Commit only when the
owner authorizes it; record the authorization in the commit body as prior governed commits do.

End by printing the new ID, its path, and any files you updated to cite it.
