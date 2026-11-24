# Spec: review outcomes by agent (the agent track record)

**Status:** draft 2, 2026-11-24, built. Draft 1 was reviewed once by a non-author (§7: NEEDS_WORK,
every HIGH and MED folded in at build). The owner said "lets do both" (this and the documentation pass) after
the session-provenance build. The defaults in §5 are taken, so building can start; the owner can still
change them. **Implements:** `CI-agent-track-record-t6m3vb` → `BR-agent-track-record-k8w2qf` O-1. **Contract rows:** CLI-INV-73..74.

## 0. The problem, in one line

The Audit room says who made each commit, and each control says what went wrong in it. Nothing answers
the question an engineering lead or auditor actually asks about an agent: **"how has this agent's work
held up when someone checked it?"**

## 1. What it shows

For a range `base..HEAD`, one row per **agent**. Agents are the same as in "Who did the work": the
author, or a `Co-Authored-By` naming a coding agent. Every count comes from records committed at HEAD:

| Column | From |
|---|---|
| Commits; claimed a spec; code without a spec | the commit classes (as today) |
| Reviewed; ACCEPTABLE; NEEDS_WORK; not reviewed | each claimed commit's `OV-<commit12>-*.md` verdicts at HEAD (NEEDS_WORK if any of them is) |
| Lines outside the area: rejected, accepted | each commit's signed `AREA-` record |
| Parts the tests did not catch: rejected, accepted; whole change not caught | each commit's signed `TESTS-` record |
| Lines by a recorded session; lines with no recorded agent edit | each commit's signed `PROV-` record |
| Open exceptions | the Audit room's exceptions without a verified outcome |

A record that does not verify against HEAD's declared keys is not counted; its commit shows under the
control's exception instead (as today).

**A second table, one row per recorded session (from PROV records).** It shows:
- the session id;
- its commits;
- the lines it wrote;
- how those commits were reviewed (ACCEPTABLE, NEEDS_WORK, not reviewed).

It is the closest Proofbound comes to "this run of the agent": the agent name is self-declared, but a
session is what the logs showed.

## 2. Where it shows

- **`proofbound report agents <base>`** prints both tables as text, and `--json` prints them as JSON.
  It is one command the owner can run on a real repository (Law 11).
- **The Audit room** shows both tables, under "Who did the work". The audit transcript JSON is not
  changed, so attestations stay as they are.

## 3. What it is not

- **Not a score or a ranking.** It shows counts, never a rate or a grade. Small numbers mislead, and
  the reader decides what matters.
- **Not proof of identity.** Agent names are as git records them, proven only for commits signed by a
  declared key (C9). Session rows are what the logs on the agent's machine showed (C12's limit).
- **Not a judgement on the model.** It describes how this repository's reviews treated this agent's
  changes, in this range.

## 4. Tests (planned)

- **`TestTrackRecord`:**
  - an agent commit reviewed ACCEPTABLE;
  - an agent commit reviewed NEEDS_WORK, with a rejected out-of-area line, a rejected uncaught hunk
    and an unrecorded line;
  - an unreviewed claim;
  - code without a spec;
  - a person's commit, which counts for the person only;
  - two sessions;
  - an unsigned or tampered record, which is not counted.
- **`TestReportAgents`:** the text and the JSON, through `run`.
- **Acceptance:** in the agent demo, `report agents demo-base` gives one result on each branch:
  - **On the first attempt's branch:** Claude with 1 claimed commit, NEEDS_WORK, 1 out-of-area line
    rejected, 2 uncaught parts rejected and 1 line with no recorded agent edit.
  - **On the redo's branch:** 1 claimed commit, ACCEPTABLE, with its session.

  (The redo starts from `main`, so its branch does not hold the first attempt.)

## 5. Decisions (defaults taken; the owner may change them)

1. **Counts, not a score.** The alternative is a rate per agent ("NEEDS_WORK share"), which invites
   ranking on small numbers.
2. **A CLI report and a section of the Audit room, not a new page.** The alternative is a sixth
   workspace in `ui`.
3. **Only records at HEAD count.** A review written in another branch does not count until it is
   merged, which is the same rule the controls follow.

## 6. Contract rows (planned)

- **CLI-INV-73:** the track record: rows, counts, verified records only.
- **CLI-INV-74:** `report agents` and the Audit room section.

## 7. Review and amendments (owner-visible)

Round 1, one non-author, on draft 1: **NEEDS_WORK** (2 HIGH, 4 MED, 3 LOW). Every finding was folded
in at build:

- **HIGH-1.** Reading the records again could count what the controls reject: a record signed by the
  agent's own key verifies, but C10 does not trust it. **Amendment:** C10, C11 and C12 hand over the
  counts they trusted, per commit, matched to what git lists now. The track record reads only those,
  and tests show a record signed by the agent counts for nothing.
- **HIGH-2.** On the redo's branch the session table also shows the first attempt's session, because
  identical lines match. **Amendment:**
  - The page and the CLI say that a line can count for more than one session.
  - An agent's "lines by a recorded session" counts each line once.
  - The acceptance on the redo's branch shows two sessions: 9 lines for 2b9e04d6, 3 for 7f3c2a91
     (the agent row counts 9 lines in all).
- **MED-1.** Review state:
  - A verdict counts only when it parses at its own path and names that commit.
  - A review that is not independent under C8 gets its own column, never ACCEPTABLE.
- **MED-2.** The notes say that NEEDS_WORK already includes rejected hunks and parts.
- **MED-3.** A commit counts once for every agent named on it, so the agents' counts do not add up to the range;
  this is said and tested. People without an agent get no row.
- **MED-4.** The session table is titled "Sessions in the provenance records", with a note on how it
  differs from the time-linked "Agent sessions" table.
- **LOW-1.** Units are named: "hunks outside the area", "parts no test caught", "lines".
- **LOW-2.** It is renamed **"Review outcomes by agent"**, and the caveats are printed first in the
  terminal too.
- **LOW-3.** The missing tests were added: a non-independent record, a verdict for another commit,
  several verdicts, a line in two sessions, and an unknown base.

**Acceptance, as run on the agent demo:**
- **First attempt's branch:** Claude has 1 claimed commit, NEEDS_WORK, out-of-area hunks 1 (1
  rejected), parts no test caught 2 (2 rejected), whole change not caught 1, lines by a session 4,
  lines with no recorded agent edit 1. Session 7f3c2a91 has 4 lines, NEEDS_WORK.
- **Redo's branch:** Claude has 1 claimed commit, ACCEPTABLE, 9 lines. Sessions 2b9e04d6 (9) and
  7f3c2a91 (3) are both ACCEPTABLE.

## 8. Code review (owner-visible)

`agent-track-record-review-20261124`: NEEDS_WORK (M1, M2 MED fixed; L2, L3 fixed; L1, L4 (part), L5
LOW recorded).

- **M1 (MED), fixed.** `report agents` counted every exception as open: it did not attach signed
  outcomes, as the Audit room does.
- **M2 (MED), fixed.**
  - `report agents` now syncs first, as the Audit room does.
  - A verdict's own declared reviewer is judged by C8's rule, so a self-review the ledger has not yet
    seen is never ACCEPTABLE (tested).
- **L2, fixed.** The Audit room table has the Commits, Code without a spec and Open exceptions columns.
- **L3, fixed.** A session's commit reviewed without independence gets its own column, never "not
  reviewed".
- **L4, partly fixed.** A diff that fails now fails the audit instead of reading as 0 lines.
- **Recorded:**
  - **L1:** two of the mutation survivors are test gaps, not equivalent mutants. They are a verdict
    re-digested to another commit's path, and a verdict deleted after the as-of commit in the time
    rail.
  - **L4:** a provenance record's session counts are not re-matched to the current diff.
  - **L5:** an area set by an agent counts its hunks against that area.
  - M1's fix is the shared call that the Audit room already makes, and no CLI test covers it.
