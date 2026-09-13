#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

# The fixture suite proves the checker; this invocation keeps the governed
# production result bound to the evidence currently tracked by HEAD/worktree.
scripts/p6-census.sh --check
