#!/usr/bin/env bash
# Self-test for scripts/p7-postgres-source-verify.sh: a neutral control that MUST be accepted
# (pristine input) and positive controls that MUST be rejected (tamper, wrong size, manifest/checksum
# divergence, bytes-vs-pin mismatch, missing input). Auto-discovered by `make hooks-test`.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
verify=scripts/p7-postgres-source-verify.sh

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
dir="$tmp/dl"
mkdir -p "$dir"
name=postgresql-16.4.tar.bz2

# Deterministic stand-in "tarball".
printf 'proofbound p7 b1 source fixture\n' >"$dir/$name"
h=$(sha256sum "$dir/$name" | awk '{print $1}')
s=$(stat -c %s "$dir/$name")
printf '%s  %s\n' "$h" "$name" >"$dir/$name.sha256"
wrong=$(printf '%064d' 0)

mkman() { # $1=published_sha256 $2=size ; prints manifest path
  local m="$tmp/manifest.$1.$2.json"
  cat >"$m" <<EOF
{"inputs":[{"role":"postgres-source","platform":"source/any","format":"tar.bz2","source_url":"https://ftp.postgresql.org/pub/source/v16.4/$name","checksum_url":"x","published_sha256":"$1","download_sha256":"$1","download_size_bytes":$2}]}
EOF
  printf '%s' "$m"
}

reject() { # $1=message ; remaining args = verify args
  local msg=$1; shift
  if bash "$verify" "$@" >/dev/null 2>&1; then
    echo "$msg" >&2
    exit 1
  fi
}

man=$(mkman "$h" "$s")

# Neutral control: pristine input is accepted.
bash "$verify" --manifest "$man" --input-dir "$dir" >/dev/null \
  || { echo 'pristine input was rejected' >&2; exit 1; }

# Positive control that MUST die: a tampered tarball byte.
cp "$dir/$name" "$dir/$name.bak"
printf 'x' >>"$dir/$name"
reject 'tampered tarball was accepted' --manifest "$man" --input-dir "$dir"
mv "$dir/$name.bak" "$dir/$name"

# Positive control: manifest pins the wrong size.
reject 'wrong pinned size was accepted' --manifest "$(mkman "$h" $((s + 1)))" --input-dir "$dir"

# Positive control: manifest digest diverges from the published checksum file.
reject 'manifest/checksum divergence was accepted' --manifest "$(mkman "$wrong" "$s")" --input-dir "$dir"

# Positive control: manifest and checksum agree but neither matches the bytes.
printf '%s  %s\n' "$wrong" "$name" >"$dir/$name.sha256"
reject 'bytes-vs-pin mismatch was accepted' --manifest "$(mkman "$wrong" "$s")" --input-dir "$dir"
printf '%s  %s\n' "$h" "$name" >"$dir/$name.sha256"

# Positive control that ISOLATES the `! -L` guard (Law 3: the test must fail if the guard is
# removed). A plain symlink to correct bytes is NOT a valid isolation control: `stat -c %s` on a
# symlink reports the length of its *target path* (not the dereferenced content size), so the size
# check rejects it on its own and the guard could rot undetected. So we deliberately defeat the size
# check too: make the target-path length EQUAL the real content size. Then `sha256sum` (which DOES
# deref) matches the pin AND `stat -c %s` (the target-path length) equals the pinned size, so both
# other checks pass and ONLY the `! -L` guard can reject it. Verified by stripping the guard from a
# copy of the verifier: with it gone this fixture is accepted and this control fails.
gdir="$tmp/guard"
mkdir -p "$gdir"
payload='proofbound guard-isolation fixture bytes'   # length computed below, not asserted here
gsize=${#payload}
gtarget=$(printf 'a%.0s' $(seq 1 "$gsize"))           # target filename of exactly gsize characters
printf '%s' "$payload" >"$gdir/$gtarget"
gh=$(sha256sum "$gdir/$gtarget" | awk '{print $1}')
ln -s "$gtarget" "$gdir/$name"                        # relative symlink; readlink length == gsize
printf '%s  %s\n' "$gh" "$name" >"$gdir/$name.sha256"
reject 'symlink passing both hash and size (guard isolation) was accepted' \
  --manifest "$(mkman "$gh" "$gsize")" --input-dir "$gdir"

# Positive control: missing tarball.
rm -f "$dir/$name"
reject 'missing tarball was accepted' --manifest "$man" --input-dir "$dir"

echo 'p7-postgres-source-verify.test: ok'
