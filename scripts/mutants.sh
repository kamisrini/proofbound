#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
pkg=${1:?usage: scripts/mutants.sh <kernel-relative-package>}
tags=${MUTANT_TEST_TAGS:-}
repo=$PWD
cd tools/mutants
args=( -root "$repo" -pkg "$pkg" )
if [ -n "$tags" ]; then args+=( -tags "$tags" ); fi
if [ "${MUTANT_FAIL_FAST:-0}" = 1 ]; then args+=( -fail-fast ); fi
if [ -n "${MUTANT_START:-}" ]; then args+=( -start "$MUTANT_START" ); fi
if [ -n "${MUTANT_END:-}" ]; then args+=( -end "$MUTANT_END" ); fi
exec env GOCACHE=${GOCACHE:-/tmp/proofbound-mutant-cache} go run . "${args[@]}"
