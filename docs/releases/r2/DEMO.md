# Proofbound — Demo Hub (Release 2, in progress)

**Release 2 is the SDD Warranty Connector:** Proofbound reads the specs a spec-driven-development
tool writes ([OpenSpec](https://github.com/Fission-AI/OpenSpec) or [Spec Kit](https://github.com/github/spec-kit)), binds each commit to the exact spec revision it claims to implement,
and refuses to vouch for work nobody has verified. This page shows it on a real project.

| I want to… | Do this | Time |
|---|---|---|
| **See Release 2 on a real OpenSpec project** | Setup below, then `bash docs/releases/r2/demo-openspec.sh` | ~4 min |
| See what Release 2 includes and its caveats | [Release Notes](RELEASE-NOTES.md) | — |
| See what Release 1 does (the trust record) | [Release 1 Demo Hub](../r1/DEMO.md) | ~5 min |

## Setup

The same one-time setup as Release 1: a Linux (WSL) terminal, Git, and Go 1.26+. See
[Release 1, §0 One-time setup](../r1/DEMO.md#0-one-time-setup). You also need network access to
GitHub and the embedded database's port, 55432, free. You do not need to build or install anything
first; the demo builds its own copy of the CLI.

## What the demo does

```bash
bash docs/releases/r2/demo-openspec.sh            # works in ${TMPDIR:-/tmp}/proofbound-openspec-demo
bash docs/releases/r2/demo-openspec.sh ~/pb-demo  # or a directory of your choice
```

1. **Clones [OpenSpec](https://github.com/Fission-AI/OpenSpec)** at a pinned commit. OpenSpec is
   developed with OpenSpec, so its history holds real change proposals and specs.
2. **Commits a Proofbound profile into every commit of the local clone.** The profile tells
   Proofbound where OpenSpec keeps its changes and who owns them. Release 2 takes owners only from a
   committed profile, so the demo puts it into history. (Commit IDs in the clone therefore differ
   from upstream. Nothing is pushed.)
3. **Reads the whole history:** about 650 spec revisions from about 1,060 commits.
4. **Adds two demo commits:** one says `Intent: sdd:install-skills-globally` (a real active change),
   the other claims a change that does not exist. Each gets a readiness gate.
5. **Prints what Proofbound concludes.**

## What you will see

```
claim-skip unknown-change sdd:no-such-change at <commit>
gate=bogus state=BLOCKED ... reason="commit binds no change intent"
gate=claimed state=BLOCKED ... reason="0 of 4 obligations verified SATISFIED; 0 of 4 requirement reviews VERIFIABLE"
intent=CI-install-skills-globally-57r52f ... state=IMPLEMENTED_UNVERIFIED ...
  target=...#O-ADD-INSTALL-THE-RUNNING-C-13EA48 ... outcome=UNVERIFIED statement="Init SHALL install skills ..."
```

- **The claimed commit binds** to the exact revision of `install-skills-globally`. The report lists
  that change's four obligations, word for word from OpenSpec's own spec.
- **Its gate is BLOCKED, and says why:** none of the four obligations has been verified. A commit
  is warranted only when every obligation it claims has a SATISFIED verdict and a VERIFIABLE review.
- **The bogus claim is reported and BLOCKED.** A commit cannot claim a change that does not exist.

Look up any change the same way: `proofbound report intent sdd:<change-name>`.

## Use it on your own OpenSpec or Spec Kit project

Build the CLI once (Release 1, [§1](../r1/DEMO.md#1-build-the-proofbound-binary)), then in your
project's repository:

```bash
proofbound init openspec          # once (or: proofbound init speckit): profile, .gitignore lines, AGENTS.md block,
                                  # a Claude Code reviewer subagent, and a commit-msg hook
git add .proofbound/sdd-profile.json .gitignore AGENTS.md .claude
PROOFBOUND_NO_INTENT=1 git commit -m "add proofbound"   # implements no change, so no trailer
```

Then work as usual with any agent. Every commit ends with `Intent: sdd:<name>`, where `<name>` is the
OpenSpec change folder or the Spec Kit feature folder such as `001-export` (the hook
reminds you; AGENTS.md tells your agent). To warrant a commit, a reviewer who did not write it
(another person, or the `proofbound-reviewer` subagent in a separate session) runs:

```bash
proofbound sync all
proofbound review draft <commit>      # prints each obligation; writes .proofbound/drafts/<commit>/draft.json
# decide each outcome in draft.json, set declared_reviewer
proofbound review seal .proofbound/drafts/<commit>
git add docs/verification/verdicts gates
PROOFBOUND_NO_INTENT=1 git commit -m "review <commit>"
proofbound sync all && proofbound gates canary
```

The commit's `ready-<commit12>` gate says **PASS** only when every obligation is SATISFIED and every
requirement is VERIFIABLE, by a reviewer other than the spec's owner; otherwise BLOCKED, with the
counts. Independence is declared, not proven: the record says who claimed to verify.

## Cleaning up

```bash
rm -rf "${TMPDIR:-/tmp}/proofbound-openspec-demo"
```

The demo's database lives inside that directory; its server log is `OpenSpec/.proofbound/db.log`.

## See drift: when a verified spec changes

```bash
proofbound sync all && proofbound report drift
```

For every change with a warranty, this prints CURRENT, or STALE when the change now promises
something other than what was verified, listing each obligation `added`, `removed` or `changed`
(with the old and new wording). It compares what the change promises, not file digests, so
archiving a change or editing its tasks is not drift. A STALE warranty is CURRENT again once a
reviewer verifies a commit against the current promise. On OpenSpec's own repository, rewording
one obligation of a verified change gives:

```
drift change=CI-install-skills-globally-57r52f state=STALE verified_commit=c772ac085ec2 ...
  changed O-ADD-INSTALL-THE-RUNNING-C-13EA48 "Init SHALL install skills ... scope and workflow profile. ..." -> "Init SHALL install skills ... scope, workflow profile and a pinned checksum. ..."
summary warranted=1 current=0 stale=1
```
