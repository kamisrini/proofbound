#!/usr/bin/env bash
set -euo pipefail
repo=$(git rev-parse --show-toplevel)
tmp=$(mktemp); trap 'rm -f "$tmp"' EXIT
cd "$repo/kernel"
GOCACHE=${GOCACHE:-/tmp/proofbound-go-build} go build ./...
if ! GOCACHE=${GOCACHE:-/tmp/proofbound-go-build} go test ./... -count=1 -v >"$tmp" 2>&1; then cat "$tmp"; exit 1; fi
cd "$repo"
scripts/skip-lint.sh "$tmp"
cd "$repo/kernel"
# golangci-lint may be installed to GOPATH/bin and not on PATH; resolve it, and
# fail closed with a clear message if it is genuinely absent (Law 3).
gcl=$(command -v golangci-lint || true)
if [[ -z $gcl ]]; then
  cand="$(go env GOPATH)/bin/golangci-lint"
  [[ -x $cand ]] && gcl=$cand
fi
[[ -n $gcl ]] || { echo 'kernel-check: golangci-lint not found on PATH or in $(go env GOPATH)/bin' >&2; exit 1; }
GOCACHE=${GOCACHE:-/tmp/proofbound-go-build} GOLANGCI_LINT_CACHE=${GOLANGCI_LINT_CACHE:-/tmp/proofbound-golangci} "$gcl" run ./...
