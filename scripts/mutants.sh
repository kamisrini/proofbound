#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
pkg=${1:?usage: scripts/mutants.sh <kernel-relative-package>}
exec env GO111MODULE=off go run ./tools/mutants -root "$PWD" -pkg "$pkg"
