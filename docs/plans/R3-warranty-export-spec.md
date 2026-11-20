# R3 slice 1: the portable warranty (spec draft 2)

**Status:** draft 2, ACCEPTABLE at round 2 (`r3-warranty-spec-recheck-20261119`), 2026-11-19. The owner approved D1–D3 as recommended ("go"); recorded in
`VD-r3-warranty-first-s59ktv`, under `CI-warranty-slice1-7t4iok` → `BR-portable-warranty-oeffoe`
O-1..O-5. Draft 1's non-author review (`r3-warranty-spec-review-20261119`) was NEEDS_WORK, with
3 HIGH and 4 MED, all answered below (§10). This is review round 2 of 2 (Law 11).

## 1. Payoff (the one command)

```bash
proofbound warranty export <commit> -o warranty.json                # in the project, gate PASS only
proofbound warranty verify warranty.json --repo <clone> [--trust-key <fingerprint>]
```

`verify` needs no database and no network. It prints exactly one verdict line and exits:

| Line | Exit | Meaning |
|---|---|---|
| `VALID key=<fpr> commit=<sha> change=<kit>:<name>` | 0 | every check in §5 passed |
| `UNVERIFIED key=<fpr>: signature only; pass --repo to check the claims` | 2 | without `--repo`: the signature is good and the file is internally consistent, but nothing ties it to real bytes. **Never a pass.** |
| `INVALID: <reason>` | 1 | some check failed, or the file is malformed |

The verifier must get the clone **from a source they trust** (the project's public remote), not
from the sender. The notes say so. `--trust-key` pins the expected key fingerprint as well.

## 2. What already exists (grounded 2026-11-19, review round 1)

- `core.Canonicalize` (`core.go:75`) gives canonical JSON. It rejects duplicate keys and unsafe
  numbers.
- `cli.ledgerDraft` (`starter.go:275`) walks the ledger from a commit to its bound change intents,
  their requirement revisions and obligations, and each target's `ObligationIDs`. This is the
  export's skeleton.
- Reviews parsers, in `connector/reviews` `v2.go`:
  - `ParseObligationVerdict` returns `reviewed_commit`, the change-intent and requirement references
    with their digests, the reviewer, `status` and outcomes;
  - `ParseRequirementReview` returns the requirement reference with its digest, the reviewer and
    outcomes. **It has no commit.**
  - `ArtifactSHA256()` (`v2.go:316`) is the self-digest.
- The readiness gate (`gates.go` SQL, about line 300) counts SATISFIED and VERIFIABLE by joining on
  commit, intent digest, requirement digest and obligation id. The reviewer ≠ owner rule is applied
  at ingest (`projections.go:555-561`) and at seal (`starter.go`, about line 413).
- `sdd.New(reader, now)` with `NewGitReader(gitcmd.Repo)` maps a commit's state from git alone, and
  reads the profile from that commit. `now` is not in any payload or digest, so it has no effect on
  ids or digests. **`StateRevision` returns only the change intent**, so this slice adds a sibling
  accessor (§6).
- No signing exists anywhere in the kernel. `crypto/ed25519` is in the standard library, so this adds
  no dependency (Law 8).

## 3. Keys

- **Key file.** `.proofbound/keys/<name>.pub` is a single line: `ed25519 <64 hex public key>\n`. The
  fingerprint is the first 16 hex characters of the sha256 of the 32 raw key bytes.
- **`proofbound keys init [--name <name>]`**:
  - writes the private key (the 64-byte seed and public key, hex) to
    `${XDG_CONFIG_HOME:-$HOME/.config}/proofbound/signing-key`, with mode 0600;
  - refuses to overwrite an existing key;
  - refuses a key path that resolves inside any git work tree;
  - writes `<name>.pub` into the project for the owner to commit.
  - `<name>` defaults to `git config user.name`, slugged.
- **At `export`,** the key is refused when the file is missing, is a symlink, is not mode 0600, or
  fails to parse.

## 4. Format: `proofbound.warranty.v1`

`{"body": {...}, "signature": {"key": "<64 hex>", "sig": "<128 hex>"}}`. The signed bytes are
`"proofbound.warranty.v1\n" + core.Canonicalize(body)`, so the domain is separated from the content.
The body holds:

- `schema: "proofbound.warranty.v1"`, `commit` (40 hex), `issued_at`, `issuer` (declared name),
  `repository_hint` (free text, never trusted);
- `kit` and `change_name`;
- `change_intent`: the full `ChangeIntent` payload (its `targets` give each requirement's
  `record_id`, `artifact_sha256` and `obligation_ids`);
- `requirements[]`: the full `Requirement` payloads (with `declared_owner` and the obligations,
  each with `id` and `statement`);
- `reviews[]`: the full parsed `RequirementReview` of each cited artifact;
- `verdicts[]`: the full parsed `ObligationVerdict` of each cited artifact. Each outcome carries
  `requirement_id`, `artifact_sha256` and `obligation_id`, because obligation ids repeat across
  requirements (round-2 N1).

Carrying the full records means §5.6 and §5.7 compare whole records, not chosen fields.
Ledger-local ids (event ids, seq) are left out because git cannot check them. Verify parses the file
strictly: it rejects unknown fields and duplicate keys, re-canonicalizes the body, and requires the
result to be byte-identical before checking the signature.

## 5. Verify: the checks, in order (the first failure is the reason)

1. **Parse:** strict, as above, with `schema` equal to `proofbound.warranty.v1`. Otherwise
   `INVALID: malformed`.
2. **Signature:** an ed25519 check over the signed bytes. Otherwise `INVALID: signature`.
3. **The rule on the file's own contents** (the readiness rule, made explicit):
   - Every `(requirement, obligation)` in `change.targets` has a verdict outcome `SATISFIED`. That
     verdict must have `reviewed_commit == commit`, a change-intent digest equal to `change`'s, and a
     requirement digest equal to the target's.
   - Every target obligation has a review outcome `VERIFIABLE` whose requirement digest equals the
     target's.
   - Every requirement in `requirements[]` is targeted, and every target has a requirement.
   - No review or verdict reviewer equals the reviewed requirement's `declared_owner`.
   - Every cited verdict's `reviewed_commit` equals `commit`. This blocks replaying a warranty for
     another commit.
   - Otherwise `INVALID: <which>`.
4. **Without `--repo`,** verify stops here with `UNVERIFIED`, exit 2.
5. **Key:** HEAD of the clone must descend from `commit` (`git merge-base --is-ancestor`). The
   key's `.pub` file must be committed at HEAD under `.proofbound/keys/`, with content equal to
   `signature.key`. With `--trust-key`, the fingerprint must also match. Otherwise `INVALID: key not
   declared by the repository` (or `fingerprint mismatch`). HEAD means the clone's `HEAD` as resolved
   by `git rev-parse HEAD`, and a detached HEAD is fine.
6. **Re-derivation:** `sdd.StateRecords(commit, change.name)` (§6) must reproduce `change.record_id`,
   `artifact_sha256` and `targets`, and every requirement's `record_id`, `artifact_sha256`,
   `declared_owner` and obligation statements, exactly. If the mapping refuses (unknown, archived
   before the commit, unmappable), the reason is `INVALID: spec at <commit>: <sdd error>`.
7. **Artifacts:** every cited review and verdict must exist at its `artifact_path` at HEAD, as a
   regular blob. Its `ArtifactSHA256()` must equal the cited digest. It must re-parse with its parser,
   and its reviewer, outcomes, references, `reviewed_commit` and `status` must equal the body's. A
   later edit or deletion at HEAD makes the warranty INVALID: a warranty holds only while its evidence
   is still in the repository.
8. Otherwise `VALID`, exit 0.

## 6. One sdd accessor

`(*Provider).StateRecords(ctx, commit, name) (change intent.Revision, requirements []intent.Revision,
err error)` reuses `changeAt`. It returns the change intent and the requirement revisions of the same
mapped state that the change's targets name (`stateResult.revisions` filtered by record id). It does
not change mapping, ids or digests. It is SDD-INV-72, proven against `StateRevision` and the ledger on
the e2e fixture.

## 7. Export

`proofbound warranty export <commit> [-o <file>]`:
- refuses (writes nothing) unless the commit's `ready-<commit12>` readiness result is PASS, computed
  the way `gates canary` computes it;
- builds the body from the ledger: `ledgerDraft`, plus the ledgered review and verdict events for
  that commit, chosen exactly as the gate counted them;
- refuses when the signing key's `.pub` is not committed at HEAD;
- runs its own `verify` with `--repo .` before writing. It never writes a file that would not
  verify.

## 8. Invariants (`kernel/internal/warranty/SPEC.md`, Law 6)

- **W-INV-1** Changing any byte of the body, or the signature, makes `verify` INVALID.
- **W-INV-2** `export` writes nothing unless the gate is PASS and its own verify of the result is
  VALID.
- **W-INV-3** §5.3 holds on the file alone. A body re-signed with an unsatisfied outcome, a missing
  target, the owner as reviewer, or a verdict for another commit is INVALID.
- **W-INV-4** Without `--repo`, a signature-good, consistent file is UNVERIFIED (exit 2), never VALID.
- **W-INV-5** With `--repo`, it is INVALID when any id, digest, owner or statement differs from §6's
  re-derivation, or when any cited artifact is missing, edited, or disagrees with the body.
- **W-INV-6** With `--repo`, it is INVALID when the key is not committed at HEAD (a HEAD that
  descends from `commit`), or when it does not match `--trust-key`.
- **W-INV-7** The `warranty` package's own imports exclude `internal/store`, pgx and the network
  packages, checked by a test on its import list. `Verify` takes no store and no database URL. The
  CLI binary links the store for its other commands; `warranty verify` never opens it.
- **W-INV-8** `keys init` never writes the private key inside a git work tree, and `export` refuses
  a key file that is missing, a symlink, or not 0600.
- **SDD-INV-72** (§6).

## 9. Demo (acceptance): `docs/releases/r3/demo-warranty.sh {openspec|speckit}`

1. Run the R2 loop to PASS, with `keys init` in a temporary `XDG_CONFIG_HOME`, and commit the `.pub`.
2. `warranty export`.
3. `verify` without `--repo`: `UNVERIFIED`.
4. `verify --repo` on a fresh `git clone` of the project: `VALID`.
5. Edit one statement in the file: `INVALID: signature`.
6. Re-sign a body with one outcome changed to INCONCLUSIVE, using the test key: `INVALID: obligation
   … not SATISFIED`.
7. Sign with a second key from another temporary keystore that is never committed: `INVALID: key
   not declared by the repository`.

## 10. Round-1 findings: dispositions

- **F1 HIGH (requirements cannot be re-derived):** §6 adds `StateRecords`. §5.6 gives the inputs and
  error mapping.
- **F2 HIGH (VALID without the repository):** without `--repo` the result is UNVERIFIED, exit 2
  (W-INV-4).
- **F3 HIGH (anchor):**
  - HEAD must descend from `commit`, and the key must be committed at HEAD with matching content;
  - `--trust-key` pins the fingerprint;
  - the clone must come from a trusted source, which the notes state;
  - export and verify use the same HEAD rule;
  - the `.pub` format is defined in §3.
- **F4 MED (bindings):** a review binds by its requirement digest and a verdict by `reviewed_commit`
  plus both digests. §5.3, §5.6 and §5.7 check all of them.
- **F5 MED (rule inputs):** the body carries `targets` and `declared_owner`, and both are
  re-derived. The reviewer ≠ owner check applies to every review and every verdict.
- **F6 MED (artifacts):** the artifact must be a blob at HEAD, checked with `ArtifactSHA256()`, then
  re-parsed and compared field by field. An edit or deletion is INVALID.
- **F7 MED (replay and canonical form):** the strict parse is followed by a byte-identical
  re-canonicalization check. The signed bytes carry a domain tag. Ledger ids are dropped. Every
  verdict's `reviewed_commit` must equal `commit`.
- **F8 LOW, F9 LOW:** taken in §3 and §9 at no cost (the XDG fallback, the key-file checks, the
  forbidden-import list).

## 10a. Round-2 LOWs (recorded; the cheap ones taken)

- **N1:** verdict outcomes carry the requirement id and digest. Taken in §4.
- **N2:** the reference to archived changes does not apply to `StateRecords`, and a commit absent
  from the clone is reported as `INVALID: commit not in clone` before the key check. Taken.
- **N3:** a verdict whose reviewer is the requirement's owner passes the gate but fails `export`.
  This fails closed, and the notes say so.

## 11. Non-goals (slice 1)

- Reviewer signatures (D1-b), multi-signature, and trust anchors beyond the repository.
- Revocation, rotation, and a transparency log.
- Release 1 gates other than readiness, and non-SDD warranties.
- Drift inside verify: `report drift` already answers that.
- Signing ledger events.
