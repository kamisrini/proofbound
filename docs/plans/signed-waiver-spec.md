# Spec: signed exception waivers

**Status:** draft 1, 2026-11-23, accepted by the owner with the three §6 defaults; not yet reviewed. **Implements:** `CI-signed-waiver-6oqz6f` →
`BR-signed-waiver-15j8f9` O-1..O-3 (requirement `proposed`). **Resolves:** `DEF-NO-INTENT-SELF-WAIVER-001`.
**Contract rows (to add):** CLI-INV-56..CLI-INV-59. **Authorized by:** `BD-audit-room-ui-k4m2p8`.

## 0. The problem, in one line

Today, a commit that changes code without claiming a spec is UNCOVERED, and the pull-request check
fails. That is right for agent work. But real projects need rare, legitimate exceptions, such as a
hotfix or a vendored file. The only way out used to be a self-granted bypass, and it was retired on
2026-11-22. A waiver is that exception done honestly:
- approved **before** the work;
- signed by someone who **did not write** the change;
- limited to named paths;
- with a reason and an expiry;
- shown wherever a reader looks.

## 1. The record (O-1)

**Location:** `.proofbound/waivers/<id>.json`. `<id>` matches `^[a-z0-9]+(-[a-z0-9]+)*$` and is at most
64 characters.

**Schema:** `proofbound.waiver.v1`. All fields are strings:

| Field | Meaning |
|---|---|
| `schema`, `id` | the schema tag and the waiver id (equal to the file name) |
| `paths` | the patterns it covers, newline-joined, in the `scope.paths` grammar (CLI-INV-52); at least one |
| `reason` | 1..500 characters, one line (`checkReason`) |
| `expires` | `YYYY-MM-DD`, the last day it can be used |
| `by`, `key`, `at` | the signer's declared key name, fingerprint, and the signing time |
| `signature` | Ed25519, exactly as exception outcomes are signed (`signRecord` / `verifyRecord`) |

**Creating one:** `proofbound waiver sign <id> --path <pattern> [--path <pattern>]... --reason <text>
--expires <YYYY-MM-DD>`.
- It signs with this machine's key. The key must be declared at HEAD (`localActor`).
- `expires` must be today or later, and at most **90 days** ahead.
- The file is written with O_EXCL, so it never overwrites an existing waiver.
- It prints "commit this file before the work it waives". **It does not commit.**

There is no UI act in this slice.

## 2. When a waiver covers a commit (O-2)

The rule applies only to commits that are UNCOVERED today (CLI-INV-33): no claim, and at least one
non-housekeeping path. A waiver at path `P` is **valid for** commit `X` when all of the following hold.

1. **Approved before the work.** `P` is in `X`'s first parent. The record is read from there and never
   from `X` or HEAD. A waiver added in the same commit, or after it, does not count. Deleting the file
   later does not un-waive a commit it already covered, because history is read as it was.
2. **Verifies** against the keys declared in `X`'s first parent (`keysBefore`, the outcome rule).
3. **Signed by someone independent of the commit.** The signer's key name is not:
   - the git author;
   - an agent named on the commit (author or `Co-Authored-By`, `agentName`);
   - the commit's verified signer.

   This uses the C8 `independence` rule with no spec owners. **An author-signed waiver is refused.**
4. **Not expired.** `expires` is on or after the UTC date of `X`'s committer date.
   - A git date is self-declared. Condition 1 is the stronger fact: the waiver existed before the
     commit.
   - The check output says which date was compared.

**`X` is WAIVED** when the union of the paths of its valid waivers matches **every** non-housekeeping
path it changes. Otherwise it stays UNCOVERED, and the reason names the unwaived paths and why each
candidate waiver did not count (not in the parent, does not verify, signed by the author, expired).

The function is `waiversFor(ctx, root, sha, code []string, info commitInfo) (waived []waiverUse,
unwaived []string, notes []string)`.

## 3. Where it is shown (O-3)

**`proofbound check`:**
- A WAIVED commit is a line in state `WAIVED`: `waived by <by> (waiver <id>, until <expires>):
  <reason>`.
- It does not make the pull request NOT WARRANTED.
- When any commit is waived, the headline adds `; N commit(s) waived`. Example: `Proofbound: WARRANTED
  (2 of 2 claiming commits verified; 1 commit waived)`.
- The same applies to `--markdown`.

**`warranty bundle`** (the release proof):
- `index.json` gains `waived`, a list of `{commit, waiver, by, expires, reason}`. The field is omitted
  when the list is empty, so existing bundles are byte-identical.
- `VERIFY.md` lists the waived commits under their own heading.
- The badge text gains ` · N waived`.

Waived commits carry no warranty file: they claim no spec.

**Audit room:**
- **C2:** a waived commit is evidence (`waived by …`), **not** an exception. The count says how many
  were waived.
- **C6:** lists every commit that adds, changes or deletes a file under `.proofbound/waivers/`, as it
  already does for keys. A waiver is a change to what is allowed.

## 4. Non-goals

- **Out of scope for waivers:**
  - lines outside a spec's area (the reviewer decides those, CLI-INV-54);
  - a claiming commit's failed gate.
- **No revocation** other than "not in the parent". A waiver deleted before the commit does not cover
  it.
- **No designated-approver list:** any declared key independent of the commit may sign.
- **No UI act for signing**, and no change to warranty files or the reviews connector.
- **No new dependency.**

## 5. Tests (derived from §1–§3)

| Test | Proves |
|---|---|
| `TestWaiverSign` | The record verifies. Bad id, empty path, invalid pattern, bad reason, a past date or more than 90 days ahead are refused. An existing file is never overwritten. An undeclared local key is refused. |
| `TestWaiverCovers` | Each case:<ul><li>Valid → WAIVED.</li><li>Only in the commit itself, not the parent → UNCOVERED.</li><li>**Signed by the author → UNCOVERED (the deferral's done-when).**</li><li>Signed by an agent on the commit → UNCOVERED.</li><li>Undeclared key → UNCOVERED.</li><li>Expired → UNCOVERED.</li><li>Covering only some paths → UNCOVERED, naming the rest.</li><li>Two waivers together covering all paths → WAIVED.</li></ul> |
| `TestCheckWaived` | Plain and Markdown output: the WAIVED line, the headline suffix, and exit 0 with an otherwise warranted range. |
| `TestBundleWaived` | `index.json` `waived` and the badge text; no `waived` field when none. |
| `TestAuditWaived` | C2 shows evidence and raises no exception for a waived commit. C6 lists the waiver commit. |

**Acceptance:** on the tour's `shop` repository:
1. The reviewer signs a waiver for `vendor/`.
2. It is committed.
3. A human commits a vendored file with no claim.
4. `proofbound check` reads WARRANTED with `1 commit waived`.
5. The bundle badge says `waived`.
6. The same commit by the waiver's own signer stays UNCOVERED.

## 6. Decisions for the owner (recommendation marked)

1. **Expiry limit:**
   - *Recommended:* at most 90 days, so no waiver is permanent.
   - Alternative: no limit.
2. **Who may sign:**
   - *Recommended:* any key the repository declares, provided the signer is independent of the
     commit.
   - Alternative: a designated approver list in the profile (later).
3. **Signing surface:**
   - *Recommended:* CLI only in this slice; the Audit room shows waivers but cannot sign them.
   - Alternative: a signing act in the UI.
