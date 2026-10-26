# VD-p7-nonruntime-manifest-2026-10-26: bound non-runtime manifest preparation

**Status:** Accepted
**Date:** 2026-10-26

## Decision

The P7 `internal/twin/manifest` package is a separately authorized, non-runtime preparation slice
under `P7-NONRUNTIME-MANIFEST-1` and
[`CI-p7-nonruntime-manifest-1a2b3c`](../intent/records/change-intents/CI-p7-nonruntime-manifest-1a2b3c/0001.md).
It may define and validate the canonical projector/build manifest shape, exact source/toolchain
identity declarations, and synthetic digest/revision controls.

The package is classified as P7 work rather than added to the frozen P6 package-acceptance universe.
Independent non-author review and the authorized slice’s repository checks remain required before
acceptance is claimed.

## Boundary

No runtime fixture, network fetch, external artifact retrieval, binary build or validation,
PostgreSQL execution or approval, `DEF-P7-001` activation, L12/L17 result, behavior-lock
acceptance, clean-clone proof, or runtime verdict is authorized by this decision. The declared
toolchain and module digests are identities only; they do not prove retrieval, licensing,
execution, or reproducibility.
