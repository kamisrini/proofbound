#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d)
# On failure, show every captured workflow output before cleanup deletes it (gate review round 2,
# F3: this self-test used to exit 1 with no output at all).
trap 'status=$?; if ((status)); then for f in "$tmp"/*.out; do [[ -f $f ]] && { echo "--- ${f##*/}"; cat "$f"; }; done >&2; fi; rm -rf "$tmp" .proofbound/delivery.lock' EXIT
bin="$tmp/bin"; mkdir "$bin"
ln -s "$PWD/scripts/tests/delivery-shim.sh" "$bin/make"
ln -s "$PWD/scripts/tests/delivery-shim.sh" "$bin/go"
printf '#!/usr/bin/env bash\nprintf "gate\\n" >>"$PROOFBOUND_TEST_LOG"\n[[ ${PROOFBOUND_TEST_FAIL_TARGET:-} != gate ]]\n' >"$bin/gate"; chmod +x "$bin/gate"
export PATH="$bin:$PATH" PROOFBOUND_TEST_LOG="$tmp/log" PROOFBOUND_INTENT_GATE="$bin/gate"

PROOFBOUND_TEST_SLEEP=2 bash scripts/delivery-enforce.sh >"$tmp/first.out" 2>&1 &
first=$!
for _ in {1..40}; do [[ -d .proofbound/delivery.lock ]] && break; sleep 0.05; done
if [[ ! -d .proofbound/delivery.lock ]]; then echo 'workflow did not acquire lock' >&2; exit 1; fi
if bash scripts/delivery-enforce.sh >"$tmp/second.out" 2>&1; then
  echo 'concurrent workflow was accepted' >&2
  exit 1
fi
wait "$first" || { echo 'first delivery workflow failed' >&2; exit 1; }
if [[ -d .proofbound/delivery.lock ]]; then echo 'lock was not cleaned after success' >&2; exit 1; fi
mapfile -t actual <"$tmp/log"
expected=(
  'make index-check-witnessed'
  'make law-citation-witnessed'
  'make spec-numbering-witnessed'
  'make invariant-table-witnessed'
  'make link-witnessed'
  'make kernel-check-witnessed'
  'make check-witnessed'
  'go all'
  'gate'
  'go enforce'
)
if [[ ${#actual[@]} -ne ${#expected[@]} ]]; then
  printf 'delivery command count mismatch: %s\n' "${actual[*]}" >&2
  exit 1
fi
for i in "${!expected[@]}"; do
  if [[ ${actual[$i]} != "${expected[$i]}" ]]; then
    printf 'delivery order mismatch at %d: got %s want %s\n' "$i" "${actual[$i]}" "${expected[$i]}" >&2
    exit 1
  fi
done

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

# The trailer gate runs after sync and before enforcement, and its failure stops enforcement.
: >"$tmp/log"
set +e
PROOFBOUND_TEST_FAIL_TARGET=gate bash scripts/delivery-enforce.sh >"$tmp/gate.out" 2>&1
gate_status=$?
set -e
if [[ $gate_status -eq 0 ]]; then echo 'failing intent gate was accepted' >&2; exit 1; fi
if ! grep -qx 'gate' "$tmp/log" || grep -qx 'go enforce' "$tmp/log"; then
  echo 'intent gate did not run, or enforcement ran after it failed' >&2; exit 1
fi
if [[ -d .proofbound/delivery.lock ]]; then echo 'lock was not cleaned after gate failure' >&2; exit 1; fi
