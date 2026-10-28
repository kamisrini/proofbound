# P7 source/build input binding — non-runtime review packet

**Intent:** `P7-NONRUNTIME-SOURCE-BUILD-INPUT-BINDING-1`
**Owner:** `sk`
**Target artifact:** `proofbound.p7.projector-manifest.v1`
**Evidence revision:** `fa11ed3cdc33262a4a3886f7e896f6a2570dcff9`
**Status:** `PENDING` independent non-author review

## Review question

Does the bounded non-runtime inventory reproducibly bind only committed candidate source files
and the previously accepted module source corpus, disclose how module pairs were inferred,
and leave all unavailable build closure, build settings, and toolchain evidence explicitly
`UNVERIFIED` or `PENDING`?

## Inventory and reproduction

The [generated inventory](p7-source-build-input-inventory-20261028.json) is a documentary index,
not a canonical projector manifest. Its SHA-256 is
`a6925fb673afec09aa506127d33fb88d6b9e0b9592fc9e057520cb502bfb0b50`.
The [read-only generator](../../scripts/p7-source-build-input-binding.py) has SHA-256
`feb94a9f2681e3b4444a5b0e0bedf1ac0b2dfbd53f47a5ae764d0733157de01f`.

From the repository root, reproduce the inventory bytes with:

```text
python3 scripts/p7-source-build-input-binding.py --revision fa11ed3cdc33262a4a3886f7e896f6a2570dcff9 | sha256sum
```

Captured output: `a6925fb673afec09aa506127d33fb88d6b9e0b9592fc9e057520cb502bfb0b50  -`.
The command reads tracked files, uses Git to require clean working bytes relative to the named
revision, and reads ZIP central-directory names only. It does not extract archives, invoke Go,
retrieve external bytes, or execute or validate a binary. The corresponding on-disk JSON has
the same SHA-256.

The inventory records 42 sorted committed candidate source files with Git modes and SHA-256:
production `.go` paths under `kernel/` excluding `_test.go` and `_windows.go`, `kernel/go.mod`,
`kernel/go.sum`, and `kernel/internal/store/migrations/001_ledger.sql`. This is a conservative
candidate set for Linux/amd64, not the exact package import or compiled-file closure. Each
source identity is bound to the named Git revision; untracked `kernel/cmd/p6dbprobe/` is outside
the inventory.

The inventory records 77 module/version pairs. Each `.mod` and `.zip` artifact is a Git-tracked
regular file at a SHA-256-named path in the already [accepted source corpus](verdicts/p7-go-module-corpus-round1-20261023.md).
The generator checks each file's digest against its path, reads the `.mod` `module` directive,
reads the ZIP top-level `module@version` prefix from its central directory, and pairs them only
where each module name has exactly one `.mod` and one `.zip`. All 77 pairs were unambiguous;
the inventory's `issues` array is empty. Unique names establish a documentary pairing, not
an independent proof that the ZIP version and `.mod` content represent the complete selected
build graph. No module archive member was executed or treated as a runtime input.

## Declared build settings and gaps

| Field | Current source | Status |
|---|---|---|
| Module root | `kernel/go.mod`: `github.com/kamisrini/proofbound/kernel` | `BOUND` to committed text |
| Go language/toolchain declaration | `kernel/go.mod`: `go 1.26`, `toolchain go1.26.7` | `BOUND` to committed text; archive and extracted toolchain `UNVERIFIED` |
| Historical command | `GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH go build -o /tmp/proofbound-p7-projector-probe ./cmd/proofbound` in [Phase-A probe](p7-replay-lock-fixtures/phase-a-build-probe-20261020.md) | `BOUND` to historic note; effective environment and current build `UNVERIFIED` |
| Historical platform | `go version go1.26.7 linux/amd64` in Phase-A probe | `BOUND` to historic note; current effective platform `UNVERIFIED` |
| `CGO_ENABLED`, `gcflags`, `ldflags`, tags, other effective environment | Not fully recorded by Phase-A probe | `PENDING` |
| Complete compiled source closure | Candidate source entries only; no import graph proof | `PENDING` |
| Complete selected module graph and `.mod`/`.zip` verification against it | Unique pairing of accepted source artifacts; no full graph reconstruction or Go checksum proof | `PENDING` |
| Verified toolchain archive/tree and built executable | Not supplied by this slice | `UNVERIFIED` |
| Concrete complete canonical projector manifest | Not emitted by this documentary inventory | `PENDING` |

`kernel/go.mod` SHA-256 remains
`0cefb499cbd9ef6cfe917ae58d3857c0456481d01c45161a6e358c396d9d1393`;
`kernel/go.sum` remains
`3b5ef08294f1dbcb6a1acc9fc011410f7ebcb78becc0e8b6c0aeb191c776f1bf`.
Their committed identities alone do not promote the historical probe to reproducible-build
evidence.

## Reviewer checks and non-claims

Verify the final frozen implementation revision and SHA-256 of the intent, generator, inventory,
and this packet. Reproduce the inventory digest; inspect source inclusion and exclusions, ZIP
module naming, pair cardinality, and gap statuses; run the repository deferrals,
artifact-integrity, and whitespace checks. Return `ACCEPT`, `AMEND`, or `REJECT` for this exact
documentary scope and identify the reviewed revision.

This packet does not accept a complete projector/build manifest, reproducible build, external
toolchain, binary, PostgreSQL JAR, runtime fixture, L12/L17 result, behavior-lock acceptance,
clean-clone proof, or runtime verdict. `DEF-P7-001` remains dormant.
