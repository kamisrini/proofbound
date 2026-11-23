# Spec: the declared change area (line-level scope)

**Status:** draft 1, 2026-11-23, accepted by the owner with both §7 defaults; reviewed with the code by one non-author (§8). **Implements:** `CI-change-area-cpcm5w` →
`BR-change-area-p59cyd` O-1..O-3 (requirement `proposed`). **Resolves:** `DEF-CLAIMED-CODE-SCOPE-001`.
**Contract rows (to add):** CLI-INV-52..CLI-INV-55. **Authorized by:** `BD-audit-room-ui-k4m2p8`
(whether agents' changes stayed inside what was authorized).

## 0. The problem, in one line

A commit that correctly claims `Intent: sdd:<name>` can still change code its spec never called for.
Today only a reviewer reading the whole diff catches that. After this change, the record names the exact
lines outside the spec's area, the reviewer must decide on each before sealing, and the Audit room shows
what was decided.

## 1. Declaring the area (O-1)

A spec folder may hold one optional file, `scope.paths`:

- OpenSpec: `openspec/changes/<name>/scope.paths` (or the archived folder `specs`/`C7` already finds).
- Spec Kit: `specs/<name>/scope.paths`.

**Format:**
- One pattern per line, UTF-8. Blank lines and lines starting with `#` are ignored.
- A pattern is a repository-relative path in which:
  - `*` matches within one path segment;
  - `**` matches any number of whole segments;
  - a trailing `/` means "everything under this directory" (not a file of that name).
  - `**` must be a whole segment; it cannot appear inside one.
- Examples: `src/cart/`, `src/**/*_test.go`, `README.md`.
- If a line is absolute, contains `..`, a backslash, or a control character, the **whole file is
  invalid**. An invalid file is reported as an exception (§3), never as "everything allowed".
- The matcher is our own, about 40 lines. It is not git pathspec, so the meaning does not depend on git
  version or config, and it adds no dependency.

**Where the area is read from:** the claiming commit's **first parent**. This is the same rule C7 uses
for the spec itself, so a commit cannot widen its own area. If a commit claims several changes, its area
is the union of their areas. If any claimed change has no `scope.paths`, that commit's area is **not
declared**.

## 2. Finding the lines outside the area (O-1)

Each changed path of the claiming commit (diffed against its first parent; a root commit is diffed
against the empty tree) falls into one of three groups:
- `housekeepingPath` (CLI-INV-33: specs, gates, `.proofbound/`, docs, Markdown): ignored;
- matched by the area: inside;
- otherwise: **outside**.

For each outside path, `git diff -U0 --no-renames <parent> <commit> -- <path>` gives one entry per hunk:
- `lines 12-18`: new-side lines added or changed;
- `removed lines 40-44`: a pure deletion, given as old-side lines;
- `file deleted`, `binary file`, or `mode change`, when there are no text hunks.

Hard limits: at most 50 entries per file. The rest are counted as "and N more hunks" in one final entry
that is decided like any other.

The function is `outOfArea(ctx, root, commit, claims) (declared bool, hunks []areaHunk, err error)`.
It is pure git and reads nothing from the ledger.

## 3. Accounting at seal (O-2)

**At draft.** `review draft` adds `"out_of_area": [{"path","lines","decision":"","reason":""}]` to
`draft.json`. The list is built from git as in §2, and it is also printed (`outside area <path>
<lines>`). If the area is not declared, the list is empty and draft prints `area: not declared by
<name>`.

**At seal.** Seal rebuilds the list from git (the "truth", F1 rule). It never takes the list from the
editable draft. It then matches decisions by `(path, lines)`. For **every** entry, the decision must be:
- `accept`, with a non-empty single-line `reason`;
- or `reject`.

A missing, empty or other decision **refuses the seal, and nothing is written**. This covers a draft
written before this change, which has no `out_of_area` field at all. The message names each undecided
entry and says to fill it in or run `review draft` again.

**What a decision does:**
- Any `reject` makes every OV verdict in this seal `NEEDS_WORK`.
- When the list is non-empty, seal writes `docs/verification/verdicts/AREA-<commit12>.json` in the same
  all-or-nothing batch.
  - Its content is `{"schema":"proofbound.area-review.v1","commit","by","key","at","hunks":[...]}`.
  - It is signed exactly as exception outcomes are: canonical JSON, Ed25519 over the tag line plus the
    unsigned bytes, using `signRecord`.
  - If the reviewer has no declared key, the record is written unsigned, like an unsigned review, and
    the audit treats it as not accounted for (§4).

**Review bench.** The Review bench shows the same list, each row with **accept / reject + reason**, and
sends the decisions through the existing seal. Without this, a browser seal of such a commit could never
succeed.

## 4. Audit control C10 (O-3)

**C10: "Each agent change stayed inside the area its spec declares."** For each claiming commit in the
period:

**Area not declared:**
- Evidence: "area not declared".
- **Not an exception.** Existing projects and the tour have no `scope.paths`, so making this an
  exception would turn every current audit red. The count shows it instead.

**Area file invalid:** exception of kind `area`.

**Area set by an agent:** exception of kind `area`.
- This applies when the area file of a claimed change was last changed, inside the period and before
  the commit, by a commit the audit names an agent on (`agentName`).
- Otherwise an agent could widen its own area in an unclaimed housekeeping commit just before.

**Every line inside:** evidence "inside its area".

**Lines outside the area:**
- All are accepted in an `AREA-` record whose signature verifies against a key declared at HEAD, by a
  reviewer who passes the C8 independence rule: evidence "N hunks outside, accepted by X: <reasons>".
- Otherwise: one exception of kind `area` per commit, listing each undecided or rejected path and line
  range.
- The exception closes through the existing signed outcome act (CLI-INV-41), like any other exception.

**State:**
- `NONE`: no claiming commit declares an area.
- `HELD`: every declared commit is inside or accepted.
- `EXCEPTION`: anything else.

**Count:** "k of m declared; n not declared".

**Shared paths:** `outOfArea` is shared by draft, seal and the audit, so all three show the same lines.

## 5. Non-goals

- **No mapping of individual lines to obligations.** The area is per spec, not per obligation; the
  deferral's further direction stays open:
  - mutation-checked tests per obligation;
  - a per-obligation warranty statement.
- **No change to:** `proofbound check`, the warranty or badge, the reviews connector's verdict format,
  or intent records. A `reject` already reaches them through `NEEDS_WORK`.
- **No function- or symbol-level areas.** That would need a parser per language.
- **No new dependency.**

## 6. Tests (derived from §1–§4)

| Test | Proves |
|---|---|
| `TestAreaPatterns` | `*`, `**` and trailing `/`; comments and blanks; each invalid form makes the whole file invalid. |
| `TestOutOfArea` | A scripted repository. One claiming commit changes:<ul><li>an in-area file;</li><li>an outside file with two hunks;</li><li>a housekeeping file;</li><li>an outside deletion.</li></ul>Exactly the two hunks and the deletion are listed, with correct ranges. An area widened **in the same commit** is ignored (read from the parent). No `scope.paths` → not declared. |
| `TestSealAccountsForArea` | The deferral's done-when.<ul><li>An undecided hunk blocks seal and nothing is written.</li><li>An old draft without the field is refused.</li><li>`accept` without a reason is refused.</li><li>`reject` → OV `NEEDS_WORK`.</li><li>`accept` → `AREA-` record written, signed, and verifying.</li></ul> |
| `TestAuditC10` | NONE / HELD / EXCEPTION, the "set by an agent" case and the invalid-file case. An `AREA-` record signed by the author does not count (C8). |
| `TestUIAreaDecisions` | The bench shows the ranges, refuses a preview until each is decided, refuses a decision for another range, and passes the decisions to seal. |

**Acceptance:** on the tour's `shop` repository:
1. Add `scope.paths` to a change.
2. Make an agent commit that also touches a file outside it.
3. `review draft` lists the lines.
4. Seal refuses until they are decided.
5. The Audit room shows C10.

`proofbound sync all` must still accept the change folder with the new file. The change intent's digest
changes, which is expected: the new file is part of the folder.

## 7. Decisions for the owner (recommendation marked)

1. **Is "area not declared" an exception?**
   - *Recommended:* no. Show it in the count only, so existing projects stay green.
   - Alternative: a `require_area` switch in `sdd-profile.json`, later.
2. **Granularity:**
   - *Recommended:* paths and globs, with the exact out-of-area line ranges reported.
   - Line ranges *in the declaration* are rejected: they go stale on the next edit.

## 8. Review (`change-area-review-20261123`, NEEDS_WORK: F1..F5 MED, F6..F10 LOW)

**MED findings, all fixed with a regression test (`TestAreaReviewFindings`):**
- **F1:** both diffs pass `--ignore-submodules=none`, so a submodule bump outside the area cannot be
  hidden by `ignore = all`.
- **F2:** the "set by an agent" check looks at each area file on its own, so a later human edit of
  one file cannot hide an agent's change to another.
- **F3:** an archived folder counts only in its exact dated form `YYYY-MM-DD-<name>` (as the sdd
  connector reads it). `2026-01-01-add-cart` is no longer part of `cart`'s area. C7 shares this rule
  through `specDirsAt`.
- **F4:** the list is capped at 200 entries in all, the last standing for the rest, and the Review
  bench accepts requests up to 1 MiB.
- **F5:** the agent check runs first, over every place a claimed change's `scope.paths` could be,
  present or deleted. An agent that deletes the area file no longer turns its next change into "not
  declared".

**LOW findings, recorded (Law 11):**
- **F6:** a `scope.paths` that is a directory is read as git's tree listing.
- **F7:** line ranges follow the local diff configuration and git version.
- **F8:** a new empty file is labelled "mode change".
- **F9:** some paths have no test:
  - the invalid case in C10;
  - unsigned and wrong-commit records;
  - the all-inside HELD path;
  - the per-file cap;
  - the binary and mode labels;
  - `review draft` output.
- **F10:** §1 says "contains `..`" where the code rejects a `..` segment, and a trailing tab is
  trimmed before the control-character check.

**Not changed:** C5 has the same submodule blind spot as F1. It predates this change and is noted
for the owner.
