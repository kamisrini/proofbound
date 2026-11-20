# MVP step 4: shareable proof — spec

**Status:** draft 2, answering spec round 1 (NEEDS_WORK: F1-F3 MED, answered; LOWs F4-F8 taken where cheap; §6). Round 2 is the last spec round (Law 11). Built **locally only**: the owner chose
2026-11-20 to keep pushes held, so nothing is published; every output is a set of files a release
could attach. **Decision:** `VD-vertical-mvp-first-ai3dhn` step 4. **Intent:**
`CI-mvp-share-proof-w7c4na` → `BR-vertical-mvp-xc31gp` O-4.

## 1. Payoff

```
$ proofbound warranty bundle v0.2.0 -o dist/warranties
warranty dist/warranties/warranty-406704b40182.json commit=406704b… change=sdd:add-csv-export VALID reviewers=alice:9c1e…
not warranted 0f9e8d7c6b5a sdd:add-pdf-export NO REVIEW
bundle dist/warranties (a1b2c3d4e5f6..406704b40182): 1 of 2 claims verified (1 VALID, 0 UNREVIEWED); VERIFY.md, index.json, badge.svg
```

A maintainer attaches `dist/warranties/` to a release. A buyer or auditor opens `VERIFY.md` and can
check every claim with two commands; the README shows `badge.svg`.

## 2. Interface: `proofbound warranty bundle <base> [-o <dir>]`

Run inside the repository after `proofbound sync all`, with the exporter's key (`KeyPath()`).
`-o` defaults to `warranties/`; an existing non-empty directory is refused (exit 1, nothing written); an
existing empty directory is replaced. It also refuses (exit 1) when `gates/` or `.proofbound/keys/`
differ from HEAD, so every state is reproducible from a clean clone at `head` (round 1 F6).

1. Classify every claiming commit in `<base>..HEAD` exactly as `proofbound check` does (CLI-INV-26).
2. For each `PASS` commit, export its warranty exactly as `warranty export` does (gate PASS, signed,
   self-verified; W-INV-2) into `<dir>/warranty-<commit12>.json`, and record its verify result:
   `VALID` or `UNREVIEWED`. A commit that cannot be exported (e.g. two claims, refused by export) is
   listed as not warranted with export's error as the reason; the bundle still completes.
3. Write `<dir>/index.json` (canonical JSON): `{"schema":"proofbound.warranty-bundle.v1",
   "unsigned":true,"note":…,"base":<sha>,"head":<sha>,"exporter_key":<fingerprint>,
   "non_claiming_commits":<n>,"entries":[{"commit","claims","state","file"?,"file_sha256"?,
   "reviewers_proven"?,"reviewers_missing"?,"reason"?}]}` — `state` ∈ `VALID`, `UNREVIEWED`, `BLOCKED`,
   `NO REVIEW`, `UNBOUND`, `NOT EXPORTABLE`; entries in range order. `base`/`head` are resolved SHAs.
4. Write `<dir>/VERIFY.md`, stating first that it and `index.json` are **unsigned** — the exporter's local
   result at `head` — and that only the signed files checked with `warranty verify` are authority
   (round 1 F2). Then: the range (`base..head` SHAs) and that commits without an `Intent: sdd:` claim
   (`non_claiming_commits`) are not covered (F1); one table (commit, claims, state, file, proven and
   missing reviewers); and the steps: install, clone from a trusted source and check out `head`,
   `proofbound warranty verify <file> --repo <clone> [--trust-key <exporter fingerprint>]`, with the
   rule that the fingerprint must come **out of band** (the page shows it only to compare), and what
   VALID / UNREVIEWED / UNVERIFIED / INVALID mean. Every cell from commit text, gate reasons or
   artifacts is escaped exactly as `check --markdown` does, backticks included (F3).
5. Write `<dir>/badge.svg`: a flat badge `proofbound | <v> of <n> claims verified`, `v` = VALID entries,
   `n` = entries; colours over **index states**: green `#4c1` when `v == n > 0`, amber `#dfb317` when
   every entry is VALID or UNREVIEWED and some is UNREVIEWED, red `#e05d44` otherwise (NOT EXPORTABLE
   is never warranted), grey `#9f9f9f` `no claims` when `n == 0`. Its `<title>` names the range and says
   it is an unsigned local result (F1). Same index, same bytes.
6. Exit 0 when every entry is VALID or there are none; **exit 3** when the bundle was written but some
   entry is not VALID (F4); exit 1 only for an error or refusal, which writes nothing.

All files are written into a fresh temporary directory beside `<dir>` and renamed into place at the
end, so a failure leaves nothing behind.

## 3. Invariants

| ID | Statement | Proving test |
|---|---|---|
| CLI-INV-28 | `<dir>` holds exactly `index.json`, `VERIFY.md`, `badge.svg` and one `warranty-<commit12>.json` per VALID/UNREVIEWED entry (both directions); each such file verifies to its entry's state and matches `file_sha256`; every claiming commit appears exactly once; VERIFY.md escapes untrusted cells | sdd `TestWarrantyBundleEndToEnd` + `cli/bundle_test.go` |
| CLI-INV-29 | The badge's text, colours and title follow §2.5; badge, index and VERIFY.md are deterministic for the same index (warranty files embed `issued_at` and are not) | `cli/bundle_test.go` |
| CLI-INV-30 | On error return nothing is written and no temporary directory remains (the temporary directory is a sibling of `<dir>`, same filesystem); a non-empty `<dir>` and a dirty `gates/` or `.proofbound/keys/` are refused | `cli/bundle_test.go` + e2e |

## 4. Non-goals (local-only step)

Uploading to GitHub releases, hosting a badge or page, signing the bundle itself (each warranty is
already signed; `index.json` and `VERIFY.md` are conveniences that point at signed files and say so),
regenerating a bundle after a later cosign (re-run `bundle` into a new directory).

## 6. Round 1 dispositions

F1 MED → badge says "claims verified", title carries the range and "unsigned"; VERIFY.md states what is
not covered; `non_claiming_commits`. F2 MED → unsigned statement first; out-of-band fingerprint rule;
reviewer pinning dropped; `reviewers_proven` / `reviewers_missing`; `file_sha256`. F3 MED → escaping as
`check --markdown`, tested. F4 LOW → exit 3. F5 LOW → badge over index states. F6 LOW → resolved SHAs;
dirty `gates/` / `.proofbound/keys/` refused; check out `head`. F7 LOW → exact contents, determinism
scope, temp dir is a sibling. F8 LOW → `head` stated on the page. Not taken: `claims` as an array (the
joined string is what `check` prints).
