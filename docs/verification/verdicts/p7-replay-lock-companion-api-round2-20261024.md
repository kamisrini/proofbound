# P7 replay-lock companion API/SPEC boundary — round 2 documentary verdict

**Reviewer identity:** `NDR-P7-20261024-B`

**Review role:** independent non-author documentary API/SPEC reviewer

**Review date:** 2026-10-24

**Reviewed commit:** `2c1ca8c21cf7dc56d1240567221395d899d1f0d9`

**Review packet:** [`p7-replay-lock-companion-api-review-packet-20261024.md`](../p7-replay-lock-companion-api-review-packet-20261024.md)

**Review packet SHA-256:** `4871a7741d393e4005fa493d2d80cc2ecc7e8b0ba690691cfab18e04375844cd`

## VERDICT: AMEND

All nine packet-corpus SHA-256 values match the corresponding blobs in the reviewed commit. The
single in-memory inventory, staged-versus-launched distinction, cleanup semantics, names, and
existing gates were satisfied.

### Finding

The amended boundary closes xz input-byte consumption, concatenated/padded/trailing xz rejection,
extension headers, path grammar, duplicate paths, and regular-file restrictions. It does not yet
fully specify the decompressed POSIX `ustar` framing:

- whether the tar reader must consume the entire decompressed payload;
- the exact accepted `ustar` magic/version, checksum, numeric-field, block-alignment, and end-marker
  rules;
- rejection of extra tar end blocks, trailing decompressed bytes, or concatenated tar payloads; and
- explicit rejection of a regular file that is an ancestor of another entry requiring an implicit
  directory.

Define those rules in the boundary and recommendation before claiming a fully exact tar grammar.

## Scope limits

This is a documentary boundary verdict only. It does not adopt or edit the live SPEC, authorize
runtime fixtures, validate binaries or PostgreSQL, close L12/L17, accept a behavior lock, or issue a
runtime verdict. `DEF-P7-001` remains dormant and the pinned PostgreSQL JAR remains unapproved.
