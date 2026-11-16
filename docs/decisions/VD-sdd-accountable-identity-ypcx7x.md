# VD-sdd-accountable-identity-ypcx7x: SDD owner and sponsor are declared in the tenant's committed profile

**Status:** Accepted
**Date:** 2026-11-16

**Context:** The canonical intent model requires a non-empty `declared_owner` on every requirement and
a non-empty `declared_sponsor` on every change intent; a warranty needs someone accountable. The R2
M0 mapping spike ([`r2-m0-mapping-spike.md`](../verification/r2-m0-mapping-spike.md)) found that
neither OpenSpec nor Spec Kit records an owner or sponsor anywhere. The existing `specdir` provider
solves this by requiring `Owner:` / `Sponsor:` lines in its own format, which SDD kits do not have.
The forcing question: where does accountable identity come from without inventing it?

**Decision:** The `intent.sdd` provider takes `declared_owner` and `declared_sponsor` **only** from
the tenant's **committed profile**, which declares identity per artifact path pattern (for example,
`specs/** → owner: team-payments`). The declaration is resolved at the **same git tree** as the
artifact it covers, so it is point-in-time and reviewable in the tenant's own PRs. An artifact that
no declaration covers **fails closed** (`ErrUnmappable`); the provider never substitutes a default,
a git author, or a placeholder. This is `BR-sdd-warranty-i9miv7` obligation O-3.

**Consequences:** Makes easy — Proofbound works on an unmodified OpenSpec or Spec Kit repo (the only
tenant addition is one profile file); accountability is explicit and versioned; changing who owns a
spec is itself a reviewable commit. Makes hard / gave up — identity is per path pattern, not per
author, so a profile that is too coarse names a team where a person was meant; a tenant must commit
the profile before any warranty exists (no zero-config start). Rejected: **git author** (observed, not
declared — it would quietly change what the `declared_*` fields mean) and **mandatory
`Owner:` / `Sponsor:` lines in every spec** (honest, but friction on every spec and it breaks
"works on an unmodified repo").

**Revisit when:** an SDD kit starts recording owner or sponsor natively (then prefer the kit's own
declaration over the profile's, as a profile option); OR tenants consistently need per-person rather
than per-path accountability; OR a code/API provider (the escape hatch) needs identity from a system
of record instead of a file.
