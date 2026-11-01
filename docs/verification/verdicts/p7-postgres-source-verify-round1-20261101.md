---
schema: vera.verdict.v1
verdict_id: p7-postgres-source-verify-round1-20261101
status: ACCEPTABLE
reviewed_commit: 652bb5f2beb970acd15820c018bae80ac2af313f
findings: []
artifact_path: docs/verification/verdicts/p7-postgres-source-verify-round1-20261101.md
artifact_sha: dbb9616233d83f5cafbab51f46c00713f5f2066ead14d5834645742a2639ab0f
---

# P7 B1 unit 1 — PostgreSQL 16.4 source provenance — non-author verdict

## Provenance

- **Under review:** B1 unit 1 (PostgreSQL 16.4 source provenance, light-touch SHA-256/HTTPS
  mechanism) at commit `652bb5f2beb970acd15820c018bae80ac2af313f` — the verifier
  `scripts/p7-postgres-source-verify.sh` (committed `00af5f3`), its self-test
  `scripts/tests/p7-postgres-source-verify.test.sh`, the source manifest
  `docs/verification/p7-replay-lock-fixtures/postgres-source-16.4.json` (committed `9540ffc`), the
  evidence record `docs/verification/p7-replay-lock-prerequisites/postgres-source-16.4-verify-20261101.md`,
  and the governing decision `docs/decisions/VD-p7-postgres-source-route-2026-11-01.md`.
- **HEAD under review:** `652bb5f2beb970acd15820c018bae80ac2af313f`.
- **Driving mechanism:** `/proofbound-review` (CLAUDE.md pre-merge review), connector/peripheral tier
  (a shell verifier + JSON manifest + evidence doc; I/O-shaped, not foundational — does not touch
  `internal/core`/`internal/store`, locking, ordering, crash semantics, or the ledger write path).
- **Reviewer identity:** an independent reviewer dispatched via the Agent tool by the authoring
  session, which did NOT write the code. Law 9 is satisfied: this is an acceptance, not a self
  pre-check.
- **Code author/committer:** `VERA Maintainer` (authoring session).
- **Cleanroom redaction:** none required; the report below contains no absolute paths or usernames.

## Verdict (verbatim reviewer report)

status: ACCEPTABLE
reviewed_commit: 652bb5f2beb970acd15820c018bae80ac2af313f

what I ran:
- git rev-parse HEAD → 652bb5f2beb970acd15820c018bae80ac2af313f
- Read all five files under review.
- Self-test: `bash scripts/tests/p7-postgres-source-verify.test.sh` → "p7-postgres-source-verify.test: ok", EXIT=0.
- Independent adversarial battery (scratch dir, my own fixture bytes + manifest + .sha256). Pristine input ACCEPTED. All of these correctly REJECTED (non-zero): missing checksum file; role absent (--role nope); two inputs sharing the role (count=2); uppercase-hex digest in manifest; uppercase-hex digest in checksum file; 63-char short digest; wrong hash on line 1 of checksum file (right one on line 2); symlink pointing to wrong-content file; non-JSON manifest; manifest missing published_sha256 (jq emits "null"); non-numeric size; source_url ending in "/"; empty PATH (no jq/sha256sum → fails closed). Also correctly ACCEPTED harmless variants: CRLF checksum line with filename (hash is $1, unaffected), extra garbage lines after a valid first line (only NR==1 read).
- Independent upstream re-fetch over HTTPS (curl --proto '=https' --max-redirs 0): .sha256 → HTTP 200, 0 redirects, digest `971766d645aa73e93b9ef4e3be44201b4f45b5477095b049125403f9f3386d6f` == manifest published_sha256 exactly. Tarball → HTTP 200, 0 redirects, size 24765786 == pin, sha256 == pin.
- Ran the ACTUAL committed verifier against the real download: exits 0 with "OK role=postgres-source file=postgresql-16.4.tar.bz2 sha256=971766d6…386d6f size=24765786".
- Gate wiring: `make check` → first dep `hooks-test`; `hooks-test` loops `scripts/tests/*.test.sh`, so the self-test is auto-wired into the gate. Verifier dies on missing jq (line 9) / sha256sum (line 10); empty-PATH run confirmed fail-closed.
- grep for the digest/size/md5 and for `download_sha256`; checked all five files for `@generated` (none).

findings:
1. LOW — scripts/p7-postgres-source-verify.sh:55 vs :60 — symlink dereference inconsistency. `sha256sum "$tarball"` follows symlinks (hashes the target) but `stat -c %s "$tarball"` does NOT (reports the symlink's own path-length, missing `-L`). Failing scenario: an input-dir where the tarball is a symlink to the correct-content file is FALSE-REJECTED ("size mismatch: got 129, pinned 35" in my run). This is a false-reject only, never a false-accept: a matching hash requires the target to be the correct bytes, whose real size can never equal the tiny symlink path length, so the two checks can't both pass on a symlink. Real downloads (curl -o) are regular files, so the real flow never hits this. Cosmetic/robustness; optionally add `-L` to stat or reject non-regular files up front.
2. LOW — docs/verification/p7-replay-lock-fixtures/postgres-source-16.4.json:10-11 — `published_sha256` and `download_sha256` hold the same digest, but the verifier only ever reads `published_sha256` (lines 33-34); nothing asserts the two are equal. Per Law 2 this is a same-file drift seed: editing one and not the other goes undetected. It cannot affect the verification outcome (the unused field is ignored), and the field is a shared schema field meaningful in sibling manifests (setup-sources.json legitimately has them differ). Consider dropping it here or asserting equality. The evidence doc, journal, and state.md also carry the digest, but those are legitimate one-time evidence/narrative citations, not competing authoritative homes — Law 2 is satisfied (the manifest is the single home the verifier reads).

scenarios I tried and could not break (no false-accept found): tampered tarball byte; wrong pinned size; manifest↔checksum-file digest divergence; bytes-vs-pin mismatch; missing tarball; missing checksum file; absent role; duplicate role; uppercase / 63-char / odd hex in both manifest and checksum file; CRLF and extra-line checksum files; wrong-hash-first-line checksum; symlink to wrong content; malformed/non-JSON manifest; null/missing fields; non-numeric size; trailing-slash source_url; missing jq/sha256sum (fails closed). No way found to make the verifier return 0 when it should not.

build-laws check: Law 1 — none of the five files carry `@generated`; all are hand-authored source, correct. Law 2 — see finding 2 (LOW). Law 3 — self-test auto-wired into `make check` via `hooks-test`, and the verifier fails closed on missing jq/sha256sum (confirmed by inspection + empty-PATH run). Law 9 — the evidence doc and decision correctly state this unit is NOT yet accepted and that a non-author verdict is still owed.

Honesty of the evidence doc / decision / manifest: confirmed accurate and not overstated. The evidence doc (lines 19-24) explicitly states the guarantee is "byte-for-byte integrity … plus official-domain TLS trust" and is "not an offline, independently-authenticated signer-identity signature," and names the residual ("the checksum and the tarball come from the same origin, so a compromise of that origin/TLS path is not independently detectable"), calling it a deliberate owner-accepted residual. The manifest comment and the VD amendment (2026-11-01) say the same. No claim of a GPG signature, KEYS file, or authenticated signer identity anywhere.

reviewer summary: The verifier is fail-closed and honest; its self-test is genuinely wired into the gate, its positive/neutral controls are calibrated (pristine accepted, tampered/divergent/missing rejected), and I independently reproduced the upstream fact and ran the real verifier against the real tarball with a clean accept. I tried ~20 break attempts including the specific hex/CRLF/duplicate-key/symlink/parsing vectors requested and found no false-accept. The two findings are both LOW and non-blocking (a symlink size/hash inconsistency that can only over-reject, and an unenforced in-file digest duplicate that the verifier ignores). High confidence this artifact is ACCEPTABLE as a light-touch integrity+TLS provenance check; it does not and does not claim to provide signer-identity authentication.

## Disposition (author note)

Accepted at the reviewed commit `652bb5f`. Both findings are LOW and fail-closed — finding 1 can only
ever *over*-reject (never a false accept) and is unreachable by the real curl-downloaded-regular-file
flow; finding 2 is an in-file digest duplicate the verifier never reads. Following the precedent of
the prior P7 verdict (a fail-closed LOW is not worth moving the tree off an accepted commit for), both
are recorded here as optional future hardening rather than fixed under this acceptance: finding 1 →
add `stat -L` or reject non-regular inputs up front; finding 2 → drop `download_sha256` from this
manifest (or assert it equals `published_sha256`). Per Law 9 the author may apply these and a
non-author re-verifies; neither blocks unit 1's acceptance. Unit 1 (source provenance) is accepted;
`DEF-P7-001` stays OPEN/BLOCKED pending unit 2 (build + license/notice inventory) and unit 3 (runtime
fixture).
