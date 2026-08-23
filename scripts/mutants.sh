#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
pkg=${1:?usage: scripts/mutants.sh <kernel-relative-package>}
tags=${MUTANT_TEST_TAGS:-}
repo=$PWD
cd tools/mutants
args=( -root "$repo" -pkg "$pkg" )
if [ -n "$tags" ]; then args+=( -tags "$tags" ); fi
exec env GOCACHE=${GOCACHE:-/tmp/vera-mutant-cache} go run . "${args[@]}"
