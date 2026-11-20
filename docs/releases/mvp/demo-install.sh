#!/usr/bin/env bash
# MVP step 1: install Proofbound the way a stranger would, then reach a first gate PASS.
#
#   bash docs/releases/mvp/demo-install.sh [workdir]   # default: ${TMPDIR:-/tmp}/proofbound-install-demo
#
# Pushes are held, so nothing is published: this builds the release archives locally
# (scripts/release-build.sh), serves them to install.sh as a file:// "release", installs into
# <workdir>/bin, and runs the Release 2 loop (docs/releases/r2/demo-loop.sh) with the INSTALLED
# binary: no go build in the loop. It times install -> PASS. Building the archives needs go, GNU tar
# and sha256sum, so this demo runs on Linux (or WSL); using the archives needs none of them. Needs the embedded database's port 55432 free. Walkthrough: QUICKSTART.md.
set -euo pipefail
repo=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
work=${1:-${TMPDIR:-/tmp}/proofbound-install-demo}
marker=.proofbound-install-demo
if [[ -e $work ]]; then
  [[ -f $work/$marker ]] || { echo "demo-install: $work exists and is not a demo workdir; pass another path" >&2; exit 1; }
  rm -rf "$work"
fi
mkdir -p "$work"
touch "$work/$marker"
say() { printf '\n== %s\n' "$*"; }

case $(uname -m) in x86_64 | amd64) arch=amd64 ;; aarch64 | arm64) arch=arm64 ;; *) echo "unsupported machine" >&2; exit 1 ;; esac
case $(uname -s) in Linux) os=linux ;; Darwin) os=darwin ;; *) echo "unsupported system" >&2; exit 1 ;; esac

say "maintainer side: build the release archives (all five targets) and SHA256SUMS"
bash "$repo/scripts/release-build.sh" --allow-dirty v0.0.0-demo "$work/release"
cat "$work/release/SHA256SUMS"

say "user side: install with install.sh (a file:// release stands in for GitHub; nothing needs go from here)"
start=$(date +%s)
PROOFBOUND_BASE_URL="file://$work/release" PROOFBOUND_INSTALL_DIR="$work/bin" sh "$repo/install.sh"
"$work/bin/proofbound" version

say "user side: the Release 2 loop with the installed binary (init, spec, claim, review, PASS)"
PROOFBOUND_BIN="$work/bin/proofbound" bash "$repo/docs/releases/r2/demo-loop.sh" openspec "$work/loop" >"$work/loop.log" 2>&1 || { cat "$work/loop.log"; exit 1; }
grep -E "using the installed|state=PASS" "$work/loop.log"
end=$(date +%s)

say "result"
echo "install -> first PASS: $((end - start))s (target: under 300s)"
((end - start < 300)) || { echo "demo-install: slower than 5 minutes" >&2; exit 1; }
echo "Workdir: $work (delete it when done). Full loop log: $work/loop.log"
