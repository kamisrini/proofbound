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

> **Not published yet.** No public release exists until the maintainer publishes one. To try the
> exact install path today, from a clone: `bash docs/releases/mvp/demo-install.sh` builds the
> archives locally, installs from them, and runs the loop below with the installed binary.

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
[Spec Kit](https://github.com/github/spec-kit):

```bash
proofbound init openspec          # or: proofbound init speckit
git add .proofbound/sdd-profile.json .gitignore AGENTS.md .claude
PROOFBOUND_NO_INTENT=1 git commit -m "add proofbound"
```

Then end each commit that implements a spec with `Intent: sdd:<change-or-feature-name>`, have
someone who is not the author review it (`proofbound review draft <commit>`, decide, `review seal`),
and run `proofbound sync all`. The commit's gate turns **PASS** when every obligation is verified.
`proofbound warranty export <commit>` then writes a signed `warranty.json` that anyone can check
with `proofbound warranty verify warranty.json --repo <clone>`.

Step-by-step, with what each command prints: [Release 2 demo hub §4](../r2/DEMO.md#4-use-it-on-your-own-openspec-or-spec-kit-project)
and the [warranty demo](../r3/DEMO.md).

## 4. See it run first (from a clone, on Linux or WSL)

`bash docs/releases/mvp/demo-install.sh` needs Go (to build the release archives), GNU tar, and port
55432 free. It builds the archives for all five platforms, then times the user's part: install →
init → spec → claimed commit → independent review → **PASS**. That part took 27 s on 2026-11-20 on a
Linux/WSL machine; the archive build before it took longer with a cold Go cache.
