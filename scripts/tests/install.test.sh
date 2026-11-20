#!/usr/bin/env bash
# INST-1 / INST-2 of docs/plans/mvp-1-install-spec.md, against a local file:// release.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
repo=$PWD
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
case $(uname -m) in x86_64 | amd64) arch=amd64 ;; aarch64 | arm64) arch=arm64 ;; *) echo "install.test: skip machine" >&2; exit 1 ;; esac
case $(uname -s) in Linux) os=linux ;; Darwin) os=darwin ;; *) echo "install.test: unsupported system" >&2; exit 1 ;; esac
archive=proofbound_${os}_$arch.tar.gz

# A fake release: a shell "binary" that answers `version`.
make_release() { # <dir> <version>
  mkdir -p "$1/src"
  printf '#!/bin/sh\necho "proofbound %s"\n' "$2" >"$1/src/proofbound"
  chmod 0755 "$1/src/proofbound"
  printf 'license\n' >"$1/src/LICENSE"
  tar -C "$1/src" -czf "$1/$archive" proofbound LICENSE
  (cd "$1" && sha256sum "$archive" >SHA256SUMS)
}
run_install() { # <base-dir> <dest> [extra env...]
  local base=$1 dest=$2; shift 2
  env PROOFBOUND_BASE_URL="file://$base" PROOFBOUND_INSTALL_DIR="$dest" "$@" sh "$repo/install.sh"
}
refuses() { # <label> <base> <dest> [extra env...]
  local label=$1; shift
  if run_install "$@" >"$tmp/out" 2>&1; then echo "install.test: $label installed" >&2; cat "$tmp/out" >&2; exit 1; fi
  [[ ! -e $2/proofbound ]] || { echo "install.test: $label left a binary" >&2; exit 1; }
}

# INST-1: bad checksum, missing line, no sha tool -> nothing installed.
make_release "$tmp/bad" v1.0.0
printf '%064d  %s\n' 0 "$archive" >"$tmp/bad/SHA256SUMS"
refuses "bad checksum" "$tmp/bad" "$tmp/d1"
grep -q "checksum mismatch" "$tmp/out"
make_release "$tmp/noline" v1.0.0
printf '%064d  other.tar.gz\n' 0 >"$tmp/noline/SHA256SUMS"
refuses "missing line" "$tmp/noline" "$tmp/d2"
grep -q "no line for $archive" "$tmp/out"
make_release "$tmp/good" v1.0.0
mkdir "$tmp/bin"
for t in sh curl tar gzip mktemp uname awk cut mkdir cp chmod mv rm cat; do ln -s "$(command -v "$t")" "$tmp/bin/$t"; done
refuses "no sha tool" "$tmp/good" "$tmp/d3" PATH="$tmp/bin"
grep -q "no sha256sum or shasum" "$tmp/out"

# INST-2: a good release installs, prints its version, and replaces an older binary.
run_install "$tmp/good" "$tmp/d4" >"$tmp/out"
[[ -x $tmp/d4/proofbound && $("$tmp/d4/proofbound" version) == "proofbound v1.0.0" ]]
grep -q "installed proofbound v1.0.0" "$tmp/out"
make_release "$tmp/newer" v1.1.0
run_install "$tmp/newer" "$tmp/d4" >/dev/null
[[ $("$tmp/d4/proofbound" version) == "proofbound v1.1.0" ]]
[[ -z $(find "$tmp/d4" -name '.proofbound.tmp.*') ]]
echo "install.test: ok"
