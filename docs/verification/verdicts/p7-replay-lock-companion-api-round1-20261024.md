# P7 replay-lock companion API/SPEC boundary — round 1 documentary verdict

**Reviewer identity:** `NDR-P7-20261024-A`

**Review role:** independent non-author documentary API/SPEC reviewer

**Review date:** 2026-10-24

**Documentation freeze commit:** `b2982d0c8b35cdb7cb71e70bab1989c2d07c9b8a`

**Packet commit:** `406e141b19dc57e78939839bb92b05bc40ed7974`

**Review packet:** [`p7-replay-lock-companion-api-review-packet-20261023.md`](../p7-replay-lock-companion-api-review-packet-20261023.md)

**Review packet SHA-256:** `71384e2ea2bec44d5894b89eb67bce6566d066194d83c2a55efd87d7a88de342`

## VERDICT: AMEND

The packet's eight recorded corpus SHA-256 values match the corresponding blobs at the packet
commit. The review is bound to that exact corpus; later owner-acceptance and handoff edits were
not treated as reviewed inputs.

### Findings

1. **Archive grammar is not fully closed.** The boundary identifies an xz-compressed tar stream and
   rejects links, unsafe paths, unsupported types, duplicate regular files, and invalid modes, but
   it does not define accepted xz framing or rejection of trailing/concatenated data, accepted tar
   metadata/header forms, or canonical handling of empty, repeated, `./`, and equivalent path
   components. Specify those rules before claiming a complete fail-closed archive grammar.

2. **Cleanup semantics need one more precise layer.** The temporary-root contract covers the named
   success and failure paths, but it must state that root removal is attempted after every store-close
   attempt, including close failure; that cleanup is independent of caller cancellation; and how
   primary and cleanup errors are combined.

3. **The frozen live-SPEC corpus is ambiguous about runtime acceptance.** Draft 2 correctly reserves
   observation of the actually launched executable for future runtime evidence, while the frozen
   `kernel/internal/twin/SPEC.md` says the companion “binds the launched” digest and presents that
   as `TWIN-INV-14`. Mark that live-SPEC wording as future/unadopted or reconcile it through a
   separately authorized SPEC update. This verdict does not authorize that update.

### Question results

- Inventory ownership: **resolved**; one complete in-memory `Inventory` is selected.
- Archive/extraction policy: **AMEND** for the missing exact grammar details above.
- Staged versus launched executable: **AMEND** for the live-SPEC wording inconsistency.
- Temporary-root ownership and cleanup: **AMEND** for close/cancellation/error-combination semantics.
- Proposed names and future wording: conditionally stable for separately authorized adoption after
  the findings are resolved.
- Existing gates: **preserved**. No runtime fixtures, binary validation, L12/L17 result,
  behavior-lock acceptance, PostgreSQL approval, or runtime verdict is established; `DEF-P7-001`
  remains dormant and the pinned PostgreSQL JAR remains unapproved.

## Scope limits

This is a documentary boundary verdict only. It does not adopt the live contract, authorize an
implementation slice, add runtime fixtures, validate binaries, approve PostgreSQL, resolve
`DEF-P7-001`, establish clean-clone proof, or issue an independent runtime verdict.
