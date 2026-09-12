# P6 debt-census SPEC

## Purpose and authority

This specification defines Task 0 of the ratified
[`P6-CONSOLIDATION-PLAN-draft2.md`](P6-CONSOLIDATION-PLAN-draft2.md). The census is a closed,
mechanically checked work queue for deep completion of P1–P5. It does not infer promises from
unbounded prose and it does not claim that a passing probe proves more than the row states.

The canonical result is `docs/plans/p6-census.md`. Stable row definitions live in
`docs/plans/p6-census-rows.tsv`; that registry is hand-classified input, not derived state. The
result is generated only by `scripts/p6-census.sh`. A newly discovered subject stops generation
until it receives a never-before-used row ID and disposition in the registry.

## Command contract

```text
scripts/p6-census.sh --write [--root PATH]
scripts/p6-census.sh --check [--root PATH]
scripts/p6-census.sh --render [--root PATH]
```

- `--write` scans, validates, executes probes, and atomically replaces the canonical result.
- `--check` performs the same work in a temporary file and exits nonzero unless it is byte-equal to
  the committed result.
- `--render` writes the candidate result to stdout without changing the repository.
- Unknown options, missing inputs, duplicate rows, unknown scan subjects, invalid enum values,
  failing closed-row probes, or malformed evidence fail closed with a diagnostic.
- Output ordering is category number then numeric row ordinal. Locale is fixed to `C`; timestamps,
  local absolute paths, and other nondeterministic values are forbidden in generated output.

The script accepts only named probe forms defined below. It never evaluates registry text as shell
source. That keeps a census row from becoming an arbitrary-code surface.

## Closed scan universe

The scanner reads exactly these live-authority classes:

1. `CLAUDE.md`, `README.md`, `ROADMAP.md`, `notes/state.md`, and `docs/gates.md`;
2. accepted execution plans listed in the registry preamble;
3. direct `docs/decisions/VD-*.md` files whose metadata says `Status: Accepted`;
4. every tracked `kernel/internal/**/SPEC.md`;
5. actual Make targets from `make -qp`, tracked paths from `git ls-files`, gate definitions from
   `gates/*.yaml`, packages from `go list ./...`, production packages from `go list ./internal/...`
   that contain a non-test Go file, and source/kind pairs from the core registry proving helper;
6. acceptance artifacts under `docs/verification/`, direct verdict/ratification artifacts under
   `docs/verification/verdicts/`, and commits reachable from `HEAD`.

Historical adjudications, verdict exhibits, journal entries, superseded drafts, and frozen payload
vectors are inventoried for reference integrity but are not current-truth authorities. The accepted
P1 and P5 plans contain labeled historical baselines; only their explicit current status, position,
DoD, owed-mechanism, and standing-rule sections are live claims. P6 draft 2 is live in full through
the semantic VD that adopts its exact digest.

Recognized live mechanism tokens are:

- a backtick-quoted `make <target>` where `<target>` is `[A-Za-z0-9][A-Za-z0-9_-]*`;
- a backtick-quoted path rooted at `scripts/`, `kernel/scripts/`, `.claude/hooks/`,
  `.github/workflows/`, `gates/`, or `kernel/internal/`;
- a gate ID parsed from tracked `gates/*.yaml`;
- a `file_test.go::TestName` citation in a package SPEC; and
- public targets parsed from the Makefile plus production packages discovered as described above.

Markdown link destinations, glob examples, placeholders containing `<`, `>`, `*`, or `?`, and
tokens inside fenced examples explicitly labeled historical are durable exclusions. Every other
recognized token is either mapped to a registry row or reported as an unknown subject. There are no
silent regular-expression exceptions.

## Stable registry contract

`docs/plans/p6-census-rows.tsv` is UTF-8, LF-terminated, and contains one header followed by rows
with these tab-separated fields:

```text
row_id category subject observation probe disposition decision evidence
```

Tabs and newlines are forbidden inside fields. Markdown table delimiters are forbidden because the
registry is rendered into a pipe table. `row_id` is `C<number>-<three-digit ordinal>`, its prefix
must match `category`, and an ID is never reused. Deleted discoveries remain as closed historical
rows or are retired through a decision; rows do not disappear. The script rejects duplicate IDs and
duplicate `(category, subject)` pairs.

`category` is exactly C1 through C8. `disposition` is exactly `close-in-P6`, `defer-to-P7`, or
`wontfix`. `decision` is `—` for `close-in-P6`; either other disposition requires one exact accepted
direct path under `docs/decisions/`, and the decision must name the row subject or its governing
live promise. `evidence` is `—` while a row's probe is open. A registry evidence value is allowed
only when every listed path, commit, command-result record, event ID, or named test resolves.

The first registry preamble also names the accepted-plan paths. This is the only manual phase-plan
allowlist. Adding or removing one is a reviewable census-scope change.

## Probe language and state derivation

The rendered `probe` cell is an executable command or a named Go/shell test. The registry stores a
closed probe identifier and arguments; the script renders and executes it without `eval`:

| Probe form | Pass condition |
|---|---|
| `path:<tracked-path>` | exact path is tracked and exists |
| `make-target:<name>` | exact public target exists in `make -qp` output |
| `make-edge:<parent>:<child>` | child is a direct or transitive prerequisite of parent |
| `shell-test:<path>` | tracked test exists and exits zero |
| `go-test:<package>:<test>` | exact named test exists and `go test -run '^<test>$'` exits zero |
| `artifact:<tracked-path>` | artifact exists, is tracked, and all exact paths/commits it declares resolve |
| `acceptance:<package>` | current tree object, calibrated mutation result, and non-author verdict all bind |
| `connector:<name>` | connector has the live/synthetic/none classification evidence required by C4 |
| `route:<source>:<kind>:<consumer>` | generated route matrix has one consume/ignore/reject cell and a proving test |
| `intent-history:<anchor>` | every post-anchor commit is classified by the ratified path rule and evidenced |
| `p5-result:<name>` | named P5 measurement/falsifier has one allowed closed value and typed provenance |
| `fresh-clone:<platform>` | committed run record binds a clean clone, commit, platform, and required commands |
| `decision-skip:<VD-path>:<subject>` | accepted VD explicitly skips/supersedes the named subject |

A passing probe renders `state=closed`; a failing probe renders `state=open`. A closed row requires
non-`—` evidence. An open row requires `evidence=—`; stale claimed evidence is an error rather than
being silently dropped. Deferral and wontfix probes can close only through `decision-skip`, and that
decision must update all live authorities carrying the promise.

Some acceptance probes are intentionally open during Task 0. In particular, C3 records all current
production packages but cannot close until Task 8 freezes the final implementation commit. C5 route
cells, C6 coverage/review rows, C7 result rows, and C8 platform/clean-clone rows remain open until
their owning task creates the required artifacts. Their presence and classification—not premature
green—is Task 0's success condition.

## Category-specific completeness

- **C1:** every recognized live token and every public Make target/gate documentation obligation is
  either a row or a proven existing surface; phase-status assertions are explicit rows.
- **C2:** every `docs/gates.md` row records existence, self-test, Make-DAG boundary, mode, owner, and
  expiry; all named draft-2 baseline gaps and the legacy alias advisory are rows.
- **C3:** every production `internal/...` package has one row; test-only packages and command
  wrappers have explicit exclusion rows and never vanish from the count.
- **C4:** every production connector package has one classification row and separate rows for its
  accepted narrowing/non-goals. Sessions and GitHub retain the numeric checks in the ratified plan.
- **C5:** the Cartesian product of every registered source/kind pair and projection apply/rebuild,
  CLI verify, twin replay, and gate evaluation has one generated route subject.
- **C6:** every post-`f426ca8` commit and every current active requirement, obligation, exact-revision
  review, and self-hosted delivery chain is represented.
- **C7:** all nine P5 measurements and seven P5 falsifiers are named individually.
- **C8:** every claimed platform, fresh-clone workflow, verdict/ratification digest, referenced
  commit, acceptance-artifact path, and the P6 motivating vision assessment is represented.

## Canonical result schema

`docs/plans/p6-census.md` starts with the Law 1 generated marker and exact regeneration command,
then generation-input commit and count summaries, followed by exactly one Markdown table:

```text
row_id | category | subject | observation | probe | state | disposition | decision | evidence
```

No other pipe table may appear. Counts include total/open/closed and C1–C8 totals. A category with
zero rows is a hard failure. Zero unclassified scan subjects is printed explicitly.

## Invariants and proving tests

| Invariant | Statement | Proving test |
|---|---|---|
| P6C-INV-1 | Unknown scan subjects and incomplete C1–C8 discovery fail closed | scripts/tests/p6-census.test.sh::unknown-and-missing-universe |
| P6C-INV-2 | Row IDs and category/subject identities are unique, stable, and enum-closed | scripts/tests/p6-census.test.sh::registry-schema |
| P6C-INV-3 | Closed rows require passing named probes and resolving evidence | scripts/tests/p6-census.test.sh::closed-evidence |
| P6C-INV-4 | Deferral or wontfix closes only through an accepted promise-changing decision | scripts/tests/p6-census.test.sh::decision-closure |
| P6C-INV-5 | `--check` detects any byte drift from deterministic rendered output | scripts/tests/p6-census.test.sh::generated-freshness |
| P6C-INV-6 | The first result contains C1–C8, unique rows, valid dispositions, and no unclassified subject | scripts/tests/p6-census.test.sh::task-zero-dod |

The shell test uses isolated temporary Git repositories and fixture registries. It must prove one
valid pass and one failure for each invariant. Production probes may be replaced only through an
explicit fixture root interface; tests never weaken validation with a skip flag.

## Non-goals

- The census does not accept P1–P5, interpret unrestricted prose, or make independent judgments.
- Task 0 does not close debt merely because it was discovered.
- The mechanism does not add a provider, event kind, UI, behavior lock, governor, warranty, or P7+
  capability.
- The census does not rename frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, or any
  pinned vector.
