# Proofbound — Demo Hub (Release 2)

**Release 2 is the SDD Warranty Connector.** Proofbound reads the specs a spec-driven-development
tool writes ([OpenSpec](https://github.com/Fission-AI/OpenSpec) or
[Spec Kit](https://github.com/github/spec-kit)), binds each commit to the exact spec revision it
claims to implement, says whether anyone other than the author has verified it, and flags a
warranty when the spec changes afterwards. It works beside any coding agent, because it reads what
the agent and the spec tool leave in git.

| I want to… | Run | Time |
|---|---|---|
| **See the whole loop on OpenSpec**: claim → verify → PASS → drift | `bash docs/releases/r2/demo-loop.sh openspec` | ~2 min |
| **See the whole loop on Spec Kit** | `bash docs/releases/r2/demo-loop.sh speckit` | ~2 min |
| **See it at scale on a real project**: OpenSpec's own repository, 1,059 commits | `bash docs/releases/r2/demo-openspec.sh` | ~4 min |
| Use it on my own project | [§4](#4-use-it-on-your-own-openspec-or-spec-kit-project) | ~10 min |
| What Release 2 includes, and its caveats | [Release Notes](RELEASE-NOTES.md) | — |
| Release 1 (the trust record underneath) | [Release 1 Demo Hub](../r1/DEMO.md) | ~5 min |

## 1. Setup (once)

- A Linux or WSL terminal, Git, Go 1.26+, and `jq`. On Windows, follow
  [Release 1, §0 One-time setup](../r1/DEMO.md#0-one-time-setup) first: use WSL from your Linux home,
  not a `/mnt/c` path.
- Run the demos from a clone of this repository. Each demo builds its own copy of the CLI, so
  nothing needs installing first.
- The embedded database uses port **55432**; it must be free (no other Proofbound command running).
- `demo-openspec.sh` also needs network access to GitHub. The loop demos fetch nothing.

## 2. The whole loop, on a tiny project (OpenSpec or Spec Kit)

```bash
bash docs/releases/r2/demo-loop.sh openspec    # works in ${TMPDIR:-/tmp}/proofbound-openspec-loop
bash docs/releases/r2/demo-loop.sh speckit     # works in ${TMPDIR:-/tmp}/proofbound-speckit-loop
```

Both create a project, run `proofbound init <kit>`, and then:

1. **Write a spec** with two obligations: an OpenSpec change `add-csv-export`, or a Spec Kit feature
   `001-csv-export` (`- **FR-001**: …`, `- **FR-002**: …`).
2. **Implement it, with a test,** in a commit that ends `Intent: sdd:<name>`.
3. **Sync and report:** the claim binds to the spec's exact revision, and every obligation is
   `UNVERIFIED`.
4. **Review:** the demo runs the feature's test, records each outcome, and runs
   `proofbound review seal`. The demo plays the reviewer and says so. In real use, a person or the
   `proofbound-reviewer` agent who did not write the commit decides.
5. **The gate says PASS**, with the counts.
6. **The spec owner rewords one obligation.** `report drift` flags exactly that one.

What you see at the end (Spec Kit; OpenSpec's is the same, with `O-ADD-HEADER-ROW`):

```
gate=ready-<commit12> state=PASS ... reason="2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE"
drift change=CI-001-csv-export-... state=CURRENT ...
...
drift change=CI-001-csv-export-... state=STALE ...
  changed O-FR-002 "System MUST start every export with a header row." -> "System MUST start every export with a header row naming each column in lower case."
summary warranted=1 current=0 stale=1 other=0
```

## 3. At scale: OpenSpec's own repository

```bash
bash docs/releases/r2/demo-openspec.sh            # works in ${TMPDIR:-/tmp}/proofbound-openspec-demo
bash docs/releases/r2/demo-openspec.sh ~/pb-demo  # or a directory of your choice
```

1. **Clones [OpenSpec](https://github.com/Fission-AI/OpenSpec)** at a pinned commit. OpenSpec is
   developed with OpenSpec, so its history holds real change proposals and specs.
2. **Commits a Proofbound profile into every commit of the local clone.** Release 2 takes owners
   only from a committed profile, so the demo puts it into history. Commit IDs in the clone differ
   from upstream as a result. Nothing is pushed.
3. **Reads the whole history:** about 650 spec revisions from about 1,060 commits, in about a minute.
4. **Adds two demo commits:** one claims a real active change (`Intent: sdd:install-skills-globally`),
   the other a change that does not exist. Each gets a readiness gate.

```
claim-skip unknown-change sdd:no-such-change at <commit>
gate=bogus state=BLOCKED ... reason="commit binds no change intent"
gate=claimed state=BLOCKED ... reason="0 of 4 obligations verified SATISFIED; 0 of 4 requirement reviews VERIFIABLE"
intent=CI-install-skills-globally-57r52f ... state=IMPLEMENTED_UNVERIFIED ...
  target=...#O-ADD-INSTALL-THE-RUNNING-C-13EA48 ... outcome=UNVERIFIED statement="Init SHALL install skills ..."
```

- **The claimed commit binds** to the exact revision of `install-skills-globally`. Its four
  obligations come word for word from OpenSpec's own spec.
- **Its gate is BLOCKED, and says why:** the demo commit implements nothing, so no one has verified
  anything, and this demo does not pretend otherwise. §2 shows the path to PASS.
- **The bogus claim is reported and BLOCKED.** A commit cannot claim a change that does not exist.

## 4. Use it on your own OpenSpec or Spec Kit project

**Build the CLI once** and put it on your PATH ([Release 1, §1](../r1/DEMO.md#1-build-the-proofbound-binary)):

```bash
cd <this-repo>/kernel && CGO_ENABLED=0 go build -o "$HOME/bin/proofbound" ./cmd/proofbound
```

**Set up your project** (in its repository, once):

```bash
proofbound init openspec            # or: proofbound init speckit
git add .proofbound/sdd-profile.json .gitignore AGENTS.md .claude
PROOFBOUND_NO_INTENT=1 git commit -m "add proofbound"   # implements no spec, so no trailer
```

`init` writes the profile (owner and sponsor = your `git config user.name`, or `--owner <name>`), two
`.gitignore` lines, an AGENTS.md block that tells any agent the commit rule, a Claude Code reviewer
subagent (`.claude/agents/proofbound-reviewer.md`), and a `commit-msg` hook. It never overwrites
anything, so running it twice is safe.

**Work as usual.** Sync specs with `proofbound sync intent sdd`: `sync all` deliberately leaves them out
(only a named sync reads them), so run both. End every commit that implements a spec with `Intent: sdd:<name>`, where
`<name>` is the folder: the OpenSpec change under `openspec/changes/`, or the Spec Kit feature under
`specs/` (for example `001-export`). Spec Kit counts a feature once its `tasks.md` exists. Commits
that implement no spec use `PROOFBOUND_NO_INTENT=1 git commit …`.

**Warrant a commit.** Someone who did not write it (another person, or the `proofbound-reviewer`
subagent in a separate session) runs:

```bash
proofbound sync intent sdd && proofbound sync all
proofbound review draft <commit>                # prints every obligation; writes .proofbound/drafts/<commit>/draft.json
# in draft.json: set declared_reviewer; set each outcome (SATISFIED / NOT_SATISFIED / INCONCLUSIVE,
# and VERIFIABLE / AMBIGUOUS / UNTESTABLE / CONTRADICTORY for the requirement review)
proofbound review seal .proofbound/drafts/<commit>   # the folder, not draft.json
git add docs/verification/verdicts gates
PROOFBOUND_NO_INTENT=1 git commit -m "review <commit>"
proofbound sync all && proofbound gates canary   # reviews only need sync all
```

`seal` takes only the reviewer's decisions from the draft and rebuilds everything else from the
ledger. It refuses the spec's owner as reviewer, an undecided outcome, or a draft that no longer
matches. The commit's `ready-<commit12>` gate says **PASS** only when every obligation is SATISFIED
and every requirement VERIFIABLE; otherwise BLOCKED, with the counts.

**Watch for drift** whenever specs change:

```bash
proofbound sync intent sdd && proofbound sync all && proofbound report drift
```

Each warranted change is CURRENT, or STALE when it now promises something other than what was
verified, with each obligation `added`, `removed` or `changed` (old and new wording). It compares
what is promised, not files, so archiving a change or editing `tasks.md` is not drift. REMOVED,
UNSYNCED and UNKNOWN cover a change that is gone at HEAD, not yet synced, or unreadable.

**Useful lookups:** `proofbound report intent sdd:<name>` shows one change's obligations and state.

## 5. Command reference (Release 2)

| Command | What it does |
|---|---|
| `proofbound init {openspec\|speckit} [--owner <name>]` | Set up a project (never overwrites) |
| `proofbound sync intent sdd` | Record every spec revision in the project's history |
| `proofbound sync git` | Record commits; bind `Intent: sdd:<name>` claims |
| `proofbound sync all` | Every other sync (git, checks, sessions, reviews, records/specdir intents). It does **not** read specs: run `sync intent sdd` first whenever specs changed |
| `proofbound report intent sdd:<name>` | One change: obligations, verification state, bound commits |
| `proofbound review draft <commit>` / `review seal <folder>` | Record an independent review of a commit |
| `proofbound gates canary` | Every gate's state, with its reason |
| `proofbound report drift` | Which warranties no longer match their spec |

## 6. Cleaning up

```bash
rm -rf "${TMPDIR:-/tmp}"/proofbound-openspec-demo "${TMPDIR:-/tmp}"/proofbound-*-loop
```

Each demo's database lives inside its own directory; the database server's log is
`.proofbound/db.log` in the demo project.
