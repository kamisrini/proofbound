#!/bin/sh
# Install the proofbound CLI.
#
#   curl -fsSL https://raw.githubusercontent.com/kamisrini/proofbound/main/install.sh | sh
#
# Environment (all optional):
#   PROOFBOUND_VERSION      a release tag such as v0.3.0 (default: the latest release)
#   PROOFBOUND_BASE_URL     where the archives and SHA256SUMS live; any URL curl accepts, file:// too
#   PROOFBOUND_INSTALL_DIR  where to put the binary (default: $HOME/.local/bin)
#
# It checks the archive against SHA256SUMS before installing and installs nothing if that fails.
# The checksum comes from the same server as the archive, so it proves the download is intact, not
# that the server is honest. It never uses sudo. Contract: docs/plans/mvp-1-install-spec.md §2.3.
set -eu

fail() { echo "install.sh: $*" >&2; exit 1; }

repo_url=https://github.com/kamisrini/proofbound/releases
if [ -n "${PROOFBOUND_BASE_URL:-}" ]; then
  base=$PROOFBOUND_BASE_URL
elif [ -n "${PROOFBOUND_VERSION:-}" ]; then
  base=$repo_url/download/$PROOFBOUND_VERSION
else
  base=$repo_url/latest/download
fi
dest=${PROOFBOUND_INSTALL_DIR:-${HOME:?HOME is not set}/.local/bin}

case $(uname -s) in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) fail "unsupported system $(uname -s); download an archive from $repo_url by hand" ;;
esac
case $(uname -m) in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) fail "unsupported machine $(uname -m); download an archive from $repo_url by hand" ;;
esac
archive=proofbound_${os}_$arch.tar.gz

if command -v sha256sum >/dev/null 2>&1; then
  sha() { sha256sum "$1" | cut -d' ' -f1; }
elif command -v shasum >/dev/null 2>&1; then
  sha() { shasum -a 256 "$1" | cut -d' ' -f1; }
else
  fail "no sha256sum or shasum found; cannot check the download, so nothing was installed"
fi
command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

echo "downloading $archive from $base"
curl -fsSL "$base/$archive" -o "$tmp/$archive" || fail "could not download $base/$archive"
curl -fsSL "$base/SHA256SUMS" -o "$tmp/SHA256SUMS" || fail "could not download $base/SHA256SUMS"

want=$(awk -v f="$archive" '$2 == f || $2 == "*" f { print $1 }' "$tmp/SHA256SUMS")
[ -n "$want" ] || fail "SHA256SUMS has no line for $archive; nothing was installed"
got=$(sha "$tmp/$archive")
[ "$got" = "$want" ] || fail "checksum mismatch for $archive (expected $want, got $got); nothing was installed"

mkdir "$tmp/x"
tar -xzf "$tmp/$archive" -C "$tmp/x" proofbound || fail "archive has no proofbound binary"
mkdir -p "$dest"
cp "$tmp/x/proofbound" "$dest/.proofbound.tmp.$$"
chmod 0755 "$dest/.proofbound.tmp.$$"
mv -f "$dest/.proofbound.tmp.$$" "$dest/proofbound"

echo "installed $("$dest/proofbound" version) to $dest/proofbound"
case ":$PATH:" in
  *":$dest:"*)
    # An older proofbound earlier on PATH would run instead of this one (DEF-INSTALL-PATH-SHADOW-001).
    found=$(command -v proofbound 2>/dev/null || true)
    if [ -n "$found" ] && [ "$found" != "$dest/proofbound" ]; then
      echo "warning: $found comes first on your PATH and will run instead ($("$found" version 2>/dev/null || echo "unknown version")); remove it or put $dest earlier on PATH"
    fi
    ;;
  *) echo "note: $dest is not on your PATH; add it, e.g.  export PATH=\"$dest:\$PATH\"" ;;
esac
