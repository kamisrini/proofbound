# Proofbound — start here (current release: v0.3.2)

Proofbound is a warranty layer for code written by AI agents. For each change an agent makes, it
shows what the agent was asked to do, which lines it changed outside that, whether its tests notice
when its code is undone, which agent session wrote each line, and whether someone other than the agent
checked every promise. You get a signed warranty file that anyone can verify against a clone of your
repository. Proofbound does not make the model safe: it makes the agent's work visible and checkable,
with checks that hold even when the agent is wrong. It works with **OpenSpec** and **GitHub Spec Kit**,
and runs entirely on your machine and in your CI.

**New here, using agents?** Install (section 1), then follow the
[agent demo](DEMO-AGENT.md) (about 15 minutes, every step typed by you).

This page is the one place to start. Older releases' pages are linked at the bottom as history; you do
not need them to use Proofbound.

## 1. Install (Linux or macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/kamisrini/proofbound/main/install.sh | sh
proofbound version   # proofbound v0.3.2
```

The script checks the archive against `SHA256SUMS` and never uses `sudo`. Windows: download
`proofbound_windows_amd64.tar.gz` from the
[v0.3.2 release](https://github.com/kamisrini/proofbound/releases/tag/v0.3.2).

## 2. Use it on your repository

Follow the [quickstart](../mvp/QUICKSTART.md): `proofbound init`, `proofbound task start <change>` so
your agent's commits claim the spec automatically, have someone else review it, and get your first PASS
and warranty. It also sets up the pull-request check.

## 3. What you can do

| Goal | Command | Details |
|---|---|---|
| Set up a repository for OpenSpec or Spec Kit (also writes the GitHub workflow) | `proofbound init openspec` or `init speckit` | [quickstart](../mvp/QUICKSTART.md) |
| Tell the hooks which spec an agent is working on, so its code commits claim it | `proofbound task start <change>`, `task stop`, `task` | [tour, chapter 13](DEMO.md#13-agents-the-claim-is-added-for-them-and-unclaimed-code-fails-the-check) |
| Declare which files a change may touch; every line outside is named and needs a signed decision (C10) | `scope.paths` in the spec folder | [agent demo, step 3](DEMO-AGENT.md#step-3--the-spec-and-the-area-it-allows) |
| Stop a Claude Code agent's edit outside that area as it is made, and log which session wrote which lines | the hooks `init` writes to `.claude/settings.json` (`hook pre-tool-use`, `hook post-tool-use`) | [agent demo, steps 4, 5 and 8](DEMO-AGENT.md#step-4--the-agent-starts-its-task) |
| Show whether your tests notice when each part of a change is undone (C11) | `.proofbound/tests.json`, then `review draft` | [agent demo, step 8](DEMO-AGENT.md#step-8--the-review-draft-a-line-outside-the-area-and-a-test-that-notices-nothing) |
| See which agent sessions wrote a change's lines, and which lines no recorded agent edit wrote (C12) | `review draft`, `review seal` | [agent demo, step 8](DEMO-AGENT.md#step-8--the-review-draft-a-line-outside-the-area-and-a-test-that-notices-nothing) |
| See how each agent's changes fared in review: outcomes, out-of-area hunks, uncaught parts, line provenance | `proofbound report agents <base>` (or the Audit room) | [agent demo, step 12](DEMO-AGENT.md#step-12--the-audit-room-what-an-auditor-sees) |
| See what each control gives evidence toward in SOC 2, ISO 27001, NIST SSDF and the EU AI Act | the Audit room, or [FRAMEWORK-MAP.md](../../compliance/FRAMEWORK-MAP.md) | [release notes](RELEASE-NOTES.md) |
| Record history: commits, specs, reviews | `proofbound sync all` | [quickstart](../mvp/QUICKSTART.md) |
| Review a commit against its spec, signed by the reviewer's key | `proofbound review draft <commit>`, then `review seal` | [quickstart](../mvp/QUICKSTART.md) |
| Check a pull request: is every spec-claiming commit verified, and does every code commit claim a spec? | `proofbound check <base>`, or the GitHub Action | [quickstart](../mvp/QUICKSTART.md) |
| Export a signed warranty, and verify one with no database | `proofbound warranty export`, `warranty verify` | [tour](DEMO.md) |
| Add a reviewer's signature to an existing warranty | `proofbound warranty cosign` | [release notes](RELEASE-NOTES.md) |
| Attach proof to a release: signed files, a verification page, an index, a badge | `proofbound warranty bundle <base>` | [release notes](RELEASE-NOTES.md) |
| See which verified specs have since changed | `proofbound report drift` | [R2 demo](../r2/DEMO.md) |
| Audit, review and track promises in a browser: Audit room, Now, Review bench, Promise map, Agent surface | `proofbound ui` | [manual, step 15](DEMO-MANUAL.md#step-15--the-audit-room-audit-the-release-in-a-browser-and-close-the-exception) |
| Make your declared key sign your git commits, so the Audit room proves who made each one | `proofbound keys git-sign` | [release notes](RELEASE-NOTES.md) |

## 4. See it work (demos)

Each script builds a throwaway project in a temporary directory and prints what it shows. They need git,
jq, Go to build the CLI (the loop demo accepts `PROOFBOUND_BIN=<installed binary>` instead), and port 55432 free for the embedded database.

| Demo | Shows | Run |
|---|---|---|
| **The agent goes off-script (agents start here)** | An agent's change edits a file outside its spec's area and ships a test that cannot fail: the Claude Code hook refuses the edit, the shell edit that gets around it is named as written by no recorded agent session, Proofbound shows the test notices nothing when each part of the change is undone, the reviewer's findings are signed, nothing is warranted until the redo, which verifies VALID from a clone. You type every step, in [DEMO-AGENT.md](DEMO-AGENT.md) | follow [DEMO-AGENT.md](DEMO-AGENT.md) (about 15 minutes), or `bash docs/releases/r3/agent-demo.sh` |
| **The full tour (start here)** | Every feature in order on one Spec Kit project, explained chapter by chapter with why each matters, in [DEMO.md](DEMO.md) | `bash docs/releases/r3/tour.sh` |
| **By hand, step by step** | The same tour, but you type every command, with what each step does and its business value, in [DEMO-MANUAL.md](DEMO-MANUAL.md) | follow [DEMO-MANUAL.md](DEMO-MANUAL.md) (about 20–30 minutes) |
| Install like a stranger | Builds the release archives, installs from them with `install.sh`, and times install → first gate PASS (Linux or WSL) | `bash docs/releases/mvp/demo-install.sh` |
| The warranty loop | Spec → claiming commit → independent review → gate PASS → drift when the spec changes (OpenSpec or Spec Kit) | `bash docs/releases/r2/demo-loop.sh speckit` |
| Portable signed warranty | Export a warranty, verify it on a fresh clone, UNREVIEWED until the reviewer signs, then VALID | `bash docs/releases/r3/demo-warranty.sh speckit` |
| Shareable proof for a release | `warranty bundle`: signed files, `VERIFY.md`, `index.json`, badge, checked from a fresh clone | `bash docs/releases/mvp/demo-share.sh` |
| A real OpenSpec repository at scale | Proofbound run over OpenSpec's own history (clones it from GitHub) | `bash docs/releases/r2/demo-openspec.sh` |

Walkthroughs: [the full tour](DEMO.md), and the [R2 demo](../r2/DEMO.md) for the real OpenSpec repository.

## 5. What changed, and known limits

- [v0.3 release notes](RELEASE-NOTES.md): v0.3.0, the v0.3.1 fix (the pull-request check passes in
  CI) and v0.3.2 (the agent features: line-level scope, the write-time perimeter, tests that can fail,
  session provenance, the Audit room and framework mapping; unclaimed code fails the check), behaviour
  changes, and caveats. Read the caveats before relying on a warranty.

## History: earlier releases

These pages describe earlier releases. Everything they shipped is still in v0.3.2.

| Release | What it added | Pages |
|---|---|---|
| Release 2, the SDD Warranty Connector (accepted 2026-11-19) | Bind each commit to the exact OpenSpec or Spec Kit revision it claims; independent review per obligation; drift | [release notes](../r2/RELEASE-NOTES.md), [demo](../r2/DEMO.md) |
| Release 1, the Trust Record | The permanent, reproducible record: commits, decisions, reviews, CI witnesses, agent sessions, gates | [release notes](../r1/RELEASE-NOTES.md), [user guide](../r1/USER-GUIDE.md), [demo hub](../r1/DEMO.md), [full tour](../r1/FULL-TOUR.md) |
