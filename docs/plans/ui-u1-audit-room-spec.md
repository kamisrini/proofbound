# Spec: U1, `proofbound ui` with the Audit room

**Status:** draft 1, 2026-11-22, for one non-author spec review (Law 11).
**Implements:** `CI-audit-room-u1-h5n9d2` → `BR-audit-room-ui-r7t3w1` O-1, O-2.
**Decision:** `VD-audit-room-ui-q2v6x3`. **Plan:** [ui-local-plan.md](ui-local-plan.md) section 2.
**Contract rows:** CLI-INV-37..41 in `kernel/internal/cli/SPEC.md` (this page is their long form).

## 1. The command

```
proofbound ui [--base <ref>] [--port <n>] [--open]
```

- Serves on `127.0.0.1:<n>`; `--port` defaults to `0` (a free port the OS picks). It prints exactly
  one line, `ui http://127.0.0.1:<port>/ (Ctrl-C to stop)`, then serves until interrupted, and exits
  0 on interrupt. `--open` also asks the OS to open that URL (`xdg-open`, `open` or `cmd /c start`);
  failing to open is a printed note, not an error.
- `--base` is the default range start. Without it, the base is the newest tag reachable from `HEAD^`
  (`git describe --tags --abbrev=0 HEAD^`); with no such tag, the repository's root commit. The page
  can change the base (`/audit?base=<ref>`); any ref `rev-parse` resolves to a commit is accepted, and
  an unresolvable one is a 400 page naming it.
- Each flag at most once; anything else is a usage error (exit 2).

## 2. Freshness and the ledger lock

The UI never holds the store between requests. On each audit request it (a) runs the same steps as
`proofbound sync all` (output discarded; a sync error is shown on the page, and the audit is still
computed from the ledger as it stands), then (b) opens the store, computes the audit (section 3) and
closes it. While another command holds the ledger, the request answers `503` with the page "The
ledger is busy: another proofbound command is running. Reload in a few seconds."; it never waits
indefinitely. Requests are serialized by one mutex, so two browser tabs cannot open the store at once.

## 3. The audit (O-1)

For the range `base..HEAD` (commits listed by `git rev-list --reverse base..HEAD`; the base itself is
not in the range), `computeAudit` returns a report with the head, base, every commit's class, the
controls below in this order, and the exceptions. Every control has an id, a question, an answer
state (`HELD`, `EXCEPTION`, `LISTED` or `NONE`), a count line, a basis sentence, the evidence rows
(commit, text), and the terminal command that gives the same answer. Exceptions are the union of the
rows marked as exceptions below; each has a stable id `<kind>-<commit12>`.

| # | Control | How it is computed | Exceptions (kind) |
|---|---|---|---|
| C1 | Every commit that claims a spec passed its readiness gate. | `classifiedRange` (the `check` states). `NONE` when nothing is claimed; `HELD` when all PASS; else `EXCEPTION`. | each non-PASS claiming commit (`gate`) |
| C2 | No commit changed code without claiming a spec. | `uncoveredCommits` (CLI-INV-33). | each UNCOVERED commit (`uncovered`) |
| C3 | Every passing claim carries a warranty that verifies against this repository. | `buildWarranty` per PASS commit, as `warranty bundle` does, kept in memory only: VALID is held; UNREVIEWED and NOT EXPORTABLE are exceptions. With no local signing key the control is `NONE`, basis "not checked: this machine has no signing key (`proofbound keys init`)". | each non-VALID warranty (`warranty`) |
| C4 | Promises that changed after they were verified are disclosed. | the `report drift` lines: STALE, REMOVED, UNSYNCED and UNKNOWN entries are listed; `HELD` when none, else `LISTED`. Disclosure is the control, so drift is not an exception. | none |
| C5 | Every commit in the range is accounted for. | each commit gets exactly one class: `claimed` (a claim or claim-like line), `uncovered`, or `housekeeping` (no claim, and every path `housekeepingPath` exempts, including a commit with no path of its own); the counts are shown and must sum to the `rev-list` count, else the control is `EXCEPTION` (an internal error made visible, never expected). | none |
| C6 | Changes to who may sign, and to the check itself, are listed for review. | every commit in the range whose combined diff (as CLI-INV-33) touches a path under `.proofbound/keys/` or `.github/workflows/`, or modifies or deletes (not adds) a file under `gates/` (`review seal` adds gates; changing one changes what a review must prove). `HELD` when none, else `LISTED`. | each such commit (`trust`) |
| C7 | Each claimed spec existed before the code that claims it. | for each claiming commit with exactly one well-formed claim `sdd:<name>`: its first parent's tree holds a path under `openspec/changes/<name>/`, `openspec/changes/archive/*-<name>/`, or `specs/<name>/`. A root commit has no parent and fails. `NONE` when nothing is claimed. | each claiming commit whose spec first appears with it or later (`order`) |

The headline is `<h> of <c> controls held for this range.` followed by `<e> exceptions, <d> with a
signed outcome.` (or `No exceptions.`), where `c` counts controls whose state is not `NONE`, `h` counts
`HELD`, and `LISTED` counts neither held nor exception.

Every untrusted string (commit subjects, paths, reasons, record fields) reaches the page only through
`html/template` escaping and reaches plain text only through `printableUnit`.

## 4. Signed acts (O-2)

**Identity.** The actor is the local signing key (`pbwarranty.KeyPath`, as `warranty cosign` uses)
whose public key equals a key declared at HEAD (`git show HEAD:.proofbound/keys/<name>.pub`, line
`ed25519 <64 lowercase hex>\n`); the name is that file's stem. With no local key, or a key not
declared at HEAD, the page shows why and how to fix it (`proofbound keys init --name <you>`, then
commit the `.pub` file) and every act button is disabled; the server refuses acts too.

**Records.** Both are one canonical-JSON object (`core.Canonicalize`) plus newline, with a
`signature` field: lowercase hex Ed25519 over the canonical bytes of the same object without
`signature`, prefixed by the tag line for its schema (`proofbound.exception-outcome.v1\n` or
`proofbound.audit-attestation.v1\n`), so a signature cannot be replayed as another record type.

- Exception outcome, `.proofbound/outcomes/<exception-id>.json`:
  `{"schema":"proofbound.exception-outcome.v1","exception":<id>,"commit":<40-hex>,"outcome":"accepted"|"fixed"|"reverted","reason":<1..500 chars, no control characters>,"fixed_by":<40-hex, only when fixed>,"by":<key name>,"key":<fingerprint>,"at":<RFC3339 UTC>,"signature":<hex>}`.
  The exception must be one the current audit lists; `fixed_by` must resolve to a commit.
- Audit attestation, `.proofbound/attestations/<base12>-<head12>.json`:
  `{"schema":"proofbound.audit-attestation.v1","base":<40>,"head":<40>,"transcript_sha256":<hex>,"headline":<string>,"by","key","at","signature"}`,
  where the transcript is the canonical JSON of the audit report (section 3, plus each exception's
  verified outcome) served at `/transcript?base=`.

**Flow.** `POST /act/preview` returns the exact unsigned record and its path; `POST /act/sign` with the
same fields re-validates against a freshly computed audit, signs, writes the file (refusing to replace
an existing one), then runs `git add -f -- <path>` and `git commit -m "proofbound: <record kind> <id>"
-- <path>` (the commit holds only that path; the repository's hooks run as usual, and the path is
housekeeping, so no claim is needed). If the commit fails, the file stays, and the page says it was
written but not committed, with git's message. A record that already exists for an exception makes
that exception read as closed by it; it is never overwritten from the page.

**Reading records back.** Outcomes and attestations are read from HEAD's tree, never the working tree.
Each is verified: well-formed, schema and tag match, `by`'s declared key at HEAD has the stated
fingerprint, and the signature verifies. A verified outcome closes its exception (shown with its
outcome, reason, signer and time); one that does not verify is shown in red as "does not verify" and
closes nothing. An attestation is shown with its range and whether its `transcript_sha256` equals the
transcript of that same range now.

**Write safety (threat: another web page or a DNS-rebound name triggering a signature).** Every
request must carry `Host` equal to `127.0.0.1:<port>` or `localhost:<port>`, else `421`. Every `POST`
must also carry `Origin` equal to `http://127.0.0.1:<port>` or `http://localhost:<port>` and header
`X-Proofbound-Token` equal to the 32-byte random hex token generated at start and embedded in the
page, else `403`; bodies are JSON of at most 16 KiB. No GET request writes anything. Responses carry
`Content-Security-Policy: default-src 'self'; frame-ancestors 'none'`, `X-Content-Type-Options:
nosniff`, and `Referrer-Policy: no-referrer`.

## 5. Surface

- `GET /` redirects to `/audit`. `GET /audit?base=<ref>` renders the Audit room; with `Accept:
  application/json` (or `format=json`) the same handler returns the report as JSON, so the page and
  its JSON twin cannot disagree. `GET /transcript?base=<ref>` downloads the transcript JSON.
  `GET /static/*` serves the embedded CSS and JS.
- The page: headline; scope (base input with the repository's tags, Apply); the commit-class bar; the
  controls, each expandable to its basis, evidence and command (`<details>`, so it works without
  JavaScript); a filter "All / Exceptions only"; the exceptions, each with "Record an outcome" (choose
  accepted, fixed in commit, or reverted; a reason; "Review what will be written" shows the record;
  "Sign and record"); the transcript panel (download, print or save as PDF via the browser, countersign,
  existing attestations); and the panel "What this room cannot see yet" (line-level scope inside a
  claimed commit; anything an agent did outside the repository).

## 6. Not in U1

The other workspaces, the time rail and as-of replay, in-browser signature checks (the page says the
binary verified each signature), the sample draw, comparing two releases, the evidence-bundle viewer
(`--bundle`), the role grid (who authored, reviewed and owns each change: identities are not yet
joined across git authors and key names), and rewritten-history detection.

## 7. Tests the rows derive

- Parsing: flags once each, unknown flag, port range (`TestParseUI`).
- Audit on a scripted repository: one PASS claim, one UNCOVERED commit, one key change, one claim whose
  spec arrives in the same commit; each control's state, counts and exception ids (`TestAuditControls`,
  with the class sum checked).
- Records: sign then verify; a changed byte, a wrong tag, a key not declared at HEAD, and a record read
  from the working tree but not committed all fail to close an exception (`TestOutcomeRecords`).
- Server: wrong `Host` 421; POST without token, with a wrong token, or with a foreign `Origin` 403; GET
  never writes; the JSON twin equals the report the page rendered (`TestUIRequestGuards`,
  `TestUIJSONTwin`).
