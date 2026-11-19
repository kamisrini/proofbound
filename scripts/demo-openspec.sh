#!/usr/bin/env bash
# R2 real-repo demo (Law 11 payoff): Proofbound warrants a real OpenSpec project.
#
#   scripts/demo-openspec.sh [workdir]      # default: ${TMPDIR:-/tmp}/proofbound-openspec-demo
#
# 1. Clones Fission-AI/OpenSpec (OpenSpec is built with OpenSpec) at a pinned commit.
# 2. Rewrites the LOCAL clone so every commit carries .proofbound/sdd-profile.json. The R2 contract
#    (BR-sdd-warranty-i9miv7 O-3) takes owners only from a COMMITTED profile, so the profile is put
#    into history rather than supplied from outside; commit ids in the clone differ from upstream.
# 3. Syncs every spec revision in the project's real history.
# 4. Adds two local demo commits: one claims a real active change (Intent: sdd:<change>), one claims
#    a change that does not exist; adds a readiness gate for each.
# 5. Prints what Proofbound knows: the claim results, the gate verdicts, and the warranty report.
# Nothing is pushed anywhere. Needs git, go, network access to github.com, and the embedded
# database's port 55432 free.
set -euo pipefail

repo=$(cd "$(dirname "$0")/.." && pwd)
work=${1:-${TMPDIR:-/tmp}/proofbound-openspec-demo}
upstream=https://github.com/Fission-AI/OpenSpec.git
pin=e6539de023915e36848b351ea737ca8e192e8538 # upstream main, 2026-11-19
change=install-skills-globally               # an active change at the pin
intent=CI-install-skills-globally-57r52f     # the sdd connector's change-intent id for it (sdd ids.go)
marker=.proofbound-openspec-demo

if [[ -e $work ]]; then
  [[ -f $work/$marker ]] || { echo "demo-openspec: $work exists and is not a demo workdir; pass another path" >&2; exit 1; }
  rm -rf "$work"
fi
mkdir -p "$work"
touch "$work/$marker"
unset DATABASE_URL

say() { printf '\n== %s\n' "$*"; }
# The embedded database prints its server log on stdout; keep only Proofbound's own lines.
pb() { "$work/proofbound" "$@" 2>/dev/null | grep -Ev 'LOG: |FATAL: |^Data page checksums|^waiting for server|^ ?done$|^server (started|stopped)$|^(selecting|creating|running|performing|syncing|fixing|initdb|The |This |You can|Success|    |$)'; }

say "building the proofbound CLI"
(cd "$repo/kernel" && go build -o "$work/proofbound" ./cmd/proofbound)

say "cloning OpenSpec at ${pin:0:12}"
git clone -q "$upstream" "$work/OpenSpec"
cd "$work/OpenSpec"
git -c advice.detachedHead=false checkout -q "$pin"
git checkout -q -b proofbound-demo
export GIT_AUTHOR_NAME=proofbound-demo GIT_AUTHOR_EMAIL=demo@example.invalid
export GIT_COMMITTER_NAME=proofbound-demo GIT_COMMITTER_EMAIL=demo@example.invalid

say "committing the OpenSpec profile into every commit of the local clone (about a minute)"
profile=$(cat <<'EOF'
{
  "schema": "proofbound.sdd-profile.v1",
  "kit": "openspec",
  "requirements": { "paths": ["openspec/changes/*/specs/*/spec.md", "openspec/changes/archive/*/specs/*/spec.md"] },
  "changes": { "mode": "delta-sections", "root": "openspec/changes", "archive": "openspec/changes/archive", "status": "accepted" },
  "owners":   [ { "pattern": "openspec/**", "owner": "openspec-maintainers" } ],
  "sponsors": [ { "pattern": "openspec/changes/**", "sponsor": "openspec-maintainers" } ]
}
EOF
)
blob=$(printf '%s\n' "$profile" | git hash-object -w --stdin)
FILTER_BRANCH_SQUELCH_WARNING=1 git filter-branch -f --index-filter \
  "git update-index --add --cacheinfo 100644,$blob,.proofbound/sdd-profile.json" proofbound-demo >/dev/null 2>&1
echo "commits: $(git rev-list --count HEAD)"

say "syncing every spec revision in OpenSpec's history"
pb sync intent sdd

say "two demo commits: one claims a real change, one claims a change that does not exist"
echo "<!-- demo work for $change -->" >>README.md
git commit -qam "demo: implement $change" -m "Intent: sdd:$change"
claimed=$(git rev-parse HEAD)
echo "<!-- demo: unrelated work -->" >>README.md
git commit -qam "demo: claim a change that does not exist" -m "Intent: sdd:no-such-change"
bogus=$(git rev-parse HEAD)
mkdir -p gates
for g in "claimed:$claimed" "bogus:$bogus"; do
  printf '{"schema":"proofbound.gate.v1","id":"%s","description":"readiness of a demo commit","expires":"2099-01-01","mode":"canary","rule":"intent-delivery-readiness","scope_commit":"%s"}\n' "${g%%:*}" "${g#*:}" >"gates/${g%%:*}.yaml"
done
git add gates && git commit -qm "demo: readiness gates for the two demo commits"
echo "claimed=$claimed (Intent: sdd:$change)"
echo "bogus=$bogus (Intent: sdd:no-such-change)"
pb sync git

say "readiness gate: is each demo commit warranted?"
pb gates canary

say "warranty report for the claimed change (every obligation comes from OpenSpec's own spec)"
pb report intent "$intent"

cat <<EOF

== what this shows
- Proofbound read OpenSpec's real history and recorded every spec revision, with owners from the
  committed profile and exact content digests.
- A commit that claims a real change binds to that change's exact revision. Its gate is BLOCKED
  because no one has verified the obligations yet (state IMPLEMENTED_UNVERIFIED): the warranty
  refuses to vouch for unverified work.
- A commit that claims a change that does not exist is reported (claim-skip unknown-change) and
  BLOCKED.
Workdir: $work (delete it when done; the demo database lives inside it).
EOF
