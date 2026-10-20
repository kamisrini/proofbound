# P7 Phase-A Go/projector build probe — 2026-10-20

**Status:** `UNVERIFIED` evidence capture; not a projector/build manifest and not runtime
acceptance.

**Evidence revision:** `007d71bdb94b165f93ce696f01595276f8042a6d`

This note records the first reproducibility-foundation probe authorized after the companion and
synthetic archive-control slices were accepted. It captures the pinned Go environment and a local
build of `cmd/proofbound`. It does not validate, download, approve, or execute the deferred
PostgreSQL binary, and it does not claim L12, L17, or a passing case manifest.

## Inputs

| Input | Value |
|---|---|
| `kernel/go.mod` SHA-256 | `0cefb499cbd9ef6cfe917ae58d3857c0456481d01c45161a6e358c396d9d1393` |
| `kernel/go.sum` SHA-256 | `3b5ef08294f1dbcb6a1acc9fc011410f7ebcb78becc0e8b6c0aeb191c776f1bf` |
| setup-source manifest SHA-256 | `e312bf28cea20d516e3b21704a53ba3ebc8ac84ff988eccb7da555317b900b33` |
| setup-source README SHA-256 | `5d200e59ac1b58e613e9fdc5946a77cec3b8afc7a7a57426c4134811e87c29e9` |
| GOOS / GOARCH | `linux` / `amd64` |
| Go version | `go1.26.7` |
| GOROOT | `/home/thamm/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.7.linux-amd64` |
| GOMODCACHE | `/home/thamm/go/pkg/mod` |
| GOPATH | `/home/thamm/go` |

## Probe commands and results

Executed from `kernel/` with the pinned environment:

```text
GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH go version
go version go1.26.7 linux/amd64

GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH go list -m all
PASS — selected module graph resolved locally

GOTOOLCHAIN=go1.26.7 PATH=/home/thamm/go/bin:$PATH go build -o /tmp/proofbound-p7-projector-probe ./cmd/proofbound
PASS — local build completed
```

The resulting untracked temporary executable was outside the repository and measured:

```text
SHA-256: 4d992a5650620e46efc9f4a209cd0bd4c183ca7475f58a9c13fa229c29328222
Size:    17362505 bytes
```

The executable is not committed and is not a reviewed projector artifact.

## Missing evidence and boundaries

This probe remains `UNVERIFIED` because the accepted P7 evidence contract still lacks:

- a canonical projector/build manifest with the complete tracked-kernel file corpus;
- a separately hashed canonical Go toolchain manifest and verified toolchain archive;
- Git-tracked `.mod`/`.zip` bytes for the complete selected module graph;
- a reviewed executable artifact bound to those exact inputs;
- a PostgreSQL archive and executable identity, which remain under dormant `DEF-P7-001`;
- runtime settings, ledger before/after witnesses, network observation, cleanup evidence, and a
  bound case manifest; and
- independent review of this evidence capture and of any later complete manifest.

The local module cache and temporary executable are observations only. They do not satisfy the
clean-clone or content-addressed-byte requirements and must not be relabeled as `PASS`.
