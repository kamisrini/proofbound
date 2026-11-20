---
schema: vera.verdict.v1
verdict_id: r3-warranty-code-review-20261119
status: ACCEPTABLE
reviewed_commit: f9c8593c3365ff40a69d9725daf1bbcda96be476
findings:
  - finding_id: r3-warranty-code-review-20261119-F1
    severity: LOW
  - finding_id: r3-warranty-code-review-20261119-F2
    severity: LOW
  - finding_id: r3-warranty-code-review-20261119-F3
    severity: LOW
  - finding_id: r3-warranty-code-review-20261119-F4
    severity: LOW
  - finding_id: r3-warranty-code-review-20261119-F5
    severity: LOW
  - finding_id: r3-warranty-code-review-20261119-F6
    severity: LOW
  - finding_id: r3-warranty-code-review-20261119-F7
    severity: LOW
  - finding_id: r3-warranty-code-review-20261119-F8
    severity: LOW
artifact_path: docs/verification/verdicts/r3-warranty-code-review-20261119.md
artifact_sha: af9cdb15191108a5bee9907657e27badb01d13d6e996d68e72442766a78a481d
---
## Provenance
Non-author reviewer (Law 9): a separate reviewer agent (Sonnet) that did not write the code. One code
round (Law 11) on R3 slice 1: the staged implementation on top of `reviewed_commit`, committed together
with this verdict under `CI-warranty-slice1-7t4iok`. It ran `go test ./internal/warranty` (pass) and
`go vet`. It did not run the sdd or embedded-Postgres tests. Findings are condensed from its report.

## Status
ACCEPTABLE: no HIGH or MED findings.

## Findings (all LOW, recorded and not fixed: Law 11)
- **F1 LOW (`warranty.go` CheckRule).** The rule never requires at least one targeted obligation,
  although the gate requires targets > 0. A target with empty `ObligationIDs` would pass with
  `obligations=0`. The sdd mapper never emits one, so this is defence in depth.
- **F2 LOW (`keys.go` insideWorkTree).** It fails open: a git error (git missing, dubious ownership,
  `GIT_DIR` set) reads as "not a work tree", and it does not strip `GIT_*`. W-INV-8 could then be
  bypassed when git refuses to answer.
- **F3 LOW (`keys init`).**
  - The private key is written before the `.pub`, so a failed `.pub` write orphans the key.
  - A second repository on the same machine dead-ends on "exists" with no way to print its `.pub`.
- **F4 LOW (`cli.go`).** A usage error exits 2, the same code as UNVERIFIED. It is never read as VALID.
- **F5 LOW (`verify.go`).** Verify does not check that the commit's message claims
  `Intent: sdd:<change_name>`, which export checks. A verifier proves that the tree holds the change
  and its reviews, not that the commit claimed it. Only the trusted signer can exploit this.
- **F6 LOW (rule keys).** `Source` and `Relation` are ignored, and a NOT_SATISFIED outcome beside a
  SATISFIED one is ignored. Both match the gate's EXISTS semantics. The SPEC should say so.
- **F7 LOW (tests).** The forgery cases assert only `err != nil`, not the expected reason. W-INV-2,
  W-INV-5 and W-INV-6 rest on the sdd e2e test, which the reviewer did not run.
- **F8 LOW (demo step 5).** The tamper `sed` asserts nothing. If the pattern missed, the demo would
  print VALID under a "tamper" heading and still exit 0.

## What was checked and found sound
- **Strict parse.** It rejects duplicate keys, unknown fields and non-round-tripping shapes.
- **Signed bytes.** The signature covers the domain-tagged canonical body.
- **Requirement order** is normalised.
- **Binding.** Each verdict is bound to the commit and the change-intent revision, and each review to
  the requirement revision. A reviewer who is the owner is refused.
- **Re-derivation** compares whole canonical records.
- **The key** must be a direct non-symlink `.pub` blob at HEAD, and HEAD must descend from the commit.
  Cited artifacts must be non-symlink blobs that parse to the same record.
- **Isolation.** `git()` strips `GIT_*`.
- **Export** self-verifies before writing, with O_EXCL and removal of partial writes.
- **Key creation and loading.** `InitKey` uses O_EXCL with 0600. `LoadKey` refuses symlinks and any
  mode other than 0600.
- **Flag parsing** is correct, and `warranty verify` runs before the repository-root lookup and
  opens no store.

## Author's evidence (not part of the verdict)
- Four hand mutants were all killed:
  - the owner check removed;
  - no-repo returning VALID;
  - re-derivation skipped;
  - the key check skipped.
- `TestWarrantyEndToEnd` and both demos (openspec and speckit) ran clean.
