# Spec: U1, `proofbound ui` with the Audit room

**Status:** draft 2, 2026-11-22. Draft 1 was reviewed by a non-author
(`ui-u1-spec-review-20261122`, NEEDS_WORK: F1, F2 HIGH; F3..F11 MED; F12 LOW); draft 2 fixes every HIGH
and MED finding as listed in section 8. Law 11 allows a second spec round; the code review covers it.
**Implements:** `CI-audit-room-u1-h5n9d2` → `BR-audit-room-ui-r7t3w1` O-1, O-2.
**Decision:** `VD-audit-room-ui-q2v6x3`. **Plan:** [ui-local-plan.md](ui-local-plan.md) section 2.
**Contract rows:** CLI-INV-37..41 in `kernel/internal/cli/SPEC.md` (this page is their long form).

## 1. The command

```
proofbound ui [--base <ref>] [--port <n>] [--open]
```

- Serves on `127.0.0.1:<n>`; `--port` defaults to `0` (a free port the OS picks). It prints exactly
  one line, `ui http://127.0.0.1:<port>/?k=<token> (Ctrl-C to stop; …)`, where `<token>` is 32 random
  bytes in hex made at start, then serves until interrupted, and exits 0 on interrupt. `--open` also asks the OS to open that URL (`xdg-open`, `open` or `cmd /c start`);
  failing to open is a printed note, not an error.
- `--base` is the default range start. Without it, the base is the newest tag reachable from `HEAD^`
  (`git describe --tags --abbrev=0 HEAD^`); with no such tag, the repository's root commit (which is
  then itself outside the range; the page shows the range). The page can change the base
  (`/audit?base=<ref>`); the ref is resolved with `rev-parse --end-of-options` to a commit, which must
  be an ancestor of HEAD, else a 400 page naming it; only the resolved id is passed on to git.
- Each flag at most once; anything else is a usage error (exit 2).

## 2. Freshness and the ledger lock

The UI never holds the store between requests. On each audit request it (a) runs the same steps as
`proofbound sync all` (output discarded; a sync error is shown on the page, and the audit is still
computed from the ledger as it stands), then (b) opens the store, computes the audit (section 3) and
closes it. Syncing appends to the local ledger under `.proofbound/` (gitignored); no GET writes a
repository file. While another command holds the ledger, the request answers `503` ("The ledger is
busy …"). One request runs at a time: a request that arrives while another runs answers `503` at
once (`TryLock`), never waiting behind it.

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
| C3 | Every passing claim carries a warranty that verifies against this repository. | `buildWarranty` per PASS commit, as `warranty bundle` does, kept in memory only: VALID is held; UNREVIEWED and NOT EXPORTABLE are exceptions. A warranty must be signed by a key HEAD declares, so the control is computed only when this machine's key is declared at HEAD (one name); otherwise it is `NONE` with the reason. What it tests is repository-derived (the reviewers' signatures and the rule); only the throwaway exporter signature uses this machine's key. Like `check`, it reads `gates/` from the working tree. | each non-VALID warranty (`warranty`) |
| C4 | Promises that changed after they were verified are disclosed. | the `report drift` lines at HEAD (ledger-wide, not range-scoped): STALE, REMOVED, UNSYNCED and UNKNOWN entries are listed; `HELD` when none, else `LISTED`. Disclosure is the control, so drift is not an exception. | none |
| C5 | Every commit in the range is accounted for. | each commit gets exactly one class: `claimed` (a claim or claim-like line), `uncovered`, or `housekeeping` (everything else). This is a disclosure of the whole population, not a test (it cannot fail), so its state is always `LISTED`. | none |
| C6 | Changes to who may sign, and to the check itself, are listed for review. | every commit in the range whose combined diff (`git diff-tree -r --cc --name-status --no-renames --root`) touches a path under `.proofbound/keys/` or `.github/workflows/`, or `.proofbound/sdd-profile.json`, or modifies or deletes (status other than `A` against some parent) a file under `gates/` (`review seal` adds gates; changing one changes what a review must prove). `HELD` when none, else `LISTED`. | each such commit (`trust`) |
| C7 | Each claimed spec existed before the code that claims it. | for each commit with at least one exact claim, every claimed name must match the `task start` name rule and have a folder in the commit's first parent at `openspec/changes/<name>/`, `specs/<name>/`, `openspec/changes/archive/<name>/` or `openspec/changes/archive/<date>-<name>/`. A root commit has no parent and fails. `NONE` when nothing is claimed. | each claiming commit with a claim whose spec first appears with it or later (`order`) |

The headline is `<h> of <c> controls held for this range.` followed by `<e> exception(s), <d> with a
signed outcome.` (or `No exceptions.`), where `c` counts the controls that are tests (`HELD` or
`EXCEPTION`) and `h` counts `HELD`; `LISTED` and `NONE` are not counted. A range with no commits sets
every control to `NONE` and the headline to `The range holds no commits.`

Every untrusted string (commit subjects, paths, reasons, record fields) reaches the page only through
`html/template` escaping and reaches plain text only through `printableUnit`.

## 4. Signed acts (O-2)

**Identity.** The actor is the local signing key (`pbwarranty.KeyPath`, as `warranty cosign` uses)
whose public key equals exactly one key declared at HEAD (`git show HEAD:.proofbound/keys/<name>.pub`,
line `ed25519 <64 lowercase hex>\n`); the name is that file's stem. A key declared under two names
is refused. **Who may sign for a range:** only a key already declared at the range's base, so a key
added inside the range cannot close the range's own exceptions; and never the author of the
exception's commit (the commit's author name in `keys init`'s key-name form), so no one closes their
own exception (Law 9). Both rules apply when signing and again when reading records back. With no local key, or a key not
declared at HEAD, the page shows why and how to fix it (`proofbound keys init --name <you>`, then
commit the `.pub` file) and every act button is disabled; the server refuses acts too.

**Records.** Both are one canonical-JSON object (`core.Canonicalize`) plus newline, with a
`signature` field: lowercase hex Ed25519 over the canonical bytes of the same object without
`signature`, prefixed by the tag line for its schema (`proofbound.exception-outcome.v1\n` or
`proofbound.audit-attestation.v1\n`), so a signature cannot be replayed as another record type.

- Exception outcome, `.proofbound/outcomes/<exception-id>.json`:
  `{"schema":"proofbound.exception-outcome.v1","exception":<id>,"commit":<40-hex>,"outcome":"accepted"|"fixed"|"reverted","reason":<1..500 chars, no control characters>,"fixed_by":<40-hex, only when fixed>,"by":<key name>,"key":<fingerprint>,"at":<RFC3339 UTC>,"signature":<hex>}`.
  The exception must be one the current audit lists; `fixed_by` is present exactly when the outcome is
  `fixed` and must be a later commit (a descendant of the exception's commit) in HEAD's history. The
  reason may not hold control or invisible formatting (Unicode `Cf`, e.g. bidi) characters. An
  outcome is the signer's signed statement: "reverted" and "accepted" are not checked against the code,
  and the page says the record is a signed statement.
- Audit attestation, `.proofbound/attestations/<base12>-<head12>.json`:
  `{"schema":"proofbound.audit-attestation.v1","base":<40>,"head":<40>,"transcript_sha256":<hex>,"headline":<string>,"by","key","at","signature"}`,
  where the transcript is the canonical JSON of the audit report (section 3, plus each exception's
  verified outcome) served at `/transcript?base=`.

**Flow.** `POST /act/preview` returns the exact unsigned record, its path and its `at` time; `POST
/act/sign` sends the same fields plus that `at` (accepted only if well formed and from the last 15
minutes) and the preview's `digest` (sha256 of the unsigned canonical record); sign rebuilds the
record and refuses with 409 if its digest differs, so the record signed is the record previewed. Under the one request lock, sign recomputes
the audit, re-validates, signs, writes the file (refusing one that exists in the working tree or at
HEAD), then runs `git add -f -- <path>` and `git commit -q -m "proofbound: <record kind> <id>" --
<path>` (the commit holds only that path, whatever else is staged; the repository's hooks run as
usual, and the path is housekeeping, so no claim is needed). If either git step fails (a merge in
progress, no identity, an unborn HEAD, a hook refusal), the path is unstaged and the file removed, and
the page shows git's message, so a retry starts clean.

**Reading records back.** Outcomes and attestations are read from HEAD's tree, never the working tree.
Each is verified: byte-exact canonical JSON, schema and tag match, only that schema's fields, the
`exception` and `commit` fields equal the exception the file name and the audit give, `by`'s key
declared at the range's base has the stated fingerprint, `by` is not the commit's author, a `fixed_by`
checks as above, and the signature verifies. A verified outcome closes its exception (shown with its
outcome, reason, signer and time); one that does not verify is shown in red as "does not verify" and
closes nothing. An attestation covers its own `base..head` only: it is shown with its range, the headline it
signed, and how many commits HEAD has moved since and how many of those only add Proofbound records
(committing the attestation itself is one). The transcript is what this machine computed (C3 depends on a
declared local key, C4 on the local ledger), so an attestation is a signed statement of what its
signer's run showed, not a value every clone reproduces byte for byte.

**Write safety (threats: another web page, a DNS-rebound name, or another local process or user
reaching the loopback port).** Every request must carry `Host` equal to `127.0.0.1:<port>` or
`localhost:<port>`, else `421`. The token exists only in the terminal link: `GET /?k=<token>` sets an
`HttpOnly`, `SameSite=Strict` cookie and redirects; every request except `/static/*` must carry that
cookie, else `401` ("open the link printed in your terminal"); the token is never written into a page
or a response. Every `POST` must also carry `Origin` equal to `http://127.0.0.1:<port>` or
`http://localhost:<port>`, else `403`; bodies are JSON of at most 16 KiB with no unknown fields. A
cross-site request carries no cookie (`SameSite=Strict`), so it cannot even trigger a sync. Responses
carry `Content-Security-Policy: default-src 'self'; base-uri 'none'; form-action 'self';
frame-ancestors 'none'`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer` and
`Cache-Control: no-store`. The page's script inserts text only through `textContent`. Residual risk:
anyone who can read the user's terminal or browser profile can act as them, as with their key file.

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

The row tests are unit tests with scripted git repositories and a stubbed audit; the ledger-derived
states reuse `check` (CLI-INV-26, 33) and `warranty bundle` (CLI-INV-28), whose rows prove them. The
whole command is accepted by running it on the tour's `shop` repository (the change intent's acceptance boundary).

- Parsing: flags once each, unknown flag, port range (`TestParseUI`).
- Audit on a scripted repository: one PASS claim, one UNCOVERED commit, one key change, one claim whose
  spec arrives in the same commit; each control's state, counts and exception ids (`TestAuditControls`,
  with the class sum checked).
- Records: sign then verify; a changed byte, a wrong tag, a key not declared at HEAD, and a record read
  from the working tree but not committed all fail to close an exception (`TestOutcomeRecords`).
- Server: wrong `Host` 421; POST without token, with a wrong token, or with a foreign `Origin` 403; GET
  never writes; the JSON twin equals the report the page rendered (`TestUIRequestGuards`,
  `TestUIJSONTwin`).

## 8. Spec review draft 1: what draft 2 changed

| Finding | Fix |
|---|---|
| F1 HIGH token in the page | the token travels only in the terminal link and a `SameSite=Strict`, `HttpOnly` cookie; every non-static request needs it |
| F2 HIGH any declared key closes any exception | signer's key must be declared at the base; the commit's author cannot sign; `sdd-profile.json` joins C6. Records committed inside the range are not C6 items: they are verified against base keys instead |
| F3 replay, fixed_by | `exception` and `commit` must match; strict field set; `fixed_by` must be a later commit in HEAD's history; outcomes are labelled signed statements |
| F4 preview/sign drift, bidi | sign uses the previewed `at`; `Cf` characters refused |
| F5 base | resolved with `--end-of-options`, must be an ancestor of HEAD; empty range is NONE everywhere; root-commit base shown as such |
| F6 GET writes | wording fixed (sync writes only the local ledger); cross-site GETs carry no cookie |
| F7 C3 machine-dependent | computed only with a declared local key, else NONE with the reason; what it tests stated |
| F8 C5 tautology; C6, C7 | C5 is LISTED, never counted as held; `--name-status` named; C7 checks every claimed name and both archive forms |
| F9 transcript reproducibility | stated as a signed statement of the signer's run |
| F10 failed commit, HEAD moving | a failed git step removes the file; audit, validation and write happen under one lock |
| F11 mutex | `TryLock`, 503 at once |
| F12 LOW | recorded; `textContent`, `base-uri`, `form-action` and a two-name key refusal adopted anyway |

## 9. Code review: what changed

`ui-u1-code-review-20261122` (NEEDS_WORK: F1..F4 MED, F5..F9 LOW). F1: sign carries the previewed
digest. F2: the unreachable "current" attestation state is replaced by "covers its own head" plus
commits since. F3: the cross-tag test now keeps the schema and changes only the tag. F4: read-back,
attestations, the sign flow, the git-failure undo and the base refusals now have tests
(`TestOutcomeReadBack`, `TestUISignFlow`). LOW adopted: F5 (format checks on read-back), F6 (act
buttons disabled without a usable key), F8 (cookie named for the port), F9 (U+2028/2029 refused);
LOW recorded: F7 (a record that does not verify must be removed with git), F9 (`computeAudit` as a
whole and merge commits in C6 are proven at acceptance, not by a unit test).
