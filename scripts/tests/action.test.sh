#!/usr/bin/env bash
# ACT-1 of docs/plans/mvp-2-pr-check-spec.md: the action body syncs, then checks base..HEAD, writes
# Markdown to the job summary, and exits with the check's code; a missing base is a usage error.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
script=$PWD/scripts/action-check.sh
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
git -C "$tmp" init -q -b main
git -C "$tmp" -c user.name=f -c user.email=f@example.invalid commit -q --allow-empty -m one
mkdir "$tmp/bin"
# A stub proofbound: logs its arguments; `check` exits $STUB_CODE and prints a format marker.
cat >"$tmp/bin/proofbound" <<'STUB'
#!/usr/bin/env bash
echo "$*" >>"$STUB_LOG"
if [[ $1 == check ]]; then
  [[ " $* " == *" --markdown "* ]] && echo "### markdown" || echo "plain"
  exit "$STUB_CODE"
fi
STUB
chmod +x "$tmp/bin/proofbound"
run() { (cd "$tmp" && env PATH="$tmp/bin:$PATH" STUB_LOG="$tmp/log" GITHUB_STEP_SUMMARY="$tmp/summary" "$@" bash "$script"); }
fail() { echo "action.test: $*" >&2; exit 1; }

: >"$tmp/log"; : >"$tmp/summary"
run PB_BASE=main STUB_CODE=0 >"$tmp/out" || fail "a passing check failed the step"
[[ $(cat "$tmp/log") == $'sync all\ncheck main\ncheck main --markdown' ]] || fail "calls: $(cat "$tmp/log")"
grep -q plain "$tmp/out" && grep -q '### markdown' "$tmp/summary" || fail "log or summary missing"

: >"$tmp/log"
code=0; run PB_BASE= PB_DEFAULT_BASE=main PB_REQUIRE_CLAIMS=true STUB_CODE=1 >/dev/null || code=$?
[[ $code == 1 ]] || fail "a failing check exited $code"
grep -qx 'check main --require-claims' "$tmp/log" || fail "require-claims not passed: $(cat "$tmp/log")"

: >"$tmp/log"
code=0; run PB_BASE= PB_DEFAULT_BASE= STUB_CODE=0 >/dev/null 2>&1 || code=$?
[[ $code == 2 && ! -s $tmp/log ]] || fail "no base: exit $code, calls $(cat "$tmp/log")"
code=0; run PB_BASE=no-such-ref STUB_CODE=0 >/dev/null 2>&1 || code=$?
[[ $code == 2 && ! -s $tmp/log ]] || fail "unknown base: exit $code"
echo "action.test: ok"
