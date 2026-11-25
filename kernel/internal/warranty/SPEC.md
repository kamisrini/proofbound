# warranty — SPEC (R3 slice 1: the portable signed warranty)

Contract source: [`docs/plans/R3-warranty-export-spec.md`](../../../docs/plans/R3-warranty-export-spec.md)
draft 2. It was accepted at round 2 (`r3-warranty-spec-recheck-20261119`) under
`CI-warranty-slice1-7t4iok` → `BR-portable-warranty-oeffoe` O-1..O-5, with the decision record
`VD-r3-warranty-first-s59ktv`. This file states the interface and the invariants. The plan holds the
rationale and the review dispositions (one home per datum).

**MVP step 3 (reviewer proof)** extends it under `CI-mvp-reviewer-sign-5tqv2c` →
`BR-vertical-mvp-xc31gp` O-3; contract [`docs/plans/mvp-3-reviewer-signed-spec.md`](../../../docs/plans/mvp-3-reviewer-signed-spec.md)
draft 3 (ACCEPTABLE at spec round 2). It adds check 9 and the rows W-INV-9..12 below, and amends
W-INV-1, 2 and 3.

## Interface

- `Body`, `Signature`, `File`; `Schema = "proofbound.warranty.v1"`.
- `Sign(body, key) ([]byte, error)`: the file bytes. The signed bytes are
  `Schema + "\n" + core.Canonicalize(body)`.
- `Parse(data) (File, error)`: strict. It rejects duplicate keys, unknown fields, trailing data, and a
  body that does not round-trip to the same canonical bytes.
- `File.CheckSignature()`; `CheckRule(body)` (the readiness rule on the body alone).
- `Verify(ctx, file, repo, trustKey)`. It returns `nil` (VALID), `ErrUnverified` (when `repo == ""`), or
  the first failing check, in the plan's §5 order:
  1. signature;
  2. rule;
  3. pin;
  4. commit in the clone;
  5. HEAD descends from the commit;
  6. key declared at HEAD;
  7. re-derivation through `sdd.StateRecords` (SDD-INV-72);
  8. every cited artifact at HEAD re-parses to exactly the cited record.
  9. (MVP step 3) every named reviewer proven by their declared key, through `.sig` files beside their
     artifacts (A) or a cosignature (B); else `ErrUnreviewed`. INVALID anywhere wins.
- `VerifyReport` is `Verify` plus the proven reviewers. `File.Cosignatures`; `Cosign(data, key)`;
  `SignArtifact(data, key)`; `KeyName(name)`; `ErrUnreviewed`.
- `KeyPath()` (`$PROOFBOUND_KEY` verbatim when set), `InitKey(path)`, `LoadKey(path)`, `PubLine(pub)`,
  `Fingerprint(pub)`; `KeysDir`.

The CLI commands `keys init`, `warranty export` and `warranty verify` live in `internal/cli/warranty.go`.
`warranty verify` needs no repository root and no database, and exits 0 VALID, 2 UNVERIFIED,
3 UNREVIEWED, 1 INVALID. `warranty cosign` (MVP step 3) lives there too.

## Invariants

| ID | Invariant | Proving test |
|---|---|---|
| W-INV-1 | Changing any byte of the body, or the signature, fails verification. Duplicate keys, unknown fields (body or outer) and non-canonical bodies are malformed. The outer object may also hold `cosignatures`, validated by W-INV-11 (MVP step 3) | warranty_test.go::TestSignatureCoversEveryByte |
| W-INV-2 | `warranty export` writes nothing unless the commit's readiness gate is PASS and its own `Verify` of the signed result against this repository is VALID or, amended for MVP step 3, UNREVIEWED (printing a warning). Proven end to end by the sdd package's `TestWarrantyEndToEnd` (export before review is refused and writes no file) | — |
| W-INV-3 | `CheckRule` holds on the body alone. A body re-signed by the same key fails when any target obligation lacks a SATISFIED verdict outcome or a VERIFIABLE review outcome for its exact requirement revision, a reviewer is the requirement's declared owner, a verdict is for another commit or change-intent revision, or a requirement is missing, untargeted, or lacks a targeted obligation. Amended for MVP step 3: identity is `KeyName`, so the rule also fails on an empty reviewer or owner key name, a reviewer whose key name is the owner's, and two spellings of one key name (also proven by reviewer_test.go::TestRuleUsesKeyNames). Amended 2026-11-24 (`CI-needs-work-blocks-h4r8wq`): the rule also fails on a verdict whose status is not ACCEPTABLE, so a warranty citing a NEEDS_WORK verdict is INVALID in `warranty verify` and `ui --bundle` | warranty_test.go::TestRuleRejectsReSignedForgeries |
| W-INV-4 | Without a repository, a good file is `ErrUnverified`, never VALID. A `--trust-key` pin that does not match fails even then | warranty_test.go::TestNoRepositoryIsNeverValid |
| W-INV-5 | With a repository, verification fails when the change intent or any requirement differs from what `sdd.StateRecords` derives from git at the commit (compared as whole canonical records), or when any cited review or verdict is missing at HEAD, edited, or parses to a different record. Proven end to end by the sdd package's `TestWarrantyEndToEnd` (a reworded obligation re-signed by the declared key; an edited review at HEAD) | — |
| W-INV-6 | With a repository, verification fails unless HEAD descends from the commit and HEAD commits, under `.proofbound/keys/`, a `.pub` regular file whose content is exactly `ed25519 <key hex>\n`. Proven end to end by the sdd package's `TestWarrantyEndToEnd` (the key removed at HEAD) and by `docs/releases/r3/demo-warranty.sh` (an undeclared key is refused at export) | — |
| W-INV-7 | The package's own imports exclude `internal/store`, pgx and the network packages | warranty_test.go::TestNoDatabaseOrNetworkImports |
| W-INV-8 | `InitKey` never creates a key inside a git work tree and never replaces an existing key. `LoadKey` refuses a missing file, a symlink, or a mode other than 0600 | warranty_test.go::TestKeyFileSafety |
| W-INV-9 | With a repository, VALID requires every named reviewer proven by their declared key (a regular file at HEAD that is exactly `PubLine`): every artifact of theirs carries a verifying `.sig` (A), or the file carries their cosignature (B). A reviewer with no proof is `ErrUnreviewed`; check 9 evaluates everything first, and any INVALID condition wins over UNREVIEWED | reviewer_test.go::TestReviewerProofByArtifactSignatures |
| W-INV-9a | A present `.sig` is never ignored: one by another key, over other bytes, malformed, a symlink, or present for a reviewer with no declared key is INVALID; a declared key file that is not exactly a key line is INVALID | reviewer_test.go::TestBadArtifactSignaturesAreInvalid |
| W-INV-10 | A reviewer's declared key equal to the exporter's key, a requirement owner's declared key, or another reviewer's declared key is INVALID | reviewer_test.go::TestReviewerKeysMustBeDistinct |
| W-INV-11 | `Parse` accepts `cosignatures` only as 1-16 entries of lowercase hex (64-char key, 128-char sig), strictly sorted by key, no other fields; `[]` and `null` are malformed; absent parses as before | reviewer_test.go::TestCosignatureParsingIsStrict |
| W-INV-12 | A cosignature (over the domain tag, the exporter's key and the canonical body) by a reviewer's declared key proves that reviewer; one carried to another exporter's file does not verify (INVALID even without a repository); one by a key that is no reviewer's declared key is INVALID; `Cosign` refuses the exporter's key and a repeat. Also proven end to end by the sdd package's `TestWarrantyCosignEndToEnd` | reviewer_test.go::TestCosignatureProvesAReviewer |

## Non-goals

These are in the plan's §11:
- multi-signature beyond "every named reviewer" (reviewer signatures arrived in MVP step 3);
- anchors beyond the repository, revocation and rotation, a transparency log;
- non-SDD warranties;
- drift inside verify.

**Trust caveat:** whoever can commit to the repository can declare a key. A verifier must clone from
a source they trust, and may pin `--trust-key`.
