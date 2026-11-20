#!/usr/bin/env bash
# REL-1 / REL-2 of docs/plans/mvp-1-install-spec.md.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
repo=$PWD
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT

expect_fail() { # <code> <message-substring> <cmd...>
  local want=$1 msg=$2; shift 2
  local code=0
  "$@" >"$tmp/out" 2>&1 || code=$?
  [[ $code == "$want" ]] || { echo "release-build.test: want exit $want, got $code: $*" >&2; cat "$tmp/out" >&2; exit 1; }
  grep -q -- "$msg" "$tmp/out" || { echo "release-build.test: missing '$msg': $*" >&2; cat "$tmp/out" >&2; exit 1; }
}

# REL-1: refusals, in a fixture repository holding only the script (it never reaches go build).
fx=$tmp/fixture
mkdir -p "$fx/scripts"
cp scripts/release-build.sh "$fx/scripts/"
printf 'x\n' >"$fx/LICENSE"
git -C "$fx" init -q
git -C "$fx" config user.email fixture@example.invalid
git -C "$fx" config user.name fixture
git -C "$fx" add . && git -C "$fx" commit -qm fixture
expect_fail 2 usage bash "$fx/scripts/release-build.sh" 1.2.3 "$tmp/o1"
expect_fail 2 usage bash "$fx/scripts/release-build.sh" v1.2 "$tmp/o1"
mkdir -p "$tmp/full" && touch "$tmp/full/x"
expect_fail 1 "not empty" bash "$fx/scripts/release-build.sh" v1.2.3 "$tmp/full"
printf 'y\n' >>"$fx/LICENSE"
expect_fail 1 "differ from HEAD" bash "$fx/scripts/release-build.sh" v1.2.3 "$tmp/o2"
[[ ! -e $tmp/o2 ]] || { echo "release-build.test: refused build created its outdir" >&2; exit 1; }

# REL-2: two builds of the same tree are byte-identical; version stamped; archive holds two entries.
export PROOFBOUND_RELEASE_TARGETS=linux/amd64
scripts/release-build.sh --allow-dirty v0.0.0-test "$tmp/a" >/dev/null
scripts/release-build.sh --allow-dirty v0.0.0-test "$tmp/b" >/dev/null
cmp "$tmp/a/SHA256SUMS" "$tmp/b/SHA256SUMS" || { echo "release-build.test: SHA256SUMS differ" >&2; exit 1; }
cmp "$tmp/a/proofbound_linux_amd64.tar.gz" "$tmp/b/proofbound_linux_amd64.tar.gz"
(cd "$tmp/a" && sha256sum -c --quiet SHA256SUMS)
entries=$(tar -tzf "$tmp/a/proofbound_linux_amd64.tar.gz" | tr '\n' ' ')
[[ $entries == "LICENSE proofbound " ]] || { echo "release-build.test: entries '$entries'" >&2; exit 1; }
mkdir "$tmp/x" && tar -xzf "$tmp/a/proofbound_linux_amd64.tar.gz" -C "$tmp/x"
[[ $(cd "$tmp" && "$tmp/x/proofbound" version) == "proofbound v0.0.0-test" ]] || { echo "release-build.test: version not stamped" >&2; exit 1; }
cmp -s "$tmp/x/LICENSE" "$repo/LICENSE"
# No git state in the binary: otherwise a clone with an untracked file builds different bytes.
if go version -m "$tmp/x/proofbound" | grep -q 'vcs\.'; then echo "release-build.test: binary embeds vcs.* stamps" >&2; exit 1; fi
echo "release-build.test: ok"
