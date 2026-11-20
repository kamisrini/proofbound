# MVP step 1: installable Proofbound — spec

**Status:** draft 1, for one non-author review (Law 11: MVP steps get one spec round).
**Decision:** `VD-vertical-mvp-first-ai3dhn` step 1. **Intent:** `CI-mvp-install-k99zpl` →
`BR-vertical-mvp-xc31gp` O-1.

## 1. Payoff

One command a stranger can run on Linux or macOS, then a first gate PASS on a fresh repository,
in under 5 minutes, with no Go toolchain and no Proofbound source checkout:

```bash
curl -fsSL https://raw.githubusercontent.com/kamisrini/proofbound/main/install.sh | sh
proofbound version
```

Proven locally (pushes are held, so nothing is published) by `docs/releases/mvp/demo-install.sh`,
which builds the archives, installs from them through `install.sh` with a `file://` base URL, and
runs the R2 claim → review → PASS loop with the installed binary.

## 2. Interface

### 2.1 `proofbound version` (kernel, `internal/cli`)

Prints exactly `proofbound <version>\n` and exits 0. `<version>` is the package variable
`cli.Version`, default `dev`, set at build time with
`-ldflags "-X github.com/kamisrini/proofbound/kernel/internal/cli.Version=<v>"`. It runs anywhere: no
repository root, no database, no network (it is dispatched before the repository lookup, like
`warranty verify`). Any extra argument is a usage error (exit 2).

### 2.2 `scripts/release-build.sh <version> [outdir]`

- `<version>` must match `^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$`, else exit 2. `outdir`
  defaults to `dist/` (gitignored); an existing non-empty `outdir` is refused.
- Refuses (exit 1) when tracked files differ from HEAD, unless `--allow-dirty` is passed first;
  untracked files do not count.
- Targets: `linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64`;
  `PROOFBOUND_RELEASE_TARGETS` (space-separated `os/arch`) narrows them, for tests.
- Each target: `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -buildid= -X …cli.Version=<v>"`
  of `./cmd/proofbound`, packed as `proofbound_<os>_<arch>.tar.gz` holding `proofbound` (or
  `proofbound.exe`) and `LICENSE`. Archive names carry no version, so GitHub's
  `releases/latest/download/<file>` URL is stable; the version lives in the binary and the release.
- Archives are reproducible: entries sorted, mtime = HEAD's commit time, owner/group 0, `gzip -n`.
- Writes `SHA256SUMS` (`<hex>  <file>` per archive, sorted by file name).
- No PostgreSQL binary, archive or data directory is ever packed (binary-exclusion law unchanged).

### 2.3 `install.sh` (repository root)

POSIX `sh`. Environment:
- `PROOFBOUND_BASE_URL`: where archives and `SHA256SUMS` live. Default
  `https://github.com/kamisrini/proofbound/releases/latest/download`, or `…/releases/download/<v>`
  when `PROOFBOUND_VERSION=<v>` is set. Any URL `curl` accepts, including `file://`.
- `PROOFBOUND_INSTALL_DIR`: default `$HOME/.local/bin`.

Steps: map `uname -s`/`uname -m` to `linux|darwin` / `amd64|arm64` (anything else: exit 1 with a
message naming the manual download); download the archive and `SHA256SUMS` into a fresh temporary
directory; take the archive's line from `SHA256SUMS` and compare with `sha256sum` or
`shasum -a 256`; extract; move `proofbound` into the install dir (created if missing) via a
temporary name and `mv`; run `proofbound version`; print a PATH hint if the dir is not on `PATH`.
It never uses `sudo` and never writes outside the temporary dir and the install dir.

## 3. Invariants and proving tests

| ID | Statement | Proving test |
|---|---|---|
| CLI-INV-25 | `version` prints `proofbound <cli.Version>` and exits 0 outside any repository without opening a store; `version x` is a usage error | `kernel/internal/cli/version_test.go::TestVersionRunsAnywhere` |
| INST-1 | `install.sh` installs nothing and exits non-zero when the archive's checksum does not match, when `SHA256SUMS` has no line for it, or when no SHA-256 tool exists | `scripts/tests/install.test.sh` |
| INST-2 | On a matching checksum it installs an executable `proofbound` into `PROOFBOUND_INSTALL_DIR`, replacing an older one, and prints its version | `scripts/tests/install.test.sh` |
| REL-1 | `release-build.sh` refuses a bad version, a non-empty outdir, and a dirty tracked tree without `--allow-dirty` | `scripts/tests/release-build.test.sh` |
| REL-2 | Two builds of the same commit give byte-identical archives and `SHA256SUMS`; the binary reports the stamped version; no archive entry other than the binary and `LICENSE` | `scripts/tests/release-build.test.sh` (one target, for speed) |

## 4. Non-goals

- Publishing a GitHub release (pushes are held, `DEF-PUSH-HOLD-20261116`), Homebrew, Windows
  installer (Windows users download the archive; `tar` ships with Windows 10+).
- Signing the archives. A checksum fetched from the same server proves the download is intact, not
  that the server is honest; the quickstart says so. Signed releases come with step 4 (shareable proof).
- Bundling PostgreSQL. The CLI is unchanged: `DATABASE_URL` if set, otherwise the embedded database,
  which downloads a PostgreSQL build from Maven Central on first use. That artifact's signature did
  not verify (`DEF-P7-001`); the quickstart says so and shows `DATABASE_URL` as the alternative.

## 5. Docs

`docs/releases/mvp/QUICKSTART.md`: install, first-run database note, then `init` and the R2 loop
(links `docs/releases/r2/DEMO.md`). `README.md` gains a three-line "Install" section linking it.
`docs/releases/r2/demo-loop.sh` honours `PROOFBOUND_BIN` (use that binary instead of building).
