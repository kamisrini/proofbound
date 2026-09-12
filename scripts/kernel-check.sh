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
GOCACHE=${GOCACHE:-/tmp/proofbound-go-build} GOLANGCI_LINT_CACHE=${GOLANGCI_LINT_CACHE:-/tmp/proofbound-golangci} golangci-lint run ./...
