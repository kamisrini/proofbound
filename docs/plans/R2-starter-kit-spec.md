# R2 starter kit and reviewer workflow — spec

**Status:** owner-approved scope 2026-11-19 ("lets do it as part of r2, go ahead"). Law 11 budget:
this spec gets at most 2 review rounds; the code gets 1.

**Payoff (the one command the owner runs):** on a fresh personal OpenSpec project,
`proofbound init openspec` once; then work as usual with any agent; then the reviewer agent (or a
person) runs `proofbound review draft <commit>`, decides each outcome, runs
`proofbound review seal <draft-dir>`, commits; `proofbound sync all && proofbound gates canary`
shows the reviewed commit's gate **PASS** (or BLOCKED with its reason).

## 1. `proofbound init openspec`

Run in a git repository. Writes only files that do not exist (prints `exists <path>` and leaves an
existing file untouched, except the two marked appends below). Prints one line per action.

| Path | Content |
|---|---|
| `.proofbound/sdd-profile.json` | the OpenSpec reference profile (same as the demo's); owner and sponsor = `--owner <name>`, default `git config user.name`, else `project-owner` |
| `.gitignore` | appends `.proofbound/*` and `!.proofbound/sdd-profile.json` if either line is missing |
| `AGENTS.md` | appends a block between `<!-- proofbound:begin -->` and `<!-- proofbound:end -->` if absent: the trailer rule and the review workflow |
| `.claude/agents/proofbound-reviewer.md` | a Claude Code subagent that performs §3 (other harnesses read the AGENTS.md block) |
| `.git/hooks/commit-msg` | rejects a commit message with no `Intent:` trailer, unless `PROOFBOUND_NO_INTENT=1`; never replaces an existing hook |

## 2. `proofbound review draft <commit>` and `proofbound review seal <dir>`

`draft` needs the commit synced (`sync all`). For the commit's latest git event it reads every
bound `sdd` ref (provider `sdd`); for each it reads the ledgered change intent and every target
requirement revision. It writes one file, `.proofbound/drafts/<commit>/draft.json` (built
2026-11-19 as one file rather than one per artifact, which is simpler for an agent to edit):
`declared_reviewer: ""`; one review per target requirement listing every obligation of that revision
with its statement, the requirement's declared owner, and `outcome: ""`; one verdict per bound change
intent listing every target obligation with its statement, `outcome: ""`, and `evidence_event_ids`
pre-filled with the commit's git event id. It refuses to replace an existing draft.
It prints each obligation's statement. A commit with no bound sdd ref fails with a message.

`seal` validates and writes the real artifacts. For each draft it requires: `declared_reviewer`
non-blank and different from the requirement's declared owner; every outcome in the schema's
closed set; evidence non-empty for `SATISFIED`. It sets the verdict `status` (`ACCEPTABLE` iff
every outcome is `SATISFIED`, else `NEEDS_WORK`), drops the `_` helper fields, fixes ids
(`RR-<commit12>-<n>`, `OV-<commit12>-<n>`) and `artifact_path`
(`docs/verification/verdicts/<id>.md`), computes the self-digest, and re-parses each artifact with
the reviews connector's own parser before writing it. It also writes
`gates/ready-<commit12>.yaml` (readiness of exactly that commit; a HEAD-scoped gate would judge the
verdict commit instead). It never overwrites an existing artifact or gate. Any failure writes
nothing.

## 3. The reviewer workflow (AGENTS.md block and subagent)

1. `proofbound sync all`; `proofbound review draft <commit>`.
2. Read the change (`openspec/changes/<name>/`) and the commit (`git show <commit>`); run the tests.
3. Decide each outcome honestly. `SATISFIED` only with evidence the code does it; otherwise
   `NOT_SATISFIED` or `INCONCLUSIVE`. Review outcomes: `VERIFIABLE`, `AMBIGUOUS`, `UNTESTABLE`,
   `CONTRADICTORY`. Set `declared_reviewer` to a name that is not the owner's.
4. `proofbound review seal <dir>`; commit with `PROOFBOUND_NO_INTENT=1`; `proofbound sync all`;
   `proofbound gates canary`.

## 4. Invariants (CLI SPEC rows)

- **CLI-INV-18** `init openspec` never overwrites an existing file and is idempotent (a second run
  changes nothing).
- **CLI-INV-19** a sealed set re-parses with the reviews connector's parsers, and on a tenant that
  ran the workflow with all outcomes SATISFIED/VERIFIABLE by a non-owner, the reviewed commit's
  readiness gate is PASS; with any other outcome it is BLOCKED.
- **CLI-INV-20** `seal` writes nothing when any check fails (undecided outcome, blank or owner
  reviewer, SATISFIED without evidence, an existing target file).

## 5. Non-goals and stated limits

- Independence is declared, not authenticated (reviews SPEC): a solo developer's reviewer agent
  declares its own name. The warranty says *who claimed* to verify, not that they are independent.
- Evidence is the reviewed commit's git event only; check runs are a later addition.
- The hook is a convenience, not enforcement (`--no-verify` skips it).
- Spec Kit, drift (M4), and main specs are out of scope.
