#!/usr/bin/env bash
set -euo pipefail
target="${@: -1}"
printf '%s %s\n' "${0##*/}" "$target" >>"$VERA_TEST_LOG"
if [[ ${VERA_TEST_SLEEP:-0} != 0 ]]; then sleep "$VERA_TEST_SLEEP"; fi
if [[ ${VERA_TEST_FAIL_TARGET:-} == "$target" ]]; then exit 9; fi
