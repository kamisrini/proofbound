# MVP step 3: the reviewer-signed warranty — spec

**Status:** draft 3, answering spec round 1 (NEEDS_WORK: F1 HIGH, F2-F6 MED, all answered in §9; LOWs F7-F9 taken where cheap). Second and last spec round (Law 11). The owner chose 2026-11-20: **"first A and
then B"** — build A (the reviewer signs at `review seal`), then B (the reviewer cosigns the warranty
file); either proves the reviewer. Both touch signing and verification, so `internal/warranty` gets a
calibrated mutation sweep (`VD-vertical-mvp-first-ai3dhn`). **Intent:** `CI-mvp-reviewer-sign-5tqv2c`
→ `BR-vertical-mvp-xc31gp` O-3.

## 1. Payoff

```
$ proofbound warranty verify warranty.json --repo ./clone
VALID key=f32d609d003537e9 commit=406704b… change=sdd:add-csv-export obligations=2 reviewers=alice:9c1e44ab07d2f310
```

Today VALID means "the exporter signed a body whose rule holds". The rule checks that the reviewer
*named* in the verdicts is not the owner, but that name is text the exporter could write. After this
step, VALID also requires **each named reviewer's own declared key to have signed**.

## 2. Keys and names (shared by A and B)

- **One key per person per machine**, at `KeyPath()`. New: `PROOFBOUND_KEY=<path>` overrides it (for
  a second identity on one machine, e.g. a reviewer agent). All W-INV-8 safety rules still apply.
- **A name's key** is the blob committed at HEAD as `.proofbound/keys/<KeyName(name)>.pub`, where
  `KeyName` lower-cases and turns every run of characters outside `[a-z0-9-]` into `-`, trimming `-`
  (the slug `keys init --name` already uses, moved into `internal/warranty`).
- **Reviewer** = each distinct `declared_reviewer` across the body's reviews and verdicts.
- **Identity is `KeyName`** (round 1 F1). `CheckRule` (body only, W-INV-3) additionally rejects: a
  reviewer or owner name whose `KeyName` is empty; a reviewer whose `KeyName` equals the `KeyName` of
  the requirement's declared owner (so `Alice` cannot review `alice`'s requirement); and two different
  raw reviewer spellings with one `KeyName` in one body. `review seal` refuses the same cases.
- **A declared key** is a regular file (not a symlink) committed at HEAD whose bytes are exactly
  `PubLine(key)` = `ed25519 <64 lowercase hex>\n`; anything else at that path is INVALID. Seal reads the
  working tree with `Lstat` and the same exact match.

## 3. Option A: the reviewer signs at `review seal`

- For every artifact it seals (requirement reviews and obligation verdicts, not the gate file),
  `review seal` also writes `<artifact path>.sig`, one line:
  `ed25519 <public key hex> <signature hex>\n`, the signature over
  `"proofbound.review-signature.v1\n" + <artifact bytes>`.
- **When it signs:** the working tree declares a key for `KeyName(reviewer)` and the local key
  (`LoadKey(KeyPath())`) is that key. **When it refuses (writes nothing):** a key is declared but the
  local key is missing, unloadable (mode, symlink, bad seed: never a silent fallback) or different
  ("this machine is not <reviewer>"); the declared file is not an exact key line. **Otherwise** (no key
  declared for the reviewer) it writes the artifacts unsigned and prints
  `unsigned: no key declared for <reviewer>; warranties for this commit will verify UNREVIEWED`.
- `PROOFBOUND_KEY` is used verbatim when set: a missing file is an error, and `keys init` applies
  W-INV-8's work-tree refusal to it.
- The `.sig` files are part of the same all-or-nothing write and existence check (CLI-INV-20) and are
  committed with the review. The reviews connector and this repository's verdict scanners read only
  `.md` files, so the ledger is unchanged.
- A signature covers only its artifact's bytes. A requirement review names no commit, so a review and
  its `.sig` can be replayed into another repository that declares the same public key; that proves
  "this key reviewed this requirement text", which is what it says (round 1 F7).

## 4. Option B: the reviewer cosigns the warranty file

- `proofbound warranty cosign <file> [-o <out>]`, run by a reviewer inside a clone (a repository is
  required). Preconditions (round 1 F4): `Verify` against the clone returns VALID or `ErrUnreviewed`
  (checks 1-8 pass; every other result refuses); the local key is the declared key of a reviewer named
  in the body; that key has not cosigned already (a repeat is an error). It appends a cosignature, an
  ed25519 signature by that key over
  `"proofbound.warranty-cosignature.v1\n" + <exporter key hex> + "\n" + <canonical body>` (binding the
  exporter's key, round 1 F7), re-verifies the result and refuses unless it is VALID or UNREVIEWED, then
  writes `<out>` with `O_EXCL`, or replaces `<file>` through a temporary file and rename.
- File format: the outer object gains an optional `"cosignatures": [{"key": hex, "sig": hex}, …]`;
  key `^[0-9a-f]{64}$`, sig `^[0-9a-f]{128}$`, strictly sorted by key (so unique), 1 to 16 entries,
  no other fields; `[]` and `null` are malformed (absent is the only empty form). The body and its
  `schema` are unchanged (`proofbound.warranty.v1`). A slice-1 binary reports a cosigned file as
  malformed (fails closed).

## 5. Verify

Unchanged checks 1-8. New, without a repository: a non-verifying cosignature is INVALID.
New check 9, with a repository, after check 8. It evaluates **every** reviewer, signature and
cosignature before deciding; any INVALID condition wins over UNREVIEWED (round 1 F3). For each
reviewer R:
- **R is proven by B** if a cosignature verifies over the body with R's declared key; or
- **R is proven by A** if every cited artifact whose `declared_reviewer` is R has a `<path>.sig` at
  HEAD that verifies over that artifact's committed bytes with R's declared key.
- **INVALID:** a `.sig` that is present but malformed, does not verify, or is not by the declared key
  of its artifact's reviewer — including when that reviewer has no declared key (a present signature
  is never ignored); a cosignature whose key is not the declared key of some reviewer of this body
  (bystander or exporter); a reviewer's declared key equal to the exporter's key, to any requirement
  owner's declared key, or to another reviewer's declared key (round 1 F6).
- A reviewer with no declared key and no present signature is unproven (UNREVIEWED).
- If nothing is INVALID but some reviewer is proven by neither, the result is `ErrUnreviewed`:
  printed `UNREVIEWED key=… reviewer <R> has not signed`, **exit 3** (round 1 F2: exit 2 stays
  "UNVERIFIED, pass --repo") — never VALID.
- VALID prints `reviewers=<name>:<fingerprint>,…`.

`warranty export` writes a file whose own Verify is VALID **or UNREVIEWED**, never anything else
(**W-INV-2 amended**); for UNREVIEWED it prints a warning line, and a reviewer can still cosign (B).

**Compatibility (round 1 F2).** This tightens VALID. A slice-1 warranty whose reviewer has neither
`.sig` files at HEAD nor a cosignature now verifies UNREVIEWED (exit 3), with the same schema string.
Updated with it: `TestWarrantyEndToEnd` (keyed reviewer → VALID; plus an unkeyed run → UNREVIEWED),
the package fixture (its reviewer signs), `docs/releases/r3/demo-warranty.sh` and its DEMO page. `proofbound check` (step 2) is unchanged in this
step: its PASS is the readiness gate; reviewer signatures are a warranty property (recorded as a
follow-up, not built).

## 6. Invariants

| ID | Statement | Proving test |
|---|---|---|
| W-INV-1 amended | as before; the outer object may also hold `cosignatures`, validated by W-INV-11 | warranty_test.go::TestSignatureCoversEveryByte |
| W-INV-2 amended | export writes nothing unless the gate is PASS and its own Verify is VALID or UNREVIEWED | sdd `TestWarrantyEndToEnd` |
| W-INV-3 amended | the rule also rejects empty `KeyName`s, a reviewer whose `KeyName` is the owner's, and two spellings of one `KeyName` | warranty_test.go::TestRuleRejectsReSignedForgeries |
| W-INV-9 | with every reviewer's artifacts `.sig`-signed by their declared key, the result is VALID; a reviewer with no proof is UNREVIEWED; INVALID wins over UNREVIEWED | reviewer_test.go::TestReviewerProofByArtifactSignatures |
| W-INV-9a | INVALID: a `.sig` by another key, over an edited artifact, malformed, or present for a reviewer with no declared key | reviewer_test.go::TestBadArtifactSignaturesAreInvalid |
| W-INV-9b | a reviewer with some artifacts unsigned is unproven by A | reviewer_test.go::TestReviewerProofByArtifactSignatures |
| W-INV-10 | INVALID: a reviewer's declared key equal to the exporter's, an owner's, or another reviewer's | reviewer_test.go::TestReviewerKeysMustBeDistinct |
| W-INV-11 | `Parse` rejects cosignatures that are not lowercase-hex sized, not strictly sorted, empty-but-present, over 16, or carry unknown fields | reviewer_test.go::TestCosignatureParsingIsStrict |
| W-INV-12 | B: a cosignature by a reviewer's declared key proves that reviewer; one over another body or exporter key, or by a non-reviewer key, is INVALID; `Cosign` refuses the exporter's key and a repeat | reviewer_test.go::TestCosignatureProvesAReviewer |
| CLI-INV-19/20 amended | `review seal` signs, refuses, or writes unsigned exactly per §3, and writes nothing on any failure (including an existing `.sig`) | starter_test.go |
| CLI-INV-27 | `warranty cosign` enforces §4's preconditions and never writes a partial file | sdd `TestWarrantyCosignEndToEnd` |

## 7. Trust caveat (unchanged in kind)

Whoever can commit to the repository can declare any key under any name, so this proves "the key
declared as alice signed", not "the human alice signed". It makes a forged review require a forged
key declaration — a visible commit — instead of a text edit. Anchoring keys outside the repository
stays out of scope (`VD-r3-warranty-first-s59ktv` revisit trigger).

## 8. Non-goals

Multi-reviewer quorum beyond "every named reviewer", key rotation and revocation, CI (OIDC) identities,
reviewer signatures in `proofbound check`.

## 9. Round 1 dispositions

F1 HIGH → §2 identity is `KeyName`; W-INV-3 amended. F2 MED → W-INV-2 amended, exit 3, compatibility
paragraph. F3 MED → §5 precedence and the no-key cases; exact declared-key blob (§2). F4 MED → §4
preconditions, post-append re-verify, repeat is an error. F5 MED → W-INV-11 exact hex, `[]`/`null`,
cap 16; non-reviewer cosignature INVALID. F6 MED → W-INV-10 widened. F7 LOW → cosignature binds the
exporter key; A replay stated. F8 LOW → §3 no silent fallback, `PROOFBOUND_KEY` exact, `.sig` in the
existence check. F9 LOW → invariants split with named tests.
