---
schema: vera.verdict.v1
verdict_id: mvp-1-install-review-20261120
status: ACCEPTABLE
reviewed_commit: 3e941772675b0c8b996824f5a1992a2b075e8b75
findings:
  - finding_id: mvp-1-install-review-20261120-F1
    severity: MED
  - finding_id: mvp-1-install-review-20261120-F2
    severity: MED
  - finding_id: mvp-1-install-review-20261120-F3
    severity: LOW
  - finding_id: mvp-1-install-review-20261120-F4
    severity: LOW
  - finding_id: mvp-1-install-review-20261120-F5
    severity: LOW
  - finding_id: mvp-1-install-review-20261120-F6
    severity: LOW
  - finding_id: mvp-1-install-review-20261120-F7
    severity: LOW
  - finding_id: mvp-1-install-review-20261120-F8
    severity: LOW
artifact_path: docs/verification/verdicts/mvp-1-install-review-20261120.md
artifact_sha: 344f7b2474b176bba64a597f5f0f9bd9a1848229adee8d3497457b1fe78af16c
---
## Provenance
Non-author reviewer (Law 9): a separate reviewer agent (Sonnet) that did not write the change. One
combined spec + code round (Law 11, `VD-vertical-mvp-first-ai3dhn`) on MVP step 1, the staged change on
top of `reviewed_commit`, committed with this verdict under `CI-mvp-install-k99zpl`. It ran both shell
tests and `TestVersionRunsAnywhere` (pass); shellcheck was not installed. Findings are condensed from
its report; the author (scribe) records the fixes.

## Status
ACCEPTABLE: no HIGH. Both MEDs fixed before commit (Law 11).

## Findings
- **F1 MED, fixed (`release-build.sh` go build).** Go's VCS stamping was on, so `vcs.modified` made a
  clone with an untracked file build different bytes, breaking REL-2 across machines. Fix:
  `-buildvcs=false`; `release-build.test.sh` now fails if the binary embeds any `vcs.*` stamp
  (calibrated: removing the flag fails the test).
- **F2 MED, fixed (`release-build.sh`, demo, QUICKSTART §4).** GNU-only tar flags and `sha256sum`, while
  the demo implied macOS worked. Fix: the script refuses without GNU tar and `sha256sum` and says
  "build releases on Linux"; the demo and QUICKSTART §4 say Linux/WSL.
- **F3 LOW, fixed with F2's rewrite.** QUICKSTART §4 said "2 minutes, nothing to prepare"; it now
  names Go, GNU tar and port 55432 and says the 27 s covers install to PASS only.
- **F4 LOW, recorded.** `trap … EXIT INT TERM` in dash resumes after INT; fail-closed under `set -e`.
- **F5 LOW, recorded.** `echo "installed $(… version)"` ignores a failing binary's exit status.
- **F6 LOW, recorded.** No `main` wrapper against a truncated `curl | sh` download (impact bounded).
- **F7 LOW, recorded.** Tests do not exercise the `shasum` branch, `*file` lines, or owner/umask leakage.
- **F8 LOW, recorded.** Rosetta reports x86_64 on Apple Silicon; `PROOFBOUND_VERSION` is unvalidated
  in the URL (user-controlled).
