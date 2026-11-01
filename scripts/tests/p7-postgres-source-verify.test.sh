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

# Positive control: missing tarball.
rm -f "$dir/$name"
reject 'missing tarball was accepted' --manifest "$man" --input-dir "$dir"

echo 'p7-postgres-source-verify.test: ok'
