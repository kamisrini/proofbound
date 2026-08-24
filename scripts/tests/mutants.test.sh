#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$repo_root/tools/mutants"
GOCACHE=${GOCACHE:-/tmp/vera-mutants-selftest} go test ./... -count=1
