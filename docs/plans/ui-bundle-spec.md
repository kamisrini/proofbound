# Spec: `proofbound ui --bundle`, release evidence in a browser without a clone

**Status:** draft 1, 2026-11-23, accepted by the owner with the three §6 defaults; reviewed with the code by one non-author (§7). **Implements:** `CI-ui-bundle-tbfsbu` →
`BR-ui-bundle-da6ue0` O-1..O-3 (requirement `proposed`). **Contract rows (to add):** CLI-INV-60..CLI-INV-62.
**Authorized by:** `BD-audit-room-ui-k4m2p8`.

## 0. The problem, in one line

`proofbound warranty bundle` produces a folder of signed warranty files, an index, a verification
page and a badge for a release. The person it is for (a buyer, an auditor, a compliance reviewer)
must today read JSON, or clone the repository and type `warranty verify` per file. They often do
not have, and should not need, the source repository. `ui --bundle` opens that folder as a page they
can read and check, honestly labelled with how much was checked.

## 1. The command (O-1)

`proofbound ui --bundle <dir> [--trust-key <fingerprint>] [--repo <clone>] [--port <n>] [--open]`

- **Runs anywhere**, like `warranty verify`. It needs:
  - no repository root;
  - no store or `DATABASE_URL`;
  - no git, unless `--repo` is given.
- **Server:** the same loopback server and guards as `ui`:
  - a token link that sets the HttpOnly cookie;
  - host check 421, origin check, the same CSP, and 401 without the cookie.
- **Read-only:** no act, no POST route, no agent token, no discovery file.
- **Routes:** only `/` (the bundle page, with a JSON twin through `Accept: application/json` or
  `?format=json`) and `/static/*`. Every other path is 404.
- **Folder rules:**
  - `<dir>` must be a directory.
  - Only regular files directly inside it are read; symlinks are refused.
  - It is read **once** at start, and each file is read at most 4 MiB.
  - If `index.json` is missing or unreadable, the command fails before serving.
- `--base` is refused with `--bundle`.

## 2. What is checked, and how it is labelled (O-2)

**The index is unsigned:** `index.json`, its `waived` list, and the bundle's own `VERIFY.md` and
badge. The page says so at the top, the same way `VERIFY.md` does. The page never renders
`VERIFY.md` or `badge.svg`; it builds its own view from the index and the files.

**Each warranty file** the index names (`warranty-<12 hex>.json`), and any such file in the folder that
the index does not name, is checked exactly as `warranty verify` checks it, through
`pbwarranty.VerifyReport`:

| Situation | Result shown |
|---|---|
| No `--repo` and the file's own checks pass (signature, cosignatures, rule, and the `--trust-key` pin when given) | **SIGNATURE ONLY**: "the file is intact and signed by key `<fp>`; its claims were not checked against the source" |
| `--repo` given, full check passes | **VALID** |
| `--repo` given, only a named reviewer's signature is missing | **UNREVIEWED** |
| A check fails | **INVALID**, with the first failure |
| The index names a file that is not there, or whose SHA-256 differs from `file_sha256` | **MISSING** or **ALTERED** |
| A warranty file the index does not name | Listed and checked as above, marked "not in the index" |

**Exporter key:**
- Without `--trust-key`, the page shows the exporter key's fingerprint, with the instruction to compare
  it with one obtained from the project out of band.
- With `--trust-key`, a different key makes the file **INVALID**.

## 3. What the page shows (O-3)

**Head:**
- the range (`base..head`);
- the counts per result;
- the unsigned-index notice;
- the trust-key state: pinned or not pinned.

**Per warranty**, from the signed body only, never from the index:
- the commit and the change name;
- the issuer, the issue time, and the repository hint;
- each **promise**: the requirement id and digest, each obligation id with its statement, and the
  outcome of each obligation verdict;
- each requirement review's outcomes;
- the reviewers named, proven (`--repo`) or not;
- the cosignatures.

**Waived commits:** read from the index, under their own heading, labelled unsigned.

**In-browser check:** for each warranty file, the page carries the signed bytes (`Schema\n` +
canonical body), the signature and the exporter key. The embedded script re-verifies them with
WebCrypto Ed25519, as U3 does for records (CLI-INV-51), and marks each file:
- verified in this browser;
- does not verify;
- cannot check.

Cosignatures are not re-checked in the browser.

**Encoding:** everything goes through `html/template`, and the script uses `textContent` only.

## 4. Non-goals

- No fetching a bundle from a URL or a zip. It reads a local folder only.
- No hosted mode.
- No change to the bundle format, `warranty verify`, or the warranty package.
- No rendering of `VERIFY.md` or `badge.svg`.
- No acts.
- No new dependency.

## 5. Tests

| Test | Proves |
|---|---|
| `TestParseUIBundle` | Flags. `--bundle` with `--base` is refused, as is each flag given twice, and a dir must be given. `--trust-key` and `--repo` are accepted only with `--bundle`. |
| `TestBundleView` | Built from a warranty signed over a scripted repository (the warranty package's own fixture, ported) in a bundle folder with its index, then four folder variants:<ul><li>untouched: SIGNATURE ONLY, VALID with `--repo`, INVALID under a wrong `--trust-key`;</li><li>an edited file: ALTERED and INVALID;</li><li>a deleted file: MISSING;</li><li>an extra signed file: "not in the index".</li></ul>A symlink is refused, and a missing `index.json` fails the start. |
| `TestBundleRunsOutsideARepository` | The built binary serves a bundle from a folder that is not a repository. |
| `TestBundleServer` | Only `/` and `/static/*` answer. No POST. The guards hold: 401 without the cookie, 421 for a wrong host. The page carries the browser proof, and a Go re-check of those bytes verifies. The waived list is shown as unsigned. |

**Acceptance:**
1. On the tour's `shop` repository, `warranty bundle` is made.
2. The folder is copied somewhere with no repository.
3. `proofbound ui --bundle <copy>` is opened in a browser: the warranties read SIGNATURE ONLY, the
   browser check reads "verified in this browser", and the promises and reviewers are readable.
4. With `--repo <clone>` the warranties read VALID.

## 6. Decisions for the owner (recommendation marked)

1. **Without `--repo`:**
   - *Recommended:* show SIGNATURE ONLY, so the page works with no clone and says exactly what was not
     checked.
   - Alternative: refuse to start without a clone.
2. **In-browser signature check:**
   - *Recommended:* include it, the same as the Audit room's.
   - Alternative: leave it out.
3. **Input:**
   - *Recommended:* a local folder only.
   - Alternative: also a URL or a zip, later.

## 7. Review (`ui-bundle-review-20261123`, NEEDS_WORK: F1..F3 MED, F4..F8 LOW)

**Fixed, each with a regression test in `TestBundleReviewFindings`:**

**F1: an INVALID warranty's contents were shown.** An edited statement, or a file signed by a key other
than the pinned one, still displayed its promises, verdicts and reviewers. Now:
- The signed body is shown only when the file reaches SIGNATURE ONLY, VALID or UNREVIEWED.
- An INVALID card shows only the key and why it failed ("the signature does not match the file's
  contents: it was edited or forged, so none of it is shown").
- A commit shown on such a card is labelled as coming from the unsigned index.

**F2: the index, the file name and the signed commit were not tied together.** A row is now ALTERED
when:
- the signed commit differs from the commit in its file name;
- the signed commit differs from the commit its index entry names;
- the index names the same file twice, or the same commit twice.

Each file is counted once.

**F3: the range came from the unsigned index without a label.** The title, heading and range now carry
"from the unsigned index". With `--repo`, each warranty's commit must be in the index's `base..head`
in the clone, or the row is ALTERED. The notice now says exactly what the page takes from the index.

**LOW findings, recorded (Law 11):**
- **F4:** a race between the check and the open. A file swapped for a symlink or FIFO between `Lstat`
  and `Open` would be followed, and a file that grew is read at max + 1 bytes without an error.
- **F5:** no cap on how many entries or files a folder may hold.
- **F6:** SIGNATURE ONLY shares VALID's green, and reviewers carry no "not proven" mark without a
  clone. The browser badge says "verified" the same way whether or not the key is pinned.
- **F7:** the index's commit and reason strings are not validated or made printable, and `short` cuts
  bytes, not characters.
- **F8:** gaps in the tests:
  - `TestBundleRunsOutsideARepository` skips if the build fails, and does not fetch the page;
  - no test for a symlinked `index.json`, UNREVIEWED with a clone, or an unusable `DATABASE_URL`.
