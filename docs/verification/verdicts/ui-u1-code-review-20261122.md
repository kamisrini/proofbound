---
schema: vera.verdict.v1
verdict_id: ui-u1-code-review-20261122
status: NEEDS_WORK
reviewed_commit: 381edea8b57ca94b061c27894782672d025f9ce3
findings:
  - finding_id: ui-u1-code-review-20261122-F1
    severity: MED
  - finding_id: ui-u1-code-review-20261122-F2
    severity: MED
  - finding_id: ui-u1-code-review-20261122-F3
    severity: MED
  - finding_id: ui-u1-code-review-20261122-F4
    severity: MED
  - finding_id: ui-u1-code-review-20261122-F5
    severity: LOW
  - finding_id: ui-u1-code-review-20261122-F6
    severity: LOW
  - finding_id: ui-u1-code-review-20261122-F7
    severity: LOW
  - finding_id: ui-u1-code-review-20261122-F8
    severity: LOW
  - finding_id: ui-u1-code-review-20261122-F9
    severity: LOW
artifact_path: docs/verification/verdicts/ui-u1-code-review-20261122.md
artifact_sha: 03daf1f691e7c629144d62b54602c6df034bc47309ad574afda49f4b43debeff
---
## Provenance
Non-author code reviewer (Build Law 9), one round (Law 11). I reviewed the staged diff (`git diff --cached`) on top of
reviewed_commit 381edea: ui.go, audit.go, acts.go, ui_test.go, the cli.go routing, uikit/audit.html and
uikit/static/app.js against docs/plans/ui-u1-audit-room-spec.md (draft 2) and CLI-INV-37..41. I ran
`go vet ./internal/cli/` (clean) and the five named tests (pass). I did not run `make check` or start Postgres. I
also tried `git diff-tree --cc --name-status -z` on a scratch merge (clean merge printed nothing, as the code assumes).

## Status
NEEDS_WORK: four MED findings (one untrue "exact record" claim, one dead feature, two test gaps); no HIGH. The
Host, cookie and Origin guards, canonical form, key-at-base, author rule, lock and template escaping all read correct.

## Findings
- **F1 (MED).** ui.go:serveAct (sign path) and acts.go:buildAct (attestation) — the UI says "This exact record will be signed", but sign recomputes the audit (after `sync all`), so an attestation's `head`, `transcript_sha256` and `headline` can differ from what the preview showed; only `at` is carried. The signer attests to a digest they never saw. Fix: have the preview return `transcript_sha256` and `head`, send them back on sign, and refuse with "the record changed since the preview" when the rebuilt fields differ; test it.
- **F2 (MED).** acts.go:readAttestations / writeAndCommit — "Current" can never be true through the UI's own flow: signing commits the attestation, so HEAD moves one past the attested head (and the commit adds to Counts.Total, changing the transcript). The page therefore always says "1 commit(s) since", and the `Current` branch (CLI-INV-41) is dead and untested. Fix: judge currency by recomputing the transcript for the attested base..head (or treat commits that only add `.proofbound/{outcomes,attestations}` as not moving it), and add a test that signs, commits and reads back Current.
- **F3 (MED).** ui_test.go:TestOutcomeRecords "retagged" case — cannot fail on its purpose. It changes `schema` to the attestation schema, so `verifyRecord(forged, outcomeSchema)` is rejected by the schema check before any signature is checked; deleting `recordTag` from sign and verify would pass every test. Fix: sign the outcome's own fields (schema unchanged) under the attestation tag (or no tag) and assert verify fails; that is the replay the tag exists to stop.
- **F4 (MED).** ui_test.go — rows CLI-INV-39/40/41 claim behaviours no test reaches. Untested: the read-back rules in `attachOutcomes` (author rule, exception/commit mismatch, fixed_by, unknown outcome, key-at-base via `signerKeys`; the in-range-key test only goes through `buildAct`, so switching `signerKeys` to HEAD would pass); `readAttestations`; the whole of `serveAct`/`signTime` wiring/`writeAndCommit` (commit only that path, refuse an existing record, undo on git failure, preview then sign); the 400 for a non-ancestor base and the default base (INV-37). TestUIRequestGuards only reaches `serveAct` for the oversize body. Fix: add tests that commit forged records (author-signed, wrong commit, in-range key, bad fixed_by) and expect "does not verify", and an end-to-end preview/sign/commit/replay/undo test on a scripted repo with a stubbed `compute`.
- **F5 (LOW).** acts.go:37-38 and verifyRecord — `sha40RE` and `hex64RE` are declared and never used, so read-back does not check field formats. `fixed_by` (acts.go attachOutcomes -> fixedByOK) and an attestation's `head` (`rev-list v.Head+"..HEAD"`) reach git without `--end-of-options`, so a signed record can carry `-...` or a symbolic name such as `HEAD` (accepted as "a later commit"). Needs a trusted signer, hence LOW. Fix: validate sha40/hex64 for commit, fixed_by, base, head, transcript_sha256, key in verifyRecord.
- **F6 (LOW).** uikit/audit.html — "Record an outcome" and "Countersign" are shown enabled when `Actor.Problem` is set; the spec says every act button is disabled. The server refuses, so it is cosmetic. Fix: `{{if not .Actor.Name}}disabled{{end}}`.
- **F7 (LOW).** acts.go:writeAndCommit / audit.html — a record at HEAD that does not verify (or a stale one) can never be replaced and the page offers no button, so that exception is closable only by a manual `git rm` commit. Fix: say so on the page ("remove it with git, then record again"), or allow replacing an unverifying record.
- **F8 (LOW).** ui.go:ServeHTTP — cookies are not port-scoped, so the browser sends the `proofbound_ui` token to any other service on 127.0.0.1/localhost, which could replay it with forged Host/Origin; only a shared-machine threat, but the spec names "another local user". Fix: state the residual risk in section 4, or shorten the cookie's life (`Max-Age`) and rotate the token per sign.
- **F9 (LOW).** audit.go/ui_test.go — `computeAudit` (C1..C7 states, empty range NONE, class sum), merge commits in `trustPaths`, and `checkReason` accepting U+2028/2029 (Zl/Zp; printableUnit covers plain text, templates escape HTML) are untested or unrefused. Recorded only; the audit is accepted by the tour run per spec section 7.
