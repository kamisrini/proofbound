#!/usr/bin/env bash
# Self-test for scripts/gate-expiry.sh (Law 3): prove it fires on an expired/dateless gate,
# does not false-positive on a fresh one, is boundary-inclusive, and fails closed on a missing
# directory. Then run it for real so `make hooks-test` (in `make check`) enforces that none of
# this repo's committed gate definitions is expired.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
hook=scripts/gate-expiry.sh
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

mk() { printf '{"schema":"proofbound.gate.v1","id":"t","expires":"%s","mode":"enforce"}\n' "$1" > "$tmp/t.yaml"; }

# positive control: an expired gate MUST fail
mk 2020-01-01
if GATE_EXPIRY_DIR="$tmp" bash "$hook" 2>/dev/null; then echo "gate-expiry.test: expired gate not caught" >&2; exit 1; fi

# neutral control: a far-future gate MUST pass
mk 2099-01-01
if ! GATE_EXPIRY_DIR="$tmp" bash "$hook" 2>/dev/null; then echo "gate-expiry.test: fresh gate wrongly failed" >&2; exit 1; fi

# boundary: a gate expiring today is still valid today
if ! GATE_EXPIRY_DIR="$tmp" GATE_EXPIRY_TODAY=2099-01-01 bash "$hook" 2>/dev/null; then
  echo "gate-expiry.test: today-expiry boundary wrongly failed" >&2; exit 1; fi

# dateless gate MUST fail closed
printf '{"schema":"proofbound.gate.v1","id":"t","mode":"enforce"}\n' > "$tmp/t.yaml"
if GATE_EXPIRY_DIR="$tmp" bash "$hook" 2>/dev/null; then echo "gate-expiry.test: dateless gate not caught" >&2; exit 1; fi

# missing directory MUST fail closed
if GATE_EXPIRY_DIR="$tmp/absent" bash "$hook" 2>/dev/null; then echo "gate-expiry.test: missing dir not fail-closed" >&2; exit 1; fi

# REAL enforcement: every committed gate in gates/ must be unexpired today
if ! bash "$hook"; then echo "gate-expiry.test: a committed gate definition is expired" >&2; exit 1; fi

echo "ok gate-expiry"
