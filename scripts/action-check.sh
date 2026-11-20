#!/usr/bin/env bash
# The body of the Proofbound GitHub Action (action.yml): sync the checkout, run `proofbound check`
# on base..HEAD, write the Markdown result to the job summary and the plain one to the log, and exit
# with the check's code. Contract: docs/plans/mvp-2-pr-check-spec.md §2.2 (ACT-1).
#   PB_BASE / PB_DEFAULT_BASE   the base ref (input, else origin/<pull request base>)
#   PB_REQUIRE_CLAIMS           "true" makes a pull request with no claims fail
#   GITHUB_STEP_SUMMARY         where the Markdown goes (optional outside Actions)
set -euo pipefail
base=${PB_BASE:-${PB_DEFAULT_BASE:-}}
if [[ -z $base ]]; then
  echo "proofbound check: no base ref; set the action's 'base' input when not running on pull_request" >&2
  exit 2
fi
if ! git rev-parse --verify --quiet "$base^{commit}" >/dev/null; then
  echo "proofbound check: base $base is not in this checkout; use actions/checkout with fetch-depth: 0" >&2
  exit 2
fi
args=("$base")
[[ ${PB_REQUIRE_CLAIMS:-false} == true ]] && args+=(--require-claims)

proofbound sync all
code=0
proofbound check "${args[@]}" || code=$?
if [[ -n ${GITHUB_STEP_SUMMARY:-} ]]; then
  proofbound check "${args[@]}" --markdown >>"$GITHUB_STEP_SUMMARY" || true
fi
exit "$code"
