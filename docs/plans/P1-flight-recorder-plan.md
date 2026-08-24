# P1 Execution Plan — Flight Recorder Kernel (self-hosted)

**Status:** VERIFIED (3-reviewer adversarial pass 2026-08-07: kernel architect + fresh-executor simulation + laws adversary; all findings folded in) · **Stack:** Go per VD-stack-go-fid9mi · **Target:** ~6 weeks from start
**DoD authority:** ROADMAP.md § P1 is the single home of the phase DoD; `vera verify` implements it; Task 0 reconciles the wording (see 0.5).
**Audience:** a FRESH Claude session (Opus, multi-agent fan-out via subagents) in the repository root with NO prior conversation context. Everything needed is in this file + the repo. Read CLAUDE.md first (auto-loaded), then this file top to bottom. "CI" in this document means *a test that runs under `make check`* — there is no CI pipeline in P1.

## Goal (one paragraph)

Build the smallest honest Flight Recorder and point it at ourselves: an append-only **event ledger** (Postgres via pgx/v5; embedded-postgres for dev/test) fed by three **connectors** (git commits, `make check` witness files, Claude agent-session metadata), with **rebuildable projections** and a CLI that answers *"what happened this week, with proof"*. Zero external services. Every DoD item is an executable command, not a judgment.

## Non-goals (reject scope creep on sight)

- NO web UI (the CLI report is the P1 view) · NO gates engine (P2) · NO external connectors (P3) · NO twin (P4)
- NO CI pipeline (GitHub Actions etc.) — `make check` is the only gate in P1
- NO hand-authored fact events — connectors are the only writers
- NO Postgres-server install requirement for dev/test (embedded-postgres; `DATABASE_URL` escape hatch only)
- NO cryptographic signing of witnesses in P1 — content digests only; signing arrives with P2 verifier identities (vision Eight Laws, Law 2)
- NO graph database of any kind (VD-no-graph-2aa4vz)

## Task 0 — Preflight

**0.1 Backup posture (SETTLED — no action needed, see VD-local-only-backup-mjic4a):** the repo stays local; off-machine durability is `make backup` (git bundle → `~/Backups/`), run at the end of any session that produced commits. The gates.md backup row already reflects this (expiry 2026-10-16 — graduate the cadence to mechanical or delete). Adding a remote later needs no permission, but record it when it happens.

**0.2 Toolchain:** Go ≥ 1.26 (pin with the `toolchain` directive in `kernel/go.mod`); `golangci-lint` installed and on PATH (required — `make check` fails without it); journal `go version` + `golangci-lint version`.

**0.3 Decisions:** the stack AND the baseline dependency set are already recorded in **VD-stack-go-fid9mi** (pgx/v5, embedded-postgres, goose, oklog/ulid/v2, gowebpki/jcs, stdlib `flag`). Nothing to mint at preflight. **Any dependency beyond that set requires its own `/vera-decide` BEFORE `go get`.**

**0.4 Makefile contract (verbatim — do NOT edit the Makefile):** when `kernel/go.mod` exists, `make check` runs, exactly: `cd kernel && go build ./... && go test ./... -count=1 && golangci-lint run ./...`. Your job is to create the module so those commands pass.

**0.5 Reconcile ROADMAP § P1 wording (one commit, journal the rationale):** confirm ROADMAP § P1 reads "witness JSON (content-digested; signing deferred to P2 verifier identities)" (already applied); change "byte-identical view" → "row-set-identical projections (defined in the P1 plan)"; witness DoD → "a witness event for every `make check-witnessed` run since Task 5 landed". Also amend `.claude/commands/vera-wrap.md` step 3 to run `make check-witnessed` once Task 5 ships (note it in the command file at Task 5, not before).

**0.6 Defuse expiry time-bombs inside the plan window:** the stop-check advisory (gates.md) expires 2026-09-15 — graduate it NOW per its own stated path (make `/vera-wrap` treat a dirty-worktree/stale-state.md finish as blocking) or re-date it past P1 close with a journal note. Note: Task 9 must land before 2026-09-18 or the spec-first row must be re-dated first.

## Architecture — one Go module (build in this order)

*Continuity discipline: this plan implements the bootstrap chain VD → SPEC → test → code → witness — see [docs/design/continuity-chain.md](../design/continuity-chain.md) for how each hop is enforced and why the chain is derivation-bound, not citation-maintained.*

```
kernel/
├── go.mod / go.sum            # module github.com/<owner>/vera-kernel (or local path name); toolchain pinned
├── .golangci.yml              # start from a minimal strict config
├── scripts/check-witness.sh   # plain bash, kernel-absent-safe (product-classified — lives under kernel/)
├── cmd/vera/                  # CLI main: sync | rebuild | verify | report (stdlib flag + subcommand switch)
└── internal/
    ├── core/                  # envelope struct, RFC-8785 canonical hashing, ulid ids, kinds registry (no I/O)
    ├── store/                 # THE ONLY package that opens the DB. ledger + sync_runs + lock. goose migrations (embed.FS).
    ├── connector/git/         # git rev-list --all → commit events
    ├── connector/checks/      # ingest .vera/spool/*.json witnesses → check events
    ├── connector/sessions/    # Claude session JSONL metadata → session events (BEST-EFFORT)
    └── projections/           # reducers → natural-key tables; rebuild-from-genesis
```

Go conventions for this repo: standard library first; constructor-injected dependencies via a config struct; small interfaces declared where they are consumed; errors wrapped with `%w` and checked with `errors.Is`; contexts passed as the first argument; table-driven tests with hand-rolled fakes rather than mock libraries.

**Event envelope (core):** `{ event_id (ulid — identity, NEVER ordering), source, native_id, kind, occurred_at, recorded_at, payload (json.RawMessage), content_sha, connector_version }`.
- **native_id per connector:** git = commit sha · checks = `run_id` (a ulid minted by check-witness.sh and written INTO the witness JSON) · sessions = session id.
- **Idempotency key:** UNIQUE `(source, native_id, content_sha)`. A repeated `(source, native_id)` with a NEW content_sha is a **revision event** (legitimate — e.g. a session grew); projections fold revisions **last-write-wins in `seq` order**.
- **Canonical JSON = RFC 8785 (JCS)** via `gowebpki/jcs` — never hand-rolled. `content_sha` is computed ONCE at ingest from canonical bytes and **never re-derived from stored JSONB** (JSONB normalizes key order/numbers). Core tests: key-order stability, unicode, numbers (−0, exponents, >2^53 rejection), nesting, plus ONE pinned vector (known input → known sha) to catch library swaps.
- **kinds registry:** `core/kinds.go` typed string consts (`commit.recorded`, `check.run`, `session.observed`, …) — the stable strings P2 gate definitions will match on.

**Ledger rules (store):**
- `events` carries **`seq BIGSERIAL` — the sole replay order** (`ORDER BY seq`); `event_id` (ulid) is identity, never ordering.
- Append-only: the package exposes NO update/delete for events. **Append contract:** `Append(ctx, event) → (row, inserted bool, err)` — `INSERT … ON CONFLICT DO NOTHING RETURNING *` with fallback SELECT on conflict; `sync_runs.events_appended` counts `inserted=true` only.
- `sync_runs`: `(connector, cursor_json, started_at, finished_at, events_appended, error)` — observability, not correctness.
- **Single-data-dir DB (embedded-postgres reality):** dev/test runs `fergusstrange/embedded-postgres` as a child process against `.vera/db/` — two processes must NEVER start it on the same data dir. Therefore EVERY CLI command that opens the ledger acquires `.vera/sync.lock` (O_EXCL create via `os.OpenFile`, pid+timestamp). Stale takeover requires BOTH age > 10 min AND a pid-liveness probe (`syscall.Kill(pid, 0)` fails) confirming the pid is dead; long syncs refresh the lock mtime as a heartbeat. **store owns the single `*pgxpool.Pool`** (over embedded PG or `DATABASE_URL`); connectors, projections, and cli receive it injected — none of them import pgx directly.
- **Read API (the P1→P2 seam):** store exports `Append`, `ReadEvents(ctx, Filter{Source, Kind, SinceSeq, OccurredAfter}) → iterator in seq order`, and query helpers for projection tables. The P2 gates engine becomes just another `ReadEvents` consumer.
- **Schema tiers:** LEDGER schema (events, sync_runs) is goose-migrated — committed SQL files in `store/migrations/` via `embed.FS`, applied with `goose.Up` at CLI startup and in a from-empty test (fresh temp data dir). Migration SQL is generate-once/append-only, hand-edit forbidden — covered by the gates.md "Known accepted bypasses" row (Task 3 DoD adds it). PROJECTION tables are **derived state**: created idempotently by the projections package (`CREATE TABLE IF NOT EXISTS` + `projection_version`), NEVER in the migration stream — rebuild may drop them freely.

**connector/git:** `git rev-list --all` in reverse Git date-order with parents before children (not a strict timestamp sort) every sync — **the events UNIQUE index IS the seen-set**; dedupe is the idempotent append, not a cursor (no watermark bugs from amend/rebase/branch-switch). `sync_runs.cursor_json` records the last-scanned tip set for observability only. Payload: sha, author/committer (name+email), committer date, message subject, files touched (numstat), cited `VD-` ids parsed from the message.

**Witness emission (kernel/scripts/check-witness.sh):** plain bash wrapping `make check` — MUST work when the kernel binary doesn't exist yet (degrades to file write only; it never invokes `vera`). Mints `run_id` (ulid format via `od`/`uuidgen` fallback is fine — it only needs uniqueness + lexical sortability; document the choice in the script header), runs `make check`, writes ONE JSON file to `.vera/spool/`. **Schema (v1, single home hereafter = internal/connector/checks/SPEC.md):**

```json
{ "schema": "vera.witness.v1", "run_id": "<ulid>", "command": "make check",
  "exit_code": 0, "started_at": "<iso>", "finished_at": "<iso>", "duration_ms": 12345,
  "output_sha256": "<sha of full combined output>", "git_sha": "<HEAD>", "git_dirty": true,
  "tool_versions": { "go": "…", "golangci_lint": "…", "make": "…" } }
```

New make target `check-witnessed` calls the script; **plain `make check` stays unchanged** (the gate must never depend on the product it gates). Spool lifecycle: the connector **ingests, never deletes** (marks by ledger presence). Recovery scope stated honestly: git and session JSONL are re-syncable sources; spool files live in gitignored `.vera/` — check-run history is lost if `.vera/` is wiped before ingest. Accepted for P1 (witnesses are re-creatable evidence, not the gate); revisit at P2 when gates read historical check events.

**connector/sessions (BEST-EFFORT, OOD-tolerant):** source dir = `~/.claude/projects/<encoded>/` where `<encoded>` = the absolute repo path with every `/` and `.` replaced by `-` — **compute from the repo root at runtime, never hardcode**; the dir is created by the first Claude session run from the repository root, so early syncs may find little. Metadata ONLY (session id, start/end, message counts, tool-call counts, files-written count) — never content. Skip-and-count unparseable lines; record `parse_coverage` on each event; coverage < 50% → emit nothing, log. **Quiescence rule: only ingest JSONL files whose mtime is > 10 minutes old** — live sessions are invisible to sync, which is what makes `vera verify`'s double-sync deterministic (the executing session's own file is excluded).

**Projections:** `commits_view`, `checks_view`, `sessions_view` + `week_report` query. Reducers are pure functions over seq-ordered events. **Projection tables use natural keys only (commit sha, run_id, session id) — no serials, no wall-clock columns.** **row-set identical :=** per projection table, the multiset of full rows (ordered by natural key, all columns, compared as canonical-JSON row digests) is equal between incremental state and a from-genesis rebuild.

**CLI (`vera`):** `cmd/vera` with stdlib `flag` + subcommand switch. **Canonical invocation: `cd kernel && go run ./cmd/vera <args>`** (optional `make vera ARGS=…` passthrough may be added — with a gates.md note if it gains checks). All DoD commands below mean that invocation.
- `vera sync [git|checks|sessions|all]` · `vera rebuild` · `vera verify` · `vera report week`
- `vera verify` executes the phase DoD as code: (a) sync twice → 0 new events on pass two (holds under the sessions quiescence rule); (b) snapshot projections → rebuild → row-set-identical diff; (c) ≥1 witness event for the latest `make check-witnessed` run.
- `vera report week`: commits (count, files, VD- ids cited), check runs (pass/fail, durations), sessions (count, tool calls) — every line carries its event ids (the proof). Commits no longer reachable from any ref are still reported, marked `[superseded]` — the recorder keeps history git rewrote.

## Task sequence (one commit per task; wrap discipline applies)

| # | Task | Mechanical DoD |
|---|---|---|
| 0 | Preflight 0.1–0.6 | 0.1 settled (local-only, bundles); toolchain versions journaled; ROADMAP reconciled; expiry defused. Task 0 needs NO human input |
| 1 | Go module scaffold | `make check` green WITH kernel wired (go.mod, .golangci.yml, one placeholder test per package); `.vera/` added to .gitignore **and** to link-lint's --exclude-dir list (scaffolding edit — note in commit msg); gates.md row for the kernel build+test+lint check |
| 2 | `/vera-spec` core → implement | JCS vector tests pass; **this SPEC pins the machine-parseable invariant-table format** (`\| INV-<n> \| <statement> \| <test file>::<TestName> \|`) and updates `.claude/commands/vera-spec.md` § 5 to require it (same commit) |
| 3 | `/vera-spec` store → implement | **re-confirm the DB/locking section of this plan against embedded-postgres behavior in the spec**; idempotent-append test (dup → inserted=false, 1 row); no update/delete API surface test; goose migrations apply from empty (fresh temp dir); **lock test: second process attempting any DB command exits non-zero without touching the data dir**; gates.md "Known accepted bypasses" row for migration SQL |
| 4 | git connector | fixture repo (built in `t.TempDir()` at test runtime, never committed): second sync → 0 events; **amend, rebase of N commits, second-branch commit + switch — each followed by a sync appending only the new shas and a second sync appending 0** |
| 5 | witness emitter + checks connector | `make check-witnessed` writes valid v1 spool JSON with the kernel binary absent (test in a tree where `go` build outputs are unavailable / script never calls vera); `vera sync checks` ingests + dedupes; `/vera-wrap` step 3 amended to `make check-witnessed`; gates.md row for witness emission |
| 6 | projections + `vera verify` | `vera verify` passes end-to-end on this repo; rebuild==incremental row-digest test; gates.md row for `vera verify` (BLOCKING once wired into a make target) |
| 7 | sessions connector (best-effort) | parse_coverage recorded; corrupted-fixture test (synthetic JSONL only — NEVER real session files; any decision-id-shaped fixture string is runtime-constructed or uses `VD-<slug>-????`); droppable ONLY with a same-commit ROADMAP § P1 amendment + journal entry — never silently |
| 8 | `vera report week` | report renders on real repo data; every line carries event ids; `[superseded]` marking works (fixture) |
| 9 | P1 close | full `vera verify` run journaled; **spec-first graduation**: gates.md row ADVISORY→BLOCKING with enforcement = the named Go test (runs under `make check`) that asserts every `internal/*` package has SPEC.md whose invariant table (pinned format from Task 2) names ≥1 existing test — expiry cell cleared; journal/week_report boundary: journal stays the authoritative factual record through P1 — at close, either mint the VD-no-graph-2aa4vz revisit (retiring the journal's factual half; amend /vera-wrap) or record in state.md that the trigger is not yet met; state.md rewritten; ROADMAP P1 marked done (before 2026-09-18, see 0.6) |

## Execution notes for the fresh Opus session

- **Per-task loop:** `/vera-spec` (new packages) → implement → `/vera-review` → `make check` → commit → `/vera-wrap` at session end. One task per commit.
- **Multi-agent fan-out that saves tokens:** parallelize only *within* a task (spec-writer + test-writer from the spec, then implementer, then adversarial verifier). Tasks 2→3 are strictly sequential (store depends on core). Tasks 4/5/7 may run as parallel parallel worktree sessions AFTER Task 3 — but `notes/` stays owned by the main session (CLAUDE.md parallel rules).
- **Context economy:** each subagent gets the package SPEC.md + this plan's relevant section only — never conversation history or the vision docs.
- **When blocked:** mint mechanical decisions via `/vera-decide`; ask the human only for genuine value tradeoffs. Any dependency beyond VD-stack-go-fid9mi's set → `/vera-decide` BEFORE `go get`.
- **Fixture hygiene:** committed fixtures are synthetic only; fixture git repos are built in `t.TempDir()` at test runtime (exempt from ref-lint); no real session JSONL ever committed.

## Risks and pre-agreed mitigations

| Risk | Mitigation |
|---|---|
| embedded-postgres startup cost (~1-2s per command) | **THIS CELL IS THE SINGLE HOME FOR THE SUITE-DURATION FIGURE.** Current, store suite: **~40-52s idle** (three runs 2026-08-11: 40.8 / 40.3 / 52.4). The previously recorded 86.5s was measured while three reviewer agents were running the same suite against the same postgres binaries — i.e. under load, not at rest, and stating it as *the* figure overstated the problem by ~2x. **Read this as a range, not a number:** the package runs ~10 embedded postmasters in parallel, so the wall-clock is dominated by contention and is genuinely load-dependent; a single measurement is not a property of the suite. VD-stack-go-fid9mi's ~30s revisit trigger is fired either way. Optimisations already MEASURED (not assumed): a count instead of a quadratic row-materialisation (helped); a dedicated cluster for the new INV-48 tests (made it WORSE — two more postmasters); trial counts trimmed to the minimum that still kills the mutants (helped). **Owed at Task 6:** `-short` gating the heaviest DB tests out of the inner loop, or a shared read-only fixture cluster for the read path — decided against a measurement taken at rest AND under load, since the two disagree by 2x. |
| single-data-dir corruption (two embedded PGs) | every-command lock + pid-liveness takeover + store-owned single pool (see Ledger rules) |
| Claude JSONL format shifts | metadata-only, skip-and-count, parse_coverage gate, quiescence rule; Task 7 droppable only with ROADMAP amendment |
| Witness wrapper breaks the gate | `make check` NEVER depends on kernel; `check-witnessed` is additive; kernel-absent test in Task 5 |
| Ledger loss during experiments | `.vera/` recoverable: git + sessions re-sync from sources; check history only since last ingest is lost (accepted, stated above) |
| Scope creep | Non-goals list; `/vera-review` checks against it |
| Expiring advisories mid-phase | defused in 0.6; Task 9 deadline noted |

**P1 is done when:** `vera verify` passes, `vera report week` tells the truth about this repo with event-id proof, `make check` is green with the kernel wired, every internal package has a spec whose invariants map to passing tests (mechanically checked from Task 9 on), and ROADMAP § P1 DoD items are each satisfied by a named executable command.

## Position — 2026-08-14 (validated against the repo, not memory)

| # | Task | Status |
|---|---|---|
| 0 | Preflight | ✓ DONE — 0.1 settled local-only (bundles); 0.6 satisfied early: the stop-check advisory GRADUATED 2026-08-08, so the 2026-09-15 bomb named above no longer exists |
| 1 | Go module scaffold | ✓ DONE |
| 2 | core (JCS, ULIDs, envelope) | ✓ DONE — accepted after ONE adversarial round |
| 3 | store (ledger, lock, migrations) | ✓ DONE — shipped after NINE rounds with an enumerated residual list (VD-fix-discipline-0e0tnz is that story); its `rows.Err()` survivor-candidate folds into the store mutation sweep below |
| 4 | git connector | ✓ DONE — accepted under Law 9 in current Round 4; frozen `d794ff7` received an independent ACCEPTABLE verdict with no findings; final author evidence was gitcmd 77/77 and combined 95/95 killed |
| 5 | witness emitter + checks connector | ◐ ROUND-2 REMEDIATED — spec-first implementation, real shell emission, DB-backed `vera sync checks` replay, connector 37/37 and CLI 17/17 killed; two `NEEDS_WORK` verdicts are committed and a Round 3 verdict is required. The prescribed `.claude/commands/vera-wrap.md` does not exist in this checkout, so its step 3 cannot be amended here and remains explicit migration debt |
| 6 | projections + `vera verify` | ○ NOT STARTED (placeholder). The `-short`/fixture-cluster decision owed here per the Risks table stands |
| 7 | sessions connector | ○ NOT STARTED (placeholder) |
| 8 | `vera report week` | ○ NOT STARTED |
| 9 | P1 close | ○ NOT STARTED — hard-linked to the spec-first advisory expiry **2026-09-18** (~5 weeks out; Task 4 must close this week for the chain to hold) |

**Acceptance bar, amended (additive to every task's DoD from Task 4 on, per Law 9 and the measured
verification asymmetry):** a package is accepted only when (a) `make mutants` is GREEN for it — every
survivor killed or declared with a reason in `docs/allowed-survivors.txt` — and (b) an adversarial
round returns a NON-AUTHOR verdict of ACCEPTABLE, committed on receipt under
`docs/verification/verdicts/`. Green-plus-self-swept predicts nothing: three rounds each found 12–21
real defects in exactly that state. Task 4 is the first package walking this bar end-to-end;
`git`, `core` and `store` owe retroactive mutation sweeps (store's sweep inherits the `rows.Err()`
survivor-candidate its SPEC records).

**Owed-mechanism register (validated 2026-08-14; live queue and ordering live in `notes/state.md`):**
the comment-claims detector (`cannot/never/always` in code comments — the claim-outruns-evidence
class is over Law 7's threshold under the honest taxonomy); a staged-content mode for
`cleanroom-lint` (it scans tracked content, so it caught the verdict leak one commit late); the
round-3 gate-hygiene residues (bound `state-freshness`'s dirty-note escape, `prescription-lint`
wrong-root guard, `spec-numbering-lint` § 4-table scope). Validated STALE and closed: the
"no gate audited for scope blindness" debt (superseded by round 3's gate-audit table);
`figure-provenance`'s narrow scope (by design per its registry row); `index-check` self-test and
the lesson-tag registry (both shipped).
