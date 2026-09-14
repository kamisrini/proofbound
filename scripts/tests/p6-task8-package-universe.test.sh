#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

spec=docs/plans/p6-task8-package-acceptance-SPEC.md
[[ -f $spec ]] || { echo 'Task 8 SPEC is missing' >&2; exit 1; }
rg -q '16 packages' "$spec" || { echo 'production package count is not frozen' >&2; exit 1; }
rg -q 'internal/specfirst' "$spec" || { echo 'test-only package is not explicitly classified' >&2; exit 1; }

mapfile -t production < <(git ls-files 'kernel/internal/**/*.go' | rg -v '_test\.go$' | sed 's#/[^/]*$##' | sort -u)
[[ ${#production[@]} -eq 16 ]] || { echo "expected 16 production packages, got ${#production[@]}" >&2; exit 1; }
[[ " ${production[*]} " != *' kernel/internal/specfirst '* ]] || { echo 'test-only package entered production set' >&2; exit 1; }

awk -F '\t' '$1 == "C3-017" && $2 == "C3" && $3 == "package-test-only:internal/specfirst" { found=1 } END { exit !found }' docs/plans/p6-census-rows.tsv
echo 'p6-task8-package-universe: production/test-only classification passed'
