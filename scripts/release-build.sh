#!/usr/bin/env bash
# Build reproducible per-OS release archives of the proofbound CLI plus SHA256SUMS.
#
#   scripts/release-build.sh [--allow-dirty] <version> [outdir]    # outdir defaults to dist/
#
# Contract: docs/plans/mvp-1-install-spec.md §2.2 (REL-1, REL-2). Archive names carry no version
# (proofbound_<os>_<arch>.tar.gz) so a "latest" download URL is stable; the version is stamped into
# the binary. PROOFBOUND_RELEASE_TARGETS ("os/arch ...") narrows the target list, for tests.
# No PostgreSQL binary is ever packed: the archive holds the CLI and LICENSE only.
# Maintainer side, Linux only: it needs GNU tar and sha256sum (macOS's bsdtar lacks the flags that
# make archives reproducible). Installing (install.sh) works on Linux and macOS.
set -euo pipefail

usage() { echo "usage: $0 [--allow-dirty] <vMAJOR.MINOR.PATCH[-pre]> [outdir]" >&2; exit 2; }

allow_dirty=0
if [[ ${1:-} == --allow-dirty ]]; then allow_dirty=1; shift; fi
version=${1:-}
[[ $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$ ]] || usage
(($# <= 2)) || usage

repo=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
outdir=${2:-$repo/dist}
if [[ -d $outdir ]] && [[ -n $(ls -A "$outdir") ]]; then
  echo "release-build: $outdir is not empty; pass an empty or new directory" >&2
  exit 1
fi
if ((allow_dirty == 0)) && ! git -C "$repo" diff --quiet HEAD --; then
  echo "release-build: tracked files differ from HEAD; commit first or pass --allow-dirty" >&2
  exit 1
fi
command -v go >/dev/null || { echo "release-build: go is required" >&2; exit 1; }
command -v sha256sum >/dev/null && tar --version 2>/dev/null | grep -q 'GNU tar' || {
  echo "release-build: needs GNU tar and sha256sum (build releases on Linux)" >&2
  exit 1
}

targets=${PROOFBOUND_RELEASE_TARGETS:-"linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"}
epoch=$(git -C "$repo" log -1 --format=%ct HEAD)
mkdir -p "$outdir"
outdir=$(cd "$outdir" && pwd)
stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT

for target in $targets; do
  os=${target%/*}
  arch=${target#*/}
  [[ $os =~ ^(linux|darwin|windows)$ && $arch =~ ^(amd64|arm64)$ ]] || {
    echo "release-build: unsupported target $target" >&2
    exit 2
  }
  exe=proofbound
  [[ $os == windows ]] && exe=proofbound.exe
  dir="$stage/${os}_$arch"
  mkdir -p "$dir"
  (cd "$repo/kernel" && CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -buildvcs=false \
    -ldflags "-s -w -buildid= -X github.com/kamisrini/proofbound/kernel/internal/cli.Version=$version" \
    -o "$dir/$exe" ./cmd/proofbound)
  cp "$repo/LICENSE" "$dir/LICENSE"
  chmod 0755 "$dir/$exe"
  chmod 0644 "$dir/LICENSE"
  archive="proofbound_${os}_$arch.tar.gz"
  tar -C "$dir" --sort=name --mtime="@$epoch" --owner=0 --group=0 --numeric-owner \
    -cf - LICENSE "$exe" | gzip -n -9 >"$outdir/$archive"
  echo "built $archive"
done

(cd "$outdir" && LC_ALL=C sha256sum proofbound_*.tar.gz | LC_ALL=C sort -k2 >SHA256SUMS)
echo "wrote $outdir/SHA256SUMS ($version)"
