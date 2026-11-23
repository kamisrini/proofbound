# Spec: waivers only for signed commits (opt-in)

**Status:** draft 1, 2026-11-23, accepted by the owner ("yes" to the opt-in switch raised in
[signed-waiver-spec.md](signed-waiver-spec.md) §7); not yet reviewed. **Implements:**
`CI-waiver-signed-commits-5akinl` → `BR-waiver-signed-commits-6fioa7` O-1 (requirement `proposed`).
**Contract row (to add):** CLI-INV-63.

## 0. Why

A waiver refuses its own signer's commits by comparing the commit's writers with the waiver's
signer. On an unsigned commit, however, the author name and e-mail are whatever the committer typed,
so someone who holds a waiver key could commit under an unrelated name. Only a commit signed by a
declared key (C9, `proofbound keys git-sign`) proves who made it. Projects that want that proof can
now require it.

## 1. The switch

The switch is the file `.proofbound/waivers/policy.json`, containing exactly:

```json
{"schema":"proofbound.waiver-policy.v1","require_signed_commits":true}
```

- Because it sits under `.proofbound/waivers/`, C6 already lists every commit that changes it.
- **It counts if it exists at the range base OR in the commit's first parent.** The strictest of the
  two wins, so a pull request cannot switch it off for its own commits, and switching it on applies
  at once.
- **A policy file that is not exactly this form also counts as on** (fail closed). The page and the
  check say so.
- The file is not a waiver: the waiver listing skips it.

## 2. The rule

While the switch is on, a commit is WAIVED only if it is also signed by a key the repository
declares (git's SSH signature, verified as C9 verifies it). Every other waiver condition still
applies. An unsigned commit, or one signed by an undeclared key, stays UNCOVERED, and its reason says:
"this project waives only commits signed by a declared key (`.proofbound/waivers/policy.json`); this
commit is not".

**The commit-msg hook** (guidance only) applies the switch by asking git whether this commit will be
signed:
- It needs `commit.gpgsign` true, `gpg.format` ssh, and `user.signingkey` set, as
  `keys git-sign` configures them.
- If any is missing, it refuses and points to `proofbound keys git-sign`.

`check` still verifies the real signature from history.

## 3. Non-goals

- No command to write the policy: it is one committed file.
- No other policy fields.
- No change to how C9 verifies signatures.

## 4. Tests

`TestWaiverRequiresSignedCommits` covers:
- the switch at the base only: an unsigned commit is UNCOVERED, a signed one is WAIVED;
- the switch in the parent only;
- the switch removed inside the range: it still applies;
- a malformed policy: it applies;
- the policy file is never treated as a waiver;
- the hook refuses without signing configured and passes with it.

**Acceptance:** on the tour's `shop` repository:
1. The switch is on.
2. An unsigned human commit under a valid waiver is UNCOVERED, with the reason.
3. The same work committed after `proofbound keys git-sign` is WAIVED.
