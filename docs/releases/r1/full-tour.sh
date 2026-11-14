#!/usr/bin/env bash
# Proofbound — FULL VALUE TOUR.
#
# This is the narrated view of the demo. It is a thin wrapper over seed-demo-app.sh (the single
# source of truth for the sample app and its seed data) run in --tour mode, so the tour and the
# seeder can never drift apart.
#
#   bash docs/releases/r1/full-tour.sh            # narrate; seed into $HOME/proofbound-demo-app
#   DEMO_DIR=/path bash .../full-tour.sh          # narrate; seed into /path
#   PROOFBOUND_BIN=/path/to/proofbound bash ...    # if proofbound isn't on your PATH
#
# See seed-demo-app.sh for prerequisites and the plain (non-narrated) seeding mode.
set -euo pipefail
dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -n "${DEMO_DIR:-}" ]; then
  exec "$dir/seed-demo-app.sh" --tour "$DEMO_DIR"
else
  exec "$dir/seed-demo-app.sh" --tour
fi
