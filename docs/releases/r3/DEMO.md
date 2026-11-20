# Proofbound — Release 3, slice 1: the portable warranty

**Status: built 2026-11-19, awaiting the owner's acceptance.** Release 2 decides whether a commit is
warranted, but the evidence lives in one machine's database. This slice turns a warranted commit into
**one signed file that anyone can check against a clone of the repository, with no database and no
network.** Plan: [`R3-warranty-export-spec.md`](../../plans/R3-warranty-export-spec.md). Decision:
[`VD-r3-warranty-first-s59ktv`](../../decisions/VD-r3-warranty-first-s59ktv.md).

## Run the demo (about 2 minutes)

```bash
bash docs/releases/r3/demo-warranty.sh openspec    # or: speckit
```

It runs the Release 2 loop to a PASS gate, then:

```
VALID key=1aea35eca61c3618 commit=5842e61... change=sdd:add-csv-export obligations=2   # fresh clone, no database
UNVERIFIED key=1aea35eca61c3618: signature only; pass --repo to check the claims    # no repository: never a pass
INVALID: signature                                                                    # one word edited in the file
INVALID: fingerprint mismatch: key 1aea35eca61c3618, expected 0000000000000000        # wrong pinned key
proofbound: warranty: would not verify against this repository: key not declared by the repository   # undeclared key, refused at export
```

## Use it on your project (after Release 2's review flow)

```bash
proofbound keys init                       # private key in ~/.config/proofbound (never in the repo)
git add .gitignore .proofbound/keys && PROOFBOUND_NO_INTENT=1 git commit -m "declare warranty key"
proofbound warranty export <commit>        # only for a commit whose ready-<commit12> gate is PASS
```

Whoever receives `warranty-<commit12>.json`:

```bash
git clone <the project's public URL> project      # from a source you trust, not from the sender
proofbound warranty verify warranty-<commit12>.json --repo project [--trust-key <fingerprint>]
```

`VALID` (exit 0) means all of the following hold:
1. The file was signed by a key the repository commits under `.proofbound/keys/`.
2. Every promise in it is word for word what the spec said at that commit. `verify` re-reads the spec
   from git itself.
3. Every obligation was reviewed SATISFIED, and every requirement VERIFIABLE, by someone other than
   the spec's owner.
4. The review files are committed, byte for byte.

`UNVERIFIED` (exit 2) means no repository was given. `INVALID` (exit 1) names the first check that
failed.

## Caveats

- **The repository is the trust anchor.** Anyone who can commit to it can declare a key, so clone
  from a source you trust, and pin `--trust-key` when you know the fingerprint. There is no rotation,
  revocation or transparency log yet.
- **Reviewer independence is still declared, not proven.** The project's key signs; reviewers do not
  sign their own verdicts yet (slice 2).
- **A warranty holds only while its evidence is in the repository.** Deleting or editing a cited
  review file makes it INVALID.
- **A warranty is about its commit.** Later spec changes do not invalidate it; `proofbound report
  drift` says whether the spec has moved since.
- **A verdict whose reviewer is the spec's owner can pass the gate but is refused at export.** That
  is stricter, and it fails closed.
