# Spec: U3, the time rail and proof in the browser

**Status:** draft 1, 2026-11-22, reviewed with the code by one non-author.
**Implements:** `CI-ui-time-proof-u3-r8m4x6` → `BR-ui-time-proof-h2v9c5` O-1, O-2.
**Builds on:** [U1](ui-u1-audit-room-spec.md) and [U2](ui-u2-workspaces-spec.md); the two items U1 gave up
(`VD-audit-room-ui-q2v6x3`). **Contract rows:** CLI-INV-50, CLI-INV-51.

## 1. Time rail (O-1)

`/audit?base=<ref>&asof=<commit>`: the Audit room as it stood at `<commit>`, which must be in the
period (an ancestor of HEAD and not an ancestor of the base).

- **How it is computed.** Under the request lock, a temporary detached worktree is added at `<commit>`
  in a fresh directory (`git worktree add --detach`), `sync all` runs inside it into its own embedded
  record (its `.proofbound/`), and the audit and promises are computed there over `base..<commit>`;
  the worktree is then removed (`git worktree remove --force`) whatever happened. Everything is read
  from that checkout and that record, so no later commit, review, outcome or record can change the
  answer: the claim "the record rebuilds from git" is what makes this honest. The result is cached in
  memory per (base, commit) for the server's life. With `DATABASE_URL` set the time rail is refused
  (it would write into a shared database).
- **The page** shows a rail of the period's commits (oldest first; each a link) with "Now" at the end,
  and, while a past moment is shown, a banner ("As of `<commit12>`, `<date>`: computed from a checkout
  of that commit with its own record; nothing after it is visible"), and **Compare with now**: each
  control whose state or count differs (then and now), and exceptions present only then or only now.
  Every act button is hidden and the act endpoints ignore `asof` (they always act on HEAD).
- JSON: the same handler returns `{"as_of": <report then>, "now_differs": {...}}` with `format=json`.

## 2. Proof in the browser (O-2)

For every exception outcome and audit attestation the page shows as verified, and every review file
(`docs/verification/verdicts/RR-<commit12>-<n>.md` and `OV-…`) of a claiming commit with its `.sig`,
the page carries the exact bytes (base64), the signature, and the declared public key (hex) the server
checked against, in `data-` attributes. The embedded script, using WebCrypto Ed25519 only:
- records: parses the JSON, removes `signature`, re-serializes canonically (keys sorted by UTF-16 code
  unit, `JSON.stringify` for each string: identical to RFC 8785 for these all-string records), prefixes
  the schema's tag line, and verifies;
- review files: checks the `.sig` line's key equals the declared key, then verifies the signature over
  `proofbound.review-signature.v1\n` plus the file's bytes;
- marks each one **verified in this browser**, **does not verify in this browser** (shown as an
  error), or **this browser cannot check** (no Ed25519 in WebCrypto). Nothing is sent anywhere.
Commit signatures (git's SSH format) are not checked in the browser; the page says so.

## 3. Tests

`TestAsOfAudit` (the as-of rule on a scripted repository with a stubbed sync: a later commit and a
later outcome are invisible; the worktree is removed; `asof` outside the period is a 400; acts ignore
it), `TestBrowserProofData` (the page carries bytes, signature and declared key for an outcome and a
review file, and a canonical re-serialization in Go equals the signed bytes). The browser script is
accepted by running it in a browser on the tour's `shop` repository.
