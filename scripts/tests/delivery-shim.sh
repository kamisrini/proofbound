#!/usr/bin/env bash
set -euo pipefail
target="${@: -1}"
printf '%s %s\n' "${0##*/}" "$target" >>"$PROOFBOUND_TEST_LOG"
if [[ ${PROOFBOUND_TEST_SLEEP:-0} != 0 ]]; then sleep "$PROOFBOUND_TEST_SLEEP"; fi
if [[ ${PROOFBOUND_TEST_FAIL_TARGET:-} == "$target" ]]; then exit 9; fi
