# Proofbound quickstart: install and get your first warranty

Proofbound tells you, for each commit that claims to implement a spec, whether someone other than its
author has checked every promise in that spec — and gives you a signed file that proves it.

## 1. Install (Linux or macOS, about 10 seconds)

```bash
curl -fsSL https://raw.githubusercontent.com/kamisrini/proofbound/main/install.sh | sh
proofbound version
```

- It installs one binary into `~/.local/bin` (`PROOFBOUND_INSTALL_DIR` changes that) and never uses
  `sudo`. `PROOFBOUND_VERSION=v0.3.0` pins a release.
- It checks the download against the release's `SHA256SUMS` and installs nothing on a mismatch. That
  checksum comes from the same place as the download, so it proves the file is intact, not that the
  source is honest; signed releases are planned.
- **Windows:** download `proofbound_windows_amd64.tar.gz` from the
  [releases page](https://github.com/kamisrini/proofbound/releases) and run `tar -xzf` on it, or use
  WSL and the command above.

If an older `proofbound` comes earlier on your PATH, the installer warns and names it. To try the
install path from a clone without the network, `bash docs/releases/mvp/demo-install.sh` builds the
archives locally, installs from them, and runs the loop below with the installed binary.

## 2. The database (nothing to set up)

Proofbound keeps its record in PostgreSQL. With nothing configured, the first command starts a
private embedded PostgreSQL for the repository on port 55432, downloading a build of it (about 15 MB,
once per machine) from Maven Central. **Caveat:** that download's published signature did not
verify when we checked it (`DEF-P7-001`). If that matters to you, point Proofbound at your own
PostgreSQL instead:

```bash
export DATABASE_URL=postgres://user:pass@host:5432/proofbound
```

## 3. First warranty in your project

In a repository that uses [OpenSpec](https://github.com/Fission-AI/OpenSpec) or
[Spec Kit](https://github.com/github/spec-kit). Starting from an empty folder instead? Follow
[DEMO-MANUAL.md](../r3/DEMO-MANUAL.md) steps 0 to 3: it creates the repository and its first spec
by hand.

```bash
proofbound init openspec          # or: proofbound init speckit
git add .proofbound/sdd-profile.json .gitignore AGENTS.md .claude .github
git commit -m "add proofbound"
```

Before an agent writes code for a change, run `proofbound task start <change-or-feature-name>`:
every commit that changes code then ends with `Intent: sdd:<name>` without anyone typing it, and a
code commit with no claim is refused (`proofbound task stop` when the change is done). Commits that
only touch specs, docs, `gates/`, review records or Markdown need no claim. Then have
someone who is not the author review it (`proofbound review draft <commit>`, decide, `review seal`),
and run `proofbound sync all`. The commit's gate turns **PASS** when every obligation is verified.
`proofbound warranty export <commit>` then writes a signed `warranty.json` that anyone can check
with `proofbound warranty verify warranty.json --repo <clone>`.

Step-by-step, with what each command prints: [Release 2 demo hub §4](../r2/DEMO.md#4-use-it-on-your-own-openspec-or-spec-kit-project)
and the [warranty demo](../r3/DEMO.md).

## 4. Every pull request gets a verdict

`init` also writes `proofbound.yml` in the tenant's `.github/workflows/`. On each pull request it installs Proofbound,
runs `proofbound sync all`, then `proofbound check origin/<base>`:

```
Proofbound: NOT WARRANTED (1 of 2 claiming commits verified)
  1f0c9e2a7b3d sdd:001-export BLOCKED  …
  8d4e6a1c2b9f sdd:001-export PASS  2 of 2 obligations verified SATISFIED; 2 of 2 requirement reviews VERIFIABLE
```

The job fails unless every commit that claims a spec is verified, and it fails on any commit that
claims no spec but changes code (listed as `UNCOVERED`). Commits that only touch specs, `gates/`,
`.proofbound/`, `docs/`, Markdown files, `.gitignore` or the Proofbound workflow need no claim. A pull
request with no claims and no uncovered commit passes (set the action's `require-claims: true` to
change that). Make the `proofbound` job a required status check in the branch protection rules, or a
red check can still be merged. The same check runs locally:
`proofbound check main`.

## 5. Attach the proof to a release

```bash
proofbound warranty bundle <previous-release-tag> -o dist/warranties
```

writes one signed `warranty-<commit>.json` per verified commit since that tag, plus `VERIFY.md` (a page
telling a buyer how to check each file themselves), `index.json` and `badge.svg`
(`![proofbound](dist/warranties/badge.svg)` in a README). The page, index and badge are unsigned
conveniences and say so; the warranty files are the proof. Exit 3 means the bundle was written but some
claiming commit is not VALID. Try it locally: `bash docs/releases/mvp/demo-share.sh`.

## 6. See it run first (from a clone, on Linux or WSL)

`bash docs/releases/mvp/demo-install.sh` needs Go (to build the release archives), GNU tar, and port
55432 free. It builds the archives for all five platforms, then times the user's part: install →
init → spec → claimed commit → independent review → **PASS**. That part took 27 s on 2026-11-20 on a
Linux/WSL machine; the archive build before it took longer with a cold Go cache.
