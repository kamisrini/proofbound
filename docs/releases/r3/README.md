# Proofbound — start here (current release: v0.3.2)

Proofbound tells you, for each commit that claims to implement a spec, whether someone other than its
author has checked every promise in that spec. It gives you a signed warranty file that anyone can
verify against a clone of your repository. It works with **OpenSpec** and **GitHub Spec Kit**, and runs
entirely on your machine and in your CI.

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

Follow the [quickstart](../mvp/QUICKSTART.md): `proofbound init`, claim a spec in a commit, have someone
else review it, and get your first PASS and warranty. It also sets up the pull-request check.

## 3. What you can do

| Goal | Command | Details |
|---|---|---|
| Set up a repository for OpenSpec or Spec Kit (also writes the GitHub workflow) | `proofbound init openspec` or `init speckit` | [quickstart](../mvp/QUICKSTART.md) |
| Record history: commits, specs, reviews | `proofbound sync all` | [quickstart](../mvp/QUICKSTART.md) |
| Review a commit against its spec, signed by the reviewer's key | `proofbound review draft <commit>`, then `review seal` | [quickstart](../mvp/QUICKSTART.md) |
| Check a pull request: is every spec-claiming commit verified? | `proofbound check <base>`, or the GitHub Action | [quickstart](../mvp/QUICKSTART.md) |
| Export a signed warranty, and verify one with no database | `proofbound warranty export`, `warranty verify` | [tour](DEMO.md) |
| Add a reviewer's signature to an existing warranty | `proofbound warranty cosign` | [release notes](RELEASE-NOTES.md) |
| Attach proof to a release: signed files, a verification page, an index, a badge | `proofbound warranty bundle <base>` | [release notes](RELEASE-NOTES.md) |
| See which verified specs have since changed | `proofbound report drift` | [R2 demo](../r2/DEMO.md) |

## 4. See it work (demos)

Each script builds a throwaway project in a temporary directory and prints what it shows. They need git,
jq, Go to build the CLI (the loop demo accepts `PROOFBOUND_BIN=<installed binary>` instead), and port 55432 free for the embedded database.

| Demo | Shows | Run |
|---|---|---|
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
  CI) and v0.3.2 (unclaimed code fails the check; `proofbound task`), behaviour changes, and caveats. Read the caveats before relying on a warranty.

## History: earlier releases

These pages describe earlier releases. Everything they shipped is still in v0.3.2.

| Release | What it added | Pages |
|---|---|---|
| Release 2, the SDD Warranty Connector (accepted 2026-11-19) | Bind each commit to the exact OpenSpec or Spec Kit revision it claims; independent review per obligation; drift | [release notes](../r2/RELEASE-NOTES.md), [demo](../r2/DEMO.md) |
| Release 1, the Trust Record | The permanent, reproducible record: commits, decisions, reviews, CI witnesses, agent sessions, gates | [release notes](../r1/RELEASE-NOTES.md), [user guide](../r1/USER-GUIDE.md), [demo hub](../r1/DEMO.md), [full tour](../r1/FULL-TOUR.md) |
