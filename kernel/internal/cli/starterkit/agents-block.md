<!-- proofbound:begin -->
## Proofbound: every commit says which change it implements

This project writes specs with a spec-driven-development kit ([OpenSpec](https://github.com/Fission-AI/OpenSpec)
or [Spec Kit](https://github.com/github/spec-kit)) and uses Proofbound to warrant the code against them.

- **Before writing code for a change, run `proofbound task start <name>`**, where `<name>` is the
  folder name: the OpenSpec change under `openspec/changes/`, or the Spec Kit feature under `specs/`
  (for example `001-export`). Write the spec first; `task start` refuses a change that has no folder.
  While the task is active, every commit that changes code gets `Intent: sdd:<name>` added for you.
  Run `proofbound task stop` when the change is done, and `proofbound task` to see the active one.
- **A commit that changes code without claiming a change is refused** here, and fails the pull-request
  check. Commits that only touch specs, docs, `gates/`, review records or Markdown need no claim.
- **Stay inside the spec's area.** When a change's folder has a `scope.paths`, edit only the files it
  lists. With Claude Code, the hook `init` writes refuses an edit outside it and says why; if the change
  truly needs another file, stop and ask the spec owner to widen `scope.paths`.
- **Never claim a change the commit does not implement.** The claim is what gets verified. One change
  per task: stop and start a task before working on another change.

### Reviewing a commit (the reviewer is never the author of the commit)

1. `proofbound sync all` (it reads the specs too, because the profile is committed), then
   `proofbound review draft <commit>`. This writes
   `.proofbound/drafts/<commit>/draft.json` and prints every obligation the commit claims.
2. Read the change (`openspec/changes/<name>/` or `specs/<name>/`), read the commit (`git show <commit>`), run the tests.
3. Edit `draft.json`: set `declared_reviewer` to your name (not the owner's). For each verdict
   obligation set `outcome` to `SATISFIED` only when the code and tests show it is done; otherwise
   `NOT_SATISFIED` or `INCONCLUSIVE`. For each review obligation set `outcome` to `VERIFIABLE`,
   `AMBIGUOUS`, `UNTESTABLE` or `CONTRADICTORY`, with a one-line `finding` when it is not
   `VERIFIABLE`.
4. `proofbound review seal .proofbound/drafts/<commit>`, then commit the written files with
   `git commit -m "review <commit>"` (review records need no claim, so an active task does not stamp it).
5. `proofbound sync all && proofbound gates canary`: the commit's `ready-<commit12>` gate says PASS
   or BLOCKED, with its reason.
<!-- proofbound:end -->
