# R3 slice 1: the portable warranty (spec draft 1)

**Status:** draft 1, 2026-11-19, for the owner's go-ahead ("option 1" from `/proofbound-next`). Planning
only: no code until the owner approves. Law 11 budget: ≤2 non-author spec review rounds, 1 code
review. First slice of the Trust Engine (vision-2028 Organ 3, build-order #5), placed ahead of #2b by
owner choice (needs a decision record, D2).

## Payoff (the one command)

```bash
proofbound warranty export <commit> -o warranty.json      # in the project, after its gate says PASS
proofbound warranty verify warranty.json --repo <clone>   # anywhere: no database, no Proofbound ledger
```

`verify` prints `VALID` or `INVALID: <reason>` and exits 0 or 1. Someone who receives the file and a
clone of the repository can check, without trusting the sender's database, that:

1. the file was signed by a key the project committed, and has not been altered since;
2. the commit claimed a spec change, and the warranted obligations are word for word what that spec
   said at that commit;
3. every obligation was reviewed SATISFIED, and every requirement VERIFIABLE, by a named reviewer who
   is not the spec's declared owner (the readiness rule, re-run on the file's own contents);
4. the review artifacts the warranty cites are committed in the repository, byte for byte.

Today that evidence exists only inside one machine's Postgres ledger. This slice makes it a file
anyone can check.

## What already exists (grounded 2026-11-19)

- Events carry `content_sha` = sha256 of canonical payload bytes (`core.Canonicalize`, `core.go:75`).
- `cli.ledgerDraft` (`starter.go:275`) already walks the ledger from a commit to its bound change
  intents, their requirement revisions and obligations: the export's skeleton.
- Review and verdict artifacts are committed files with self-digests (`vera.verdict.v1`,
  `obligation-verdict.v2`, `requirement-review.v1`); the reviews connector's parsers exist.
- The readiness rule and its counts (`gates.go:358`, GATE-INV-16).
- `sdd.Provider.StateRevision(commit, name)` re-derives a change's revision, with digests, from git
  alone, without a database. This is what lets `verify --repo` re-check the spec independently.
- No signing exists anywhere in the kernel. Go's standard library has `crypto/ed25519`, so this adds
  no new dependency (Law 8).

## Format: `proofbound.warranty.v1`

One JSON file: `{"body": {...}, "signature": {"key": "<ed25519 pubkey hex>", "sig": "<hex>"}}`. The
signature covers `core.Canonicalize(body)`. The body holds:

- `commit` (40 hex), `repository_hint` (free text, never trusted);
- `change`: kit, native name, change-intent id, its `artifact_sha256`;
- `requirements[]`: record id, digest, declared owner, and obligations (`id`, `statement`);
- `reviews[]` and `verdicts[]`: artifact path, artifact sha, reviewed commit, declared reviewer, each
  obligation's outcome;
- `gate`: rule id, state `PASS`, the reason string, the gate event's id and `content_sha`;
- `issued_at`, `issuer` (declared name), and `proofbound_version`.

## Commands

1. **`proofbound keys init [--name <name>]`** generates an ed25519 key pair. It writes the private
   key to `$XDG_CONFIG_HOME/proofbound/signing-key` (mode 0600, never inside a repository; it refuses
   to overwrite one). It writes the public key to `.proofbound/keys/<name>.pub` in the current
   project, for the owner to commit. Committing the public key is how a project declares which keys
   speak for it.
2. **`proofbound warranty export <commit> [-o <file>]`** refuses unless the commit's readiness gate
   is PASS in the ledger. It builds the body from the ledger (reusing `ledgerDraft` plus the ledgered
   review and verdict events), signs it, and writes the file. It also refuses when the signing key's
   public half is not committed at the warranted commit or later on HEAD's history.
3. **`proofbound warranty verify <file> [--repo <path>]`**:
   - Without `--repo`, it checks the signature and the internal rule (3). It prints the key
     fingerprint and says plainly that the key's ownership and the spec text are unchecked.
   - With `--repo`, it also checks:
     - the commit exists;
     - the key is committed under `.proofbound/keys/` at HEAD;
     - `StateRevision(commit, name)` reproduces the change intent's id and digest, and the
       requirements' ids, digests and obligation statements;
     - every cited review and verdict artifact exists at its path, in the commit history, with the
       cited sha;
     - the cited verdicts' `reviewed_commit` is the warranted commit.
   - It needs no database and no network.

## Invariants (become `kernel/internal/warranty/SPEC.md`, Law 6)

- **W-INV-1** Any change to one byte of the body makes `verify` INVALID (signature).
- **W-INV-2** `export` refuses a commit whose gate is not PASS, and writes nothing.
- **W-INV-3** `verify` re-runs the readiness rule on the file: a body edited so that any outcome is not
  SATISFIED/VERIFIABLE, or the reviewer equals the declared owner, is INVALID even when re-signed by
  the same key.
- **W-INV-4** `verify --repo` is INVALID when any obligation statement, digest or id differs from
  what the sdd provider derives at the commit, or a cited artifact is missing or has a different sha.
- **W-INV-5** `verify --repo` is INVALID when the signing key is not committed under
  `.proofbound/keys/` at HEAD.
- **W-INV-6** `verify` opens no database and makes no network call (checked on the package's imports).
- **W-INV-7** The private key never lands in a repository: `keys init` refuses a path inside a git
  work tree, and the starter kit's `.gitignore` lines cover it.

## Demo (the acceptance run)

`docs/releases/r3/demo-warranty.sh {openspec|speckit}` runs the R2 loop to PASS, then:

1. `keys init`, and commits the public key;
2. `warranty export`;
3. copies the file to a fresh directory and runs `verify --repo` on a fresh clone: `VALID`;
4. edits one statement in the file: `INVALID: signature`;
5. re-signs a body with SATISFIED changed to INCONCLUSIVE: `INVALID: obligation O-… is not SATISFIED`;
6. signs with an uncommitted key: `INVALID: key not declared by the repository`.

## Decisions for the owner

- **D1: who signs.**
  - **(a) The exporter's key, recommended.** One signature per warranty, by the project's committed
    key. Reviewer independence stays declared, as in R2.
  - (b) The reviewer also signs at `review seal`, so every verdict carries its own signature. This is
    stronger (vision: verifier keys separate from builder keys) but adds a key per reviewer. It makes
    a natural slice 2.
- **D2: sequencing.** Put this Trust Engine slice ahead of #2b (witnessed regeneration), which the
  roadmap put first in R3. **Recommended**, because it turns R2's output into something a buyer can
  check. It needs a decision record amending `VD-release-line-r2-sdd-t5j34p`, plus a requirement and
  change intent before code.
- **D3: trust anchor.** The repository is the anchor: a key is trusted if HEAD commits it.
  **Recommended for slice 1.** Its weakness is stated in the notes: anyone who can commit to the repo
  can add a key. Certificate chains, key rotation and revocation, and transparency logs come later.

## Non-goals (slice 1)

- Release 1 gates other than readiness; non-SDD warranties.
- Reviewer signatures (D1-b); multi-signature; trust anchors beyond the repository; revocation; a
  public transparency log; the verifier marketplace, staking, decay.
- Drift inside `verify`. A warranty is about its commit. `report drift` remains the place to ask
  whether the spec moved since.
- Signing ledger events themselves.

## Size estimate

One new package `internal/warranty` (format, sign, verify, rule re-check), about 400 lines plus tests,
and three CLI commands. It reuses `ledgerDraft`, the reviews parsers, the readiness counts, and
`StateRevision`. Expect 1–2 days of build, plus one spec review round and one code review.
