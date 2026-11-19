<!-- proofbound:begin -->
## Proofbound: every commit says which change it implements

This project writes specs with a spec-driven-development kit ([OpenSpec](https://github.com/Fission-AI/OpenSpec)
or [Spec Kit](https://github.com/github/spec-kit)) and uses Proofbound to warrant the code against them.

- **End every commit message with a trailer naming the change it implements:**
  `Intent: sdd:<name>`, where `<name>` is the folder name: the OpenSpec change under
  `openspec/changes/`, or the Spec Kit feature under `specs/` (for example `001-export`). A commit
  that implements no change (docs, chores, reviews) is committed with
  `PROOFBOUND_NO_INTENT=1 git commit ...`.
- **Never claim a change the commit does not implement.** The claim is what gets verified.

### Reviewing a commit (the reviewer is never the author of the commit)

1. `proofbound sync all`, then `proofbound review draft <commit>`. This writes
   `.proofbound/drafts/<commit>/draft.json` and prints every obligation the commit claims.
2. Read the change (`openspec/changes/<name>/` or `specs/<name>/`), read the commit (`git show <commit>`), run the tests.
3. Edit `draft.json`: set `declared_reviewer` to your name (not the owner's). For each verdict
   obligation set `outcome` to `SATISFIED` only when the code and tests show it is done; otherwise
   `NOT_SATISFIED` or `INCONCLUSIVE`. For each review obligation set `outcome` to `VERIFIABLE`,
   `AMBIGUOUS`, `UNTESTABLE` or `CONTRADICTORY`, with a one-line `finding` when it is not
   `VERIFIABLE`.
4. `proofbound review seal .proofbound/drafts/<commit>`, then commit the written files with
   `PROOFBOUND_NO_INTENT=1 git commit -m "review <commit>"`.
5. `proofbound sync all && proofbound gates canary`: the commit's `ready-<commit12>` gate says PASS
   or BLOCKED, with its reason.
<!-- proofbound:end -->
