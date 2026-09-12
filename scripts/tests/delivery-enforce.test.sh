#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d); trap 'rm -rf "$tmp"; rm -rf .proofbound/delivery.lock' EXIT
bin="$tmp/bin"; mkdir "$bin"
ln -s "$PWD/scripts/tests/delivery-shim.sh" "$bin/make"
ln -s "$PWD/scripts/tests/delivery-shim.sh" "$bin/go"
export PATH="$bin:$PATH" PROOFBOUND_TEST_LOG="$tmp/log"

PROOFBOUND_TEST_SLEEP=2 bash scripts/delivery-enforce.sh >"$tmp/first.out" 2>&1 &
first=$!
for _ in {1..40}; do [[ -d .proofbound/delivery.lock ]] && break; sleep 0.05; done
if [[ ! -d .proofbound/delivery.lock ]]; then echo 'workflow did not acquire lock' >&2; exit 1; fi
if bash scripts/delivery-enforce.sh >"$tmp/second.out" 2>&1; then
  echo 'concurrent workflow was accepted' >&2
  exit 1
fi
wait "$first"
if [[ -d .proofbound/delivery.lock ]]; then echo 'lock was not cleaned after success' >&2; exit 1; fi

mkdir .proofbound/delivery.lock
printf '999999\n' >.proofbound/delivery.lock/pid
: >"$tmp/log"
set +e
PROOFBOUND_TEST_FAIL_TARGET=index-check-witnessed bash scripts/delivery-enforce.sh >"$tmp/fail.out" 2>&1
failure_status=$?
set -e
if [[ $failure_status -eq 0 ]]; then echo 'failed witness was accepted' >&2; exit 1; fi
if [[ -d .proofbound/delivery.lock ]]; then echo 'lock was not cleaned after failure' >&2; exit 1; fi
if grep -Eq '^go (all|enforce)$' "$tmp/log"; then
  echo 'failure did not stop before sync or enforcement' >&2
  exit 1
fi

: >"$tmp/log"
set +e
PROOFBOUND_TEST_FAIL_TARGET=enforce bash scripts/delivery-enforce.sh >"$tmp/readiness.out" 2>&1
readiness_status=$?
set -e
if [[ $readiness_status -eq 0 ]]; then echo 'blocked readiness gate was accepted' >&2; exit 1; fi
if ! grep -Eq '^go all$' "$tmp/log" || ! grep -Eq '^go enforce$' "$tmp/log"; then
  echo 'delivery boundary did not sync the complete chain before readiness enforcement' >&2
  exit 1
fi
if [[ -d .proofbound/delivery.lock ]]; then echo 'lock was not cleaned after readiness block' >&2; exit 1; fi
