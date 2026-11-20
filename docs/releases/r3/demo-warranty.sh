#!/usr/bin/env bash
# R3 slice 1: the portable signed warranty, for either kit:
#
#   bash docs/releases/r3/demo-warranty.sh openspec [workdir]   # default: ${TMPDIR:-/tmp}/proofbound-<kit>-warranty
#   bash docs/releases/r3/demo-warranty.sh speckit  [workdir]
#
# It runs the Release 2 loop (claim -> review -> gate PASS) and then:
#   1. creates a signing key (in a throwaway config dir, never in the repository) and commits its
#      public half, which is how the repository declares the key;
#   2. exports a signed warranty for the PASS commit;
#   3. verifies it without the repository: UNVERIFIED (a signature alone proves nothing);
#   4. verifies it against a fresh clone, with no database: VALID;
#   5. tampers with one obligation in the file: INVALID;
#   6. pins the wrong key fingerprint: INVALID;
#   7. tries to export with a key the repository never declared: refused.
# Needs git, go, jq and the embedded database's port 55432. Nothing is fetched or pushed.
set -euo pipefail

kit=${1:-}
[[ $kit == openspec || $kit == speckit ]] || { echo "usage: $0 {openspec|speckit} [workdir]" >&2; exit 2; }
repo=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
work=${2:-${TMPDIR:-/tmp}/proofbound-$kit-warranty}

say() { printf '\n== %s\n' "$*"; }

say "Release 2 loop: claim, review, gate PASS (docs/releases/r2/demo-loop.sh)"
bash "$repo/docs/releases/r2/demo-loop.sh" "$kit" "$work" >"$work.loop.log" 2>&1 || { cat "$work.loop.log"; exit 1; }
grep -E "state=PASS" "$work.loop.log"
rm -f "$work.loop.log"

pb() { "$work/proofbound" "$@"; }
cd "$work/project"
commit=$(git log --format=%H --grep='^implement CSV export' -1)
export XDG_CONFIG_HOME="$work/config"

say "1. a signing key: private part outside the repository, public part committed"
pb keys init --name spec-owner
git add .gitignore .proofbound/keys
PROOFBOUND_NO_INTENT=1 git commit -qm "declare the warranty signing key"

say "2. export a signed warranty for the PASS commit"
pb warranty export "$commit" -o "$work/warranty.json"
jq '{commit: .body.commit, change: .body.change_name, obligations: [.body.requirements[].obligations[].statement]}' "$work/warranty.json"

say "3. verify without the repository: the signature alone proves nothing"
pb warranty verify "$work/warranty.json" || echo "(exit $?)"

say "4. verify against a fresh clone, anywhere, with no database"
git clone -q "$work/project" "$work/clone"
(cd "$work" && ./proofbound warranty verify warranty.json --repo clone)

say "5. tamper with one obligation in the file"
sed 's/System MUST\|The system SHALL/The system MAY/' "$work/warranty.json" >"$work/tampered.json"
pb warranty verify "$work/tampered.json" --repo "$work/clone" || echo "(exit $?)"

say "6. pin a different key fingerprint"
pb warranty verify "$work/warranty.json" --repo "$work/clone" --trust-key 0000000000000000 || echo "(exit $?)"

say "7. export with a key the repository never declared"
XDG_CONFIG_HOME="$work/other-config" pb keys init --name intruder
rm .proofbound/keys/intruder.pub
XDG_CONFIG_HOME="$work/other-config" pb warranty export "$commit" -o "$work/forged.json" || echo "(refused, exit $?)"

cat <<MSG

== what this shows ($kit)
- A commit that passed independent review became one signed file, warranty.json.
- Anyone with that file and a clone of the repository (from a source they trust) can check it with
  no database: the promises are re-read from git at that commit, the review files must be committed
  byte for byte, and the signing key must be one the repository declares.
- Without the repository the answer is UNVERIFIED, never VALID. Any edit, a wrong key, or an
  undeclared key is INVALID or refused.
Workdir: $work (delete it when done).
MSG
