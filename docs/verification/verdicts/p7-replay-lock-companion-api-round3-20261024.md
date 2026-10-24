# P7 replay-lock companion API/SPEC boundary — round 3 documentary verdict

**Reviewer identity:** `NDR-P7-20261024-C`

**Review role:** independent non-author documentary API/SPEC reviewer

**Review date:** 2026-10-24

**Reviewed commit:** `eef63158b3247573cf4c8ee33645c127dea98d27`

**Review packet:** [`p7-replay-lock-companion-api-review-packet-20261024-round3.md`](../p7-replay-lock-companion-api-review-packet-20261024-round3.md)

**Review packet SHA-256:** `2d56f320218b196e1a98c1c1221b159cae3663a7d06f5714cd1dccd7d0fbe864`

## VERDICT: ACCEPT

All ten packet-corpus SHA-256 values match the committed blobs. All six review questions are
satisfied:

1. The boundary keeps one complete in-memory `Inventory` and creates no second runtime truth.
2. The exact one-stream xz and POSIX `ustar` rules define full decompressed-stream consumption,
   magic/version, checksum, numeric fields, 512-byte alignment, exactly two final zero blocks,
   rejection of trailing/concatenated tar data, and file/implicit-directory ancestor collisions.
3. Staged `bin/postgres` hashing is distinct from future launched-process observation, and the live
   `TWIN-INV-14` wording is explicitly future/unadopted.
4. Temporary-root ownership, close-before-remove ordering, cancellation, close failure, cleanup
   failure, and combined-error precedence are defined.
5. The proposed names and future SPEC wording are stable for separately authorized adoption.
6. Existing gates remain intact.

No findings or required edits.

## Scope limits

This is a documentary boundary verdict only. It does not edit or adopt the live SPEC, authorize
runtime fixtures, validate binaries or PostgreSQL, close L12/L17, accept a behavior lock, establish
clean-clone proof, or issue an independent runtime verdict. `DEF-P7-001` remains dormant and the
pinned PostgreSQL JAR remains unapproved.
