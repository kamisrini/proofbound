#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
pkg=${1:?usage: scripts/mutants.sh <kernel-relative-package>}
repo=$PWD
cd tools/mutants
exec env GOCACHE=${GOCACHE:-/tmp/vera-mutant-cache} go run . -root "$repo" -pkg "$pkg"
