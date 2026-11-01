#!/usr/bin/env bash
# Verify a pristine-upstream PostgreSQL source tarball against a pinned manifest and its published
# SHA-256 checksum file (P7 B1, VD-p7-postgres-source-route-2026-11-01, light-touch mechanism:
# SHA-256 over HTTPS, no detached signature exists for PostgreSQL source tarballs). Fails CLOSED:
# any mismatch, missing input, or malformed field exits non-zero with no partial trust.
set -euo pipefail

die() { echo "p7-postgres-source-verify: $*" >&2; exit 1; }
command -v jq >/dev/null || die 'jq is required'
command -v sha256sum >/dev/null || die 'sha256sum is required'

manifest=''
input_dir=''
role='postgres-source'
while (($#)); do
  case $1 in
    --manifest) shift; (($#)) || die '--manifest requires a path'; manifest=$1 ;;
    --input-dir) shift; (($#)) || die '--input-dir requires a path'; input_dir=$1 ;;
    --role) shift; (($#)) || die '--role requires a value'; role=$1 ;;
    *) die "unknown option: $1" ;;
  esac
  shift
done
[[ -n $manifest && -n $input_dir ]] || die 'usage: --manifest <json> --input-dir <dir> [--role <role>]'
[[ -f $manifest ]] || die "manifest not found: $manifest"
[[ -d $input_dir ]] || die "input dir not found: $input_dir"

count=$(jq --arg r "$role" '[.inputs[]? | select(.role == $r)] | length' "$manifest") \
  || die "manifest is not valid JSON: $manifest"
[[ $count == 1 ]] || die "expected exactly one input with role=$role, found ${count:-0}"

read -r source_url published_sha256 size < <(
  jq -r --arg r "$role" '.inputs[] | select(.role == $r)
    | "\(.source_url) \(.published_sha256) \(.download_size_bytes)"' "$manifest"
)
[[ -n ${source_url:-} && -n ${published_sha256:-} && -n ${size:-} ]] \
  || die 'manifest entry missing source_url/published_sha256/download_size_bytes'
[[ $published_sha256 =~ ^[0-9a-f]{64}$ ]] || die "published_sha256 is not 64 lowercase hex: $published_sha256"
[[ $size =~ ^[0-9]+$ ]] || die "download_size_bytes is not a non-negative integer: $size"

tarball_name=${source_url##*/}
[[ -n $tarball_name && $tarball_name != "$source_url"/ ]] || die "cannot derive tarball name from source_url: $source_url"
tarball="$input_dir/$tarball_name"
checksum="$tarball.sha256"
[[ -f $tarball ]] || die "tarball not found: $tarball"
[[ -f $checksum ]] || die "published checksum file not found: $checksum"

# 1. The published checksum file must pin the same digest as the manifest (no silent divergence).
file_hash=$(awk 'NR==1{print $1}' "$checksum")
[[ $file_hash =~ ^[0-9a-f]{64}$ ]] || die "checksum file does not contain a 64-hex digest: $checksum"
[[ $file_hash == "$published_sha256" ]] \
  || die "manifest published_sha256 ($published_sha256) != published checksum file ($file_hash)"

# 2. The tarball bytes must hash to the pinned digest.
actual_hash=$(sha256sum "$tarball" | awk '{print $1}')
[[ $actual_hash == "$published_sha256" ]] \
  || die "tarball SHA-256 mismatch: computed $actual_hash, pinned $published_sha256"

# 3. The tarball size must match the pinned size.
actual_size=$(stat -c %s "$tarball")
[[ $actual_size == "$size" ]] || die "tarball size mismatch: got $actual_size, pinned $size"

echo "p7-postgres-source-verify: OK role=$role file=$tarball_name sha256=$published_sha256 size=$size"
