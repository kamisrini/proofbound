# VD-repo-structure-55mzvd: the repo root is the mono-home; vision docs stay at root

**Status:** Accepted
**Date:** 2026-08-07
**Context:** VERA needs one home for vision, decisions, notes, and (from P1) the kernel code. Splitting them across repos would create cross-repo drift on day one — the exact failure the Build Laws exist to prevent.
**Decision:** This repository is the single home for everything: vision docs at root (they are the project's most-read artifacts and deserve top-level visibility), `docs/decisions/`, `docs/gates.md`, `docs/plans/`, `docs/design/`, `notes/` (state.md + journal/ + tmp/), `.claude/` (hooks/commands/settings), `kernel/` (the product Go module, from P1), `Makefile`, `ROADMAP.md`, `CLAUDE.md`. Sessions start from the repo root so the project hooks and commands load.
**Consequences:** One clone = the whole project; the meta-tax metric can be computed from one git log. The root directory is busier than a purist layout — accepted in exchange for the vision docs being the first thing a reader sees.
**Revisit when:** a second contributor joins (branch/PR flow and idea intake get added then), or the kernel grows large enough to warrant its own release cadence.
